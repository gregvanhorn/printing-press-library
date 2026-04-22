// Package attribution implements Meta-specific domain logic ported from
// the magoosh-founder production CLI:
//   - Purchase-action deduplication (omni_purchase vs purchase vs pixel events)
//   - Frequency-based capacity signal
//   - Completed-day date window matching Ads Manager "Last N days"
//   - Graph API error code classification
package attribution

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PurchaseActionPriority is the dedup order for purchase action types.
// The FIRST present type wins — never sum across types or double-counting results.
// Source: magoosh-founder meta_ads.py:27-33.
var PurchaseActionPriority = []string{
	"omni_purchase",
	"purchase",
	"offsite_conversion.fb_pixel_purchase",
	"offsite_conversion.purchase",
}

// LegacyPurchaseActions are the two types that commonly co-occur and cause
// the double-count bug. The verify command compares dedup vs legacy-combined.
var LegacyPurchaseActions = []string{"purchase", "omni_purchase"}

// DefaultFrequencySaturation is the frequency above which an audience is
// considered saturated (no headroom regardless of spend).
const DefaultFrequencySaturation = 3.5

// DefaultFrequencyWarning is the frequency at which we flag approaching-saturation.
const DefaultFrequencyWarning = 2.5

// BuildActionValueMap sums values per action_type, coercing non-numeric values to 0.
// Defensive against Meta returning duplicates for the same action_type.
func BuildActionValueMap(actions []map[string]any) map[string]float64 {
	out := make(map[string]float64)
	for _, a := range actions {
		t, _ := a["action_type"].(string)
		if t == "" {
			continue
		}
		var v float64
		switch x := a["value"].(type) {
		case float64:
			v = x
		case string:
			if x == "" {
				continue
			}
			if _, err := fmt.Sscanf(x, "%f", &v); err != nil {
				continue
			}
		}
		out[t] += v
	}
	return out
}

// SelectPrimaryPurchaseActionType walks PurchaseActionPriority and returns the
// first action type present in any of the provided maps. Returns "" if none.
func SelectPrimaryPurchaseActionType(maps ...map[string]float64) string {
	for _, t := range PurchaseActionPriority {
		for _, m := range maps {
			if _, ok := m[t]; ok {
				return t
			}
		}
	}
	return ""
}

// ExtractPurchaseMetrics reads a Meta insights row and returns the dedup'd
// (conversions, conversion_value, roas_override) tuple using the priority list.
// actionsField and valuesField are typically "actions" and "action_values".
// roasField is typically "purchase_roas".
func ExtractPurchaseMetrics(insight map[string]any) (conversions, value float64, roasOverride float64, selectedAction string) {
	actionsRaw, _ := insight["actions"].([]any)
	valuesRaw, _ := insight["action_values"].([]any)
	roasRaw, _ := insight["purchase_roas"].([]any)

	actionsMap := BuildActionValueMap(toMaps(actionsRaw))
	valuesMap := BuildActionValueMap(toMaps(valuesRaw))
	roasMap := BuildActionValueMap(toMaps(roasRaw))

	selectedAction = SelectPrimaryPurchaseActionType(actionsMap, valuesMap, roasMap)
	if selectedAction == "" {
		return 0, 0, 0, ""
	}
	return actionsMap[selectedAction], valuesMap[selectedAction], roasMap[selectedAction], selectedAction
}

// LegacyCombinedValue sums values for the LegacyPurchaseActions — used by verify
// to compute the double-count delta vs the deduplicated value.
func LegacyCombinedValue(insight map[string]any) float64 {
	valuesRaw, _ := insight["action_values"].([]any)
	valuesMap := BuildActionValueMap(toMaps(valuesRaw))
	var sum float64
	for _, t := range LegacyPurchaseActions {
		sum += valuesMap[t]
	}
	return sum
}

