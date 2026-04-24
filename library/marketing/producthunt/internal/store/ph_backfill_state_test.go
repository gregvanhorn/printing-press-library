package store

import (
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// TestPHSchemaV2_FreshDBHasBackfillTable verifies that a freshly opened store
// has ph_backfill_state available and stamps PHTablesSchemaVersion = 2 in
// ph_meta. This is the contract U4 and U5 rely on.
func TestPHSchemaV2_FreshDBHasBackfillTable(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()

	if err := EnsurePHTables(s); err != nil {
		t.Fatalf("EnsurePHTables: %v", err)
	}

	var name string
	err = s.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='ph_backfill_state'`).Scan(&name)
	if err != nil {
		t.Fatalf("ph_backfill_state table not present: %v", err)
	}
	if name != "ph_backfill_state" {
		t.Fatalf("expected ph_backfill_state, got %q", name)
	}

	v, err := s.GetPHMeta(PHMetaSchemaVersion)
	if err != nil {
		t.Fatalf("get schema version: %v", err)
	}
	if v != "2" {
		t.Fatalf("ph_schema_version = %q, want \"2\"", v)
	}
}

// TestPHSchemaV2_MigratesFromV1PreservesPosts verifies that a v1 store (which
// has posts, posts_fts, snapshots, snapshot_entries but no ph_backfill_state)
// migrates cleanly to v2 without losing data on the next EnsurePHTables call.
func TestPHSchemaV2_MigratesFromV1PreservesPosts(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	// Simulate a v1 store: run EnsurePHTables at v2, then manually drop
	// ph_backfill_state and re-stamp version to 1. The other tables stay.
	if err := EnsurePHTables(s); err != nil {
		t.Fatalf("initial EnsurePHTables: %v", err)
	}
	if _, err := s.db.Exec(`DROP TABLE ph_backfill_state`); err != nil {
		t.Fatalf("simulate v1 by dropping table: %v", err)
	}
	if err := s.SetPHMeta(PHMetaSchemaVersion, "1"); err != nil {
		t.Fatalf("restamp to v1: %v", err)
	}

	// Insert a sample post directly through the public API so we can verify
	// the migration preserves it.
	tx, err := s.db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := UpsertPost(tx, Post{
		PostID: 12345, Slug: "acme-widget", Title: "Acme Widget",
		Tagline: "Widgets for acme", Author: "Road Runner",
		DiscussionURL: "https://www.producthunt.com/products/acme-widget",
		PublishedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("upsert post: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	s.Close()

	// Re-open: EnsurePHTables runs again and must migrate to v2 cleanly.
	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	if err := EnsurePHTables(s2); err != nil {
		t.Fatalf("migration EnsurePHTables: %v", err)
	}

	// Original post survived.
	got, err := s2.GetPostBySlug("acme-widget")
	if err != nil {
		t.Fatalf("post lost in migration: %v", err)
	}
	if got.Title != "Acme Widget" {
		t.Fatalf("title corrupted: got %q", got.Title)
	}

	// Version is now 2.
	v, err := s2.GetPHMeta(PHMetaSchemaVersion)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if v != "2" {
		t.Fatalf("post-migration version = %q, want \"2\"", v)
	}

	// ph_backfill_state is present.
	var name string
	if err := s2.db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='ph_backfill_state'`,
	).Scan(&name); err != nil {
		t.Fatalf("ph_backfill_state not present after migration: %v", err)
	}
}

// TestPHSchemaV2_EnsureIsIdempotent verifies calling EnsurePHTables twice
// in a row does not error and does not corrupt state.
func TestPHSchemaV2_EnsureIsIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()

	for i := 0; i < 3; i++ {
		if err := EnsurePHTables(s); err != nil {
			t.Fatalf("EnsurePHTables call %d: %v", i, err)
		}
	}
}

