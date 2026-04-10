package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newCartCheckoutCmd(flags *rootFlags) *cobra.Command {
	var couponCodes []string
	var tipAmount float64

	cmd := &cobra.Command{
		Use:   "checkout",
		Short: "Validate, price, and place the active cart as an order",
		Long: `Converts the active cart into a Domino's order, validates it,
shows a price breakdown, and places it on confirmation.

Requires authentication (run 'auth login' first). Uses card-on-file
payment from your Domino's account.`,
		Example: "  dominos-pp-cli cart checkout\n" +
			"  dominos-pp-cli cart checkout --coupon 1121\n" +
			"  dominos-pp-cli cart checkout --coupon 1121 --tip 5.00\n" +
			"  dominos-pp-cli cart checkout --yes  # skip confirmation (agents)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load cart
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
			if len(items) == 0 {
				return usageErr(fmt.Errorf("cart is empty; add items with 'cart add' first"))
			}

			// Load coupons from cart + CLI flags
			coupons, err := loadCartCoupons(cart)
			if err != nil {
				return err
			}
			for _, code := range couponCodes {
				found := false
				for i := range coupons {
					if coupons[i].Code == code {
						found = true
						break
					}
				}
				if !found {
					coupons = append(coupons, cartCoupon{Code: code, Qty: 1})
				}
			}

			// Build order payload
			orderPayload := buildOrderPayload(cart, items, coupons)

			// Get API client
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			// Step 1: Validate
			fmt.Fprintln(os.Stderr, "Validating order...")
			validateBody := map[string]any{"Order": orderPayload}
			valData, _, valErr := c.Post("/power/validate-order", validateBody)
			if valErr != nil {
				return fmt.Errorf("validation failed: %w", classifyAPIError(valErr))
			}
			valStatus, valItems := parseOrderStatus(valData)
			if valStatus < 0 {
				return apiErr(fmt.Errorf("order validation failed:\n%s", formatStatusItems(valItems)))
			}
			// Check for specific warnings
			for _, item := range valItems {
				code, _ := item["Code"].(string)
				if code == "StoreClosed" {
					return apiErr(fmt.Errorf("store %s is currently closed. Try again during business hours", cart.StoreID))
				}
				if strings.Contains(strings.ToLower(code), "recaptcha") {
					return apiErr(fmt.Errorf("reCAPTCHA required. Complete your order at https://www.dominos.com or call the store"))
				}
			}

			// Step 2: Price
			fmt.Fprintln(os.Stderr, "Pricing order...")
			priceBody := map[string]any{"Order": orderPayload}
			priceData, _, priceErr := c.Post("/power/price-order", priceBody)
			if priceErr != nil {
				return fmt.Errorf("pricing failed: %w", classifyAPIError(priceErr))
			}
			priceStatus, priceItems := parseOrderStatus(priceData)
			if priceStatus < 0 {
				return apiErr(fmt.Errorf("order pricing failed:\n%s", formatStatusItems(priceItems)))
			}

			amounts := parseAmounts(priceData)
			products := parseProducts(priceData)
			estimatedWait := parseEstimatedWait(priceData)

			// Show order summary
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "Order Summary")
			fmt.Fprintln(w, strings.Repeat("-", 40))
			for _, p := range products {
				name, _ := p["Name"].(string)
				price, _ := p["Price"].(float64)
				fmt.Fprintf(w, "  %-30s %s\n", name, money(price))
			}
			if len(coupons) > 0 {
				fmt.Fprintln(w, "")
				for _, cp := range coupons {
					fmt.Fprintf(w, "  Coupon: %s\n", cp.Code)
				}
			}
			fmt.Fprintln(w, strings.Repeat("-", 40))

			menuAmt, _ := amounts["Menu"].(float64)
			discount, _ := amounts["Discount"].(float64)
			deliveryFee, _ := amounts["DeliveryFee"].(float64)
			tax, _ := amounts["Tax"].(float64)
			customerAmt, _ := amounts["Customer"].(float64)

			fmt.Fprintf(w, "  %-30s %s\n", "Subtotal:", money(menuAmt))
			if discount > 0 {
				fmt.Fprintf(w, "  %-30s -%s\n", "Discount:", money(discount))
			}
			if deliveryFee > 0 {
				fmt.Fprintf(w, "  %-30s %s\n", "Delivery Fee:", money(deliveryFee))
			}
			if tipAmount > 0 {
				fmt.Fprintf(w, "  %-30s %s\n", "Tip:", money(tipAmount))
			}
			fmt.Fprintf(w, "  %-30s %s\n", "Tax:", money(tax))
			fmt.Fprintln(w, strings.Repeat("-", 40))
			total := customerAmt + tipAmount
			fmt.Fprintf(w, "  %-30s %s\n", "Total:", money(total))
			if estimatedWait != "" {
				fmt.Fprintf(w, "\n  Estimated wait: %s min\n", estimatedWait)
			}
			fmt.Fprintf(w, "  Delivery to: %s\n", cart.ServiceMethod)
			fmt.Fprintf(w, "  Store: %s\n", cart.StoreID)
			fmt.Fprintln(w, "  Payment: card on file")
			fmt.Fprintln(w, "")

			// JSON output mode: show priced order and exit
			if flags.asJSON {
				envelope := map[string]any{
					"action":          "cart.checkout",
					"stage":           "priced",
					"cart_id":         cart.ID,
					"store_id":        cart.StoreID,
					"amounts":         amounts,
					"products":        products,
					"estimated_wait":  estimatedWait,
					"tip":             tipAmount,
					"total_with_tip":  total,
					"needs_confirm":   !flags.yes,
				}
				if !flags.yes {
					return flags.printJSON(cmd, envelope)
				}
			}

			// Step 3: Confirm
			if !flags.yes && !flags.noInput {
				fmt.Fprint(w, "Place this order? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Fprintln(os.Stderr, "Order cancelled.")
					return nil
				}
			}

			// Step 4: Place order with card-on-file payment
			fmt.Fprintln(os.Stderr, "Placing order...")
			orderPayload["Payments"] = []map[string]any{
				{
					"Type":         "CreditCard",
					"Amount":       customerAmt,
					"TipAmount":    tipAmount,
					"CardType":     "",
					"SecurityCode": "",
					"PostalCode":   "",
				},
			}
			placeBody := map[string]any{"Order": orderPayload}

			if flags.dryRun {
				envelope := map[string]any{
					"action":   "cart.checkout",
					"stage":    "dry-run",
					"cart_id":  cart.ID,
					"order":    orderPayload,
					"dry_run":  true,
				}
				return flags.printJSON(cmd, envelope)
			}

			placeData, _, placeErr := c.Post("/power/place-order", placeBody)
			if placeErr != nil {
				errMsg := placeErr.Error()
				if strings.Contains(strings.ToLower(errMsg), "recaptcha") {
					return apiErr(fmt.Errorf("reCAPTCHA required by Domino's.\nComplete your order at: https://www.dominos.com\nOr call store %s", cart.StoreID))
				}
				return fmt.Errorf("order placement failed: %w", classifyAPIError(placeErr))
			}

			// Parse response for order ID
			var placeResp map[string]any
			if err := json.Unmarshal(placeData, &placeResp); err == nil {
				if order, ok := placeResp["Order"].(map[string]any); ok {
					orderID, _ := order["OrderID"].(string)
					estWait, _ := order["EstimatedWaitMinutes"].(string)
					fmt.Fprintln(w, green("Order placed successfully!"))
					if orderID != "" {
						fmt.Fprintf(w, "  Order ID: %s\n", orderID)
					}
					if estWait != "" {
						fmt.Fprintf(w, "  Estimated wait: %s min\n", estWait)
					}
					fmt.Fprintf(w, "  Track at: https://www.dominos.com/pages/tracker\n")

					if flags.asJSON {
						envelope := map[string]any{
							"action":         "cart.checkout",
							"stage":          "placed",
							"cart_id":        cart.ID,
							"order_id":       orderID,
							"estimated_wait": estWait,
							"success":        true,
						}
						return flags.printJSON(cmd, envelope)
					}
				}
			}

			return nil
		},
	}
	cmd.Flags().StringArrayVar(&couponCodes, "coupon", nil, "Coupon code to apply (can be repeated)")
	cmd.Flags().Float64Var(&tipAmount, "tip", 0, "Tip amount in dollars")

	return cmd
}

