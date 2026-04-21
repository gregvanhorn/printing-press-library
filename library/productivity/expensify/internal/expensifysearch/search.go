// Copyright 2026 matt-van-horn. Licensed under Apache-2.0. See LICENSE.

// Package expensifysearch defines the typed filter DSL and response envelope
// for Expensify's /Search command. The jsonQuery payload is a nested tree of
// {operator, left, right} nodes; modeling it as a typed Go struct tree keeps
// call sites readable and catches DSL typos at compile time.
//
// This package intentionally avoids any HTTP concerns — the actual wire call
// lives in the client package. Callers build a Query, hand it to client.Search,
// and walk the parsed Response.OnyxData entries to extract what they want.
package expensifysearch

import (
	"encoding/json"
	"fmt"
)

// Query mirrors the jsonQuery shape Expensify's web UI sends to /Search.
//
// Fields use pointer/zero-aware JSON tags so an empty Query round-trips with
// `"filters":null` rather than `{}` — this matches the web UI's observed
// behavior when no filter tree is attached (e.g., the initial view load).
type Query struct {
	Type                  string  `json:"type"`
	Status                string  `json:"status,omitempty"`
	SortBy                string  `json:"sortBy,omitempty"`
	SortOrder             string  `json:"sortOrder,omitempty"`
	View                  string  `json:"view,omitempty"`
	Filters               *Filter `json:"filters"`
	InputQuery            string  `json:"inputQuery,omitempty"`
	IsViewExplicitlySet   bool    `json:"isViewExplicitlySet,omitempty"`
	Hash                  int     `json:"hash,omitempty"`
	RecentSearchHash      int     `json:"recentSearchHash,omitempty"`
	SimilarSearchHash     int     `json:"similarSearchHash,omitempty"`
	SearchKey             string  `json:"searchKey,omitempty"`
	Offset                int     `json:"offset,omitempty"`
	ShouldCalculateTotals bool    `json:"shouldCalculateTotals,omitempty"`
}

// Filter is a recursive node in the jsonQuery filter tree. Left and Right may
// be strings (field name / value) or nested *Filter nodes.
//
// Example tree for `action:submit AND from:20631946`:
//
//	And(Eq("action","submit"), Eq("from","20631946"))
//
// Marshals to:
//
//	{"operator":"and",
//	 "left":  {"operator":"eq","left":"action","right":"submit"},
//	 "right": {"operator":"eq","left":"from","right":"20631946"}}
type Filter struct {
	Operator string `json:"operator"`
	Left     any    `json:"left"`
	Right    any    `json:"right"`
}

// Eq builds a leaf comparison filter: field == value.
// Value is typed as any so callers can pass strings, numbers, or bools.
func Eq(field string, value any) *Filter {
	return &Filter{Operator: "eq", Left: field, Right: value}
}

// And combines two filters into a logical AND node.
func And(a, b *Filter) *Filter {
	return &Filter{Operator: "and", Left: a, Right: b}
}

// Or combines two filters into a logical OR node.
func Or(a, b *Filter) *Filter {
	return &Filter{Operator: "or", Left: a, Right: b}
}

// Response is the parsed envelope returned by /Search.
//
// JSONCode is Expensify's numeric status ("200" for success; "401" for invalid
// query; "407" for an expired session). Message carries the human-readable
// error when JSONCode != 200. OnyxData is an array of patch entries — the
// caller walks these to extract typed rows (reports, transactions, etc.).
type Response struct {
	JSONCode int         `json:"jsonCode"`
	Message  string      `json:"message,omitempty"`
	OnyxData []OnyxEntry `json:"onyxData"`
}

// OnyxEntry is one patch in the Onyx stream. Value is left as a raw message
// so callers can parse into their own typed structs (or walk as generic maps)
// — mirroring the pattern in internal/cli/sync.go.
type OnyxEntry struct {
	Key        string          `json:"key"`
	OnyxMethod string          `json:"onyxMethod"`
	Value      json.RawMessage `json:"value"`
}

// SearchError is the typed error returned when /Search responds with a
// non-200 jsonCode. The calling command handler can inspect JSONCode to
// decide the right CLI exit code (4 auth, 5 API, etc.) without parsing
// a free-form string.
type SearchError struct {
	JSONCode int
	Message  string
	Query    Query
}

func (e *SearchError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "Search failed"
	}
	if e.Query.InputQuery != "" {
		return fmt.Sprintf("expensify /Search jsonCode %d: %s (inputQuery=%q)", e.JSONCode, msg, e.Query.InputQuery)
	}
	return fmt.Sprintf("expensify /Search jsonCode %d: %s", e.JSONCode, msg)
}
