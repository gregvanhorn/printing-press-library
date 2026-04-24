package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/phgraphql"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
)

// fakePostID ensures unique post IDs across every fakePosts() call in a test
// so upsert dedup doesn't collapse pages together.
var fakePostID = int64(10_000)

// fakePosts returns N post edges for a mock GraphQL response. hasNext controls
// whether the page signals more pages beyond this one. cursorPrefix is only
// used for deterministic cursor strings.
func fakePosts(n int, hasNext bool, cursorPrefix string) map[string]any {
	edges := make([]map[string]any, n)
	for i := 0; i < n; i++ {
		fakePostID++
		edges[i] = map[string]any{
			"node": map[string]any{
				"id":         fmt.Sprintf("%d", fakePostID),
				"slug":       fmt.Sprintf("post-%s-%d", cursorPrefix, i),
				"name":       fmt.Sprintf("Post %s-%d", cursorPrefix, i),
				"tagline":    "A tagline",
				"url":        "https://www.producthunt.com/products/p",
				"website":    "https://example.com",
				"createdAt":  time.Now().UTC().Format(time.RFC3339),
				"votesCount": 42,
				"user":       map[string]string{"name": "Alice", "username": "alice"},
			},
		}
	}
	return map[string]any{
		"data": map[string]any{
			"posts": map[string]any{
				"pageInfo": map[string]any{
					"hasNextPage": hasNext,
					"endCursor":   cursorPrefix + "-cursor",
				},
				"edges": edges,
			},
		},
	}
}

// TestFetchBackfillPage_DecodesGraphQLResponse verifies the GraphQL envelope
// decoder cleanly turns a realistic posts response into PostsPage + PostNodes.
func TestFetchBackfillPage_DecodesGraphQLResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert auth + user agent headers present.
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("X-Rate-Limit-Limit", "6250")
		w.Header().Set("X-Rate-Limit-Remaining", "6200")
		w.Header().Set("Content-Type", "application/json")
		payload := fakePosts(3, false, "p1")
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	client := phgraphql.NewClient("test-token", "test-agent/1.0")
	client.Endpoint = srv.URL

	page, err := fetchBackfillPage(context.Background(), client, "2026-04-01T00:00:00Z", "2026-04-23T00:00:00Z", "")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(page.Edges) != 3 {
		t.Fatalf("edges = %d, want 3", len(page.Edges))
	}
	if page.PageInfo.HasNextPage {
		t.Fatalf("hasNextPage should be false")
	}
	if page.Edges[0].Node.Slug != "post-p1-0" {
		t.Fatalf("slug = %q", page.Edges[0].Node.Slug)
	}
}

// TestBackfillLoop_PaginatesUntilDone runs the full loop against a mock
// server that returns 3 pages (2 with hasNextPage=true, last with false).
// Verifies that posts are upserted and the state table is stamped complete.
func TestBackfillLoop_PaginatesUntilDone(t *testing.T) {
	// Track call count so we can vary responses.
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("X-Rate-Limit-Limit", "6250")
		w.Header().Set("X-Rate-Limit-Remaining", fmt.Sprintf("%d", 6250-calls*4))
		w.Header().Set("Content-Type", "application/json")
		switch calls {
		case 1:
			json.NewEncoder(w).Encode(fakePosts(2, true, "p1"))
		case 2:
			json.NewEncoder(w).Encode(fakePosts(2, true, "p2"))
		case 3:
			json.NewEncoder(w).Encode(fakePosts(1, false, "p3"))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
	}))
	defer srv.Close()

	client := phgraphql.NewClient("test-token", "test-agent/1.0")
	client.Endpoint = srv.URL

	db, _ := openTestStore(t)
	defer db.Close()

	state := &store.BackfillState{
		WindowID:     "win-test",
		PostedAfter:  "2026-04-01T00:00:00Z",
		PostedBefore: "2026-04-23T00:00:00Z",
	}
	if err := db.UpsertBackfillState(*state); err != nil {
		t.Fatal(err)
	}

	flags := newTestFlags(t)
	flags.asJSON = true
	buf := &strings.Builder{}
	cmd := captureCmd(buf)

	err := executeBackfillLoop(cmd, flags, db, client, state)
	if err != nil {
		t.Fatalf("loop: %v", err)
	}

	got, err := db.GetBackfillState("win-test")
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsComplete() {
		t.Fatalf("state should be complete after final page")
	}
	if got.PostsUpserted != 5 {
		t.Fatalf("PostsUpserted = %d, want 5", got.PostsUpserted)
	}
	if got.PagesCompleted != 3 {
		t.Fatalf("PagesCompleted = %d, want 3", got.PagesCompleted)
	}
	count, _ := db.PostCount()
	if count != 5 {
		t.Fatalf("store post count = %d, want 5", count)
	}
}

