package cli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
)

func TestSelectAppliesToMetaWrappedResults(t *testing.T) {
	body := postsToJSON([]store.Post{{
		PostID:        42,
		Slug:          "posthog",
		Title:         "PostHog",
		Tagline:       "Open-source analytics",
		Author:        "tim",
		DiscussionURL: "https://www.producthunt.com/posts/posthog",
		PublishedAt:   time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC),
	}})
	wrapped := withMeta(body, map[string]any{
		"auth_hint": graphQLAuthHint("GraphQL auth can enrich thin local searches."),
	})

	var out bytes.Buffer
	flags := &rootFlags{asJSON: true, selectFields: "slug,title"}
	if err := printOutputWithFlags(&out, wrapped, flags); err != nil {
		t.Fatalf("print output: %v", err)
	}

	var payload struct {
		Results []map[string]any `json:"results"`
		Meta    map[string]any   `json:"_meta"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	if len(payload.Results) != 1 {
		t.Fatalf("results len = %d", len(payload.Results))
	}
	row := payload.Results[0]
	if row["slug"] != "posthog" || row["title"] != "PostHog" {
		t.Fatalf("selected row = %#v", row)
	}
	if _, ok := row["tagline"]; ok {
		t.Fatalf("unselected field survived: %#v", row)
	}
	if payload.Meta["auth_hint"] == nil {
		t.Fatalf("_meta was not preserved: %#v", payload.Meta)
	}
}

func TestAgentCompactPreservesProductHuntFieldsBeforeSelect(t *testing.T) {
	body := postsToJSON([]store.Post{{
		PostID:  42,
		Slug:    "posthog",
		Title:   "PostHog",
		Tagline: "Open-source analytics",
		Author:  "tim",
	}})
	wrapped := withMeta(body, map[string]any{"source": "test"})

	var out bytes.Buffer
	flags := &rootFlags{asJSON: true, compact: true, selectFields: "slug,title,tagline"}
	if err := printOutputWithFlags(&out, wrapped, flags); err != nil {
		t.Fatalf("print output: %v", err)
	}

	var payload struct {
		Results []map[string]any `json:"results"`
		Meta    map[string]any   `json:"_meta"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("decode output: %v\n%s", err, out.String())
	}
	row := payload.Results[0]
	if row["slug"] != "posthog" || row["title"] != "PostHog" || row["tagline"] != "Open-source analytics" {
		t.Fatalf("agent compact/select dropped PH fields: %#v", row)
	}
	if payload.Meta["source"] != "test" {
		t.Fatalf("_meta not preserved: %#v", payload.Meta)
	}
}
