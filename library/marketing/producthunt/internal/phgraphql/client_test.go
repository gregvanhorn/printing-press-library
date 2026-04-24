package phgraphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient points a Client at a mock server and gives it a token.
func newTestClient(server *httptest.Server) *Client {
	c := NewClient("test-token", "producthunt-pp-cli-test/1.0")
	c.Endpoint = server.URL
	return c
}

func TestExecute_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", got)
		}
		if got := r.Header.Get("User-Agent"); got != "producthunt-pp-cli-test/1.0" {
			t.Errorf("User-Agent = %q", got)
		}
		w.Header().Set("X-Rate-Limit-Limit", "6250")
		w.Header().Set("X-Rate-Limit-Remaining", "6245")
		w.Header().Set("X-Rate-Limit-Reset", "900")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data":{"posts":{"pageInfo":{"hasNextPage":false,"endCursor":""},"edges":[]}}}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.Execute(context.Background(), BackfillPostsQuery, map[string]any{"first": 20})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.HasErrors() {
		t.Fatalf("unexpected graphql errors: %s", resp.ErrorMessage())
	}
	budget := c.Budget()
	if !budget.Known() {
		t.Fatalf("budget should be known after successful response")
	}
	if budget.Limit != 6250 || budget.Remaining != 6245 {
		t.Fatalf("budget = %+v, want Limit=6250 Remaining=6245", budget)
	}
	if budget.PercentRemaining() < 0.99 {
		t.Fatalf("PercentRemaining = %f, want ~1.0", budget.PercentRemaining())
	}
}

func TestExecute_UnknownBudgetBeforeFirstCall(t *testing.T) {
	c := NewClient("x", "ua")
	b := c.Budget()
	if b.Known() {
		t.Fatalf("fresh client budget should be unknown, got %+v", b)
	}
	if b.PercentRemaining() != 1.0 {
		t.Fatalf("unknown budget should report 1.0 remaining, got %f", b.PercentRemaining())
	}
}

func TestExecute_429ReturnsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit-Limit", "6250")
		w.Header().Set("X-Rate-Limit-Remaining", "0")
		w.Header().Set("X-Rate-Limit-Reset", "300")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintln(w, `{"error":"rate limited"}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Execute(context.Background(), "query{}", nil)
	var rle *RateLimitedError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitedError, got %T: %v", err, err)
	}
	// ResetAt should be roughly now + 300s.
	delta := time.Until(rle.ResetAt).Round(time.Second)
	if delta < 250*time.Second || delta > 310*time.Second {
		t.Fatalf("ResetAt delta = %s, want ~300s", delta)
	}
	// Budget snapshot reflects the 429 state — remaining = 0.
	if c.Budget().Remaining != 0 {
		t.Fatalf("post-429 remaining should be 0, got %d", c.Budget().Remaining)
	}
}

func TestExecute_401ReturnsAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"error":"invalid_token"}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Execute(context.Background(), "query{}", nil)
	var ae *AuthError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AuthError, got %T: %v", err, err)
	}
	if ae.StatusCode != 401 {
		t.Fatalf("AuthError.StatusCode = %d, want 401", ae.StatusCode)
	}
}

func TestExecute_500ReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "boom")
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Execute(context.Background(), "query{}", nil)
	var ae *APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if ae.StatusCode != 500 {
		t.Fatalf("APIError.StatusCode = %d, want 500", ae.StatusCode)
	}
}

func TestExecute_NoTokenReturnsAuthError(t *testing.T) {
	c := NewClient("", "ua")
	_, err := c.Execute(context.Background(), "query{}", nil)
	var ae *AuthError
	if !errors.As(err, &ae) {
		t.Fatalf("expected AuthError for missing token, got %T: %v", err, err)
	}
}

func TestExecute_GraphQLLevelErrorsAreReturnedNotRaised(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Rate-Limit-Limit", "6250")
		w.Header().Set("X-Rate-Limit-Remaining", "6000")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data":null,"errors":[{"message":"Field 'foo' doesn't exist on type 'Query'"}]}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	resp, err := c.Execute(context.Background(), "query{}", nil)
	if err != nil {
		t.Fatalf("graphql-level errors should surface on Response, not as transport error: %v", err)
	}
	if !resp.HasErrors() {
		t.Fatalf("HasErrors should be true")
	}
	if !strings.Contains(resp.ErrorMessage(), "doesn't exist") {
		t.Fatalf("ErrorMessage = %q", resp.ErrorMessage())
	}
}

func TestExecute_MalformedBudgetHeadersArePreserved(t *testing.T) {
	// First response populates budget with real values.
	// Second response returns malformed headers — client should keep the
	// last known good values rather than zeroing out.
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call == 1 {
			w.Header().Set("X-Rate-Limit-Limit", "6250")
			w.Header().Set("X-Rate-Limit-Remaining", "6000")
			w.Header().Set("X-Rate-Limit-Reset", "900")
		} else {
			w.Header().Set("X-Rate-Limit-Limit", "not-a-number")
			w.Header().Set("X-Rate-Limit-Remaining", "")
			w.Header().Set("X-Rate-Limit-Reset", "")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data":{}}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if _, err := c.Execute(context.Background(), "q", nil); err != nil {
		t.Fatalf("first call: %v", err)
	}
	first := c.Budget()
	if _, err := c.Execute(context.Background(), "q", nil); err != nil {
		t.Fatalf("second call: %v", err)
	}
	second := c.Budget()
	// Limit/Remaining from the first call should be preserved through
	// the malformed second response.
	if second.Limit != first.Limit {
		t.Fatalf("malformed headers clobbered Limit: first=%d second=%d", first.Limit, second.Limit)
	}
}

func TestExecute_SendsUserAgentAndContentType(t *testing.T) {
	gotUA := ""
	gotCT := ""
	gotAccept := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotCT = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")
		b, _ := io.ReadAll(r.Body)
		var envelope map[string]any
		if err := json.Unmarshal(b, &envelope); err != nil {
			t.Errorf("request body not JSON: %v", err)
		}
		if _, ok := envelope["query"]; !ok {
			t.Errorf("request body missing 'query' field")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"data":{}}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	if _, err := c.Execute(context.Background(), "query Foo { me { id } }", nil); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotUA != "producthunt-pp-cli-test/1.0" {
		t.Fatalf("User-Agent header not sent: %q", gotUA)
	}
	if gotCT != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotCT)
	}
	if gotAccept != "application/json" {
		t.Fatalf("Accept = %q, want application/json", gotAccept)
	}
}

func TestExecute_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintln(w, `{"data":{}}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := c.Execute(ctx, "q", nil)
	if err == nil {
		t.Fatalf("expected context deadline error, got nil")
	}
}
