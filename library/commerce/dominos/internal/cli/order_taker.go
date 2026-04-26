package cli

// ensureOrderTaker injects OrderTaker:"Web" into a parsed order payload if absent.
// Without this field, /power/price-order returns Status -1, PulseCode 21
// ("Missing Order Taker in Order Source Element"). Discovered during sniff capture
// 2026-04-25 against order.dominos.com.
func ensureOrderTaker(parsedOrder any) any {
	m, ok := parsedOrder.(map[string]any)
	if !ok {
		return parsedOrder
	}
	if _, present := m["OrderTaker"]; !present {
		m["OrderTaker"] = "Web"
	}
	return m
}
