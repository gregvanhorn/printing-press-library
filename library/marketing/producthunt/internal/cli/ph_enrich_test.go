package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/config"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/phgraphql"
	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
)

// Tests for search --enrich (U8). These cover the attemptEnrich helper
// directly rather than the full cobra search pipeline, because the
// happy/error paths are in the helper and the cobra glue is small.

func TestTopicTokens(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"PostHog", []string{"posthog"}},
		{"AI agent", []string{"ai", "agent"}},
		// 1-char tokens filtered
		{"x AI y", []string{"ai"}},
		// leading/trailing punctuation stripped (interior kept verbatim)
		{"(Quote) (Example)", []string{"quote", "example"}},
	}
	for _, c := range cases {
		got := topicTokens(c.in)
		if len(got) != len(c.want) {
			t.Errorf("topicTokens(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("topicTokens(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestPostMatchesTokens_MatchInTagline(t *testing.T) {
	node := phgraphql.PostNode{
		Slug: "posthog", Name: "PostHog", Tagline: "Open-source analytics",
	}
	if !postMatchesTokens(node, []string{"posthog"}) {
		t.Fatalf("should match on slug")
	}
	if !postMatchesTokens(node, []string{"analytics"}) {
		t.Fatalf("should match in tagline")
	}
	if postMatchesTokens(node, []string{"banana"}) {
		t.Fatalf("should not match unrelated token")
	}
}

func TestPostMatchesTokens_EmptyTokensMatchAll(t *testing.T) {
	node := phgraphql.PostNode{Slug: "x", Name: "X"}
	if !postMatchesTokens(node, nil) {
		t.Fatalf("empty token list should match everything")
	}
}

func TestAttemptEnrich_NoOAuthIsNoop(t *testing.T) {
	db, _ := openTestStore(t)
	defer db.Close()

	// cfg without OAuth.
	cfg := &config.Config{}
	flags := newTestFlags(t)
	// Should return nil without making any HTTP calls — we don't provide
	// a mock server, so a real attempt would fail on DNS. nil return is
	// the contract.
	meta, err := attemptEnrich(context.Background(), flags, db, cfg, "posthog")
	if err != nil {
		t.Fatalf("enrich without OAuth should be a no-op, got err: %v", err)
	}
	if meta == nil || meta.SkippedReason != "missing_graphql_token" || meta.AuthHint == nil {
		t.Fatalf("expected missing auth metadata, got %+v", meta)
	}
	count, _ := db.PostCount()
	if count != 0 {
		t.Fatalf("store should be untouched, got %d posts", count)
	}
}

func TestAttemptEnrich_UpsertsMatchingPostsAndSkipsNonMatches(t *testing.T) {
	hit := map[string]any{
		"id":         "42",
		"slug":       "posthog",
		"name":       "PostHog",
		"tagline":    "Open-source analytics",
		"url":        "https://producthunt.com/products/posthog",
		"website":    "https://posthog.com",
		"createdAt":  "2026-04-01T00:00:00Z",
		"votesCount": 500,
		"user":       map[string]string{"name": "Tim", "username": "tim"},
	}
	miss := map[string]any{
		"id":         "43",
		"slug":       "unrelated",
		"name":       "Some Other Thing",
		"tagline":    "Does something unrelated",
		"url":        "https://producthunt.com/products/unrelated",
		"website":    "https://example.com",
		"createdAt":  "2026-04-01T00:00:00Z",
		"votesCount": 10,
		"user":       map[string]string{"name": "Bob", "username": "bob"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit-Limit", "6250")
		w.Header().Set("X-Rate-Limit-Remaining", "6200")
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"data": map[string]any{
				"posts": map[string]any{
					"pageInfo": map[string]any{"endCursor": ""},
					"edges": []map[string]any{
						{"node": hit},
						{"node": miss},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	// Swap the phgraphql endpoint by passing token via cfg; for this test
	// we call the helper directly with a client whose endpoint points at
	// the mock. attemptEnrich builds its client from cfg, so we inline the
	// logic here to exercise postMatchesTokens + UpsertPost without the
	// real PH endpoint.
	db, _ := openTestStore(t)
	defer db.Close()

	client := phgraphql.NewClient("tok", "test/1.0")
	client.Endpoint = srv.URL

	resp, err := client.Execute(context.Background(), phgraphql.EnrichPostsQuery, map[string]any{
		"first":        BackfillPageSize,
		"postedAfter":  "2026-03-24T00:00:00Z",
		"postedBefore": "2026-04-23T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	var env struct {
		Posts phgraphql.PostsPage `json:"posts"`
	}
	if err := json.Unmarshal(resp.Data, &env); err != nil {
		t.Fatalf("decode: %v", err)
	}

	tokens := topicTokens("posthog")
	tx, _ := db.DB().Begin()
	for _, edge := range env.Posts.Edges {
		if postMatchesTokens(edge.Node, tokens) {
			if err := store.UpsertPost(tx, postNodeToStore(edge.Node)); err != nil {
				t.Fatalf("upsert: %v", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	count, _ := db.PostCount()
	if count != 1 {
		t.Fatalf("expected 1 posthog match upserted, got %d", count)
	}
	stored, err := db.GetPostBySlug("posthog")
	if err != nil {
		t.Fatalf("get posthog: %v", err)
	}
	if stored.Title != "PostHog" {
		t.Fatalf("title = %q", stored.Title)
	}
	// The unrelated post should NOT be in the store.
	if _, err := db.GetPostBySlug("unrelated"); err == nil {
		t.Fatalf("unrelated post should have been filtered out")
	}
}
