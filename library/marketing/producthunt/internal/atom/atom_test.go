package atom

import (
	"strings"
	"testing"
	"time"
)

// atomFixture is a trimmed fragment of the real producthunt.com/feed response.
// Preserves the exact shape (tag URIs, entity-encoded content HTML, redirect
// link in the second paragraph) so the parser is exercised against the surface
// it will actually see at runtime.
const atomFixture = `<?xml version="1.0" encoding="UTF-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom">
  <id>tag:www.producthunt.com,2005:/feed</id>
  <link rel="alternate" type="text/html" href="https://www.producthunt.com"/>
  <link rel="self" type="application/atom+xml" href="https://www.producthunt.com/feed"/>
  <title>Product Hunt — Latest</title>
  <updated>2026-04-22T23:02:14-07:00</updated>
  <entry>
    <id>tag:www.producthunt.com,2005:Post/1129094</id>
    <published>2026-04-21T09:02:49-07:00</published>
    <updated>2026-04-22T23:02:14-07:00</updated>
    <link rel="alternate" type="text/html" href="https://www.producthunt.com/products/seeknal"/>
    <title>Seeknal</title>
    <content type="html">          &lt;p&gt;
            Data &amp; AI/ML CLI for pipelines and NL queries
          &lt;/p&gt;
          &lt;p&gt;
            &lt;a href="https://www.producthunt.com/products/seeknal?utm_campaign=producthunt-atom-posts-feed&amp;amp;utm_medium=rss-feed&amp;amp;utm_source=producthunt-atom-posts-feed"&gt;Discussion&lt;/a&gt;
            |
            &lt;a href="https://www.producthunt.com/r/p/1129094?app_id=339"&gt;Link&lt;/a&gt;
          &lt;/p&gt;
    </content>
    <author>
      <name>Fitra Kacamarga</name>
    </author>
  </entry>
  <entry>
    <id>tag:www.producthunt.com,2005:Post/1128832</id>
    <published>2026-04-21T04:15:49-07:00</published>
    <updated>2026-04-22T22:02:46-07:00</updated>
    <link rel="alternate" type="text/html" href="https://www.producthunt.com/products/cavalry-2"/>
    <title>Cavalry Studio</title>
    <content type="html">          &lt;p&gt;
            Free Motion Design tool by Canva
          &lt;/p&gt;
          &lt;p&gt;
            &lt;a href="https://www.producthunt.com/r/p/1128832?app_id=339"&gt;Link&lt;/a&gt;
          &lt;/p&gt;
    </content>
    <author>
      <name>Adithya Shreshti</name>
    </author>
  </entry>
  <entry>
    <id>not-a-post-tag</id>
    <title>Should be skipped</title>
    <link rel="alternate" href="https://www.producthunt.com/products/skip"/>
  </entry>
</feed>
`

func TestParse_Happy(t *testing.T) {
	feed, err := Parse([]byte(atomFixture))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := len(feed.Entries), 2; got != want {
		t.Fatalf("entry count: got %d, want %d (malformed third entry should be skipped)", got, want)
	}

	e := feed.Entries[0]
	if e.PostID != 1129094 {
		t.Errorf("PostID: got %d, want 1129094", e.PostID)
	}
	if e.Slug != "seeknal" {
		t.Errorf("Slug: got %q, want %q", e.Slug, "seeknal")
	}
	if e.Title != "Seeknal" {
		t.Errorf("Title: got %q, want %q", e.Title, "Seeknal")
	}
	if !strings.Contains(e.Tagline, "Data & AI/ML CLI") {
		t.Errorf("Tagline missing: %q", e.Tagline)
	}
	if strings.Contains(e.Tagline, "\n") {
		t.Errorf("Tagline should have whitespace collapsed: %q", e.Tagline)
	}
	if e.DiscussionURL != "https://www.producthunt.com/products/seeknal" {
		t.Errorf("DiscussionURL: %q", e.DiscussionURL)
	}
	if e.ExternalURL != "https://www.producthunt.com/r/p/1129094?app_id=339" {
		t.Errorf("ExternalURL: %q", e.ExternalURL)
	}
	if e.Author != "Fitra Kacamarga" {
		t.Errorf("Author: %q", e.Author)
	}

	wantPub, _ := time.Parse(time.RFC3339, "2026-04-21T09:02:49-07:00")
	if !e.Published.Equal(wantPub) {
		t.Errorf("Published: got %v, want %v", e.Published, wantPub)
	}
}

func TestParse_SkipsMalformedID(t *testing.T) {
	// Third fixture entry has id="not-a-post-tag" — parser must drop it silently.
	feed, err := Parse([]byte(atomFixture))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	for _, e := range feed.Entries {
		if e.Title == "Should be skipped" {
			t.Errorf("expected malformed entry to be skipped, got it in results")
		}
	}
}

func TestParse_InvalidXML(t *testing.T) {
	_, err := Parse([]byte("not xml at all"))
	if err == nil {
		t.Fatal("expected error on invalid XML, got nil")
	}
}

func TestSlugFromURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"https://www.producthunt.com/products/foo", "foo"},
		{"https://www.producthunt.com/products/foo/", "foo"},
		{"https://www.producthunt.com/products/foo?utm=x", "foo"},
		{"", ""},
		{"not-a-url-no-slash", ""}, // no slash -> empty (caller filters the entry)
	}
	for _, tc := range tests {
		if got := slugFromURL(tc.in); got != tc.want {
			t.Errorf("slugFromURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
