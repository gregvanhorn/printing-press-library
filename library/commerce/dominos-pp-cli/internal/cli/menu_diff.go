package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newMenuDiffCmd(flags *rootFlags) *cobra.Command {
	var storeID string
	cmd := &cobra.Command{
		Use:   "diff --store <id>",
		Short: "Compare a live menu against the cached SQLite snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := flags.openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			before, err := snapshotMenuItems(s, storeID)
			if err != nil {
				return err
			}
			if len(before) == 0 {
				return notFoundErr(fmt.Errorf("no cached menu snapshot for store %s; run `dominos-pp-cli menu search --store %s \"pizza\"` first", storeID, storeID))
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			data, err := c.Get("/power/store/"+storeID+"/menu", map[string]string{"lang": "en", "structured": "true"})
			if err != nil {
				return classifyAPIError(err)
			}
			after := liveMenuItems(data)
			out := diffMenuItems(storeID, before, after)
			if flags.asJSON {
				return flags.printJSON(cmd, out)
			}
			printMenuDiff(cmd, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&storeID, "store", "", "Store ID")
	_ = cmd.MarkFlagRequired("store")
	return cmd
}

func snapshotMenuItems(s interface {
	Query(string, ...any) (*sql.Rows, error)
}, storeID string) (map[string]menuItemSnap, error) {
	rows, err := s.Query(`SELECT code, name, price FROM menu_items WHERE store_id = ?`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]menuItemSnap{}
	for rows.Next() {
		var item menuItemSnap
		if rows.Scan(&item.Code, &item.Name, &item.Price) == nil {
			out[item.Code] = item
		}
	}
	return out, rows.Err()
}

type menuItemSnap struct {
	Code, Name string
	Price      float64
}

func liveMenuItems(data json.RawMessage) map[string]menuItemSnap {
	out := map[string]menuItemSnap{}
	for _, m := range findMaps(data) {
		code, name := pickS(m, "code", "Code"), pickS(m, "name", "Name")
		if code != "" && name != "" {
			out[code] = menuItemSnap{Code: code, Name: name, Price: num(m, "price", "Price")}
		}
	}
	return out
}

func diffMenuItems(storeID string, before, after map[string]menuItemSnap) map[string]any {
	newItems, removed, changed := []map[string]any{}, []map[string]any{}, []map[string]any{}
	for code, item := range after {
		old, ok := before[code]
		if !ok {
			newItems = append(newItems, map[string]any{"code": code, "name": item.Name, "price": money(item.Price)})
			continue
		}
		if old.Price != item.Price {
			changed = append(changed, map[string]any{"code": code, "name": item.Name, "old_price": money(old.Price), "new_price": money(item.Price)})
		}
	}
	for code, item := range before {
		if _, ok := after[code]; !ok {
			removed = append(removed, map[string]any{"code": code, "name": item.Name, "price": money(item.Price)})
		}
	}
	sort.Slice(newItems, func(i, j int) bool { return fmt.Sprint(newItems[i]["code"]) < fmt.Sprint(newItems[j]["code"]) })
	sort.Slice(removed, func(i, j int) bool { return fmt.Sprint(removed[i]["code"]) < fmt.Sprint(removed[j]["code"]) })
	sort.Slice(changed, func(i, j int) bool { return fmt.Sprint(changed[i]["code"]) < fmt.Sprint(changed[j]["code"]) })
	return map[string]any{"store_id": storeID, "new_items": newItems, "removed_items": removed, "price_changes": changed}
}

func printMenuDiff(cmd *cobra.Command, out map[string]any) {
	fmt.Fprintf(cmd.OutOrStdout(), "Menu diff for store %v\n", out["store_id"])
	for _, row := range out["new_items"].([]map[string]any) {
		fmt.Fprintf(cmd.OutOrStdout(), "+ %v %v %v\n", row["code"], row["name"], row["price"])
	}
	for _, row := range out["removed_items"].([]map[string]any) {
		fmt.Fprintf(cmd.OutOrStdout(), "- %v %v %v\n", row["code"], row["name"], row["price"])
	}
	for _, row := range out["price_changes"].([]map[string]any) {
		fmt.Fprintf(cmd.OutOrStdout(), "~ %v %v %v -> %v\n", row["code"], row["name"], row["old_price"], row["new_price"])
	}
}
