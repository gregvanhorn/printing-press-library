package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/atom"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/config"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
)

// StaleAfter is the age threshold for the Atom store: anything older than
// this age triggers an auto-sync before a read command serves its query.
// 24h matches normal daily-ish usage patterns; integrators that run more
// or less often still get useful warmth.
const StaleAfter = 24 * time.Hour

// AutoSyncMeta is the summary emitted when a read command's PreRunE runs
// EnsureFresh. Commands that build JSON output attach this under the
// _meta.auto_synced key so integrators can see what happened behind the
// scenes without asking for it.
type AutoSyncMeta struct {
	Ran           bool   `json:"ran"`
	Reason        string `json:"reason,omitempty"` // "fresh", "stale", "never_synced", "disabled"
	PostsUpserted int    `json:"posts_upserted,omitempty"`
	ElapsedMs     int64  `json:"elapsed_ms,omitempty"`
	Error         string `json:"error,omitempty"`
	LastSyncAt    string `json:"last_sync_at,omitempty"` // RFC3339
	Caller        string `json:"caller,omitempty"`
}

// EnsureFresh is the single entry point auto-sync-aware commands call at
// the top of their RunE. Contract:
//
//   - If auto-sync is disabled (--no-auto-sync flag, config AutoSync=false),
//     returns {Ran: false, Reason: "disabled"} without opening the store.
//   - If the store is fresh (last_sync_at within StaleAfter), returns
//     {Ran: false, Reason: "fresh", LastSyncAt: ...}.
//   - If stale or never-synced, runs a single Atom sync and returns
//     {Ran: true, Reason: "stale"|"never_synced", PostsUpserted: N, ElapsedMs: M}.
//   - Sync failures are non-fatal: they are captured in Error and the caller
//     continues serving from whatever data the store already has. The only
//     transport that raises is "can't open the store", which bubbles up
//     as a real error.
//
// After a successful sync, ph_meta.last_sync_at and ph_meta.last_caller are
// stamped so the next EnsureFresh call can decide freshness. Callers pass
// a caller string (e.g. "last30days/3.0.1") for diagnostics; empty string
// is fine for direct CLI usage.
func EnsureFresh(flags *rootFlags, dbPath string) *AutoSyncMeta {
	meta := &AutoSyncMeta{}

	// Persist caller on rootFlags so downstream helpers can include it
	// in JSON output without threading through every call site.
	if flags.caller != "" {
		meta.Caller = flags.caller
	}

	// Load config to see whether auto-sync is globally disabled.
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		// Config errors should not block reads; note and continue.
		meta.Reason = "disabled"
		meta.Error = "config load: " + err.Error()
		return meta
	}
	if flags.noAutoSync || !cfg.AutoSyncEnabled() {
		meta.Reason = "disabled"
		return meta
	}

	// Open the store to check freshness. Fail-soft: if we can't open it,
	// skip auto-sync rather than crashing the read path.
	db, err := store.Open(dbPath)
	if err != nil {
		meta.Reason = "disabled"
		meta.Error = "open store: " + err.Error()
		return meta
	}
	defer db.Close()
	if err := store.EnsurePHTables(db); err != nil {
		meta.Reason = "disabled"
		meta.Error = "ensure tables: " + err.Error()
		return meta
	}

	last, err := db.LastSyncAt()
	if err != nil {
		meta.Reason = "disabled"
		meta.Error = "read last_sync_at: " + err.Error()
		return meta
	}
	if !last.IsZero() {
		meta.LastSyncAt = last.UTC().Format(time.RFC3339)
	}

	// Fresh enough? Bail out — no network.
	if !last.IsZero() && time.Since(last) < StaleAfter {
		meta.Reason = "fresh"
		return meta
	}

	if last.IsZero() {
		meta.Reason = "never_synced"
	} else {
		meta.Reason = "stale"
	}

	// Stale or never-synced: fire a sync.
	start := time.Now()
	posts, err := runAtomSync(db, flags.caller, flags.timeout)
	meta.ElapsedMs = time.Since(start).Milliseconds()

	if err != nil {
		// Record the failure but don't propagate — caller still serves
		// whatever's in the store.
		meta.Error = err.Error()
		meta.Ran = false
		return meta
	}

	meta.Ran = true
	meta.PostsUpserted = posts
	meta.LastSyncAt = time.Now().UTC().Format(time.RFC3339)
	return meta
}

