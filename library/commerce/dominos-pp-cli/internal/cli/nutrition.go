package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newNutritionCmd(flags *rootFlags) *cobra.Command {
	var useCart bool
	var rawItems, storeID string
	cmd := &cobra.Command{
		Use:   "nutrition [--cart] [--items <code1,code2,...>] [--store <id>]",
		Short: "Calculate nutrition totals from cached menu data",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := flags.openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			items, cartStore, err := nutritionItems(s, useCart, rawItems)
			if err != nil {
				return err
			}
			if storeID == "" {
				storeID = cartStore
			}
			if storeID != "" {
				if err := syncMenuStore(flags, s, storeID); err != nil {
					return err
				}
			}
			rows, total := nutritionBreakdown(s, items, storeID)
			if len(rows) == 0 {
				return notFoundErr(fmt.Errorf("no nutrition data found; sync a store menu first"))
			}
			out := map[string]any{"store_id": storeID, "total_calories": total, "items": rows}
			if flags.asJSON {
				return flags.printJSON(cmd, out)
			}
			tab := make([][]string, 0, len(rows))
			for _, row := range rows {
				tab = append(tab, []string{fmt.Sprint(row["code"]), fmt.Sprint(row["name"]), fmt.Sprint(row["qty"]), fmt.Sprint(row["calories_each"]), fmt.Sprint(row["calories_total"])})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Total calories: %d\n", total)
			return flags.printTable(cmd, []string{"CODE", "NAME", "QTY", "CAL/EACH", "CAL/TOTAL"}, tab)
		},
	}
	cmd.Flags().BoolVar(&useCart, "cart", false, "Use the active cart")
	cmd.Flags().StringVar(&rawItems, "items", "", "Comma-separated menu item codes")
	cmd.Flags().StringVar(&storeID, "store", "", "Store ID for menu lookup")
	return cmd
}

func nutritionItems(s interface {
	ListCarts() ([]json.RawMessage, error)
}, useCart bool, rawItems string) (map[string]int, string, error) {
	if useCart {
		cartStore, items, err := cartNutritionItems(s)
		return items, cartStore, err
	}
	items := map[string]int{}
	for _, code := range parseCodeList(rawItems) {
		items[strings.ToUpper(code)]++
	}
	if len(items) == 0 {
		return nil, "", usageErr(fmt.Errorf("set --cart or --items"))
	}
	return items, "", nil
}

func cartNutritionItems(s interface {
	ListCarts() ([]json.RawMessage, error)
}) (string, map[string]int, error) {
	rows, err := s.ListCarts()
	if err != nil || len(rows) == 0 {
		return "", nil, notFoundErr(fmt.Errorf("no active cart found; run `dominos-pp-cli cart new` first"))
	}
	var cart cartRecord
	if err := json.Unmarshal(rows[0], &cart); err != nil {
		return "", nil, err
	}
	items, err := loadCartItems(&cart)
	if err != nil {
		return "", nil, err
	}
	out := map[string]int{}
	for _, it := range items {
		out[strings.ToUpper(it.Code)] += max(it.Qty, 1)
	}
	return cart.StoreID, out, nil
}

func nutritionBreakdown(s interface {
	Query(string, ...any) (*sql.Rows, error)
}, items map[string]int, storeID string) ([]map[string]any, int) {
	out, total := []map[string]any{}, 0
	for code, qty := range items {
		row, ok := lookupNutritionRow(s, code, storeID)
		if !ok {
			out = append(out, map[string]any{"code": code, "name": "(missing)", "qty": qty, "calories_each": 0, "calories_total": 0})
			continue
		}
		itemTotal := row.Calories * qty
		total += itemTotal
		out = append(out, map[string]any{"code": code, "name": row.Name, "qty": qty, "calories_each": row.Calories, "calories_total": itemTotal})
	}
	return out, total
}

func lookupNutritionRow(s interface {
	Query(string, ...any) (*sql.Rows, error)
}, code, storeID string) (struct {
	Name     string
	Calories int
}, bool) {
	rows, err := s.Query(
		`SELECT name, calories FROM menu_items
		 WHERE UPPER(code) = UPPER(?)
		 ORDER BY CASE WHEN store_id = ? THEN 0 ELSE 1 END, synced_at DESC LIMIT 1`,
		code, storeID,
	)
	if err != nil {
		return struct {
			Name     string
			Calories int
		}{}, false
	}
	defer rows.Close()
	var out struct {
		Name     string
		Calories int
	}
	if !rows.Next() || rows.Scan(&out.Name, &out.Calories) != nil {
		return out, false
	}
	return out, true
}
