// Package-level Product Hunt feed fetcher used by every command that needs
// live /feed data. Centralizes the browser-compatible User-Agent, timeout
// handling, and content-type sanity check so individual commands stay short.
package cli

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FeedEndpoint is the one and only URL this CLI reads from Product Hunt.
// HTML pages are gated by Cloudflare and are not used at runtime.
const FeedEndpoint = "https://www.producthunt.com/feed"

// ChromeLikeUA matches the current stable Chrome on macOS. Some Cloudflare
// configurations whitelist browser-ish UAs at the edge; this one is enough
// for /feed, which is also fully public without it.
const ChromeLikeUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"

// fetchFeedBody issues a single GET to /feed using the given timeout. Returns
// the raw Atom XML body on 200. Non-2xx responses propagate with their status
// in the error so the doctor command can render a useful message.
func fetchFeedBody(timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", FeedEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", ChromeLikeUA)
	req.Header.Set("Accept", "application/atom+xml, application/xml, text/xml, */*;q=0.1")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", FeedEndpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned HTTP %d (%s); first bytes: %q",
			resp.StatusCode, resp.Status, truncate(string(body), 200))
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "atom") && !strings.Contains(ct, "xml") {
		// PH occasionally sends "text/xml" — accepted above. Anything else
		// (e.g., HTML challenge page smuggled under 200) is a protocol error.
		return nil, fmt.Errorf("feed content-type %q not Atom/XML", ct)
	}

	return body, nil
}
