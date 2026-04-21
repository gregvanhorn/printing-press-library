// Copyright 2026 matt-van-horn. Licensed under Apache-2.0.
//
// Headless authentication against Expensify's internal /Authenticate endpoint.
//
// Endpoint shape (derived from prior art in open-source Expensify clients —
// old node-expensify, expensify-api, agenticledger/expensify-mcp-http, and
// the now-archived Expensify/App mobile client):
//
//   POST https://www.expensify.com/api/Authenticate
//   Content-Type: application/x-www-form-urlencoded
//
//   Request form fields:
//     partnerName       "expensify.com"           (constant — identifies the partner)
//     partnerPassword   "e21965746fd75f82bb66"    (well-known public partner secret;
//                                                  baked into every Expensify client,
//                                                  like Slack's xoxp- prefix — it
//                                                  identifies the partner app, not
//                                                  the user)
//     partnerUserID     user's email
//     partnerUserSecret user's password
//     doNotRetry        "true"
//     api_setCookie     "false"
//     platform          "web"
//     referer           "ecash"                   (matches /ReconnectApp)
//
//   Response (JSON body on HTTP 200):
//     Success:          {"jsonCode": 200, "authToken": "...", "email": "...",
//                        "accountID": <int>, "expires": "<ISO or unix seconds>"}
//     Invalid creds:    {"jsonCode": 403, "message": "..."}
//     2FA required:     {"jsonCode": 402, "message": "...twoFactorAuth...",
//                        "requiresTwoFactorAuth": true}
//     Other:            anything else — wrapped with code + message.
//
// This call is issued WITHOUT the client's existing session authToken: we're
// trying to mint a fresh one, so including a stale token would be meaningless
// and could confuse the server. The postUnauthenticated helper below bypasses
// the normal form-body auth injection.

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// partnerName is the constant partner identifier sent on every Authenticate call.
	partnerName = "expensify.com"

	// partnerPasswordPublic is the well-known public partner secret baked into
	// every Expensify client. It is NOT a credential — it identifies the partner
	// app, not the user. Distributed openly in the Expensify mobile/web source.
	partnerPasswordPublic = "e21965746fd75f82bb66"
)

// AuthenticateResult carries the fields extracted from a successful /Authenticate
// response. ExpiresAt is zero-value when the server response doesn't include a
// parseable expiry — callers fall back to the age-since-mint heuristic.
type AuthenticateResult struct {
	AuthToken string
	Email     string
	AccountID int64
	ExpiresAt time.Time
}

// ErrTwoFactorRequired is returned when the /Authenticate response indicates the
// account requires a 2FA code. The CLI layer translates this into an actionable
// "use headed auth login" message; a future plan can add TOTP input.
var ErrTwoFactorRequired = errors.New("two-factor authentication required")

// ErrInvalidCredentials is returned when Expensify rejects the email/password
// pair with jsonCode 402 (non-2FA) or 403.
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrAuthenticateMissing is returned when the server returns jsonCode 200 but
// the response body lacks an authToken field. Defensive: keeps downstream code
// from silently persisting an empty token.
var ErrAuthenticateMissing = errors.New("authenticate response missing token")

// authenticateResponse mirrors the /Authenticate JSON body. Fields not present
// in the response decode to their zero value and are treated as "absent" below.
type authenticateResponse struct {
	JSONCode              int    `json:"jsonCode"`
	Message               string `json:"message"`
	AuthToken             string `json:"authToken"`
	Email                 string `json:"email"`
	AccountID             int64  `json:"accountID"`
	RequiresTwoFactorAuth bool   `json:"requiresTwoFactorAuth"`
	// Expires can come back as an ISO 8601 string OR as a unix-seconds number.
	// Decoding as json.RawMessage lets us try both without fighting the decoder.
	Expires json.RawMessage `json:"expires"`
}

