package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mvanhorn/printing-press-library/library/commerce/dominos-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

type cartRecord struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	StoreID       string `json:"store_id"`
	ServiceMethod string `json:"service_method"`
	AddressJSON   string `json:"address_json"`
	ItemsJSON     string `json:"items_json"`
	CouponsJSON   string `json:"coupons_json"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type cartItem struct {
	Code           string        `json:"code"`
	Qty            int           `json:"qty"`
	Size           string        `json:"size,omitempty"`
	Toppings       []cartTopping `json:"toppings,omitempty"`
	EstimatedPrice float64       `json:"estimated_price,omitempty"`
}

type cartTopping struct {
	Code   string  `json:"code"`
	Side   string  `json:"side"`
	Amount float64 `json:"amount"`
}

func newCartCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "cart", Short: "Manage locally stored shopping carts"}
	cmd.AddCommand(newCartNewCmd(flags), newCartAddCmd(flags), newCartRemoveCmd(flags), newCartShowCmd(flags), newCartAddCouponCmd(flags), newCartRemoveCouponCmd(flags), newCartCheckoutCmd(flags))
	return cmd
}

func openCartStore() (*store.Store, error) {
	s, err := store.Open(defaultDBPath("dominos-pp-cli"))
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}
	return s, nil
}

func loadActiveCart(s *store.Store) (*cartRecord, error) {
	rows, err := s.ListCarts()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, notFoundErr(fmt.Errorf("no active cart found; run `dominos-pp-cli cart new` first"))
	}
	var cart cartRecord
	if err := json.Unmarshal(rows[0], &cart); err != nil {
		return nil, fmt.Errorf("parsing active cart: %w", err)
	}
	return &cart, nil
}

func loadCartItems(c *cartRecord) ([]cartItem, error) {
	if c.ItemsJSON == "" {
		return []cartItem{}, nil
	}
	var items []cartItem
	if err := json.Unmarshal([]byte(c.ItemsJSON), &items); err != nil {
		return nil, fmt.Errorf("parsing cart items: %w", err)
	}
	return items, nil
}

func saveCart(s *store.Store, c *cartRecord, items []cartItem) error {
	data, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshaling cart items: %w", err)
	}
	return s.UpsertCart(c.ID, c.Name, c.StoreID, c.ServiceMethod, c.AddressJSON, string(data), c.CouponsJSON)
}

func newUUID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating uuid: %w", err)
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	hexed := hex.EncodeToString(buf)
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexed[:8], hexed[8:12], hexed[12:16], hexed[16:20], hexed[20:32]), nil
}

func printMutationResult(cmd *cobra.Command, flags *rootFlags, action string, cart *cartRecord, items []cartItem) error {
	out := map[string]any{
		"action":   action,
		"cart_id":  cart.ID,
		"name":     cart.Name,
		"store_id": cart.StoreID,
		"items":    items,
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
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s cart %s (%d item(s))\n", status, cart.ID, len(items))
	if cart.Name != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "template: %s\n", cart.Name)
	}
	return nil
}

func warnRecentCartFallback(cmd *cobra.Command) {
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "using most recently updated cart as active cart")
}

var _ = os.Stdin