func toMaps(in []any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, item := range in {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// Capacity is the derived capacity signal for one campaign.
type Capacity struct {
	HasHeadroom bool   `json:"has_headroom"`
	Confidence  string `json:"confidence"` // high | medium | low | unknown
	Details     string `json:"details"`
}

// CapacitySignal implements the frequency-based capacity decision tree from
// magoosh-founder meta_ads.py:320-426. Order matters:
//  1. No frequency data → unknown.
//  2. Delivery status has explicit spending-limited flag → has-headroom high.
//  3. Delivery status has paused flag → no-headroom low.
//  4. Budget utilization < 0.70 → no-headroom low (KEY INSIGHT: not budget-constrained).
//  5. Frequency > saturation (default 3.5) → no-headroom high (saturated).
//  6. Frequency >= warning (default 2.5) → has-headroom medium (approaching saturation).
//  7. Otherwise → has-headroom high (healthy).
func CapacitySignal(frequency float64, haveFreq bool, spendRatio float64, deliveryStatus string, saturation, warning float64) Capacity {
	if saturation == 0 {
		saturation = DefaultFrequencySaturation
	}
	if warning == 0 {
		warning = DefaultFrequencyWarning
	}
	if !haveFreq {
		return Capacity{false, "unknown", "No frequency data available"}
	}
	ds := strings.ToLower(deliveryStatus)
	if containsAny(ds, []string{"spending_limited", "not_delivering", "limited", "learning_limited"}) {
		return Capacity{true, "high", "Delivery limited by budget — explicit spending_limited flag"}
	}
	if containsAny(ds, []string{"paused", "campaign_paused", "adset_paused", "inactive"}) {
		return Capacity{false, "low", "Campaign paused or inactive"}
	}
	if spendRatio < 0.70 {
		return Capacity{false, "low", fmt.Sprintf("Budget underutilized (%.0f%%) — not budget-constrained", spendRatio*100)}
	}
	if frequency > saturation {
		return Capacity{false, "high", fmt.Sprintf("Frequency %.2f > saturation %.2f — audience saturated", frequency, saturation)}
	}
	if frequency >= warning {
		return Capacity{true, "medium", fmt.Sprintf("Frequency %.2f >= warning %.2f — approaching saturation", frequency, warning)}
	}
	return Capacity{true, "high", fmt.Sprintf("Frequency %.2f below warning — healthy headroom", frequency)}
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// CompletedDayWindow returns the Meta-Ads-Manager-compatible "Last N days" window:
// end = yesterday (UTC), start = end - (days-1). Matches magoosh-founder meta_ads.py:82-104.
func CompletedDayWindow(days int, now time.Time) (since, until string, err error) {
	if days < 1 {
		return "", "", fmt.Errorf("days must be >= 1, got %d", days)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	end := now.AddDate(0, 0, -1)
	start := end.AddDate(0, 0, -(days - 1))
	return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
}

// TimeRangeJSON builds the JSON blob Meta's insights endpoint expects.
func TimeRangeJSON(since, until string) string {
	b, _ := json.Marshal(map[string]string{"since": since, "until": until})
	return string(b)
}

// ErrorKind is the classification of a Graph API error.
type ErrorKind int

const (
	ErrUnknown ErrorKind = iota
	ErrAuth
	ErrPermission
	ErrRateLimit
	ErrPlatform
)

// ClassifyErrorCode maps Meta Graph API error codes to ErrorKind.
// Mapping ported from magoosh-founder meta_ads.py:264-317.
func ClassifyErrorCode(code int) ErrorKind {
	switch code {
	case 190, 102, 463, 467:
		return ErrAuth
	case 10, 200, 294:
		return ErrPermission
	case 4, 17, 32, 613:
		return ErrRateLimit
	case 0:
		return ErrUnknown
	default:
		return ErrPlatform
	}
}

// CentsToDollars converts Meta's cent-denominated budgets to float dollars.
func CentsToDollars(cents int64) float64 { return float64(cents) / 100.0 }

// DollarsToCents converts float dollars to Meta's cent-denominated integer budgets.
func DollarsToCents(dollars float64) int64 { return int64(dollars * 100) }