// Authenticate POSTs the given email/password to Expensify's internal
// /Authenticate endpoint and returns the minted session token. It does NOT
// rely on the Client's existing session authToken — this is the chicken-and-egg
// fix for a CLI that wants to mint a fresh token after the old one expired.
//
// Typed errors (ErrTwoFactorRequired, ErrInvalidCredentials, ErrAuthenticateMissing)
// let callers branch without parsing server messages.
func (c *Client) Authenticate(email, password string) (*AuthenticateResult, error) {
	form := url.Values{}
	form.Set("partnerName", partnerName)
	form.Set("partnerPassword", partnerPasswordPublic)
	form.Set("partnerUserID", email)
	form.Set("partnerUserSecret", password)
	form.Set("doNotRetry", "true")
	form.Set("api_setCookie", "false")
	form.Set("platform", "web")
	form.Set("referer", "ecash")

	targetURL := c.authenticateURL()

	req, err := http.NewRequest("POST", targetURL, bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("authenticate: building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "expensify-pp-cli/0.1.0")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("authenticate: reading response: %w", err)
	}
	raw = sanitizeJSONResponse(raw)

	// Transport-level failure (bad HTTP status) surfaces as an APIError-shaped
	// message. Don't attempt to parse typed auth errors out of an HTML error
	// page or similar.
	if resp.StatusCode >= 400 && len(bytes.TrimSpace(raw)) == 0 {
		return nil, fmt.Errorf("authenticate: HTTP %d", resp.StatusCode)
	}

	var body authenticateResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("authenticate: parsing response (HTTP %d): %w", resp.StatusCode, err)
	}

	switch body.JSONCode {
	case 200:
		if body.AuthToken == "" {
			return nil, ErrAuthenticateMissing
		}
		return &AuthenticateResult{
			AuthToken: body.AuthToken,
			Email:     body.Email,
			AccountID: body.AccountID,
			ExpiresAt: parseExpires(body.Expires),
		}, nil
	case 402:
		// 402 covers both "requires 2FA" AND "invalid credentials" depending on
		// message. The requiresTwoFactorAuth flag is the authoritative signal
		// when the server sets it; fall back to string sniffing on message.
		if body.RequiresTwoFactorAuth || looksLikeTwoFactor(body.Message) {
			return nil, ErrTwoFactorRequired
		}
		return nil, ErrInvalidCredentials
	case 403:
		return nil, ErrInvalidCredentials
	default:
		if body.JSONCode == 0 {
			// Shouldn't happen on a successful HTTP parse; guard anyway.
			return nil, fmt.Errorf("authenticate: HTTP %d: %s", resp.StatusCode, truncateBody(raw))
		}
		if body.Message == "" {
			return nil, fmt.Errorf("authenticate: jsonCode %d", body.JSONCode)
		}
		return nil, fmt.Errorf("authenticate: jsonCode %d: %s", body.JSONCode, body.Message)
	}
}

// authenticateURL builds the target URL. Honors the BaseURL override when it's
// already pointing at /api (so EXPENSIFY_BASE_URL=<httptest server> works in
// unit tests without special casing).
func (c *Client) authenticateURL() string {
	base := "https://www.expensify.com/api"
	if c.Config != nil && c.Config.BaseURL != "" && !strings.Contains(c.Config.BaseURL, "www.expensify.com/api") {
		base = strings.TrimRight(c.Config.BaseURL, "/")
	}
	return base + "/Authenticate"
}

// looksLikeTwoFactor checks for common 2FA indicators in the response message.
// Expensify has historically used phrases like "twoFactorAuthRequired" and
// "Two Factor Authentication"; a generous case-insensitive match covers both.
func looksLikeTwoFactor(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "twofactor") ||
		strings.Contains(lower, "two factor") ||
		strings.Contains(lower, "two-factor") ||
		strings.Contains(lower, "2fa")
}

// parseExpires tries unix-seconds-as-number first, then ISO-8601-as-string.
// Returns zero-value time.Time when neither parse succeeds — callers should
// treat that as "unknown expiry" rather than "already expired".
func parseExpires(raw json.RawMessage) time.Time {
	if len(bytes.TrimSpace(raw)) == 0 {
		return time.Time{}
	}
	// Number form
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil && n > 0 {
		return time.Unix(n, 0).UTC()
	}
	// String form — try ISO 8601 first, then numeric-as-string as a fallback.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.UTC()
		}
		if t, err := time.Parse("2006-01-02T15:04:05Z07:00", s); err == nil {
			return t.UTC()
		}
		if sec, err := strconv.ParseInt(s, 10, 64); err == nil && sec > 0 {
			return time.Unix(sec, 0).UTC()
		}
	}
	return time.Time{}
}
