package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/commerce/dominos-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

type templateRecord struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	StoreID       string `json:"store_id"`
	ItemsJSON     string `json:"items_json"`
	AddressJSON   string `json:"address_json"`
	ServiceMethod string `json:"service_method"`
	CreatedAt     string `json:"created_at"`
}

func (f *rootFlags) openStore() (*store.Store, error) {
	s, err := store.Open(defaultDBPath("dominos-pp-cli"))
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}
	return s, nil
}

func newTemplateCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "template", Short: "Manage saved order templates"}
	cmd.AddCommand(newTemplateSaveCmd(flags), newTemplateListCmd(flags), newTemplateShowCmd(flags), newTemplateDeleteCmd(flags), newTemplateOrderCmd(flags))
	return cmd
}

func newTemplateSaveCmd(flags *rootFlags) *cobra.Command {
	var storeID, items, address, service string
	cmd := &cobra.Command{
		Use:   "save <name> --store <id> --items <code1,code2,...> --address <addr> --service <delivery|carryout>",
		Short: "Save a named order template",
		Example: "  dominos-pp-cli template save friday-night --store 4336 --items 14SCREEN,20BCOKE --address \"1 Market St, San Francisco, CA 94105\" --service delivery\n" +
			"  dominos-pp-cli template save pickup-lunch --store 6918 --items P12IPAZA,BONLESS --address \"500 Howard St, San Francisco, CA 94105\" --service carryout",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if service != "delivery" && service != "carryout" {
				return usageErr(fmt.Errorf("service must be delivery or carryout"))
			}
			id, err := newUUID()
			if err != nil {
				return err
			}
			itemsJSON, err := json.Marshal(parseTemplateItems(items))
			if err != nil {
				return fmt.Errorf("marshaling items: %w", err)
			}
			addressJSON, err := json.Marshal(map[string]any{"display": address})
			if err != nil {
				return fmt.Errorf("marshaling address: %w", err)
			}
			rec := templateRecord{ID: id, Name: args[0], StoreID: storeID, ItemsJSON: string(itemsJSON), AddressJSON: string(addressJSON), ServiceMethod: service}
			if flags.dryRun {
				return renderAction(cmd, flags, "saved", map[string]any{"action": "template.save", "template": rec, "item_count": len(parseTemplateItems(items))})
			}
			s, err := flags.openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.UpsertOrderTemplate(rec.ID, rec.Name, rec.StoreID, rec.ItemsJSON, rec.AddressJSON, rec.ServiceMethod); err != nil {
				return err
			}
			return renderAction(cmd, flags, "saved", map[string]any{"action": "template.save", "template": rec, "item_count": len(parseTemplateItems(items))})
		},
	}
	cmd.Flags().StringVar(&storeID, "store", "", "Store ID")
	cmd.Flags().StringVar(&items, "items", "", "Comma-separated menu item codes")
	cmd.Flags().StringVar(&address, "address", "", "Address text to save with the template")
	cmd.Flags().StringVar(&service, "service", "", "Service method")
	_ = cmd.MarkFlagRequired("store")
	_ = cmd.MarkFlagRequired("items")
	_ = cmd.MarkFlagRequired("address")
	_ = cmd.MarkFlagRequired("service")
	return cmd
}

func newTemplateListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List saved order templates", Example: "  dominos-pp-cli template list\n  dominos-pp-cli template list --json", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := flags.openStore()
		if err != nil {
			return err
		}
		defer s.Close()
		rows, err := s.ListOrderTemplates()
		if err != nil {
			return err
		}
		items := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			var rec templateRecord
			if json.Unmarshal(row, &rec) != nil {
				continue
			}
			var saved []cartItem
			_ = json.Unmarshal([]byte(rec.ItemsJSON), &saved)
			items = append(items, map[string]any{"name": rec.Name, "store_id": rec.StoreID, "service_method": rec.ServiceMethod, "item_count": len(saved)})
		}
		if flags.asJSON {
			return flags.printJSON(cmd, items)
		}
		rowsOut := make([][]string, 0, len(items))
		for _, item := range items {
			rowsOut = append(rowsOut, []string{fmt.Sprintf("%v", item["name"]), fmt.Sprintf("%v", item["store_id"]), fmt.Sprintf("%v", item["service_method"]), fmt.Sprintf("%v", item["item_count"])})
		}
		return flags.printTable(cmd, []string{"NAME", "STORE", "SERVICE", "ITEMS"}, rowsOut)
	}}
}

func newTemplateShowCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "show <name>", Short: "Show a saved order template", Example: "  dominos-pp-cli template show friday-night", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		s, err := flags.openStore()
		if err != nil {
			return err
		}
		defer s.Close()
		row, err := s.GetOrderTemplate(args[0])
		if err != nil {
			return notFoundErr(fmt.Errorf("template %q not found", args[0]))
		}
		return printOutputWithFlags(cmd.OutOrStdout(), row, flags)
	}}
}

func newTemplateDeleteCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "delete <name>", Short: "Delete a saved order template", Example: "  dominos-pp-cli template delete friday-night --dry-run\n  dominos-pp-cli template delete friday-night", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		out := map[string]any{"action": "template.delete", "name": args[0]}
		if flags.dryRun {
			return renderAction(cmd, flags, "saved", out)
		}
		s, err := flags.openStore()
		if err != nil {
			return err
		}
		defer s.Close()
		if err := s.DeleteOrderTemplate(args[0]); err != nil {
			return err
		}
		return renderAction(cmd, flags, "saved", out)
	}}
}

func newTemplateOrderCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "order <name>", Short: "Create a new local cart from a template", Example: "  dominos-pp-cli template order friday-night\n  dominos-pp-cli template order friday-night --dry-run", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		s, err := flags.openStore()
		if err != nil {
			return err
		}
		defer s.Close()
		row, err := s.GetOrderTemplate(args[0])
		if err != nil {
			return notFoundErr(fmt.Errorf("template %q not found", args[0]))
		}
		var rec templateRecord
		if err := json.Unmarshal(row, &rec); err != nil {
			return fmt.Errorf("parsing template: %w", err)
		}
		items, err := loadCartItems(&cartRecord{ItemsJSON: rec.ItemsJSON})
		if err != nil {
			return err
		}
		id, err := newUUID()
		if err != nil {
			return err
		}
		cart := &cartRecord{ID: id, Name: rec.Name, StoreID: rec.StoreID, ServiceMethod: rec.ServiceMethod, AddressJSON: rec.AddressJSON, ItemsJSON: rec.ItemsJSON}
		if flags.dryRun {
			return printMutationResult(cmd, flags, "template.order", cart, items)
		}
		if err := s.UpsertCart(cart.ID, cart.Name, cart.StoreID, cart.ServiceMethod, cart.AddressJSON, cart.ItemsJSON); err != nil {
			return err
		}
		return printMutationResult(cmd, flags, "template.order", cart, items)
	}}
}

func parseTemplateItems(raw string) []cartItem {
	parts := strings.Split(raw, ",")
	items := make([]cartItem, 0, len(parts))
	for _, part := range parts {
		if code := strings.TrimSpace(part); code != "" {
			items = append(items, cartItem{Code: code, Qty: 1})
		}
	}
	return items
}
