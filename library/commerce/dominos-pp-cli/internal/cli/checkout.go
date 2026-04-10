package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/commerce/dominos-pp-cli/internal/config"
	"github.com/spf13/cobra"
)

func newCheckoutCmd(flags *rootFlags) *cobra.Command {
	var paymentType string

	cmd := &cobra.Command{
		Use:   "checkout",
		Short: "Check out the active cart — validate, price, and place your order",
		Long: `Converts your local cart into a Domino's order. Automatically uses your
saved profile (name, email, phone) and payment method.

On first use, prompts for any missing profile info and saves it for next time.`,
		Example: `  # Interactive checkout
  dominos-pp-cli checkout

  # Preview order without placing
  dominos-pp-cli checkout --dry-run

  # Agent-friendly: skip prompts, pay cash
  dominos-pp-cli checkout --yes --payment-type cash --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Step 1: Load active cart
			s, err := openCartStore()
			if err != nil {
				return err
			}
			cart, err := loadActiveCart(s)
			if err != nil {
				return usageErr(fmt.Errorf("no active cart found\n\n  Create one first:\n    dominos-pp-cli cart new --store <id> --service delivery --address \"your address\"\n    dominos-pp-cli cart add <product-code>"))
			}
			items, err := loadCartItems(cart)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return usageErr(fmt.Errorf("cart is empty — add items first:\n    dominos-pp-cli cart add <product-code>"))
			}

			// Step 2: Load profile (prompt for missing info if interactive)
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}
			profile := cfg.Profile()

			// Check token validity
			token := cfg.DominosToken
			if token == "" {
				token = cfg.AccessToken
			}
			if token != "" {
				claims := config.DecodeJWTClaims(token)
				if claims.IsExpired() {
					w := cmd.ErrOrStderr()
					fmt.Fprintln(w, yellow("WARN")+" auth token is expired — order may fail")
					fmt.Fprintln(w, "  Re-authenticate: dominos-pp-cli auth set-token <new-token>")
				}
			}

			// Prompt for missing profile fields if interactive
			if !flags.noInput {
				changed := false
				reader := bufio.NewReader(os.Stdin)

				if profile.FirstName == "" {
					profile.FirstName = promptLine(reader, cmd, "First name: ")
					changed = true
				}
				if profile.LastName == "" {
					profile.LastName = promptLine(reader, cmd, "Last name: ")
					changed = true
				}
				if profile.Email == "" {
					profile.Email = promptLine(reader, cmd, "Email: ")
					changed = true
				}
				if profile.Phone == "" {
					profile.Phone = promptLine(reader, cmd, "Phone (10 digits): ")
					changed = true
				}

				if changed {
					_ = cfg.SaveProfile(profile.FirstName, profile.LastName, profile.Phone, profile.Email)
				}
			}

			// Validate we have required fields
			missing := []string{}
			if profile.FirstName == "" {
				missing = append(missing, "first_name")
			}
			if profile.LastName == "" {
				missing = append(missing, "last_name")
			}
			if profile.Email == "" {
				missing = append(missing, "email")
			}
			if profile.Phone == "" {
				missing = append(missing, "phone")
			}
			if len(missing) > 0 {
				return usageErr(fmt.Errorf("missing profile fields: %s\n  Set them in config or run checkout interactively (without --no-input)",
					strings.Join(missing, ", ")))
			}

			// Step 3: Build order object
			order := buildOrderFromCart(cart, items, profile)

			// Step 4: Validate
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			validateBody := map[string]any{"Order": order}
			validateResp, _, err := c.Post("/power/validate-order", validateBody)
			if err != nil {
				return classifyAPIError(fmt.Errorf("order validation failed: %w", err))
			}

			// Check validation status
			var validateResult struct {
				Status int `json:"Status"`
				Order  struct {
					StatusItems []struct {
						Code string `json:"Code"`
					} `json:"StatusItems"`
				} `json:"Order"`
			}
			if err := json.Unmarshal(validateResp, &validateResult); err == nil {
				if validateResult.Status < 0 {
					return apiErr(fmt.Errorf("order validation failed: %s", string(validateResp)))
				}
			}

			// Step 5: Price
			priceBody := map[string]any{"Order": order}
			priceResp, _, err := c.Post("/power/price-order", priceBody)
			if err != nil {
				return classifyAPIError(fmt.Errorf("order pricing failed: %w", err))
			}

			var priceResult struct {
				Order struct {
					Amounts struct {
						Menu        float64 `json:"Menu"`
						Customer    float64 `json:"Customer"`
						Tax         float64 `json:"Tax"`
						DeliveryFee float64 `json:"DeliveryFee"`
						Surcharge   float64 `json:"Surcharge"`
						Payment     float64 `json:"Payment"`
					} `json:"Amounts"`
					EstimatedWaitMinutes string `json:"EstimatedWaitMinutes"`
					Products             []struct {
						Name   string  `json:"Name"`
						Price  float64 `json:"Price"`
						Qty    int     `json:"Qty"`
						Code   string  `json:"Code"`
					} `json:"Products"`
					OrderID string `json:"OrderID"`
				} `json:"Order"`
			}
			if err := json.Unmarshal(priceResp, &priceResult); err != nil {
				return apiErr(fmt.Errorf("could not parse pricing response: %w", err))
			}

			// Step 6: Display summary
			pr := priceResult.Order
			w := cmd.OutOrStdout()

			if flags.asJSON {
				summary := map[string]any{
					"action":    "checkout",
					"store_id":  cart.StoreID,
					"service":   cart.ServiceMethod,
					"address":   cart.AddressJSON,
					"items":     pr.Products,
					"subtotal":  pr.Amounts.Menu,
					"delivery":  pr.Amounts.DeliveryFee,
					"surcharge": pr.Amounts.Surcharge,
					"tax":       pr.Amounts.Tax,
					"total":     pr.Amounts.Customer,
					"wait":      pr.EstimatedWaitMinutes,
					"order_id":  pr.OrderID,
					"dry_run":   flags.dryRun,
				}
				return flags.printJSON(cmd, summary)
			}

			fmt.Fprintln(w, "")
			fmt.Fprintln(w, bold("Order Summary"))
			fmt.Fprintln(w, strings.Repeat("-", 40))

			// Parse address for display
			var addrDisplay struct {
				Display string `json:"display"`
			}
			if json.Unmarshal([]byte(cart.AddressJSON), &addrDisplay) == nil && addrDisplay.Display != "" {
				fmt.Fprintf(w, "  Deliver to: %s\n", addrDisplay.Display)
			}
			fmt.Fprintf(w, "  Store: #%s\n", cart.StoreID)
			if pr.EstimatedWaitMinutes != "" {
				fmt.Fprintf(w, "  Est. wait: %s min\n", pr.EstimatedWaitMinutes)
			}
			fmt.Fprintln(w, "")

			for _, p := range pr.Products {
				fmt.Fprintf(w, "  %dx %-30s $%.2f\n", p.Qty, p.Name, p.Price)
			}

			fmt.Fprintln(w, strings.Repeat("-", 40))
			if pr.Amounts.DeliveryFee > 0 {
				fmt.Fprintf(w, "  %-32s $%.2f\n", "Delivery fee", pr.Amounts.DeliveryFee)
			}
			if pr.Amounts.Surcharge > 0 {
				fmt.Fprintf(w, "  %-32s $%.2f\n", "Surcharge", pr.Amounts.Surcharge)
			}
			fmt.Fprintf(w, "  %-32s $%.2f\n", "Tax", pr.Amounts.Tax)
			fmt.Fprintf(w, "  %-32s $%.2f\n", bold("Total"), pr.Amounts.Customer)
			fmt.Fprintln(w, "")

			// Dry run stops here
			if flags.dryRun {
				fmt.Fprintln(w, "  (dry run — order NOT placed)")
				return nil
			}

			// Step 7: Payment selection
			if paymentType == "" && !flags.noInput {
				// Check if user has cardOnFile scope
				hasCard := false
				if token != "" {
					claims := config.DecodeJWTClaims(token)
					hasCard = claims.HasScope("order:place:cardOnFile")
				}

				reader := bufio.NewReader(os.Stdin)
				if hasCard {
					fmt.Fprintln(w, "  Payment options:")
					fmt.Fprintln(w, "    1. Card on file")
					fmt.Fprintln(w, "    2. Cash")
					choice := promptLine(reader, cmd, "  Choose (1/2): ")
					if strings.TrimSpace(choice) == "2" {
						paymentType = "cash"
					} else {
						paymentType = "card"
					}
				} else {
					fmt.Fprintln(w, "  Payment: Cash (no saved card found)")
					paymentType = "cash"
				}
			}
			if paymentType == "" {
				paymentType = "cash"
			}

			// Step 8: Confirm
			if !flags.yes && !flags.noInput {
				reader := bufio.NewReader(os.Stdin)
				confirm := promptLine(reader, cmd, fmt.Sprintf("  Place order for $%.2f? (y/N): ", pr.Amounts.Customer))
				if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(confirm)), "y") {
					fmt.Fprintln(w, "  Order cancelled.")
					return nil
				}
			}

			// Step 9: Place order
			// Add payment to order
			var payments []map[string]any
			if paymentType == "card" {
				payments = []map[string]any{
					{
						"Type":   "DoorDebit",
						"Amount": pr.Amounts.Customer,
					},
				}
			} else {
				payments = []map[string]any{
					{
						"Type":   "Cash",
						"Amount": pr.Amounts.Customer,
					},
				}
			}

			// Re-parse the priced order to get the full object back
			var pricedFull map[string]any
			if err := json.Unmarshal(priceResp, &pricedFull); err != nil {
				return apiErr(fmt.Errorf("could not parse priced order: %w", err))
			}
			if orderObj, ok := pricedFull["Order"].(map[string]any); ok {
				orderObj["Payments"] = payments
				// Add customer info
				orderObj["FirstName"] = profile.FirstName
				orderObj["LastName"] = profile.LastName
				orderObj["Email"] = profile.Email
				orderObj["Phone"] = profile.Phone
			}

			placeBody := map[string]any{"Order": pricedFull["Order"]}
			placeResp, _, err := c.Post("/power/place-order", placeBody)
			if err != nil {
				return classifyAPIError(fmt.Errorf("order placement failed: %w", err))
			}

			var placeResult struct {
				Order struct {
					OrderID              string `json:"OrderID"`
					EstimatedWaitMinutes string `json:"EstimatedWaitMinutes"`
					Phone                string `json:"Phone"`
				} `json:"Order"`
			}
			if err := json.Unmarshal(placeResp, &placeResult); err == nil && placeResult.Order.OrderID != "" {
				fmt.Fprintln(w, green("OK")+" Order placed!")
				fmt.Fprintf(w, "  Order ID: %s\n", placeResult.Order.OrderID)
				if placeResult.Order.EstimatedWaitMinutes != "" {
					fmt.Fprintf(w, "  Est. wait: %s min\n", placeResult.Order.EstimatedWaitMinutes)
				}
				fmt.Fprintf(w, "  Track: dominos-pp-cli track --phone %s\n", profile.Phone)
			} else {
				// Still succeeded, just can't parse nicely
				fmt.Fprintln(w, green("OK")+" Order placed!")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&paymentType, "payment-type", "", "Payment method: card or cash")

	return cmd
}

// buildOrderFromCart constructs a Domino's API order object from local cart data.
func buildOrderFromCart(cart *cartRecord, items []cartItem, profile config.UserProfile) map[string]any {
	// Convert cart items to Domino's product format
	products := make([]map[string]any, len(items))
	for i, item := range items {
		options := map[string]any{}
		for _, t := range item.Toppings {
			side := "1/1"
			switch t.Side {
			case "left":
				side = "1/2"
			case "right":
				side = "2/2"
			}
			options[t.Code] = map[string]string{
				side: fmt.Sprintf("%.0f", t.Amount),
			}
		}

		products[i] = map[string]any{
			"Code":    item.Code,
			"Qty":     item.Qty,
			"ID":      i + 1,
			"isNew":   true,
			"Options": options,
		}
	}

	// Parse the stored address
	var addr map[string]any
	var addrObj struct {
		Display string `json:"display"`
	}
	if json.Unmarshal([]byte(cart.AddressJSON), &addrObj) == nil && addrObj.Display != "" {
		// Parse "Street, City, ST ZIP" format
		parts := strings.Split(addrObj.Display, ",")
		if len(parts) >= 2 {
			street := strings.TrimSpace(parts[0])
			rest := strings.TrimSpace(strings.Join(parts[1:], ","))
			// Try to split "City, ST ZIP" or "New York, NY 10023"
			restParts := strings.Fields(rest)
			city := ""
			region := ""
			zip := ""
			if len(restParts) >= 3 {
				// Last part is ZIP, second to last is state
				zip = restParts[len(restParts)-1]
				region = restParts[len(restParts)-2]
				city = strings.Join(restParts[:len(restParts)-2], " ")
			} else if len(restParts) == 2 {
				region = restParts[0]
				zip = restParts[1]
			}
			addr = map[string]any{
				"Street":     street,
				"City":       strings.TrimRight(city, ","),
				"Region":     region,
				"PostalCode": zip,
				"Type":       "House",
			}
		}
	}
	if addr == nil {
		addr = map[string]any{
			"Street": addrObj.Display,
			"Type":   "House",
		}
	}

	order := map[string]any{
		"Address":                addr,
		"Coupons":               []any{},
		"CustomerID":            profile.CustomerID,
		"Extension":             "",
		"OrderChannel":          "OLO",
		"OrderMethod":           "Web",
		"LanguageCode":          "en",
		"ServiceMethod":         "Delivery",
		"SourceOrganizationURI": "order.dominos.com",
		"StoreID":               cart.StoreID,
		"Tags":                  map[string]any{},
		"Version":               "1.0",
		"NoCombine":             true,
		"Partners":              map[string]any{},
		"NewUser":               true,
		"metaData":              map[string]any{},
		"Products":              products,
		"FirstName":             profile.FirstName,
		"LastName":              profile.LastName,
		"Email":                 profile.Email,
		"Phone":                 profile.Phone,
	}

	if cart.ServiceMethod != "" {
		svc := cart.ServiceMethod
		// Capitalize first letter for API
		if len(svc) > 0 {
			order["ServiceMethod"] = strings.ToUpper(svc[:1]) + svc[1:]
		}
	}

	return order
}

// promptLine prints a prompt and reads a line from stdin.
func promptLine(reader *bufio.Reader, cmd *cobra.Command, prompt string) string {
	fmt.Fprint(cmd.ErrOrStderr(), prompt)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
