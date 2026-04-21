// Copyright 2026 matt-van-horn. Licensed under Apache-2.0.
//
// Tests for Client.Authenticate use httptest.NewServer to serve canned
// responses; no live calls to Expensify happen here.

package client

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mvanhorn/printing-press-library/library/productivity/expensify/internal/config"
)

// newAuthenticateTestClient wires a Client whose BaseURL points at the given
// httptest server. The config carries a pre-set (stale) authToken so we can
// assert the Authenticate path does NOT forward it on the wire.
func newAuthenticateTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	cfg := &config.Config{
		BaseURL:            serverURL, // no "www.expensify.com/api" suffix → treated as override
		ExpensifyAuthToken: "stale-session-token-should-not-be-sent",
	}
	return New(cfg, 5*time.Second, 0 /* disable limiter for tests */)
}

// authServer sets up an httptest server that captures the last inbound request
// form for assertions, then responds with the caller-provided body + status.
type authServer struct {
	srv        *httptest.Server
	lastMethod string
	lastPath   string
	lastForm   url.Values
}

func newAuthServer(t *testing.T, status int, body string) *authServer {
	t.Helper()
	a := &authServer{}
	a.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.lastMethod = r.Method
		a.lastPath = r.URL.Path
		// ParseForm reads and parses the request body; must be called before
		// writing the response so PostForm is populated.
		_ = r.ParseForm()
		a.lastForm = r.PostForm
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(a.srv.Close)
	return a
}

func TestAuthenticate_Success(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":200,"authToken":"abc","email":"a@b.com","accountID":42,"expires":1724000000}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	got, err := c.Authenticate("a@b.com", "pw")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if got.AuthToken != "abc" {
		t.Fatalf("AuthToken = %q, want %q", got.AuthToken, "abc")
	}
	if got.Email != "a@b.com" {
		t.Fatalf("Email = %q, want %q", got.Email, "a@b.com")
	}
	if got.AccountID != 42 {
		t.Fatalf("AccountID = %d, want 42", got.AccountID)
	}
	wantTime := time.Unix(1724000000, 0).UTC()
	if !got.ExpiresAt.Equal(wantTime) {
		t.Fatalf("ExpiresAt = %v, want %v", got.ExpiresAt, wantTime)
	}
	if a.lastPath != "/Authenticate" {
		t.Fatalf("server saw path %q, want /Authenticate", a.lastPath)
	}
	if a.lastMethod != "POST" {
		t.Fatalf("server saw method %q, want POST", a.lastMethod)
	}
}

func TestAuthenticate_TwoFactor(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":402,"message":"twoFactorAuth required"}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	_, err := c.Authenticate("a@b.com", "pw")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("err = %v, want ErrTwoFactorRequired", err)
	}
}

func TestAuthenticate_TwoFactorFlag(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":402,"requiresTwoFactorAuth":true,"message":"blah"}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	_, err := c.Authenticate("a@b.com", "pw")
	if !errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("err = %v, want ErrTwoFactorRequired", err)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":403,"message":"Incorrect password"}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	_, err := c.Authenticate("a@b.com", "pw")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthenticate_InvalidCredentials402(t *testing.T) {
	// 402 without 2FA indicators should also be treated as invalid creds.
	a := newAuthServer(t, 200, `{"jsonCode":402,"message":"User account is frozen"}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	_, err := c.Authenticate("a@b.com", "pw")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthenticate_Unknown(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":500,"message":"oops"}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	_, err := c.Authenticate("a@b.com", "pw")
	if err == nil {
		t.Fatalf("expected generic error for jsonCode 500")
	}
	if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrTwoFactorRequired) {
		t.Fatalf("err = %v, should not be a typed auth error", err)
	}
	if !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "oops") {
		t.Fatalf("err = %v, want it to include code 500 and message 'oops'", err)
	}
}

func TestAuthenticate_MissingToken(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":200}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	_, err := c.Authenticate("a@b.com", "pw")
	if !errors.Is(err, ErrAuthenticateMissing) {
		t.Fatalf("err = %v, want ErrAuthenticateMissing", err)
	}
}

func TestAuthenticate_NetworkError(t *testing.T) {
	// Point at a server that immediately closes connections — any request will
	// error at the transport layer.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hijack and close without writing anything.
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Skip("httptest server doesn't support hijacking on this platform")
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			return
		}
		_ = conn.Close()
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{BaseURL: srv.URL}
	c := New(cfg, 2*time.Second, 0)

	_, err := c.Authenticate("a@b.com", "pw")
	if err == nil {
		t.Fatalf("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "authenticate:") {
		t.Fatalf("err = %v, want it to mention authenticate", err)
	}
}

func TestAuthenticate_NoSessionTokenSent(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":200,"authToken":"fresh","email":"a@b.com","accountID":1}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	if _, err := c.Authenticate("a@b.com", "pw"); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	// The stale session token from the client's config MUST NOT appear in the
	// form body of the authenticate call. Assert both key-absence and value-
	// absence so a future code change that renames the key still trips this.
	if vals, ok := a.lastForm["authToken"]; ok && len(vals) > 0 && vals[0] != "" {
		t.Fatalf("authenticate request leaked authToken=%q in form body", vals[0])
	}
	for k, vals := range a.lastForm {
		for _, v := range vals {
			if v == "stale-session-token-should-not-be-sent" {
				t.Fatalf("stale session token leaked into form field %q", k)
			}
		}
	}

	// Positive asserts: the required partner fields ARE in the form.
	wantFields := map[string]string{
		"partnerName":       partnerName,
		"partnerPassword":   partnerPasswordPublic,
		"partnerUserID":     "a@b.com",
		"partnerUserSecret": "pw",
	}
	for k, want := range wantFields {
		got := ""
		if vals, ok := a.lastForm[k]; ok && len(vals) > 0 {
			got = vals[0]
		}
		if got != want {
			t.Fatalf("form[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestAuthenticate_ISOExpires(t *testing.T) {
	a := newAuthServer(t, 200, `{"jsonCode":200,"authToken":"abc","email":"a@b.com","accountID":1,"expires":"2026-05-01T12:00:00Z"}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	got, err := c.Authenticate("a@b.com", "pw")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	want, err := time.Parse(time.RFC3339, "2026-05-01T12:00:00Z")
	if err != nil {
		t.Fatalf("parsing want: %v", err)
	}
	if !got.ExpiresAt.Equal(want) {
		t.Fatalf("ExpiresAt = %v, want %v", got.ExpiresAt, want)
	}
}

func TestAuthenticate_MissingExpires(t *testing.T) {
	// The API may omit the expires field entirely — ExpiresAt should be zero-value.
	a := newAuthServer(t, 200, `{"jsonCode":200,"authToken":"abc","email":"a@b.com","accountID":1}`)
	c := newAuthenticateTestClient(t, a.srv.URL)

	got, err := c.Authenticate("a@b.com", "pw")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if !got.ExpiresAt.IsZero() {
		t.Fatalf("ExpiresAt = %v, want zero-value", got.ExpiresAt)
	}
}
