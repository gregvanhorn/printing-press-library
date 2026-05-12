// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

// Package openfda is a thin wrapper around the openFDA device endpoints
// that understands the {meta, results} envelope used by the API.
package openfda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/client"
)

type Meta struct {
	Disclaimer  string         `json:"disclaimer,omitempty"`
	Terms       string         `json:"terms,omitempty"`
	License     string         `json:"license,omitempty"`
	LastUpdated string         `json:"last_updated,omitempty"`
	Results     map[string]any `json:"results,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Meta    Meta              `json:"meta"`
	Results []json.RawMessage `json:"results"`
	Error   *APIError         `json:"error,omitempty"`
}

// Query is the input to a single openFDA request. Zero values become empty
// query params and are dropped by the client.
type Query struct {
	Search string
	Sort   string
	Count  string
	Limit  int
	Skip   int
}

// Params turns the Query into the map[string]string the client expects.
// If the OPENFDA_API_KEY env var is set, it is added as api_key.
func (q Query) Params() map[string]string {
	p := map[string]string{}
	if q.Search != "" {
		p["search"] = q.Search
	}
	if q.Sort != "" {
		p["sort"] = q.Sort
	}
	if q.Count != "" {
		p["count"] = q.Count
	}
	if q.Limit > 0 {
		p["limit"] = strconv.Itoa(q.Limit)
	}
	if q.Skip > 0 {
		p["skip"] = strconv.Itoa(q.Skip)
	}
	if key := strings.TrimSpace(os.Getenv("OPENFDA_API_KEY")); key != "" {
		p["api_key"] = key
	}
	return p
}

// Run executes a query against a device endpoint path (e.g. "/device/510k.json")
// and returns the parsed envelope. openFDA error{code,message} bodies are
// surfaced as a Go error.
func Run(ctx context.Context, c *client.Client, path string, q Query) (*Envelope, error) {
	raw, err := c.Get(path, q.Params())
	if err != nil {
		// HTTP 404 with the standard error envelope means "no records found"
		// for a search; surface it as an empty envelope so callers can
		// distinguish empty vs. transport failure.
		if isNotFound(err) {
			return &Envelope{Results: []json.RawMessage{}}, nil
		}
		return nil, err
	}
	var env Envelope
	if uerr := json.Unmarshal(raw, &env); uerr != nil {
		return nil, fmt.Errorf("openfda: parse envelope: %w", uerr)
	}
	if env.Error != nil {
		return nil, fmt.Errorf("openfda error %s: %s", env.Error.Code, env.Error.Message)
	}
	return &env, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "HTTP 404")
}

// One returns the first result, or nil if none.
func One(ctx context.Context, c *client.Client, path string, q Query) (json.RawMessage, error) {
	if q.Limit == 0 {
		q.Limit = 1
	}
	env, err := Run(ctx, c, path, q)
	if err != nil {
		return nil, err
	}
	if len(env.Results) == 0 {
		return nil, nil
	}
	return env.Results[0], nil
}

// AllPages iterates with skip pagination until exhausted or hardCap is reached.
// openFDA caps skip at 25000; the default hardCap matches that.
func AllPages(ctx context.Context, c *client.Client, path string, q Query, hardCap int) ([]json.RawMessage, error) {
	if hardCap <= 0 {
		hardCap = 25000
	}
	if q.Limit == 0 {
		q.Limit = 100
	}
	var all []json.RawMessage
	for {
		env, err := Run(ctx, c, path, q)
		if err != nil {
			return all, err
		}
		if len(env.Results) == 0 {
			break
		}
		all = append(all, env.Results...)
		if len(all) >= hardCap {
			if len(all) > hardCap {
				all = all[:hardCap]
			}
			break
		}
		if len(env.Results) < q.Limit {
			break
		}
		q.Skip += len(env.Results)
		if q.Skip >= 25000 {
			break
		}
	}
	return all, nil
}

// SinceToYYYYMMDD parses "2y", "6m", "30d", "12w" and returns YYYYMMDD relative
// to now. Empty input returns "".
func SinceToYYYYMMDD(since string) (string, error) {
	since = strings.TrimSpace(since)
	if since == "" {
		return "", nil
	}
	if len(since) < 2 {
		return "", fmt.Errorf("invalid --since %q (expected like 2y, 6m, 30d, 12w)", since)
	}
	unit := since[len(since)-1]
	nStr := since[:len(since)-1]
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 0 {
		return "", fmt.Errorf("invalid --since %q: %v", since, err)
	}
	now := time.Now().UTC()
	var t time.Time
	switch unit {
	case 'y', 'Y':
		t = now.AddDate(-n, 0, 0)
	case 'm', 'M':
		t = now.AddDate(0, -n, 0)
	case 'w', 'W':
		t = now.AddDate(0, 0, -7*n)
	case 'd', 'D':
		t = now.AddDate(0, 0, -n)
	default:
		return "", fmt.Errorf("invalid --since unit %q (use y/m/w/d)", string(unit))
	}
	return t.Format("20060102"), nil
}

// EndpointPath maps logical names to API paths.
func EndpointPath(name string) string {
	switch strings.ToLower(name) {
	case "510k":
		return "/device/510k.json"
	case "pma":
		return "/device/pma.json"
	case "recall":
		return "/device/recall.json"
	case "maude", "event":
		return "/device/event.json"
	case "classification", "product-code":
		return "/device/classification.json"
	case "establishment", "registrationlisting":
		return "/device/registrationlisting.json"
	}
	return ""
}

// BuildLuceneAND combines clauses with " AND " separators. Empty clauses are dropped.
// Spaces (not literal '+') because Go's url.Values.Encode percent-encodes '+' to %2B,
// which openFDA's Lucene parser rejects.
func BuildLuceneAND(clauses ...string) string {
	out := make([]string, 0, len(clauses))
	for _, c := range clauses {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	return strings.Join(out, " AND ")
}
