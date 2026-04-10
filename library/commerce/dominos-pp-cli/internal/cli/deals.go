package cli

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/spf13/cobra"
)

type dealRecord struct {
	ID, StoreID, Code, Name, Description, Price, ExpiresAt, SyncedAt string
	IsMemberOnly                                                     bool
}

func newDealsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "deals", Short: "List and analyze store deals"}
	cmd.AddCommand(newDealsListCmd(flags), newDealsBestCmd(flags))
	return cmd
}

func newDealsListCmd(flags *rootFlags) *cobra.Command {
	var storeID, service string
	cmd := &cobra.Command{Use: "list --store <id>", Short: "List all deals for a store", RunE: func(cmd *cobra.Command, args []string) error {
		c, err := flags.newClient()
		if err != nil {
			return err
		}
		body := map[string]any{"operationName": "DealsList", "variables": map[string]any{"storeId": storeID, "serviceMethod": service}, "query": ""}
		data, _, err := c.Post("/web-bff/graphql", body)
		if err != nil {
			return classifyAPIError(err)
		}
		deals := normalizeDeals(data, storeID)
		s, err := flags.openStore()
		if err != nil {
			return err
		}
		defer s.Close()
		for _, d := range deals {
			if err := s.UpsertDeal(d.ID, d.StoreID, d.Code, d.Name, d.Description, d.Price, d.ExpiresAt, d.IsMemberOnly); err != nil {
				return err
			}
		}
		if flags.asJSON {
			return flags.printJSON(cmd, deals)
		}
		rows := make([][]string, 0, len(deals))
		for _, d := range deals {
			rows = append(rows, []string{d.Code, d.Name, d.Price, yesNo(d.IsMemberOnly), d.ExpiresAt})
		}
		return flags.printTable(cmd, []string{"CODE", "NAME", "PRICE", "MEMBER", "EXPIRES"}, rows)
	}}
	cmd.Flags().StringVar(&storeID, "store", "", "Store ID")
	cmd.Flags().StringVar(&service, "service", "delivery", "Service method")
	_ = cmd.MarkFlagRequired("store")
	return cmd
}

func newDealsBestCmd(flags *rootFlags) *cobra.Command {
	var storeID string
	var useCart bool
	cmd := &cobra.Command{Use: "best --cart [--store <id>]", Short: "Find the cheapest deal combination for the active cart", RunE: func(cmd *cobra.Command, args []string) error {
		if !useCart {
			return usageErr(fmt.Errorf("--cart is required"))
		}
		s, err := flags.openStore()
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
		if storeID == "" {
			storeID = cart.StoreID
		}
		rows, err := s.ListDeals(storeID)
		if err != nil {
			return err
		}
		deals := rawDeals(rows)
		chosen, total := bestDealsForCart(items, deals)
		out := map[string]any{"store_id": storeID, "cart_id": cart.ID, "cart_estimated_total": sumCart(items), "deal_total_estimate": total, "recommended_deals": chosen}
		if flags.asJSON {
			return flags.printJSON(cmd, out)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Cart %s for store %s\nEstimated cart total: %s\nDeal total estimate: %s\n\n", cart.ID, storeID, money(sumCart(items)), money(total))
		tab := make([][]string, 0, len(chosen))
		for _, d := range chosen {
			tab = append(tab, []string{d.Code, d.Name, d.Price, fmt.Sprintf("%d", dealCoverage(d, items))})
		}
		return flags.printTable(cmd, []string{"CODE", "NAME", "PRICE", "MATCHES"}, tab)
	}}
	cmd.Flags().BoolVar(&useCart, "cart", false, "Use the active cart")
	cmd.Flags().StringVar(&storeID, "store", "", "Override store ID")
	return cmd
}

func normalizeDeals(data json.RawMessage, storeID string) []dealRecord {
	seen, out := map[string]bool{}, []dealRecord{}
	for _, m := range findMaps(data) {
		d := dealRecord{ID: pickS(m, "id", "dealId", "couponId", "code", "couponCode"), StoreID: storeID, Code: pickS(m, "code", "couponCode", "dealCode"), Name: pickS(m, "name", "title"), Description: pickS(m, "description", "desc"), Price: pickS(m, "price", "displayPrice", "dealPrice", "amount"), ExpiresAt: pickS(m, "expiresAt", "expirationDate", "expiration"), IsMemberOnly: pickB(m, "memberOnly", "isMemberOnly", "loyaltyOnly")}
		if d.ID == "" {
			d.ID = firstNonEmpty(d.StoreID+":"+d.Code, d.StoreID+":"+d.Name)
		}
		if d.Name == "" || seen[d.ID] {
			continue
		}
		seen[d.ID] = true
		out = append(out, d)
	}
	return out
}

func rawDeals(rows []json.RawMessage) []dealRecord {
	out := make([]dealRecord, 0, len(rows))
	for _, row := range rows {
		var d dealRecord
		if json.Unmarshal(row, &d) == nil {
			out = append(out, d)
		}
	}
	return out
}

func bestDealsForCart(items []cartItem, deals []dealRecord) ([]dealRecord, float64) {
	left, chosen, total := map[string]int{}, []dealRecord{}, 0.0
	for _, it := range items {
		left[strings.ToUpper(it.Code)] += it.Qty
	}
	for len(left) > 0 {
		best, score := -1, math.MaxFloat64
		for i, d := range deals {
			cov := dealCoverageRemaining(d, left)
			if cov == 0 {
				continue
			}
			if s := parsePrice(d.Price) / float64(cov); s < score {
				best, score = i, s
			}
		}
		if best < 0 {
			break
		}
		chosen, total = append(chosen, deals[best]), total+parsePrice(deals[best].Price)
		for k := range left {
			if strings.Contains(strings.ToUpper(deals[best].Name+" "+deals[best].Description), k) {
				delete(left, k)
			}
		}
	}
	if len(chosen) == 0 && len(deals) > 0 {
		best := deals[0]
		for _, d := range deals[1:] {
			if parsePrice(d.Price) < parsePrice(best.Price) {
				best = d
			}
		}
		return []dealRecord{best}, parsePrice(best.Price)
	}
	return chosen, total
}

func dealCoverageRemaining(d dealRecord, left map[string]int) int {
	hay, n := strings.ToUpper(d.Name+" "+d.Description+" "+d.Code), 0
	for code, qty := range left {
		if qty > 0 && strings.Contains(hay, code) {
			n += qty
		}
	}
	return n
}
