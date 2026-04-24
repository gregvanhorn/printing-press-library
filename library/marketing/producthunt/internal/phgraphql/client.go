// Package phgraphql is a minimal GraphQL client for api.producthunt.com/v2.
//
// The design is deliberately thin: a single Execute(ctx, query, vars) method,
// typed errors for the 2 cases callers need to handle distinctly
// (RateLimitedError, AuthError), and a Budget() reader exposing the state PH
// returns via X-Rate-Limit-* response headers.
//
// Why standard net/http instead of enetx/surf (used by internal/client):
// api.producthunt.com is a clean JSON API endpoint, not CF-gated, so the
// anti-fingerprint TLS stack surf provides is unnecessary overhead here.
// The Atom runtime still uses surf; the two live side by side.
package phgraphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// DefaultEndpoint is the v2 GraphQL endpoint. Kept as a variable so tests can
// point the client at an httptest server without touching the outside world.
const DefaultEndpoint = "https://api.producthunt.com/v2/api/graphql"

// DefaultTokenEndpoint is the OAuth token exchange endpoint. Used by the
// auth register flow; exposed here so tests and callers share one source.
const DefaultTokenEndpoint = "https://api.producthunt.com/v2/oauth/token"

// Budget captures the rate-limit state PH returns on every response. Zero
// value means "unknown" — the client starts there and only populates real
// numbers after the first response.
type Budget struct {
	Limit     int       // X-Rate-Limit-Limit — total complexity points per window
	Remaining int       // X-Rate-Limit-Remaining — complexity remaining now
	Reset     time.Time // when Remaining returns to Limit (derived from X-Rate-Limit-Reset seconds-from-now)
	UpdatedAt time.Time // when this Budget was last populated from a response
}

// Known reports whether we've seen at least one successful response with
// parseable rate-limit headers. Callers treat Known() == false as "be
// conservative, don't push budget hard yet."
func (b Budget) Known() bool {
	return !b.UpdatedAt.IsZero() && b.Limit > 0
}

// PercentRemaining returns the fraction 0.0-1.0 of budget still available,
// or 1.0 when Budget is unknown (callers get a clean "plenty" default).
func (b Budget) PercentRemaining() float64 {
	if !b.Known() {
		return 1.0
	}
	return float64(b.Remaining) / float64(b.Limit)
}

// Client wraps net/http with PH-specific auth + budget tracking. Safe for
// concurrent use: the budget mutex serialises reads and writes.
type Client struct {
	Endpoint   string
	Token      string
	UserAgent  string
	HTTPClient *http.Client

	mu     sync.Mutex
	budget Budget
}

// NewClient returns a Client ready to issue GraphQL queries. token is the
// OAuth access_token obtained from client_credentials exchange. userAgent
// should identify the caller ("producthunt-pp-cli/1.1.0 (+github.com/mvanhorn/printing-press-library)").
func NewClient(token, userAgent string) *Client {
	return &Client{
		Endpoint:  DefaultEndpoint,
		Token:     token,
		UserAgent: userAgent,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Budget returns a copy of the client's current rate-limit snapshot.
func (c *Client) Budget() Budget {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.budget
}

// RateLimitedError signals that PH returned 429. ResetAt is when the caller
// may retry; callers typically save progress and exit with a pointer at
// ResetAt rather than blocking here.
type RateLimitedError struct {
	ResetAt time.Time
	Body    string
}

func (e *RateLimitedError) Error() string {
	delta := time.Until(e.ResetAt).Round(time.Second)
	if delta <= 0 {
		return "rate limited (reset window passed); retry now"
	}
	return fmt.Sprintf("rate limited; reset in %s", delta)
}

// AuthError signals 401/403 from PH. Callers map this to "run `auth register`".
type AuthError struct {
	StatusCode int
	Body       string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed (HTTP %d): %s", e.StatusCode, truncate(e.Body, 200))
}

// APIError is the catch-all typed error for non-2xx responses that aren't
// 429/401/403. Kept separate so callers can distinguish a 5xx from a bad
// query.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("HTTP %d from PH API: %s", e.StatusCode, truncate(e.Body, 200))
}

// Response is the decoded GraphQL envelope. Errors is non-nil when the API
// returned 200 with a GraphQL-level error (syntax, field resolution).
type Response struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError is one entry in the errors array.
type GraphQLError struct {
	Message string         `json:"message"`
	Path    []any          `json:"path,omitempty"`
	Extras  map[string]any `json:"extensions,omitempty"`
}

// HasErrors reports whether the response carried any GraphQL-level errors.
func (r *Response) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorMessage joins all GraphQL error messages into a single string for
// logging. Callers that need individual errors should iterate Errors directly.
func (r *Response) ErrorMessage() string {
	if !r.HasErrors() {
		return ""
	}
	msgs := make([]string, 0, len(r.Errors))
	for _, e := range r.Errors {
		msgs = append(msgs, e.Message)
	}
	out := msgs[0]
	for i := 1; i < len(msgs); i++ {
		out += "; " + msgs[i]
	}
	return out
}

// Execute issues a GraphQL query against c.Endpoint. query is the GraphQL
// document, variables is nil or a map[string]any of GraphQL variables.
//
// Returns the decoded Response on success (including when Response carries
// GraphQL-level errors — those are the caller's problem, not a transport
// failure). Returns typed error values for transport / auth / rate-limit
// failures.
func (c *Client) Execute(ctx context.Context, query string, variables map[string]any) (*Response, error) {
	if c.Token == "" {
		return nil, &AuthError{StatusCode: 0, Body: "no token configured"}
	}

	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Always update the budget snapshot — even on error, PH returns the
	// current headers on 4xx responses so callers can know how hard they
	// were throttled.
	c.absorbBudgetHeaders(resp.Header)

	switch {
	case resp.StatusCode == http.StatusOK:
		// Fall through to body parse.
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, &RateLimitedError{
			ResetAt: c.Budget().Reset,
			Body:    string(respBody),
		}
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return nil, &AuthError{StatusCode: resp.StatusCode, Body: string(respBody)}
	default:
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var out Response
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, truncate(string(respBody), 200))
	}
	return &out, nil
}

// absorbBudgetHeaders parses PH's rate-limit headers and updates the client's
// budget snapshot. Unknown / malformed headers leave previous state intact.
func (c *Client) absorbBudgetHeaders(h http.Header) {
	limit := parseIntHeader(h.Get("X-Rate-Limit-Limit"))
	remaining := parseIntHeader(h.Get("X-Rate-Limit-Remaining"))
	resetSecs := parseIntHeader(h.Get("X-Rate-Limit-Reset"))

	// If none of the headers came through, treat the response as if it
	// didn't carry budget info — don't clobber the last known values.
	if limit == 0 && remaining == 0 && resetSecs == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if limit > 0 {
		c.budget.Limit = limit
	}
	c.budget.Remaining = remaining
	if resetSecs > 0 {
		c.budget.Reset = time.Now().Add(time.Duration(resetSecs) * time.Second)
	}
	c.budget.UpdatedAt = time.Now()
}

// parseIntHeader returns the integer value of an HTTP header, or 0 if the
// header is missing or malformed. 0 is the "unknown" sentinel.
func parseIntHeader(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// truncate caps a string to n runes, adding an ellipsis if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
