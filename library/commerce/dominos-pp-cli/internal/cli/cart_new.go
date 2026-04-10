package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newCartNewCmd(flags *rootFlags) *cobra.Command {
	var storeID, service, address, name string
	cmd := &cobra.Command{
		Use:   "new --store <id> --service <delivery|carryout> --address <addr> [--name <template-name>]",
		Short: "Create a new local cart",
		Example: "  dominos-pp-cli cart new --store 1234 --service delivery --address \"1 Market St, San Francisco, CA\" --name friday-night\n" +
			"  dominos-pp-cli cart new --store 1234 --service carryout --address \"1 Market St, San Francisco, CA\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			if service != "delivery" && service != "carryout" {
				return usageErr(fmt.Errorf("service must be delivery or carryout"))
			}
			id, err := newUUID()
			if err != nil {
				return err
			}
			addr, _ := json.Marshal(map[string]any{"display": address})
			cart := &cartRecord{ID: id, Name: name, StoreID: storeID, ServiceMethod: service, AddressJSON: string(addr)}
			if flags.dryRun {
				return printMutationResult(cmd, flags, "cart.new", cart, []cartItem{})
			}
			s, err := openCartStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.UpsertCart(cart.ID, cart.Name, cart.StoreID, cart.ServiceMethod, cart.AddressJSON, "[]"); err != nil {
				return err
			}
			if name != "" {
				if err := s.UpsertOrderTemplate(cart.ID, name, storeID, "[]", cart.AddressJSON, service); err != nil {
					return err
				}
			}
			return printMutationResult(cmd, flags, "cart.new", cart, []cartItem{})
		},
	}
	cmd.Flags().StringVar(&storeID, "store", "", "Store ID")
	cmd.Flags().StringVar(&service, "service", "", "Service method")
	cmd.Flags().StringVar(&address, "address", "", "Delivery or carryout address")
	cmd.Flags().StringVar(&name, "name", "", "Optional template name")
	_ = cmd.MarkFlagRequired("store")
	_ = cmd.MarkFlagRequired("service")
	_ = cmd.MarkFlagRequired("address")
	return cmd
}
