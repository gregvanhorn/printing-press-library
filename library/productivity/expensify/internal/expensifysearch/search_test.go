// Copyright 2026 matt-van-horn. Licensed under Apache-2.0. See LICENSE.

package expensifysearch

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestQueryMarshal — a Query with a single Eq filter serializes to the
// expected nested jsonQuery shape. The DSL is the load-bearing surface;
// if the marshaling output drifts, every /Search call drifts with it.
func TestQueryMarshal(t *testing.T) {
	q := Query{Type: "expense-report", Filters: Eq("action", "submit")}
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Round-trip into a generic map so we can assert on the nested filter
	// shape without depending on Go struct field ordering.
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["type"] != "expense-report" {
		t.Errorf("type = %v, want %q", got["type"], "expense-report")
	}
	f, ok := got["filters"].(map[string]any)
	if !ok {
		t.Fatalf("filters not an object: %T", got["filters"])
	}
	if f["operator"] != "eq" {
		t.Errorf("filters.operator = %v, want eq", f["operator"])
	}
	if f["left"] != "action" {
		t.Errorf("filters.left = %v, want action", f["left"])
	}
	if f["right"] != "submit" {
		t.Errorf("filters.right = %v, want submit", f["right"])
	}
}

// TestAndFilter — And(Eq,Eq) marshals to the full nested
// {operator:and, left:{operator:eq,…}, right:{operator:eq,…}} tree.
func TestAndFilter(t *testing.T) {
	f := And(Eq("action", "submit"), Eq("from", "20631946"))
	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["operator"] != "and" {
		t.Errorf("operator = %v, want and", got["operator"])
	}
	left, ok := got["left"].(map[string]any)
	if !ok {
		t.Fatalf("left not a filter object: %T", got["left"])
	}
	if left["operator"] != "eq" || left["left"] != "action" || left["right"] != "submit" {
		t.Errorf("left branch = %v, want eq/action/submit", left)
	}
	right, ok := got["right"].(map[string]any)
	if !ok {
		t.Fatalf("right not a filter object: %T", got["right"])
	}
	if right["operator"] != "eq" || right["left"] != "from" || right["right"] != "20631946" {
		t.Errorf("right branch = %v, want eq/from/20631946", right)
	}
}

// TestOrFilter — mirror of TestAndFilter for the or-operator constructor.
func TestOrFilter(t *testing.T) {
	f := Or(Eq("status", "open"), Eq("status", "submitted"))
	b, _ := json.Marshal(f)
	if !strings.Contains(string(b), `"operator":"or"`) {
		t.Errorf("Or should marshal operator=or, got %s", string(b))
	}
}

// TestNilFilters — a Query{} with no filter tree must serialize as
// "filters":null (NOT omitted, NOT {}). The web UI sends explicit null for
// the initial view; matching that shape avoids a server-side shape mismatch.
func TestNilFilters(t *testing.T) {
	q := Query{Type: "expense-report"}
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"filters":null`) {
		t.Errorf("expected \"filters\":null in %s", s)
	}
	// Round-trip check: unmarshaling back to a Query preserves the nil.
	var rt Query
	if err := json.Unmarshal(b, &rt); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if rt.Filters != nil {
		t.Errorf("round-trip Filters = %+v, want nil", rt.Filters)
	}
}

// TestSearchError — the typed error surfaces both the jsonCode and message
// so command handlers can display and classify it without parsing strings.
func TestSearchError(t *testing.T) {
	e := &SearchError{JSONCode: 401, Message: "Invalid Query"}
	s := e.Error()
	if !strings.Contains(s, "401") {
		t.Errorf("Error() = %q, want it to contain 401", s)
	}
	if !strings.Contains(s, "Invalid Query") {
		t.Errorf("Error() = %q, want it to contain 'Invalid Query'", s)
	}
}

// TestSearchErrorWithQuery — when the Query carries an inputQuery, the error
// string includes it so logs / stderr show what the caller asked for.
func TestSearchErrorWithQuery(t *testing.T) {
	e := &SearchError{
		JSONCode: 401,
		Message:  "Invalid Query",
		Query:    Query{InputQuery: "type:expense-report action:submit"},
	}
	s := e.Error()
	if !strings.Contains(s, "type:expense-report") {
		t.Errorf("Error() = %q, want it to contain the inputQuery", s)
	}
}

// TestResponseParse — a canned /Search response body parses into a Response
// with one OnyxEntry whose Key, OnyxMethod, and raw Value are accessible.
// Mirrors the shape captured during the live dogfood session.
func TestResponseParse(t *testing.T) {
	body := []byte(`{
		"jsonCode": 200,
		"onyxData": [
			{
				"key": "snapshot_abc",
				"onyxMethod": "merge",
				"value": {"data": {"report_1": {"reportID": "1", "ownerAccountID": 20631946}}}
			}
		]
	}`)

	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.JSONCode != 200 {
		t.Errorf("JSONCode = %d, want 200", resp.JSONCode)
	}
	if len(resp.OnyxData) != 1 {
		t.Fatalf("OnyxData len = %d, want 1", len(resp.OnyxData))
	}
	entry := resp.OnyxData[0]
	if entry.Key != "snapshot_abc" {
		t.Errorf("OnyxData[0].Key = %q, want snapshot_abc", entry.Key)
	}
	if entry.OnyxMethod != "merge" {
		t.Errorf("OnyxData[0].OnyxMethod = %q, want merge", entry.OnyxMethod)
	}
	// Value is json.RawMessage; the caller walks it as needed. Confirm it
	// parses as a JSON object with the expected nested key.
	var v map[string]any
	if err := json.Unmarshal(entry.Value, &v); err != nil {
		t.Fatalf("parse value: %v", err)
	}
	data, ok := v["data"].(map[string]any)
	if !ok {
		t.Fatalf("value.data not an object: %T", v["data"])
	}
	if _, ok := data["report_1"]; !ok {
		t.Errorf("value.data missing report_1 key; got %v", data)
	}
}

// TestQueryRoundTrip — a fully-populated Query survives a marshal+unmarshal
// round-trip with field values preserved. Guards against typo in JSON tags.
func TestQueryRoundTrip(t *testing.T) {
	q := Query{
		Type:                  "expense-report",
		Status:                "all",
		SortBy:                "date",
		SortOrder:             "desc",
		View:                  "default",
		Filters:               And(Eq("action", "submit"), Eq("from", "20631946")),
		InputQuery:            "type:expense-report action:submit",
		IsViewExplicitlySet:   true,
		Hash:                  12345,
		RecentSearchHash:      67890,
		SimilarSearchHash:     11111,
		SearchKey:             "default",
		Offset:                0,
		ShouldCalculateTotals: true,
	}
	b, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var rt Query
	if err := json.Unmarshal(b, &rt); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rt.Type != q.Type || rt.Status != q.Status || rt.SortBy != q.SortBy {
		t.Errorf("string fields drifted: %+v", rt)
	}
	if rt.Hash != q.Hash || rt.RecentSearchHash != q.RecentSearchHash {
		t.Errorf("hash fields drifted: %+v", rt)
	}
	if rt.Filters == nil || rt.Filters.Operator != "and" {
		t.Errorf("filter tree drifted: %+v", rt.Filters)
	}
}
