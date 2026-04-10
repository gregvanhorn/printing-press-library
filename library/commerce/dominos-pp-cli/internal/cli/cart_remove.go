package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCartRemoveCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <item-index>",
		Short: "Remove an item from the active cart",
		Example: "  dominos-pp-cli cart remove 1\n" +
			"  dominos-pp-cli cart remove 2 --dry-run",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := parsePositiveInt(args[0])
			if err != nil {
				return err
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
			if index > len(items) {
				return usageErr(fmt.Errorf("item index %d out of range", index))
			}
			items = append(items[:index-1], items[index:]...)
			if flags.dryRun {
				return printMutationResult(cmd, flags, "cart.remove", cart, items)
			}
			if err := saveCart(s, cart, items); err != nil {
				return err
			}
			return printMutationResult(cmd, flags, "cart.remove", cart, items)
		},
	}
	return cmd
}

func parsePositiveInt(raw string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(raw, "%d", &n); err != nil || n < 1 {
		return 0, usageErr(fmt.Errorf("invalid positive integer %q", raw))
	}
	return n, nil
}
