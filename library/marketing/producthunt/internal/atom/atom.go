// Package atom parses the public Product Hunt Atom feed at
// https://www.producthunt.com/feed.
//
// The feed is stable Atom 1.0. Each <entry> carries:
//   - <id>tag:www.producthunt.com,2005:Post/{numeric}</id>
//   - <published> and <updated> ISO 8601 timestamps
//   - <link rel="alternate" type="text/html" href="https://www.producthunt.com/products/{slug}"/>
//   - <title>Product name</title>
//   - <content type="html">...tagline paragraph + Discussion/Link paragraph...</content>
//   - <author><name>Maker or hunter display name</name></author>
//
// This package does not hit the network itself; pass a []byte to Parse.
// Clients fetch /feed through the generated internal/client package.
package atom

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Feed is a parsed Atom 1.0 feed from producthunt.com/feed.
type Feed struct {
	ID      string
	Title   string
	Updated time.Time
	Entries []Entry
}

// Entry is one featured launch from the Product Hunt feed.
// Fields are derived from the raw Atom XML; see PostID for the numeric key.
type Entry struct {
	// PostID is the numeric Product Hunt ID extracted from the Atom <id>
	// tag URI (tag:www.producthunt.com,2005:Post/{N}). Stable and unique.
	PostID int64
	// Slug is the product slug extracted from the canonical alternate link.
	// The PH URL shape is /products/{slug}; slug is the last path segment.
	Slug string
	// Title is the product name from <title>.
	Title string
	// Tagline is the first HTML paragraph inside <content>, unwrapped and
	// whitespace-collapsed. Safe to display in terminal.
	Tagline string
	// DiscussionURL is the canonical PH product page (the <link rel=alternate> href).
	DiscussionURL string
	// ExternalURL is the product's own landing page, reached via the PH redirect
	// endpoint (/r/p/{PostID}?app_id=339) — PH embeds this in the content HTML.
	// Empty if the feed entry did not contain one.
	ExternalURL string
	// Author is the display name from <author><name>.
	Author string
	// Published / Updated are parsed timestamps; both are typically set.
	Published time.Time
	Updated   time.Time
}

// atomFeed mirrors the wire-level Atom XML shape.
type atomFeed struct {
	XMLName xml.Name   `xml:"feed"`
	ID      string     `xml:"id"`
	Title   string     `xml:"title"`
	Updated string     `xml:"updated"`
	Entries []atomItem `xml:"entry"`
}

type atomItem struct {
	ID        string     `xml:"id"`
	Title     string     `xml:"title"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	Links     []atomLink `xml:"link"`
	Content   atomText   `xml:"content"`
	Author    atomAuthor `xml:"author"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

type atomText struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

type atomAuthor struct {
	Name string `xml:"name"`
}

// postIDRE extracts the numeric suffix from the Atom ID tag URI.
// The feed always emits: tag:www.producthunt.com,2005:Post/{numeric}
// The regex is tight (anchored suffix) so it fails loudly on schema drift.
var postIDRE = regexp.MustCompile(`Post/(\d+)$`)

// slugFromURL returns the trailing path segment of a PH product URL.
// Inputs look like https://www.producthunt.com/products/{slug}. On malformed
// inputs it returns the empty string, signaling the caller to skip the entry.
func slugFromURL(u string) string {
	if u == "" {
		return ""
	}
	// strip trailing slash then query string
	if i := strings.Index(u, "?"); i >= 0 {
		u = u[:i]
	}
	u = strings.TrimRight(u, "/")
	if i := strings.LastIndex(u, "/"); i >= 0 {
		return u[i+1:]
	}
	return ""
}

// externalURLFromContent walks an entry's content HTML for the /r/p/{id} link.
// PH's atom feed wraps the tagline in a <p>, then places Discussion and Link
// anchors in a second <p>. The Link anchor points at /r/p/{id}, the PH redirect
// that ultimately forwards to the product's own landing page. Returning that
// URL is enough — the CLI's open/info commands can either hop the redirect
// server-side or hand it to the user's default browser.
var externalURLRE = regexp.MustCompile(`https?://www\.producthunt\.com/r/p/\d+\?app_id=\d+`)

// firstParagraphRE grabs the first <p>...</p> block from the content HTML.
// Atom feeds declare content type="html" with angle brackets entity-encoded,
// so the encoder pre-decodes; xml.Unmarshal already handed us real "<p>".
var firstParagraphRE = regexp.MustCompile(`(?s)<p>\s*(.*?)\s*</p>`)

// wsRE collapses any run of whitespace (including newlines) to a single space.
var wsRE = regexp.MustCompile(`\s+`)

// Parse decodes an Atom feed body into a strongly-typed Feed. Malformed
// entries are skipped with a best-effort recovery rather than failing the
// whole parse; the caller receives every entry the feed produced in order.
// An outright XML decode failure is returned as an error.
func Parse(body []byte) (*Feed, error) {
	var raw atomFeed
	if err := xml.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode atom: %w", err)
	}

	feed := &Feed{
		ID:    raw.ID,
		Title: raw.Title,
	}
	feed.Updated, _ = time.Parse(time.RFC3339, raw.Updated)

	for _, it := range raw.Entries {
		m := postIDRE.FindStringSubmatch(it.ID)
		if len(m) != 2 {
			// ID didn't match Post/<N>; skip rather than guess
			continue
		}
		postID, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			continue
		}

		var alternateHref string
		for _, l := range it.Links {
			if l.Rel == "alternate" && strings.Contains(l.Href, "producthunt.com") {
				alternateHref = l.Href
				break
			}
		}

		entry := Entry{
			PostID:        postID,
			Slug:          slugFromURL(alternateHref),
			Title:         strings.TrimSpace(it.Title),
			DiscussionURL: alternateHref,
			Author:        strings.TrimSpace(it.Author.Name),
		}

		// Content HTML: first <p> is the tagline; a /r/p link appears in the
		// second <p>. Both are best-effort.
		if m := firstParagraphRE.FindStringSubmatch(it.Content.Body); len(m) == 2 {
			entry.Tagline = wsRE.ReplaceAllString(m[1], " ")
		}
		if m := externalURLRE.FindString(it.Content.Body); m != "" {
			entry.ExternalURL = m
		}

		entry.Published, _ = time.Parse(time.RFC3339, it.Published)
		entry.Updated, _ = time.Parse(time.RFC3339, it.Updated)

		feed.Entries = append(feed.Entries, entry)
	}

	return feed, nil
}
