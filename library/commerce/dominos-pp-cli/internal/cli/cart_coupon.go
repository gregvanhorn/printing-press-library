package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type cartCoupon struct {
	Code string `json:"code"`
	Qty  int    `json:"qty"`
}

func loadCartCoupons(c *cartRecord) ([]cartCoupon, error) {
	if c.CouponsJSON == "" {
		return []cartCoupon{}, nil
	}
	var coupons []cartCoupon
	if err := json.Unmarshal([]byte(c.CouponsJSON), &coupons); err != nil {
		return nil, fmt.Errorf("parsing cart coupons: %w", err)
	}
	return coupons, nil
}

func saveCartCoupons(c *cartRecord, coupons []cartCoupon) error {
	data, err := json.Marshal(coupons)
	if err != nil {
		return fmt.Errorf("marshaling cart coupons: %w", err)
	}
	c.CouponsJSON = string(data)
	return nil
}

func newCartAddCouponCmd(flags *rootFlags) *cobra.Command {
	var qty int
	cmd := &cobra.Command{
		Use:   "add-coupon <coupon-code>",
		Short: "Add a coupon to the active cart",
		Example: "  dominos-pp-cli cart add-coupon 9171\n" +
			"  dominos-pp-cli cart add-coupon 1121 --qty 2",
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
			coupons, err := loadCartCoupons(cart)
			if err != nil {
				return err
			}
			// Check if this coupon is already in the cart; if so, increment qty.
			found := false
			for i := range coupons {
				if coupons[i].Code == args[0] {
					coupons[i].Qty += qty
					found = true
					break
				}
			}
			if !found {
				coupons = append(coupons, cartCoupon{Code: args[0], Qty: qty})
			}
			if err := saveCartCoupons(cart, coupons); err != nil {
				return err
			}
			items, err := loadCartItems(cart)
			if err != nil {
				return err
			}
			if flags.dryRun {
				return printCouponResult(cmd, flags, "cart.add-coupon", cart, coupons)
			}
			if err := saveCart(s, cart, items); err != nil {
				return err
			}
			return printCouponResult(cmd, flags, "cart.add-coupon", cart, coupons)
		},
	}
	cmd.Flags().IntVar(&qty, "qty", 1, "Quantity")
	return cmd
}

func newCartRemoveCouponCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-coupon <coupon-code>",
		Short: "Remove a coupon from the active cart",
		Example: "  dominos-pp-cli cart remove-coupon 9171\n" +
			"  dominos-pp-cli cart remove-coupon 1121",
		Args: cobra.ExactArgs(1),
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
			warnRecentCartFallback(cmd)
			coupons, err := loadCartCoupons(cart)
			if err != nil {
				return err
			}
			found := false
			for i, c := range coupons {
				if c.Code == args[0] {
					coupons = append(coupons[:i], coupons[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				return notFoundErr(fmt.Errorf("coupon %q not found in cart", args[0]))
			}
			if err := saveCartCoupons(cart, coupons); err != nil {
				return err
			}
			items, err := loadCartItems(cart)
			if err != nil {
				return err
			}
			if flags.dryRun {
				return printCouponResult(cmd, flags, "cart.remove-coupon", cart, coupons)
			}
			if err := saveCart(s, cart, items); err != nil {
				return err
			}
			return printCouponResult(cmd, flags, "cart.remove-coupon", cart, coupons)
		},
	}
	return cmd
}

func printCouponResult(cmd *cobra.Command, flags *rootFlags, action string, cart *cartRecord, coupons []cartCoupon) error {
	out := map[string]any{
		"action":  action,
		"cart_id": cart.ID,
		"coupons": coupons,
	}
	if flags.dryRun {
		out["dry_run"] = true
	}
	if flags.asJSON {
		return flags.printJSON(cmd, out)
	}
	status := "updated"
	if flags.dryRun {
		status = "dry-run"
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s cart %s (%d coupon(s))\n", status, cart.ID, len(coupons))
	for _, c := range coupons {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  coupon %s x%d\n", c.Code, c.Qty)
	}
	return nil
}