// runAtomSync fetches /feed, parses entries, and upserts them into the
// store as a new snapshot. Returns the count of posts upserted. Used by
// both the sync command (via its RunE) and by EnsureFresh.
//
// The caller supplies an open Store; this function does NOT close it.
// Timeout controls the HTTP fetch only; parse + upsert have no timeout.
func runAtomSync(db *store.Store, caller string, timeout time.Duration) (int, error) {
	body, err := fetchFeedBody(timeout)
	if err != nil {
		return 0, fmt.Errorf("fetch /feed: %w", err)
	}
	feed, err := atom.Parse(body)
	if err != nil {
		return 0, fmt.Errorf("parse feed: %w", err)
	}

	tx, err := db.DB().Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	rollback := true
	defer func() {
		if rollback {
			_ = tx.Rollback()
		}
	}()

	snapID, err := store.RecordSnapshot(tx, time.Now(), len(feed.Entries), "feed")
	if err != nil {
		return 0, err
	}

	var upserts int
	for i, e := range feed.Entries {
		if e.Slug == "" {
			continue
		}
		if err := store.UpsertPost(tx, store.Post{
			PostID:        e.PostID,
			Slug:          e.Slug,
			Title:         e.Title,
			Tagline:       e.Tagline,
			Author:        e.Author,
			DiscussionURL: e.DiscussionURL,
			ExternalURL:   e.ExternalURL,
			PublishedAt:   e.Published,
			UpdatedAt:     e.Updated,
		}); err != nil {
			return 0, err
		}
		rank := i + 1
		if err := store.RecordSnapshotEntry(tx, snapID, e.PostID, rank, e.ExternalURL); err != nil {
			return 0, err
		}
		upserts++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit sync tx: %w", err)
	}
	rollback = false

	// Stamp last_sync_at + caller so the next EnsureFresh call sees a
	// warm store without re-running. Non-fatal if this stamp fails —
	// we already have the data committed.
	_ = db.RecordSync(caller)
	return upserts, nil
}

// autoWarm is the one-liner that read commands call at the top of their
// RunE. Resolves the default DB path when the per-command --db flag is
// empty, runs EnsureFresh, and stashes the result on flags for later
// attachAutoSyncMeta calls.
func autoWarm(flags *rootFlags, dbPath string) {
	if dbPath == "" {
		dbPath = defaultDBPath("producthunt-pp-cli")
	}
	flags.autoSyncMeta = EnsureFresh(flags, dbPath)
}

// attachAutoSyncMeta merges meta into a top-level JSON object under the
// _meta.auto_synced key. No-op when meta is nil, or when called with a
// non-object JSON document.
//
// Usage at a command's end, before printOutputWithFlags:
//
//	out, _ := json.Marshal(result)
//	out = attachAutoSyncMeta(out, flags.autoSyncMeta)
//	return printOutputWithFlags(w, out, flags)
func attachAutoSyncMeta(body []byte, meta *AutoSyncMeta) []byte {
	if meta == nil {
		return body
	}
	// Decode, attach, re-encode. JSON tagging on AutoSyncMeta handles field
	// naming automatically.
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		// Not a JSON object (might be an array or primitive). Return the
		// body unchanged rather than mangling it.
		return body
	}

	// Build or merge existing _meta.
	var metaMap map[string]json.RawMessage
	if existing, ok := obj["_meta"]; ok {
		_ = json.Unmarshal(existing, &metaMap)
	}
	if metaMap == nil {
		metaMap = map[string]json.RawMessage{}
	}
	enc, err := json.Marshal(meta)
	if err != nil {
		return body
	}
	metaMap["auto_synced"] = enc

	encMeta, err := json.Marshal(metaMap)
	if err != nil {
		return body
	}
	obj["_meta"] = encMeta

	out, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return out
}
