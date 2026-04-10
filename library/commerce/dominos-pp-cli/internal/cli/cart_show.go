package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCartShowCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the active cart",
		Example: "  dominos-pp-cli cart show\n" +
			"  dominos-pp-cli cart show --json\n" +
			"  dominos-pp-cli cart show --compact",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := openCartStore()
			if err != nil {
				return err
			}
			defer s.Close()
			cart, err := loadActiveCart(s)
			if err != nil {
				return err
			}
			items, err := loadCartItems(cart)
			if err != nil {
				return err
			}
			coupons, err := loadCartCoupons(cart)
			if err != nil {
				return err
			}
			total := 0.0
			for _, item := range items {
				total += item.EstimatedPrice
			}
			out := map[string]any{
				"id":              cart.ID,
				"name":            cart.Name,
				"store_id":        cart.StoreID,
				"service_method":  cart.ServiceMethod,
				"address_json":    json.RawMessage(cart.AddressJSON),
				"items":           items,
				"coupons":         coupons,
				"estimated_total": total,
				"updated_at":      cart.UpdatedAt,
			}
			if flags.asJSON {
				return flags.printJSON(cmd, out)
			}
			if flags.compact {
				return flags.printTable(cmd, []string{"ID", "STORE", "ITEMS", "COUPONS", "EST_TOTAL"}, [][]string{{cart.ID, cart.StoreID, fmt.Sprintf("%d", len(items)), fmt.Sprintf("%d", len(coupons)), money(total)}})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Cart %s\nStore: %s  Service: %s\n", cart.ID, cart.StoreID, cart.ServiceMethod)
			if cart.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Template: %s\n", cart.Name)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Estimated total: %s\n\n", money(total))
			rows := make([][]string, 0, len(items))
			for i, item := range items {
				rows = append(rows, []string{
					fmt.Sprintf("%d", i+1), item.Code, fmt.Sprintf("%d", item.Qty), item.Size, formatToppings(item.Toppings), money(item.EstimatedPrice),
				})
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Cart is empty.")
			} else {
				if err := flags.printTable(cmd, []string{"#", "CODE", "QTY", "SIZE", "TOPPINGS", "EST_PRICE"}, rows); err != nil {
					return err
				}
			}
			if len(coupons) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\nCoupons:\n")
				couponRows := make([][]string, 0, len(coupons))
				for _, c := range coupons {
					couponRows = append(couponRows, []string{c.Code, fmt.Sprintf("%d", c.Qty)})
				}
				return flags.printTable(cmd, []string{"CODE", "QTY"}, couponRows)
			}
			return nil
		},
	}
	return cmd
}

func formatToppings(ts []cartTopping) string {
	if len(ts) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ts))
	for _, t := range ts {
		parts = append(parts, fmt.Sprintf("%s:%s:%.1f", t.Code, t.Side, t.Amount))
	}
	return strings.Join(parts, ",")
}

func money(v float64) string {
	return fmt.Sprintf("$%.2f", v)
}
