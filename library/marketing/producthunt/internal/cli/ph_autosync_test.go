package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/config"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
)

// newTestFlags returns a rootFlags pointed at a temp config file so tests
// don't trample the user's real config. dbPath is not used here — callers
// pass their own to EnsureFresh.
func newTestFlags(t *testing.T) *rootFlags {
	t.Helper()
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.toml")
	return &rootFlags{configPath: cfgPath, timeout: 30 * time.Second}
}

// openTestStore returns an initialised Store at a temp path.
func openTestStore(t *testing.T) (*store.Store, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "data.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.EnsurePHTables(s); err != nil {
		t.Fatalf("ensure tables: %v", err)
	}
	return s, dbPath
}

func TestEnsureFresh_WarmStoreSkipsSync(t *testing.T) {
	flags := newTestFlags(t)
	db, dbPath := openTestStore(t)
	// Simulate a fresh sync 1 minute ago.
	if err := db.RecordSync("test"); err != nil {
		t.Fatalf("record sync: %v", err)
	}
	db.Close()

	meta := EnsureFresh(flags, dbPath)
	if meta.Ran {
		t.Fatalf("warm store should not trigger sync")
	}
	if meta.Reason != "fresh" {
		t.Fatalf("Reason = %q, want fresh", meta.Reason)
	}
	if meta.LastSyncAt == "" {
		t.Fatalf("LastSyncAt should be populated when store has prior sync")
	}
}

func TestEnsureFresh_NoAutoSyncFlagDisables(t *testing.T) {
	flags := newTestFlags(t)
	flags.noAutoSync = true
	_, dbPath := openTestStore(t)

	meta := EnsureFresh(flags, dbPath)
	if meta.Ran {
		t.Fatalf("--no-auto-sync should suppress sync")
	}
	if meta.Reason != "disabled" {
		t.Fatalf("Reason = %q, want disabled", meta.Reason)
	}
}

func TestEnsureFresh_ConfigAutoSyncFalseDisables(t *testing.T) {
	flags := newTestFlags(t)

	// Write a config with auto_sync = false.
	cfg := &config.Config{Path: flags.configPath}
	disabled := false
	cfg.AutoSync = &disabled
	if err := writeTestConfig(cfg); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, dbPath := openTestStore(t)

	meta := EnsureFresh(flags, dbPath)
	if meta.Ran {
		t.Fatalf("config auto_sync=false should suppress sync")
	}
	if meta.Reason != "disabled" {
		t.Fatalf("Reason = %q, want disabled", meta.Reason)
	}
}

func TestEnsureFresh_CallerRecorded(t *testing.T) {
	flags := newTestFlags(t)
	flags.caller = "last30days/3.0.1"
	db, dbPath := openTestStore(t)
	if err := db.RecordSync("test"); err != nil {
		t.Fatalf("seed sync: %v", err)
	}
	db.Close()

	meta := EnsureFresh(flags, dbPath)
	if meta.Caller != "last30days/3.0.1" {
		t.Fatalf("Caller = %q", meta.Caller)
	}
}

func TestEnsureFresh_NeverSyncedReason(t *testing.T) {
	flags := newTestFlags(t)
	flags.noAutoSync = true // skip the actual network call; we only want to assert the reason path
	_, dbPath := openTestStore(t)

	// Even with noAutoSync, the reason should be "disabled" here since the
	// noAutoSync check wins. But with auto-sync enabled and no prior
	// sync, the reason path evaluates to "never_synced" before an
	// attempted sync. This test documents the precedence.
	meta := EnsureFresh(flags, dbPath)
	if meta.Reason != "disabled" {
		t.Fatalf("noAutoSync should take precedence, got %q", meta.Reason)
	}
}

func TestAttachAutoSyncMeta_ObjectOutput(t *testing.T) {
	in := []byte(`{"title":"Foo","count":5}`)
	meta := &AutoSyncMeta{Ran: true, Reason: "stale", PostsUpserted: 42}

	out := attachAutoSyncMeta(in, meta)

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("output not JSON: %v", err)
	}
	metaRaw, ok := parsed["_meta"]
	if !ok {
		t.Fatalf("_meta not attached: %s", out)
	}
	var metaWrap map[string]json.RawMessage
	if err := json.Unmarshal(metaRaw, &metaWrap); err != nil {
		t.Fatalf("_meta not JSON: %v", err)
	}
	if _, ok := metaWrap["auto_synced"]; !ok {
		t.Fatalf("_meta.auto_synced not present: %s", metaRaw)
	}
	// Existing fields preserved.
	if string(parsed["title"]) != `"Foo"` {
		t.Fatalf("existing field clobbered: %s", parsed["title"])
	}
}

func TestAttachAutoSyncMeta_ArrayOutputUnchanged(t *testing.T) {
	// Array-shaped outputs (list, search, today) can't carry _meta cleanly.
	// The helper must not corrupt them.
	in := []byte(`[{"id":1},{"id":2}]`)
	meta := &AutoSyncMeta{Ran: true}
	out := attachAutoSyncMeta(in, meta)
	if string(out) != string(in) {
		t.Fatalf("array output mangled: %s", out)
	}
}

func TestAttachAutoSyncMeta_NilMetaIsNoop(t *testing.T) {
	in := []byte(`{"x":1}`)
	out := attachAutoSyncMeta(in, nil)
	if string(out) != string(in) {
		t.Fatalf("nil meta should be a no-op: %s", out)
	}
}

func TestAttachAutoSyncMeta_MergesWithExistingMeta(t *testing.T) {
	in := []byte(`{"x":1,"_meta":{"source":"atom"}}`)
	meta := &AutoSyncMeta{Ran: true, Reason: "stale"}
	out := attachAutoSyncMeta(in, meta)

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	var metaObj map[string]json.RawMessage
	if err := json.Unmarshal(parsed["_meta"], &metaObj); err != nil {
		t.Fatal(err)
	}
	if _, ok := metaObj["source"]; !ok {
		t.Fatalf("existing _meta.source was clobbered: %s", parsed["_meta"])
	}
	if _, ok := metaObj["auto_synced"]; !ok {
		t.Fatalf("new _meta.auto_synced not added: %s", parsed["_meta"])
	}
}

func TestAutoWarm_StashesMetaOnFlags(t *testing.T) {
	flags := newTestFlags(t)
	flags.noAutoSync = true
	_, dbPath := openTestStore(t)

	if flags.autoSyncMeta != nil {
		t.Fatalf("flags.autoSyncMeta should start nil")
	}
	autoWarm(flags, dbPath)
	if flags.autoSyncMeta == nil {
		t.Fatalf("autoWarm did not populate flags.autoSyncMeta")
	}
	if flags.autoSyncMeta.Reason != "disabled" {
		t.Fatalf("Reason = %q, want disabled", flags.autoSyncMeta.Reason)
	}
}

// writeTestConfig marshals a config.Config to TOML and writes it to cfg.Path
// with 0600 perms. Used by tests that need to seed a specific auto_sync
// state without going through the full auth flow.
func writeTestConfig(cfg *config.Config) error {
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// Minimal TOML: just the auto_sync key when configured.
	body := "base_url = \"https://www.producthunt.com\"\n"
	if cfg.AutoSync != nil {
		if *cfg.AutoSync {
			body += "auto_sync = true\n"
		} else {
			body += "auto_sync = false\n"
		}
	}
	if strings.TrimSpace(cfg.AccessToken) != "" {
		body += "access_token = \"" + cfg.AccessToken + "\"\n"
	}
	return os.WriteFile(cfg.Path, []byte(body), 0o600)
}
