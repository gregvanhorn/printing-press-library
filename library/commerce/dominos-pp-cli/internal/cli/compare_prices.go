package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newComparePricesCmd(flags *rootFlags) *cobra.Command {
	var address, rawItems string
	cmd := &cobra.Command{
		Use:   "compare-prices --address <addr> --items <code1,code2,...>",
		Short: "Compare item pricing across nearby stores",
		RunE: func(cmd *cobra.Command, args []string) error {
			items := parseCodeList(rawItems)
			if len(items) == 0 {
				return usageErr(fmt.Errorf("--items is required"))
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			s, city := splitAddress(address)
			data, err := c.Get("/power/store-locator", map[string]string{"s": s, "c": city, "type": "Delivery"})
			if err != nil {
				return classifyAPIError(err)
			}
			rows := comparePriceRows(c, data, items)
			if len(rows) == 0 {
				return notFoundErr(fmt.Errorf("no nearby stores found for %q", address))
			}
			out := map[string]any{"address": address, "items": items, "stores": rows}
			if flags.asJSON {
				return flags.printJSON(cmd, out)
			}
			return printAutoTable(cmd.OutOrStdout(), rows)
		},
	}
	cmd.Flags().StringVar(&address, "address", "", "Street, city, state, zip")
	cmd.Flags().StringVar(&rawItems, "items", "", "Comma-separated menu item codes")
	_ = cmd.MarkFlagRequired("address")
	_ = cmd.MarkFlagRequired("items")
	return cmd
}

func comparePriceRows(c interface {
	Get(string, map[string]string) (json.RawMessage, error)
}, data json.RawMessage, items []string) []map[string]any {
	seen, out := map[string]bool{}, []map[string]any{}
	for _, m := range findMaps(data) {
		id := firstNonEmpty(pickS(m, "storeId", "StoreID", "id"), pickS(m, "StoreID"))
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		menu, err := c.Get("/power/store/"+id+"/menu", map[string]string{"lang": "en", "structured": "true"})
		if err != nil {
			continue
		}
		profile, _ := c.Get("/power/store/"+id+"/profile", nil)
		prices, total, missing := menuPriceMap(menu, items)
		row := map[string]any{
			"store_id":      id,
			"store_name":    firstNonEmpty(pickS(m, "storeName", "StoreName", "name"), id),
			"distance":      firstNonEmpty(pickS(m, "distance", "Distance", "miles"), "?"),
			"item_prices":   prices,
			"total":         total,
			"fastest_eta":   firstNonEmpty(profileETA(profile), pickS(m, "serviceIsOpen", "IsOpen"), "?"),
			"missing_items": missing,
		}
		if len(missing) > 0 {
			row["total"] = "n/a"
		} else {
			row["total"] = money(total)
		}
		out = append(out, row)
	}
	return out
}

func menuPriceMap(data json.RawMessage, items []string) (string, float64, []string) {
	lookup, total, missing, parts := map[string]float64{}, 0.0, []string{}, []string{}
	for _, m := range findMaps(data) {
		if code := strings.ToUpper(pickS(m, "code", "Code")); code != "" && pickS(m, "name", "Name") != "" {
			lookup[code] = num(m, "price", "Price")
		}
	}
	for _, code := range items {
		price, ok := lookup[strings.ToUpper(code)]
		if !ok {
			missing = append(missing, code)
			parts = append(parts, code+":-")
			continue
		}
		total += price
		parts = append(parts, fmt.Sprintf("%s:%s", code, money(price)))
	}
	return strings.Join(parts, ", "), total, missing
}

func profileETA(data json.RawMessage) string {
	for _, m := range findMaps(data) {
		for _, key := range []string{"fastestEta", "eta", "waitTime", "waitMinutes", "deliveryWaitMinutes"} {
			if v := pickS(m, key); v != "" {
				return v
			}
		}
	}
	return ""
}

func splitAddress(v string) (string, string) {
	parts := strings.SplitN(v, ",", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(v), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func parseCodeList(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
