package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newMenuSearchCmd(flags *rootFlags) *cobra.Command {
	var storeID, category string
	var limit int
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Full-text search across cached menu items",
		Example: "  dominos-pp-cli menu search pepperoni --store 7094 --limit 10",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := flags.openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if storeID != "" {
				if err := syncMenuStore(flags, s, storeID); err != nil {
					return err
				}
			}
			rows, err := s.SearchMenuItems(ftsQuery(args[0]), max(limit*5, 25))
			if err != nil {
				return err
			}
			items := filterMenu(rows, storeID, category, limit)
			if len(items) == 0 && storeID == "" {
				return notFoundErr(fmt.Errorf("no cached menu matches; rerun with --store to sync a menu first"))
			}
			if flags.asJSON {
				return flags.printJSON(cmd, items)
			}
			tab := make([][]string, 0, len(items))
			for _, it := range items {
				tab = append(tab, []string{fmt.Sprint(it["name"]), fmt.Sprint(it["code"]), fmt.Sprintf("%v", it["price"]), fmt.Sprintf("%v", it["calories"]), fmt.Sprintf("%v", it["category"])})
			}
			return flags.printTable(cmd, []string{"NAME", "CODE", "PRICE", "CALORIES", "CATEGORY"}, tab)
		},
	}
	cmd.Flags().StringVar(&storeID, "store", "", "Store ID to sync and filter")
	cmd.Flags().StringVar(&category, "category", "", "Category filter")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum results")
	return cmd
}

func syncMenuStore(flags *rootFlags, s interface {
	UpsertMenuItem(string, string, string, string, string, string, float64, int, bool, string) error
}, storeID string) error {
	c, err := flags.newClient()
	if err != nil {
		return err
	}
	data, err := c.Get("/power/store/"+storeID+"/menu", map[string]string{"lang": "en", "structured": "true"})
	if err != nil {
		return classifyAPIError(err)
	}
	var v any
	if json.Unmarshal(data, &v) != nil {
		return nil
	}
	seen := map[string]bool{}
	var walk func(any, string)
	walk = func(x any, cat string) {
		switch t := x.(type) {
		case map[string]any:
			nextCat := firstNonEmpty(pickS(t, "category", "Category", "categoryName"), cat)
			code, name := pickS(t, "code", "Code"), pickS(t, "name", "Name")
			if code != "" && name != "" && !seen[code] {
				seen[code] = true
				raw, _ := json.Marshal(t)
				_ = s.UpsertMenuItem(firstNonEmpty(pickS(t, "id", "ID"), storeID+":"+code), storeID, name, code, nextCat, pickS(t, "description", "Description"), num(t, "price", "Price"), int(num(t, "calories", "Calories")), !hasFalse(t, "available", "isAvailable"), string(raw))
			}
			for _, v := range t {
				walk(v, nextCat)
			}
		case []any:
			for _, v := range t {
				walk(v, cat)
			}
		}
	}
	walk(v, "")
	return nil
}

func filterMenu(rows []json.RawMessage, storeID, category string, limit int) []map[string]any {
	out := []map[string]any{}
	for _, row := range rows {
		var it map[string]any
		if json.Unmarshal(row, &it) != nil {
			continue
		}
		if storeID != "" && fmt.Sprint(it["store_id"]) != storeID {
			continue
		}
		if category != "" && !strings.EqualFold(fmt.Sprint(it["category"]), category) {
			continue
		}
		out = append(out, map[string]any{"name": it["name"], "code": it["code"], "price": it["price"], "calories": it["calories"], "category": it["category"]})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func ftsQuery(q string) string {
	parts := strings.Fields(strings.TrimSpace(q))
	if len(parts) == 0 {
		return q
	}
	for i, p := range parts {
		parts[i] = `"` + strings.ReplaceAll(p, `"`, "") + `"` + "*"
	}
	return strings.Join(parts, " OR ")
}

func num(m map[string]any, keys ...string) float64 {
	for _, k := range keys {
		switch v := m[k].(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return 0
}
func hasFalse(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if v, ok := m[k].(bool); ok && !v {
			return true
		}
	}
	return false
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