// TestPHMeta_RoundTrip exercises the new Get/Set/RecordSync helpers.
func TestPHMeta_RoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	if err := EnsurePHTables(s); err != nil {
		t.Fatalf("EnsurePHTables: %v", err)
	}

	// Unknown key returns empty string, no error.
	v, err := s.GetPHMeta("never_set")
	if err != nil {
		t.Fatalf("get unknown: %v", err)
	}
	if v != "" {
		t.Fatalf("unknown key should return \"\", got %q", v)
	}

	// LastSyncAt on a fresh store is zero.
	zero, err := s.LastSyncAt()
	if err != nil {
		t.Fatalf("last sync at: %v", err)
	}
	if !zero.IsZero() {
		t.Fatalf("fresh store LastSyncAt = %v, want zero", zero)
	}

	// RecordSync stamps both keys.
	if err := s.RecordSync("last30days/3.0.1"); err != nil {
		t.Fatalf("RecordSync: %v", err)
	}
	after, err := s.LastSyncAt()
	if err != nil {
		t.Fatalf("last sync at after record: %v", err)
	}
	if after.IsZero() {
		t.Fatalf("LastSyncAt after RecordSync is still zero")
	}
	caller, err := s.GetPHMeta(PHMetaLastCaller)
	if err != nil {
		t.Fatalf("get last caller: %v", err)
	}
	if caller != "last30days/3.0.1" {
		t.Fatalf("last caller = %q, want %q", caller, "last30days/3.0.1")
	}

	// RecordSync without caller preserves the previous caller value.
	if err := s.RecordSync(""); err != nil {
		t.Fatalf("RecordSync empty caller: %v", err)
	}
	caller2, err := s.GetPHMeta(PHMetaLastCaller)
	if err != nil {
		t.Fatalf("get last caller again: %v", err)
	}
	if caller2 != "last30days/3.0.1" {
		t.Fatalf("caller clobbered by empty RecordSync: got %q", caller2)
	}
}

// TestBackfillState_Lifecycle exercises Upsert / Get / Pending for
// ph_backfill_state rows.
func TestBackfillState_Lifecycle(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	if err := EnsurePHTables(s); err != nil {
		t.Fatalf("EnsurePHTables: %v", err)
	}

	// Fresh windowID returns nil, nil.
	got, err := s.GetBackfillState("win-unknown")
	if err != nil {
		t.Fatalf("get unknown: %v", err)
	}
	if got != nil {
		t.Fatalf("unknown windowID should return nil, got %+v", got)
	}

	// Upsert a partial (in-progress) row.
	in := BackfillState{
		WindowID:       "win-2026-03-24-to-2026-04-23",
		PostedAfter:    "2026-03-24",
		PostedBefore:   "2026-04-23",
		Cursor:         "abc123",
		PagesCompleted: 5,
		PostsUpserted:  100,
		LastRunAt:      time.Now().UTC(),
	}
	if err := s.UpsertBackfillState(in); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Read back.
	got, err = s.GetBackfillState(in.WindowID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatalf("expected state back, got nil")
	}
	if got.Cursor != "abc123" || got.PagesCompleted != 5 || got.PostsUpserted != 100 {
		t.Fatalf("state round-trip mismatch: %+v", got)
	}
	if got.IsComplete() {
		t.Fatalf("incomplete state reports IsComplete() = true")
	}

	// Pending list includes it.
	pending, err := s.PendingBackfillStates()
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}

	// Mark complete.
	in.CompletedAt = time.Now().UTC()
	in.Cursor = ""
	if err := s.UpsertBackfillState(in); err != nil {
		t.Fatalf("upsert complete: %v", err)
	}
	pending, err = s.PendingBackfillStates()
	if err != nil {
		t.Fatalf("pending after complete: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending after complete, got %d", len(pending))
	}
	done, err := s.GetBackfillState(in.WindowID)
	if err != nil {
		t.Fatalf("get complete: %v", err)
	}
	if !done.IsComplete() {
		t.Fatalf("complete state reports IsComplete() = false")
	}
}
