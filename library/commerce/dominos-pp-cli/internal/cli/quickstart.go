package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/commerce/dominos-pp-cli/internal/config"
	"github.com/spf13/cobra"
)

func newQuickstartCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quickstart",
		Short: "Get started with Domino's CLI — guided setup and first order",
		Long: `Walks you through setting up the CLI and placing your first order.

Steps:
  1. Check authentication (token setup)
  2. Save your delivery address
  3. Find your nearest store
  4. Browse the menu and add to cart
  5. Check out`,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return configErr(err)
			}

			// Non-interactive mode: output checklist
			if flags.noInput {
				return quickstartChecklist(cmd, cfg, flags)
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Fprintln(w, "")
			fmt.Fprintln(w, bold("Welcome to the Domino's CLI!"))
			fmt.Fprintln(w, "Let's get you set up to order pizza from your terminal.")
			fmt.Fprintln(w, "")

			// Step 1: Auth
			header := cfg.AuthHeader()
			if header == "" {
				fmt.Fprintln(w, bold("Step 1: Authentication"))
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "  You need an auth token from your Domino's account.")
				fmt.Fprintln(w, "  To get one:")
				fmt.Fprintln(w, "    1. Log in to order.dominos.com in your browser")
				fmt.Fprintln(w, "    2. Open DevTools > Application > Cookies")
				fmt.Fprintln(w, "    3. Find the 'dpz-cus-token' cookie")
				fmt.Fprintln(w, "    4. Copy its value")
				fmt.Fprintln(w, "")
				token := promptLine(reader, cmd, "  Paste your token (or press Enter to skip): ")
				if token != "" {
					cfg.DominosToken = token
					if err := cfg.SaveProfile(cfg.FirstName, cfg.LastName, cfg.Phone, cfg.Email); err != nil {
						fmt.Fprintf(w, "  %s Could not save token: %v\n", red("FAIL"), err)
					} else {
						fmt.Fprintf(w, "  %s Token saved!\n", green("OK"))
					}
				} else {
					fmt.Fprintln(w, "  Skipped — you can set it later:")
					fmt.Fprintln(w, "    dominos-pp-cli auth set-token <token>")
				}
				fmt.Fprintln(w, "")
			} else {
				profile := cfg.Profile()
				fmt.Fprintf(w, "  %s Authenticated", green("OK"))
				if profile.Email != "" {
					fmt.Fprintf(w, " (%s)", profile.Email)
				}
				fmt.Fprintln(w, "")
			}

			// Step 2: Profile info
			fmt.Fprintln(w, bold("Step 2: Your info"))
			profile := cfg.Profile()
			changed := false
			if profile.FirstName == "" {
				profile.FirstName = promptLine(reader, cmd, "  First name: ")
				changed = true
			} else {
				fmt.Fprintf(w, "  First name: %s\n", profile.FirstName)
			}
			if profile.LastName == "" {
				profile.LastName = promptLine(reader, cmd, "  Last name: ")
				changed = true
			} else {
				fmt.Fprintf(w, "  Last name: %s\n", profile.LastName)
			}
			if profile.Email == "" {
				profile.Email = promptLine(reader, cmd, "  Email: ")
				changed = true
			} else {
				fmt.Fprintf(w, "  Email: %s\n", profile.Email)
			}
			if profile.Phone == "" {
				profile.Phone = promptLine(reader, cmd, "  Phone (10 digits): ")
				changed = true
			} else {
				fmt.Fprintf(w, "  Phone: %s\n", profile.Phone)
			}
			if changed {
				_ = cfg.SaveProfile(profile.FirstName, profile.LastName, profile.Phone, profile.Email)
				fmt.Fprintf(w, "  %s Profile saved!\n", green("OK"))
			}
			fmt.Fprintln(w, "")

			// Step 3: Delivery address
			fmt.Fprintln(w, bold("Step 3: Delivery address"))
			address := promptLine(reader, cmd, "  Enter your delivery address (e.g. 2 Lincoln Square, New York, NY 10023): ")
			if address == "" {
				fmt.Fprintln(w, "  Skipped — you'll need an address when creating a cart.")
				fmt.Fprintln(w, "")
				printNextSteps(w)
				return nil
			}
			fmt.Fprintln(w, "")

			// Step 4: Find nearest store
			fmt.Fprintln(w, bold("Step 4: Finding nearby stores..."))

			// Parse address into street + city
			addrParts := strings.SplitN(address, ",", 2)
			street := strings.TrimSpace(addrParts[0])
			cityStateZip := ""
			if len(addrParts) > 1 {
				cityStateZip = strings.TrimSpace(addrParts[1])
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			storeResp, err := c.Get("/power/store-locator", map[string]string{
				"s":    street,
				"c":    cityStateZip,
				"type": "Delivery",
			})
			if err != nil {
				fmt.Fprintf(w, "  %s Could not find stores: %v\n", red("FAIL"), err)
				fmt.Fprintln(w, "")
				printNextSteps(w)
				return nil
			}

			var storeResult struct {
				Stores []struct {
					StoreID         string `json:"StoreID"`
					IsDeliveryStore bool   `json:"IsDeliveryStore"`
					MinDistance     float64 `json:"MinDistance"`
					IsOpen          bool   `json:"IsOpen"`
					Address         struct {
						Street string `json:"Street"`
						City   string `json:"City"`
						Region string `json:"Region"`
					} `json:"Address"`
				} `json:"Stores"`
			}
			if err := json.Unmarshal(storeResp, &storeResult); err != nil || len(storeResult.Stores) == 0 {
				fmt.Fprintf(w, "  %s No delivery stores found near that address\n", yellow("WARN"))
				fmt.Fprintln(w, "")
				printNextSteps(w)
				return nil
			}

			// Find closest delivery store
			store := storeResult.Stores[0]
			for _, s := range storeResult.Stores {
				if s.IsDeliveryStore && s.IsOpen {
					store = s
					break
				}
			}

			fmt.Fprintf(w, "  %s Nearest store: #%s — %s, %s, %s (%.1f mi)\n",
				green("OK"), store.StoreID, store.Address.Street, store.Address.City, store.Address.Region, store.MinDistance)
			fmt.Fprintln(w, "")

			// Step 5: Create cart and add a pepperoni pizza
			fmt.Fprintln(w, bold("Step 5: Let's add something to your cart"))
			fmt.Fprintln(w, "  Popular items:")
			fmt.Fprintln(w, "    1. Large Hand Tossed Pepperoni Pizza")
			fmt.Fprintln(w, "    2. Medium Hand Tossed Cheese Pizza")
			fmt.Fprintln(w, "    3. Pepperoni Stuffed Cheesy Bread")
			fmt.Fprintln(w, "    4. Skip — I'll browse the menu myself")

			choice := promptLine(reader, cmd, "  Choose (1-4): ")

			productCode := ""
			var toppings []cartTopping
			switch strings.TrimSpace(choice) {
			case "1":
				productCode = "14SCREEN"
				toppings = []cartTopping{{Code: "P", Side: "full", Amount: 1}}
			case "2":
				productCode = "14SCREEN"
			case "3":
				productCode = "B8PCSPP"
			default:
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "  Browse the menu:")
				fmt.Fprintf(w, "    dominos-pp-cli menu search pizza --store %s\n", store.StoreID)
				fmt.Fprintln(w, "")
				fmt.Fprintln(w, "  Then create a cart and add items:")
				fmt.Fprintf(w, "    dominos-pp-cli cart new --store %s --service delivery --address \"%s\"\n", store.StoreID, address)
				fmt.Fprintln(w, "    dominos-pp-cli cart add <product-code>")
				fmt.Fprintln(w, "    dominos-pp-cli checkout")
				return nil
			}

			// Create cart
			cartID, err := newUUID()
			if err != nil {
				return err
			}

			cs, err := openCartStore()
			if err != nil {
				return err
			}

			addrJSON, _ := json.Marshal(map[string]string{"display": address})
			cartRec := &cartRecord{
				ID:            cartID,
				StoreID:       store.StoreID,
				ServiceMethod: "delivery",
				AddressJSON:   string(addrJSON),
			}

			cartItems := []cartItem{
				{
					Code:     productCode,
					Qty:      1,
					Toppings: toppings,
				},
			}
			if err := saveCart(cs, cartRec, cartItems); err != nil {
				return fmt.Errorf("saving cart: %w", err)
			}
			fmt.Fprintf(w, "  %s Cart created with 1 item\n", green("OK"))
			fmt.Fprintln(w, "")

			// Step 6: Checkout preview
			fmt.Fprintln(w, bold("Step 6: Ready to checkout!"))
			fmt.Fprintln(w, "  Run this to review and place your order:")
			fmt.Fprintln(w, "    dominos-pp-cli checkout")
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, "  Or preview first:")
			fmt.Fprintln(w, "    dominos-pp-cli checkout --dry-run")

			return nil
		},
	}

	return cmd
}