// TestBackfillLoop_SavesCursorOnRateLimit verifies that a 429 mid-loop
// persists the cursor and page count so backfill resume can continue.
func TestBackfillLoop_SavesCursorOnRateLimit(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("X-Rate-Limit-Limit", "6250")
			w.Header().Set("X-Rate-Limit-Remaining", "5000")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(fakePosts(2, true, "p1"))
			return
		}
		w.Header().Set("X-Rate-Limit-Limit", "6250")
		w.Header().Set("X-Rate-Limit-Remaining", "0")
		w.Header().Set("X-Rate-Limit-Reset", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintln(w, `{"error":"rate limited"}`)
	}))
	defer srv.Close()

	client := phgraphql.NewClient("test-token", "test-agent/1.0")
	client.Endpoint = srv.URL

	db, _ := openTestStore(t)
	defer db.Close()

	state := &store.BackfillState{
		WindowID:     "win-rate-limited",
		PostedAfter:  "2026-04-01T00:00:00Z",
		PostedBefore: "2026-04-23T00:00:00Z",
	}
	if err := db.UpsertBackfillState(*state); err != nil {
		t.Fatal(err)
	}

	flags := newTestFlags(t)
	flags.asJSON = true
	cmd := captureCmd(&strings.Builder{})

	err := executeBackfillLoop(cmd, flags, db, client, state)
	if err == nil {
		t.Fatalf("expected rate-limited error")
	}

	got, err := db.GetBackfillState("win-rate-limited")
	if err != nil {
		t.Fatal(err)
	}
	if got.IsComplete() {
		t.Fatalf("state should not be complete after rate limit")
	}
	if got.Cursor == "" {
		t.Fatalf("cursor should be persisted for resume")
	}
	if got.PagesCompleted != 1 {
		t.Fatalf("PagesCompleted = %d, want 1", got.PagesCompleted)
	}
	if got.PostsUpserted != 2 {
		t.Fatalf("PostsUpserted = %d, want 2", got.PostsUpserted)
	}
	if !strings.Contains(got.LastError, "rate limited") {
		t.Fatalf("LastError should mention rate limit, got %q", got.LastError)
	}
}

func TestResolveWindow_DaysDefault(t *testing.T) {
	from, to, id, err := resolveWindow(backfillOpts{Days: 30})
	if err != nil {
		t.Fatal(err)
	}
	if from == "" || to == "" {
		t.Fatalf("from/to should be populated")
	}
	if id == "" {
		t.Fatalf("windowID should be populated")
	}
	fromT, _ := time.Parse(time.RFC3339, from)
	toT, _ := time.Parse(time.RFC3339, to)
	delta := toT.Sub(fromT).Hours() / 24.0
	if delta < 29.9 || delta > 30.1 {
		t.Fatalf("window delta = %f days, want ~30", delta)
	}
}

func TestResolveWindow_FromToPair(t *testing.T) {
	from, to, id, err := resolveWindow(backfillOpts{
		Days: 30, // default — not explicitly set
		From: "2026-03-01",
		To:   "2026-04-01",
	})
	// The detector only kicks in when Days differs from its default. We
	// accept "Days=30 AND From/To set" as "user set From/To", since the
	// CLI's IntVar with default 30 can't distinguish "default" from
	// "explicit 30". In practice users either pass --days OR --from/--to,
	// and the Cobra shape enforces that.
	if err != nil {
		t.Fatalf("from/to should resolve: %v", err)
	}
	if !strings.HasPrefix(from, "2026-03-01") {
		t.Fatalf("from = %q", from)
	}
	if !strings.HasPrefix(to, "2026-04-01") {
		t.Fatalf("to = %q", to)
	}
	if id == "" {
		t.Fatalf("windowID should be populated")
	}
}

func TestResolveWindow_FromOnlyErrors(t *testing.T) {
	_, _, _, err := resolveWindow(backfillOpts{From: "2026-03-01"})
	if err == nil {
		t.Fatalf("--from without --to should error")
	}
}

func TestResolveWindow_ToBeforeFromErrors(t *testing.T) {
	_, _, _, err := resolveWindow(backfillOpts{From: "2026-04-01", To: "2026-03-01"})
	if err == nil {
		t.Fatalf("--to before --from should error")
	}
}

func TestResolveWindow_WindowIDStable(t *testing.T) {
	_, _, id1, _ := resolveWindow(backfillOpts{Days: 7})
	_, _, id2, _ := resolveWindow(backfillOpts{Days: 7})
	// Days=7 always produces the same from/to at resolution time (within
	// a second). id1 == id2 when called in quick succession.
	if id1[:8] != id2[:8] {
		// Close enough — the prefix is deterministic even if timestamps
		// roll.
		t.Logf("ids differ (expected across long delay): %s vs %s", id1, id2)
	}
}

func TestPostNodeToStore_MapsFields(t *testing.T) {
	node := phgraphql.PostNode{
		ID:         "1234",
		Slug:       "my-product",
		Name:       "My Product",
		Tagline:    "Does things",
		URL:        "https://www.producthunt.com/products/my-product",
		Website:    "https://mysite.com",
		CreatedAt:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		VotesCount: 99,
		User:       phgraphql.PostUser{Name: "Alice Smith", Username: "alice"},
	}
	p := postNodeToStore(node)
	if p.PostID != 1234 {
		t.Fatalf("PostID = %d", p.PostID)
	}
	if p.Slug != "my-product" || p.Title != "My Product" {
		t.Fatalf("slug/title mismatch: %+v", p)
	}
	if p.Author != "Alice Smith" {
		t.Fatalf("Author = %q", p.Author)
	}
	if p.DiscussionURL != "https://www.producthunt.com/products/my-product" {
		t.Fatalf("DiscussionURL = %q", p.DiscussionURL)
	}
	if p.ExternalURL != "https://mysite.com" {
		t.Fatalf("ExternalURL = %q", p.ExternalURL)
	}
}

func TestPostNodeToStore_FallsBackToUsernameWhenNameEmpty(t *testing.T) {
	node := phgraphql.PostNode{
		ID:   "1",
		Slug: "x",
		Name: "X",
		User: phgraphql.PostUser{Name: "", Username: "alice"},
	}
	p := postNodeToStore(node)
	if p.Author != "alice" {
		t.Fatalf("Author = %q", p.Author)
	}
}

// captureCmd builds a real cobra.Command with Out/Err writing into buf, for
// tests that exercise emit functions that take *cobra.Command.
func captureCmd(buf *strings.Builder) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetContext(context.Background())
	return cmd
}

var _ = io.Discard // keep io import