// buildOrderPayload converts local cart data into the Domino's order API format.
func buildOrderPayload(cart *cartRecord, items []cartItem, coupons []cartCoupon) map[string]any {
	// Parse address
	var addr map[string]any
	if cart.AddressJSON != "" {
		_ = json.Unmarshal([]byte(cart.AddressJSON), &addr)
	}
	// Extract display address if it's a wrapper
	if display, ok := addr["display"].(string); ok {
		addr = map[string]any{"Street": display}
	}

	// Build products array
	products := make([]map[string]any, 0, len(items))
	for _, item := range items {
		p := map[string]any{
			"Code": item.Code,
			"Qty":  item.Qty,
		}
		if len(item.Toppings) > 0 {
			options := map[string]any{}
			for _, t := range item.Toppings {
				side := t.Side
				if side == "full" {
					side = "1/1"
				} else if side == "left" {
					side = "1/2"
				} else if side == "right" {
					side = "2/2"
				}
				options[t.Code] = map[string]any{
					side: fmt.Sprintf("%.1f", t.Amount),
				}
			}
			p["Options"] = options
		}
		products = append(products, p)
	}

	// Build coupons array
	couponsList := make([]map[string]any, 0, len(coupons))
	for _, cp := range coupons {
		couponsList = append(couponsList, map[string]any{
			"Code": cp.Code,
			"Qty":  cp.Qty,
		})
	}

	order := map[string]any{
		"ServiceMethod": cart.ServiceMethod,
		"StoreID":       cart.StoreID,
		"Products":      products,
	}
	if addr != nil {
		order["Address"] = addr
	}
	if len(couponsList) > 0 {
		order["Coupons"] = couponsList
	}

	return order
}

