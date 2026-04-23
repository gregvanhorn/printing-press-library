package cli

import (
	"encoding/json"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/producthunt/internal/store"
)

// postPayload is the JSON shape the CLI emits for a single post. Stable by
// design — every command that returns posts (today, info, list, search,
// watch, trend, etc.) uses this so `--select` paths work consistently.
type postPayload struct {
	ID            int64  `json:"id"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Tagline       string `json:"tagline,omitempty"`
	Author        string `json:"author,omitempty"`
	DiscussionURL string `json:"discussion_url,omitempty"`
	ExternalURL   string `json:"external_url,omitempty"`
	Published     string `json:"published,omitempty"`
	Updated       string `json:"updated,omitempty"`
	FirstSeen     string `json:"first_seen,omitempty"`
	LastSeen      string `json:"last_seen,omitempty"`
	SeenCount     int    `json:"seen_count,omitempty"`
	Rank          int    `json:"rank,omitempty"`
}

func postPayloadOf(p store.Post) postPayload {
	return postPayload{
		ID:            p.PostID,
		Slug:          p.Slug,
		Title:         p.Title,
		Tagline:       p.Tagline,
		Author:        p.Author,
		DiscussionURL: p.DiscussionURL,
		ExternalURL:   p.ExternalURL,
		Published:     fmtTime(p.PublishedAt),
		Updated:       fmtTime(p.UpdatedAt),
		FirstSeen:     fmtTime(p.FirstSeenAt),
		LastSeen:      fmtTime(p.LastSeenAt),
		SeenCount:     p.SeenCount,
	}
}

// fmtTime returns an RFC3339 string or "" for the zero time.
func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// postsToJSON marshals a slice of store.Post as a JSON array of postPayload.
func postsToJSON(posts []store.Post) json.RawMessage {
	out := make([]postPayload, len(posts))
	for i, p := range posts {
		out[i] = postPayloadOf(p)
	}
	buf, _ := json.Marshal(out)
	return json.RawMessage(buf)
}

// postToJSON marshals a single post. Used by commands like `info` that return
// one object rather than an array.
func postToJSON(p store.Post) json.RawMessage {
	buf, _ := json.Marshal(postPayloadOf(p))
	return json.RawMessage(buf)
}

// openStore opens the CLI's default store and runs EnsurePHTables.
// Callers `defer db.Close()` on success.
func openStore(dbPath string) (*store.Store, error) {
	if dbPath == "" {
		dbPath = defaultDBPath("producthunt-pp-cli")
	}
	db, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	if err := store.EnsurePHTables(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
