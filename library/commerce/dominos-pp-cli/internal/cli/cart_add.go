package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newCartAddCmd(flags *rootFlags) *cobra.Command {
	var qty int
	var size string
	var toppingFlags []string
	cmd := &cobra.Command{
		Use:   "add <product-code> [--qty N] [--size small|medium|large] [--topping code:side:amount]",
		Short: "Add a product to the active cart",
		Example: "  dominos-pp-cli cart add S_PIZPH --qty 2 --size medium --topping P:full:1.5\n" +
			"  dominos-pp-cli cart add 14SCREEN --topping X:left:1.0 --topping M:right:0.5",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if qty < 1 {
				return usageErr(fmt.Errorf("qty must be at least 1"))
			}
			s, err := openCartStore()
			if err != nil {
				return err
			}
			defer s.Close()
			cart, err := loadActiveCart(s)
			if err != nil {
				return err
			}
			warnRecentCartFallback(cmd)
			items, err := loadCartItems(cart)
			if err != nil {
				return err
			}
			item := cartItem{Code: args[0], Qty: qty, Size: size, EstimatedPrice: resolveEstimatedPrice(s, args[0], qty)}
			for _, raw := range toppingFlags {
				t, err := parseTopping(raw)
				if err != nil {
					return err
				}
				item.Toppings = append(item.Toppings, t)
				item.EstimatedPrice += float64(qty) * t.Amount
			}
			items = append(items, item)
			if flags.dryRun {
				return printMutationResult(cmd, flags, "cart.add", cart, items)
			}
			if err := saveCart(s, cart, items); err != nil {
				return err
			}
			return printMutationResult(cmd, flags, "cart.add", cart, items)
		},
	}
	cmd.Flags().IntVar(&qty, "qty", 1, "Quantity")
	cmd.Flags().StringVar(&size, "size", "", "Optional size")
	cmd.Flags().StringArrayVar(&toppingFlags, "topping", nil, "Topping as code:full|left|right:amount")
	return cmd
}

func parseTopping(raw string) (cartTopping, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return cartTopping{}, usageErr(fmt.Errorf("invalid topping %q; want code:full|left|right:amount", raw))
	}
	if parts[1] != "full" && parts[1] != "left" && parts[1] != "right" {
		return cartTopping{}, usageErr(fmt.Errorf("invalid topping side %q", parts[1]))
	}
	amount, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return cartTopping{}, usageErr(fmt.Errorf("invalid topping amount %q", parts[2]))
	}
	return cartTopping{Code: parts[0], Side: parts[1], Amount: amount}, nil
}

func resolveEstimatedPrice(s interface {
	SearchMenuItems(string, int) ([]json.RawMessage, error)
}, code string, qty int) float64 {
	rows, err := s.SearchMenuItems(code, 10)
	if err != nil {
		return 0
	}
	for _, row := range rows {
		var item struct {
			Code  string  `json:"code"`
			Price float64 `json:"price"`
		}
		if json.Unmarshal(row, &item) == nil && strings.EqualFold(item.Code, code) {
			return item.Price * float64(qty)
		}
	}
	return 0
}
