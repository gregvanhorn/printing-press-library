package attribution

import (
	"testing"
	"time"
)

func TestSelectPrimaryPurchaseActionType(t *testing.T) {
	tests := []struct {
		name    string
		actions map[string]float64
		values  map[string]float64
		want    string
	}{
		{"omni_purchase wins over purchase", map[string]float64{"purchase": 5, "omni_purchase": 5}, nil, "omni_purchase"},
		{"purchase wins over pixel", map[string]float64{"purchase": 3, "offsite_conversion.fb_pixel_purchase": 3}, nil, "purchase"},
		{"pixel when only pixel", map[string]float64{"offsite_conversion.fb_pixel_purchase": 7}, nil, "offsite_conversion.fb_pixel_purchase"},
		{"empty when none", map[string]float64{}, map[string]float64{}, ""},
		{"picks from values map even if actions empty", map[string]float64{}, map[string]float64{"omni_purchase": 10}, "omni_purchase"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SelectPrimaryPurchaseActionType(tc.actions, tc.values)
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestBuildActionValueMap(t *testing.T) {
	in := []map[string]any{
		{"action_type": "omni_purchase", "value": "15.5"},
		{"action_type": "omni_purchase", "value": 4.5},
		{"action_type": "purchase", "value": "10"},
		{"action_type": "link_click", "value": "not-a-number"},
		{"action_type": "", "value": "1"},
	}
	got := BuildActionValueMap(in)
	if got["omni_purchase"] != 20.0 {
		t.Errorf("omni_purchase sum: got %v want 20.0", got["omni_purchase"])
	}
	if got["purchase"] != 10.0 {
		t.Errorf("purchase: got %v want 10.0", got["purchase"])
	}
	if _, ok := got[""]; ok {
		t.Errorf("empty action_type must be dropped")
	}
}

func TestExtractPurchaseMetrics(t *testing.T) {
	insight := map[string]any{
		"actions":       []any{map[string]any{"action_type": "purchase", "value": "3"}, map[string]any{"action_type": "omni_purchase", "value": "4"}},
		"action_values": []any{map[string]any{"action_type": "purchase", "value": "30"}, map[string]any{"action_type": "omni_purchase", "value": "45"}},
		"purchase_roas": []any{map[string]any{"action_type": "omni_purchase", "value": "3.0"}},
	}
	c, v, r, sel := ExtractPurchaseMetrics(insight)
	if sel != "omni_purchase" {
		t.Errorf("selected: got %q want omni_purchase", sel)
	}
	if c != 4 {
		t.Errorf("conversions: got %v want 4", c)
	}
	if v != 45 {
		t.Errorf("value: got %v want 45", v)
	}
	if r != 3.0 {
		t.Errorf("roas: got %v want 3.0", r)
	}
}

func TestLegacyCombinedValue(t *testing.T) {
	insight := map[string]any{
		"action_values": []any{
			map[string]any{"action_type": "purchase", "value": "30"},
			map[string]any{"action_type": "omni_purchase", "value": "45"},
			map[string]any{"action_type": "offsite_conversion.fb_pixel_purchase", "value": "99"},
		},
	}
	got := LegacyCombinedValue(insight)
	if got != 75.0 {
		t.Errorf("legacy combined (purchase+omni): got %v want 75", got)
	}
}

func TestCapacitySignal(t *testing.T) {
	tests := []struct {
		name     string
		freq     float64
		haveFreq bool
		ratio    float64
		delivery string
		wantHead bool
		wantConf string
	}{
		{"no freq data", 0, false, 1.0, "", false, "unknown"},
		{"spending_limited flag wins", 1.0, true, 0.8, "spending_limited", true, "high"},
		{"paused", 1.0, true, 0.9, "paused", false, "low"},
		{"under-utilized overrides", 2.0, true, 0.50, "active", false, "low"},
		{"saturated freq", 4.2, true, 0.95, "active", false, "high"},
		{"warning freq", 2.8, true, 0.95, "active", true, "medium"},
		{"healthy", 1.5, true, 0.85, "active", true, "high"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CapacitySignal(tc.freq, tc.haveFreq, tc.ratio, tc.delivery, 0, 0)
			if got.HasHeadroom != tc.wantHead || got.Confidence != tc.wantConf {
				t.Errorf("got headroom=%v conf=%q want headroom=%v conf=%q (details: %s)",
					got.HasHeadroom, got.Confidence, tc.wantHead, tc.wantConf, got.Details)
			}
		})
	}
}

func TestCompletedDayWindow(t *testing.T) {
	now := time.Date(2026, 2, 5, 10, 0, 0, 0, time.UTC)
	since, until, err := CompletedDayWindow(14, now)
	if err != nil {
		t.Fatal(err)
	}
	// end = yesterday = 2026-02-04; start = end - 13 = 2026-01-22
	if since != "2026-01-22" || until != "2026-02-04" {
		t.Errorf("got since=%s until=%s want since=2026-01-22 until=2026-02-04", since, until)
	}
	if _, _, err := CompletedDayWindow(0, now); err == nil {
		t.Error("expected error for days=0")
	}
}

func TestClassifyErrorCode(t *testing.T) {
	tests := []struct {
		code int
		want ErrorKind
	}{
		{190, ErrAuth}, {102, ErrAuth}, {463, ErrAuth}, {467, ErrAuth},
		{10, ErrPermission}, {200, ErrPermission}, {294, ErrPermission},
		{4, ErrRateLimit}, {17, ErrRateLimit}, {32, ErrRateLimit}, {613, ErrRateLimit},
		{0, ErrUnknown},
		{500, ErrPlatform}, {999, ErrPlatform},
	}
	for _, tc := range tests {
		got := ClassifyErrorCode(tc.code)
		if got != tc.want {
			t.Errorf("code %d: got %v want %v", tc.code, got, tc.want)
		}
	}
}

func TestCentsDollars(t *testing.T) {
	if CentsToDollars(5000) != 50.0 {
		t.Errorf("cents→dollars: got %v want 50.0", CentsToDollars(5000))
	}
	if DollarsToCents(50.0) != 5000 {
		t.Errorf("dollars→cents: got %v want 5000", DollarsToCents(50.0))
	}
}
