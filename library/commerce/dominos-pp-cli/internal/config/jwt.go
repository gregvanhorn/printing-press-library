package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

// JWTClaims holds the decoded claims from a Domino's auth JWT.
type JWTClaims struct {
	Email      string   `json:"Email"`
	CustomerID string   `json:"CustomerID"`
	Scopes     []string `json:"scope"`
	Exp        int64    `json:"exp"`
}

// IsExpired returns true if the token's exp claim is in the past.
func (c JWTClaims) IsExpired() bool {
	if c.Exp == 0 {
		return false
	}
	return time.Unix(c.Exp, 0).Before(time.Now())
}

// HasScope returns true if the token contains the given scope.
func (c JWTClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// DecodeJWTClaims extracts claims from a JWT token without verifying the signature.
// The token may be prefixed with "Bearer " or "Bearer+".
func DecodeJWTClaims(token string) JWTClaims {
	// Strip Bearer prefix
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "Bearer+")

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JWTClaims{}
	}

	// Decode the payload (second part)
	payload := parts[1]
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	data, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return JWTClaims{}
	}

	var claims JWTClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return JWTClaims{}
	}
	return claims
}
