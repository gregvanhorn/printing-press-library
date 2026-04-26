package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// fetchTrackerRaw fetches raw tracker data directly from tracker.dominos.com,
// bypassing the order.dominos.com base URL. The Domino's tracker lives on a
// different host than the rest of /power/* endpoints — sniff capture
// 2026-04-25 confirmed: GET https://tracker.dominos.com/orderstorage/GetTrackerData?Phone=<phone>
// returns text/xml with order status. Discovered live during Phase 5 dogfood.
func fetchTrackerRaw(params map[string]string) (json.RawMessage, error) {
	phone := params["Phone"]
	q := url.Values{}
	if phone != "" {
		q.Set("Phone", phone)
	}
	u := "https://tracker.dominos.com/orderstorage/GetTrackerData?" + q.Encode()
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/xml")
	req.Header.Set("User-Agent", "dominos-pp-cli/1.0.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tracker request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading tracker response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("tracker returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	return json.RawMessage(body), nil
}