// parseOrderStatus extracts Status and StatusItems from the order response.
func parseOrderStatus(data []byte) (int, []map[string]any) {
	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return 0, nil
	}
	order, _ := resp["Order"].(map[string]any)
	if order == nil {
		return 0, nil
	}
	status := 0
	if s, ok := order["Status"].(float64); ok {
		status = int(s)
	}
	var items []map[string]any
	if si, ok := order["StatusItems"].([]any); ok {
		for _, item := range si {
			if m, ok := item.(map[string]any); ok {
				items = append(items, m)
			}
		}
	}
	return status, items
}

// parseAmounts extracts the Amounts map from a priced order response.
func parseAmounts(data []byte) map[string]any {
	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}
	order, _ := resp["Order"].(map[string]any)
	if order == nil {
		return nil
	}
	amounts, _ := order["Amounts"].(map[string]any)
	return amounts
}

// parseProducts extracts the Products array from a priced order response.
func parseProducts(data []byte) []map[string]any {
	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}
	order, _ := resp["Order"].(map[string]any)
	if order == nil {
		return nil
	}
	prods, ok := order["Products"].([]any)
	if !ok {
		return nil
	}
	var result []map[string]any
	for _, p := range prods {
		if m, ok := p.(map[string]any); ok {
			result = append(result, m)
		}
	}
	return result
}

// parseEstimatedWait extracts the EstimatedWaitMinutes from the order response.
func parseEstimatedWait(data []byte) string {
	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return ""
	}
	order, _ := resp["Order"].(map[string]any)
	if order == nil {
		return ""
	}
	wait, _ := order["EstimatedWaitMinutes"].(string)
	return wait
}

// formatStatusItems formats StatusItems for display.
func formatStatusItems(items []map[string]any) string {
	if len(items) == 0 {
		return "  (no details)"
	}
	var parts []string
	for _, item := range items {
		code, _ := item["Code"].(string)
		msg, _ := item["Message"].(string)
		if msg != "" {
			parts = append(parts, fmt.Sprintf("  %s: %s", code, msg))
		} else if code != "" {
			parts = append(parts, fmt.Sprintf("  %s", code))
		}
	}
	return strings.Join(parts, "\n")
}
