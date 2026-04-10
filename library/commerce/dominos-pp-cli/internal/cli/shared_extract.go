package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func findMaps(data json.RawMessage) []map[string]any {
	var v any
	if json.Unmarshal(data, &v) != nil {
		return nil
	}
	out := []map[string]any{}
	var walk func(any)
	walk = func(x any) {
		switch t := x.(type) {
		case map[string]any:
			out = append(out, t)
			for _, v := range t {
				walk(v)
			}
		case []any:
			for _, v := range t {
				walk(v)
			}
		}
	}
	walk(v)
	return out
}

func pickS(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && fmt.Sprint(v) != "" && fmt.Sprint(v) != "<nil>" {
			return fmt.Sprint(v)
		}
	}
	return ""
}

func pickB(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if v, ok := m[k].(bool); ok {
			return v
		}
	}
	return false
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func dealCoverage(d dealRecord, items []cartItem) int {
	left := map[string]int{}
	for _, it := range items {
		left[strings.ToUpper(it.Code)] += it.Qty
	}
	return dealCoverageRemaining(d, left)
}

func sumCart(items []cartItem) float64 {
	total := 0.0
	for _, it := range items {
		total += it.EstimatedPrice
	}
	return total
}

func parsePrice(s string) float64 {
	s = strings.TrimSpace(strings.TrimLeft(s, "$"))
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}