func quickstartChecklist(cmd *cobra.Command, cfg *config.Config, flags *rootFlags) error {
	w := cmd.OutOrStdout()
	issues := []string{}

	// Check auth
	header := cfg.AuthHeader()
	if header == "" {
		issues = append(issues, "auth: not configured — run: dominos-pp-cli auth set-token <token>")
	}

	// Check profile
	profile := cfg.Profile()
	if profile.FirstName == "" || profile.LastName == "" {
		issues = append(issues, "profile: name missing — set first_name/last_name in config or run quickstart interactively")
	}
	if profile.Email == "" {
		issues = append(issues, "profile: email missing — set email in config or authenticate")
	}
	if profile.Phone == "" {
		issues = append(issues, "profile: phone missing — set phone in config or run quickstart interactively")
	}

	// Check for active cart
	s, err := openCartStore()
	if err == nil {
		if _, err := loadActiveCart(s); err != nil {
			issues = append(issues, "cart: no active cart — run: dominos-pp-cli cart new --store <id> --service delivery --address \"your address\"")
		}
	}

	if flags.asJSON {
		result := map[string]any{
			"ready":  len(issues) == 0,
			"issues": issues,
		}
		return flags.printJSON(cmd, result)
	}

	if len(issues) == 0 {
		fmt.Fprintf(w, "  %s All set! Run: dominos-pp-cli checkout\n", green("OK"))
	} else {
		fmt.Fprintln(w, "  Setup checklist:")
		for _, issue := range issues {
			fmt.Fprintf(w, "    %s %s\n", red("x"), issue)
		}
	}
	return nil
}

func printNextSteps(w io.Writer) {
	fmt.Fprintln(w, "  Next steps:")
	fmt.Fprintln(w, "    dominos-pp-cli stores find_stores --s \"your street\" --c \"city, ST ZIP\"")
	fmt.Fprintln(w, "    dominos-pp-cli cart new --store <id> --service delivery --address \"your address\"")
	fmt.Fprintln(w, "    dominos-pp-cli cart add <product-code>")
	fmt.Fprintln(w, "    dominos-pp-cli checkout")
}
