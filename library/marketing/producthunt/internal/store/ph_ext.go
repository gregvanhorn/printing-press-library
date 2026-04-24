package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// PHTablesSchemaVersion is the Product-Hunt-specific schema epoch. Bumped
// whenever EnsurePHTables changes shape. Stamped into a sibling pragma-like
// row so stale binaries refuse to read newer-shaped tables rather than
// corrupting data. Independent from StoreSchemaVersion so generator bumps
// and PH-specific bumps don't collide.
//
// Version history:
//
//   1 - initial: posts, posts_fts, snapshots, snapshot_entries, ph_meta
//   2 - adds ph_backfill_state (for GraphQL backfill cursor persistence)
const PHTablesSchemaVersion = 2

// EnsurePHTables is idempotent; call once per Open(). It adds the
// Product-Hunt-specific tables on top of the generator's resources/feed tables:
//
//   - posts           (one row per unique PostID ever seen on /feed)
//   - posts_fts       (FTS5 index over title + tagline + author)
//   - snapshots       (one row per sync; drives rank-over-time commands)
//   - snapshot_entries (postID × snapshotID × rank)
//
// A persistent ph_meta row tracks PHTablesSchemaVersion.
func EnsurePHTables(s *Store) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS ph_meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			post_id INTEGER PRIMARY KEY,
			slug TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			tagline TEXT,
			author TEXT,
			discussion_url TEXT,
			external_url TEXT,
			published_at DATETIME,
			updated_at DATETIME,
			first_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			seen_count INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_published ON posts(published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(
			slug, title, tagline, author,
			tokenize='porter unicode61'
		)`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			snapshot_id INTEGER PRIMARY KEY AUTOINCREMENT,
			taken_at DATETIME NOT NULL,
			entry_count INTEGER NOT NULL DEFAULT 0,
			source TEXT NOT NULL DEFAULT 'feed'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_snapshots_taken_at ON snapshots(taken_at DESC)`,
		`CREATE TABLE IF NOT EXISTS snapshot_entries (
			snapshot_id INTEGER NOT NULL,
			post_id INTEGER NOT NULL,
			rank INTEGER NOT NULL,
			external_url TEXT NOT NULL DEFAULT '',
			PRIMARY KEY (snapshot_id, post_id),
			FOREIGN KEY (snapshot_id) REFERENCES snapshots(snapshot_id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
		)`,
		// Best-effort column add for databases created before this column
		// existed. SQLite errors on duplicate ALTER; we swallow with a
		// separate statement guarded by PRAGMA introspection.
		`CREATE INDEX IF NOT EXISTS idx_snapshot_entries_post ON snapshot_entries(post_id)`,

		// Schema v2: backfill state. Each row represents one contiguous
		// window the caller has asked the CLI to pull from GraphQL. Keyed
		// by a deterministic window_id hash so resuming the same window is
		// idempotent. cursor is opaque PH-provided pagination state.
		`CREATE TABLE IF NOT EXISTS ph_backfill_state (
			window_id TEXT PRIMARY KEY,
			posted_after TEXT NOT NULL,
			posted_before TEXT NOT NULL,
			cursor TEXT,
			pages_completed INTEGER NOT NULL DEFAULT 0,
			posts_upserted INTEGER NOT NULL DEFAULT 0,
			last_run_at DATETIME,
			last_error TEXT,
			completed_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_backfill_incomplete ON ph_backfill_state(completed_at) WHERE completed_at IS NULL`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("ensuring PH tables: %w", err)
		}
	}
	// Stamp ph_meta version
	if _, err := s.db.Exec(
		`INSERT INTO ph_meta (key, value, updated_at) VALUES ('ph_schema_version', ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		fmt.Sprintf("%d", PHTablesSchemaVersion), time.Now(),
	); err != nil {
		return fmt.Errorf("stamping ph_schema_version: %w", err)
	}
	return nil
}

// Post is the shape returned by list/get/search queries.
// PHExt uses this instead of exposing raw SQL rows.
type Post struct {
	PostID        int64
	Slug          string
	Title         string
	Tagline       string
	Author        string
	DiscussionURL string
	ExternalURL   string
	PublishedAt   time.Time
	UpdatedAt     time.Time
	FirstSeenAt   time.Time
	LastSeenAt    time.Time
	SeenCount     int
}

// Snapshot is one sync cycle's capture of the /feed.
type Snapshot struct {
	SnapshotID int64
	TakenAt    time.Time
	EntryCount int
	Source     string
}

// SnapshotEntry pairs a post_id with the rank it held within a given snapshot.
type SnapshotEntry struct {
	SnapshotID int64
	PostID     int64
	Rank       int
}

// UpsertPost writes a post to the posts table and refreshes the FTS row.
// Increments seen_count by 1 on every call; updates last_seen_at; preserves
// first_seen_at (insert-only on conflict). Pure SQL; no JSON side effects.
//
// Caller wraps in a tx so the snapshot and its entries land atomically.
func UpsertPost(tx *sql.Tx, p Post) error {
	_, err := tx.Exec(
		`INSERT INTO posts (post_id, slug, title, tagline, author, discussion_url, external_url, published_at, updated_at, first_seen_at, last_seen_at, seen_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
		 ON CONFLICT(post_id) DO UPDATE SET
		   slug = excluded.slug,
		   title = excluded.title,
		   tagline = excluded.tagline,
		   author = excluded.author,
		   discussion_url = excluded.discussion_url,
		   external_url = excluded.external_url,
		   updated_at = excluded.updated_at,
		   last_seen_at = excluded.last_seen_at,
		   seen_count = posts.seen_count + 1`,
		p.PostID, p.Slug, p.Title, p.Tagline, p.Author,
		p.DiscussionURL, p.ExternalURL,
		nullableTime(p.PublishedAt), nullableTime(p.UpdatedAt),
		time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("upsert post %d (%s): %w", p.PostID, p.Slug, err)
	}

	// FTS5 doesn't support ON CONFLICT cleanly; delete then re-insert using
	// post_id as the rowid so updates stay in lockstep.
	if _, err := tx.Exec(`DELETE FROM posts_fts WHERE rowid = ?`, p.PostID); err != nil {
		return fmt.Errorf("fts cleanup for post %d: %w", p.PostID, err)
	}
	if _, err := tx.Exec(
		`INSERT INTO posts_fts (rowid, slug, title, tagline, author) VALUES (?, ?, ?, ?, ?)`,
		p.PostID, p.Slug, p.Title, p.Tagline, p.Author,
	); err != nil {
		return fmt.Errorf("fts insert for post %d: %w", p.PostID, err)
	}
	return nil
}

// RecordSnapshot inserts a snapshot row and returns its autoincrement ID.
func RecordSnapshot(tx *sql.Tx, takenAt time.Time, entryCount int, source string) (int64, error) {
	if source == "" {
		source = "feed"
	}
	res, err := tx.Exec(
		`INSERT INTO snapshots (taken_at, entry_count, source) VALUES (?, ?, ?)`,
		takenAt, entryCount, source,
	)
	if err != nil {
		return 0, fmt.Errorf("record snapshot: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("snapshot last insert id: %w", err)
	}
	return id, nil
}

// RecordSnapshotEntry writes a post_id + rank pair. Rank is 1-indexed in
// the order the entries appeared in the feed (1 = topmost). Also records
// the post's external_url as of this snapshot so outbound-diff can detect
// per-post URL changes across time.
func RecordSnapshotEntry(tx *sql.Tx, snapshotID int64, postID int64, rank int, externalURL string) error {
	_, err := tx.Exec(
		`INSERT INTO snapshot_entries (snapshot_id, post_id, rank, external_url) VALUES (?, ?, ?, ?)`,
		snapshotID, postID, rank, externalURL,
	)
	if err != nil {
		return fmt.Errorf("record snapshot entry: %w", err)
	}
	return nil
}

// GetPostBySlug returns the stored post for the given slug, or sql.ErrNoRows.
func (s *Store) GetPostBySlug(slug string) (*Post, error) {
	row := s.db.QueryRow(
		`SELECT post_id, slug, title, COALESCE(tagline, ''), COALESCE(author, ''),
		        COALESCE(discussion_url, ''), COALESCE(external_url, ''),
		        COALESCE(published_at, ''), COALESCE(updated_at, ''),
		        first_seen_at, last_seen_at, seen_count
		 FROM posts WHERE slug = ?`,
		slug,
	)
	return scanPost(row)
}

// GetPostByID returns the stored post for the given numeric Post ID.
func (s *Store) GetPostByID(id int64) (*Post, error) {
	row := s.db.QueryRow(
		`SELECT post_id, slug, title, COALESCE(tagline, ''), COALESCE(author, ''),
		        COALESCE(discussion_url, ''), COALESCE(external_url, ''),
		        COALESCE(published_at, ''), COALESCE(updated_at, ''),
		        first_seen_at, last_seen_at, seen_count
		 FROM posts WHERE post_id = ?`,
		id,
	)
	return scanPost(row)
}

// ListPostsOpts is the filter set for ListPosts. All fields are optional.
type ListPostsOpts struct {
	Author    string    // exact match on author (case-sensitive)
	Since     time.Time // posts published at or after this time
	Until     time.Time // posts published at or before this time
	SortField string    // one of: published, updated, title, author, seen_count, first_seen
	SortDesc  bool      // default true when empty/"published"
	Limit     int       // 0 = no limit
	Offset    int
}

// ListPosts returns posts matching the filter, ordered per opts.
// Default sort is published_at DESC.
func (s *Store) ListPosts(opts ListPostsOpts) ([]Post, error) {
	var (
		where []string
		args  []any
	)
	if opts.Author != "" {
		where = append(where, "author = ?")
		args = append(args, opts.Author)
	}
	if !opts.Since.IsZero() {
		where = append(where, "published_at >= ?")
		args = append(args, opts.Since)
	}
	if !opts.Until.IsZero() {
		where = append(where, "published_at <= ?")
		args = append(args, opts.Until)
	}

	orderCol := "published_at"
	switch opts.SortField {
	case "published", "":
		orderCol = "published_at"
	case "updated":
		orderCol = "updated_at"
	case "title":
		orderCol = "title"
	case "author":
		orderCol = "author"
	case "seen_count":
		orderCol = "seen_count"
	case "first_seen":
		orderCol = "first_seen_at"
	default:
		return nil, fmt.Errorf("unknown sort field %q", opts.SortField)
	}
	dir := "DESC"
	if !opts.SortDesc && opts.SortField != "" {
		// Explicit ascending only when the caller picked a field.
		dir = "ASC"
	}

	q := `SELECT post_id, slug, title, COALESCE(tagline, ''), COALESCE(author, ''),
	             COALESCE(discussion_url, ''), COALESCE(external_url, ''),
	             COALESCE(published_at, ''), COALESCE(updated_at, ''),
	             first_seen_at, last_seen_at, seen_count
	       FROM posts`
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += fmt.Sprintf(" ORDER BY %s %s", orderCol, dir)
	if opts.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d OFFSET %d", opts.Limit, opts.Offset)
	}

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var out []Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// SearchPostsFTS runs an FTS5 MATCH against the posts_fts index.
// Returns posts ordered by FTS5 rank ascending (best match first).
func (s *Store) SearchPostsFTS(query string, limit int) ([]Post, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT p.post_id, p.slug, p.title, COALESCE(p.tagline, ''), COALESCE(p.author, ''),
		       COALESCE(p.discussion_url, ''), COALESCE(p.external_url, ''),
		       COALESCE(p.published_at, ''), COALESCE(p.updated_at, ''),
		       p.first_seen_at, p.last_seen_at, p.seen_count
		FROM posts p
		JOIN posts_fts f ON f.rowid = p.post_id
		WHERE posts_fts MATCH ?
		ORDER BY bm25(posts_fts)
		LIMIT ?`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fts search: %w", err)
	}
	defer rows.Close()

	var out []Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// SnapshotsForPost returns every (snapshot, rank) pair the post has ever
// occupied, newest snapshot first.
type PostAppearance struct {
	SnapshotID int64
	TakenAt    time.Time
	Rank       int
}

func (s *Store) SnapshotsForPost(postID int64) ([]PostAppearance, error) {
	rows, err := s.db.Query(`
		SELECT s.snapshot_id, s.taken_at, e.rank
		FROM snapshot_entries e
		JOIN snapshots s ON s.snapshot_id = e.snapshot_id
		WHERE e.post_id = ?
		ORDER BY s.taken_at DESC`, postID)
	if err != nil {
		return nil, fmt.Errorf("snapshots for post: %w", err)
	}
	defer rows.Close()
	var out []PostAppearance
	for rows.Next() {
		var a PostAppearance
		if err := rows.Scan(&a.SnapshotID, &a.TakenAt, &a.Rank); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// LatestSnapshot returns the most recent snapshot, or sql.ErrNoRows.
func (s *Store) LatestSnapshot() (*Snapshot, error) {
	row := s.db.QueryRow(`SELECT snapshot_id, taken_at, entry_count, source FROM snapshots ORDER BY taken_at DESC LIMIT 1`)
	var sp Snapshot
	if err := row.Scan(&sp.SnapshotID, &sp.TakenAt, &sp.EntryCount, &sp.Source); err != nil {
		return nil, err
	}
	return &sp, nil
}

// PostsInSnapshot returns every post present in the given snapshot,
// ordered by rank ascending (1 = top of feed).
func (s *Store) PostsInSnapshot(snapshotID int64) ([]Post, error) {
	rows, err := s.db.Query(`
		SELECT p.post_id, p.slug, p.title, COALESCE(p.tagline, ''), COALESCE(p.author, ''),
		       COALESCE(p.discussion_url, ''), COALESCE(p.external_url, ''),
		       COALESCE(p.published_at, ''), COALESCE(p.updated_at, ''),
		       p.first_seen_at, p.last_seen_at, p.seen_count
		FROM snapshot_entries e
		JOIN posts p ON p.post_id = e.post_id
		WHERE e.snapshot_id = ?
		ORDER BY e.rank ASC`, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("posts in snapshot: %w", err)
	}
	defer rows.Close()
	var out []Post
	for rows.Next() {
		p, err := scanPostRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// AuthorTally is a (author, count) aggregate used by makers/authors commands.
type AuthorTally struct {
	Author string
	Count  int
	Unique int // unique posts by this author in window
}

// TopAuthorsSince aggregates authors whose posts were seen in snapshots taken
// at or after `since`. Count is total seen_count (frequency across snapshots);
// Unique is distinct post slugs by that author.
func (s *Store) TopAuthorsSince(since time.Time, limit int) ([]AuthorTally, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(`
		SELECT p.author AS author,
		       COUNT(*) AS cnt,
		       COUNT(DISTINCT p.post_id) AS unique_posts
		FROM snapshot_entries e
		JOIN snapshots s ON s.snapshot_id = e.snapshot_id
		JOIN posts p ON p.post_id = e.post_id
		WHERE s.taken_at >= ? AND p.author IS NOT NULL AND p.author != ''
		GROUP BY p.author
		ORDER BY cnt DESC, author ASC
		LIMIT ?`, since, limit)
	if err != nil {
		return nil, fmt.Errorf("top authors: %w", err)
	}
	defer rows.Close()
	var out []AuthorTally
	for rows.Next() {
		var a AuthorTally
		if err := rows.Scan(&a.Author, &a.Count, &a.Unique); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AuthorsCoOccurring finds authors who repeatedly appear in the same feed
// snapshots as the target author. Returns (other, shared-snapshot-count) rows.
type AuthorCoOccurrence struct {
	Other           string
	SharedSnapshots int
}

func (s *Store) AuthorsCoOccurring(target string, since time.Time, limit int) ([]AuthorCoOccurrence, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(`
		WITH target_snapshots AS (
			SELECT DISTINCT e.snapshot_id
			FROM snapshot_entries e
			JOIN snapshots s ON s.snapshot_id = e.snapshot_id
			JOIN posts p ON p.post_id = e.post_id
			WHERE p.author = ? AND s.taken_at >= ?
		)
		SELECT p.author AS other, COUNT(DISTINCT e.snapshot_id) AS shared
		FROM snapshot_entries e
		JOIN posts p ON p.post_id = e.post_id
		JOIN target_snapshots t ON t.snapshot_id = e.snapshot_id
		WHERE p.author != ? AND p.author IS NOT NULL AND p.author != ''
		GROUP BY p.author
		ORDER BY shared DESC, other ASC
		LIMIT ?`, target, since, target, limit)
	if err != nil {
		return nil, fmt.Errorf("co-occurring authors: %w", err)
	}
	defer rows.Close()
	var out []AuthorCoOccurrence
	for rows.Next() {
		var a AuthorCoOccurrence
		if err := rows.Scan(&a.Other, &a.SharedSnapshots); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// OutboundDrift returns posts whose external_url changed across any two
// adjacent snapshots in the window. Implemented via sequential scan in Go
// because the snapshot_entries table doesn't store external_url directly;
// callers with many snapshots may prefer to specify a tighter window.
type URLChange struct {
	PostID    int64
	Slug      string
	Title     string
	OldURL    string
	NewURL    string
	ChangedAt time.Time
}

// OutboundDrift returns per-post URL changes detected across snapshot_entries
// in the given window. A post is flagged only when at least two snapshots in
// the window recorded different external_url values for it.
//
// The query finds, for each post, the earliest and latest external_url observed
// in the window. If they differ, it emits a URLChange with OldURL/NewURL and
// the snapshot time the URL last changed. Posts that were re-seen with no URL
// change are excluded.
func (s *Store) OutboundDrift(since time.Time) ([]URLChange, error) {
	// Strategy: group snapshot_entries by post_id within the window,
	// compute oldest URL (MIN taken_at) and newest URL (MAX taken_at), emit
	// only rows where those URLs differ and at least two snapshots landed.
	rows, err := s.db.Query(`
		WITH windowed AS (
			SELECT e.post_id, e.external_url, s.taken_at
			FROM snapshot_entries e
			JOIN snapshots s ON s.snapshot_id = e.snapshot_id
			WHERE s.taken_at >= ?
		),
		first_seen AS (
			SELECT post_id, external_url AS first_url,
			       ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY taken_at ASC) AS rn
			FROM windowed
		),
		last_seen AS (
			SELECT post_id, external_url AS last_url, taken_at AS last_taken,
			       ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY taken_at DESC) AS rn
			FROM windowed
		)
		SELECT f.post_id, p.slug, p.title, f.first_url, l.last_url, l.last_taken
		FROM first_seen f
		JOIN last_seen l ON l.post_id = f.post_id AND l.rn = 1
		JOIN posts p ON p.post_id = f.post_id
		WHERE f.rn = 1 AND f.first_url != l.last_url AND f.first_url != ''
		ORDER BY l.last_taken DESC`, since)
	if err != nil {
		return nil, fmt.Errorf("outbound drift: %w", err)
	}
	defer rows.Close()
	var out []URLChange
	for rows.Next() {
		var u URLChange
		if err := rows.Scan(&u.PostID, &u.Slug, &u.Title, &u.OldURL, &u.NewURL, &u.ChangedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// PostCount returns the total number of posts in the store.
func (s *Store) PostCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&n)
	return n, err
}

// SnapshotCount returns the total number of snapshots persisted.
func (s *Store) SnapshotCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM snapshots`).Scan(&n)
	return n, err
}

// scanPost decodes a single-row result into a Post.
func scanPost(row *sql.Row) (*Post, error) {
	var p Post
	var publishedStr, updatedStr string
	if err := row.Scan(
		&p.PostID, &p.Slug, &p.Title, &p.Tagline, &p.Author,
		&p.DiscussionURL, &p.ExternalURL,
		&publishedStr, &updatedStr,
		&p.FirstSeenAt, &p.LastSeenAt, &p.SeenCount,
	); err != nil {
		return nil, err
	}
	p.PublishedAt = parseStoredTime(publishedStr)
	p.UpdatedAt = parseStoredTime(updatedStr)
	return &p, nil
}

// scanPostRow decodes a row from a multi-row *sql.Rows into a Post.
func scanPostRow(rows *sql.Rows) (*Post, error) {
	var p Post
	var publishedStr, updatedStr string
	if err := rows.Scan(
		&p.PostID, &p.Slug, &p.Title, &p.Tagline, &p.Author,
		&p.DiscussionURL, &p.ExternalURL,
		&publishedStr, &updatedStr,
		&p.FirstSeenAt, &p.LastSeenAt, &p.SeenCount,
	); err != nil {
		return nil, err
	}
	p.PublishedAt = parseStoredTime(publishedStr)
	p.UpdatedAt = parseStoredTime(updatedStr)
	return &p, nil
}

// parseStoredTime tries several layouts because modernc.org/sqlite stores
// time.Time values in a driver-specific format ("2006-01-02 15:04:05.999999999-07:00")
// that isn't bit-compatible with RFC3339. We accept both so reads succeed
// regardless of how the write driver happened to serialize.
func parseStoredTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		// modernc.org/sqlite serializes time.Time via .String() which emits
		// "2006-01-02 15:04:05.999999999 -0700 MST" (space-separated, named
		// zone at end). Accepting both precisions here.
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// nullableTime returns a nil interface when the time is the zero value so
// INSERT stores NULL instead of "0001-01-01T00:00:00Z". Downstream parsing
// handles both cases.
func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

// ---------- ph_meta key/value helpers ----------

// Well-known ph_meta keys. These drive auto-sync freshness decisions and
// diagnostic output; keep names stable across versions.
const (
	PHMetaLastSyncAt   = "last_sync_at"
	PHMetaLastCaller   = "last_caller"
	PHMetaSchemaVersion = "ph_schema_version"
)

// GetPHMeta returns the value stored at the given ph_meta key, or empty string
// if the key does not exist. Callers treat "" as "unknown" rather than "set".
func (s *Store) GetPHMeta(key string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM ph_meta WHERE key = ?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get ph_meta %q: %w", key, err)
	}
	return v, nil
}

// SetPHMeta upserts a ph_meta key/value pair, refreshing updated_at.
func (s *Store) SetPHMeta(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO ph_meta (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("set ph_meta %q: %w", key, err)
	}
	return nil
}

// LastSyncAt returns the time of the most recent successful sync, or the zero
// time if no sync has ever run. Used by auto-sync staleness checks.
func (s *Store) LastSyncAt() (time.Time, error) {
	v, err := s.GetPHMeta(PHMetaLastSyncAt)
	if err != nil {
		return time.Time{}, err
	}
	if v == "" {
		return time.Time{}, nil
	}
	return parseStoredTime(v), nil
}

// RecordSync stamps the current time as last_sync_at and records the caller
// identity for diagnostics. caller may be empty when the sync was user-initiated
// from the terminal rather than through an integrator.
func (s *Store) RecordSync(caller string) error {
	if err := s.SetPHMeta(PHMetaLastSyncAt, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	if caller != "" {
		if err := s.SetPHMeta(PHMetaLastCaller, caller); err != nil {
			return err
		}
	}
	return nil
}

// ---------- ph_backfill_state helpers ----------

// BackfillState represents one row in ph_backfill_state. Cursor is opaque
// PH-provided pagination state; empty means "start from the beginning".
// CompletedAt is zero until the backfill finishes cleanly.
type BackfillState struct {
	WindowID       string
	PostedAfter    string
	PostedBefore   string
	Cursor         string
	PagesCompleted int
	PostsUpserted  int
	LastRunAt      time.Time
	LastError      string
	CompletedAt    time.Time
}

// IsComplete reports whether the backfill has finished successfully.
func (b *BackfillState) IsComplete() bool {
	return !b.CompletedAt.IsZero()
}

// GetBackfillState loads the row for windowID, or returns (nil, nil) when the
// window has never been seen. Callers use (nil, nil) as the signal to start a
// fresh backfill.
func (s *Store) GetBackfillState(windowID string) (*BackfillState, error) {
	row := s.db.QueryRow(
		`SELECT window_id, posted_after, posted_before,
		        COALESCE(cursor, ''),
		        pages_completed, posts_upserted,
		        COALESCE(last_run_at, ''),
		        COALESCE(last_error, ''),
		        COALESCE(completed_at, '')
		 FROM ph_backfill_state WHERE window_id = ?`,
		windowID,
	)
	var (
		b                     BackfillState
		lastRunStr, completedStr string
	)
	err := row.Scan(
		&b.WindowID, &b.PostedAfter, &b.PostedBefore,
		&b.Cursor,
		&b.PagesCompleted, &b.PostsUpserted,
		&lastRunStr, &b.LastError, &completedStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get backfill state %q: %w", windowID, err)
	}
	b.LastRunAt = parseStoredTime(lastRunStr)
	b.CompletedAt = parseStoredTime(completedStr)
	return &b, nil
}

// UpsertBackfillState persists the row. Used both for initial creation and
// per-page cursor updates during a running backfill.
func (s *Store) UpsertBackfillState(b BackfillState) error {
	_, err := s.db.Exec(
		`INSERT INTO ph_backfill_state
		   (window_id, posted_after, posted_before, cursor, pages_completed, posts_upserted, last_run_at, last_error, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(window_id) DO UPDATE SET
		   cursor = excluded.cursor,
		   pages_completed = excluded.pages_completed,
		   posts_upserted = excluded.posts_upserted,
		   last_run_at = excluded.last_run_at,
		   last_error = excluded.last_error,
		   completed_at = excluded.completed_at`,
		b.WindowID, b.PostedAfter, b.PostedBefore,
		nullableString(b.Cursor),
		b.PagesCompleted, b.PostsUpserted,
		nullableTime(b.LastRunAt),
		nullableString(b.LastError),
		nullableTime(b.CompletedAt),
	)
	if err != nil {
		return fmt.Errorf("upsert backfill state %q: %w", b.WindowID, err)
	}
	return nil
}

// PendingBackfillStates returns every row where completed_at IS NULL,
// ordered by last_run_at DESC. Used by `backfill resume` to find work.
func (s *Store) PendingBackfillStates() ([]BackfillState, error) {
	rows, err := s.db.Query(
		`SELECT window_id, posted_after, posted_before,
		        COALESCE(cursor, ''),
		        pages_completed, posts_upserted,
		        COALESCE(last_run_at, ''),
		        COALESCE(last_error, ''),
		        ''
		 FROM ph_backfill_state
		 WHERE completed_at IS NULL
		 ORDER BY last_run_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list pending backfills: %w", err)
	}
	defer rows.Close()
	var out []BackfillState
	for rows.Next() {
		var (
			b                     BackfillState
			lastRunStr, completedStr string
		)
		if err := rows.Scan(
			&b.WindowID, &b.PostedAfter, &b.PostedBefore,
			&b.Cursor,
			&b.PagesCompleted, &b.PostsUpserted,
			&lastRunStr, &b.LastError, &completedStr,
		); err != nil {
			return nil, err
		}
		b.LastRunAt = parseStoredTime(lastRunStr)
		b.CompletedAt = parseStoredTime(completedStr)
		out = append(out, b)
	}
	return out, rows.Err()
}

// nullableString returns a nil interface when the string is empty so INSERT
// stores NULL instead of "". Kept symmetric with nullableTime above.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
