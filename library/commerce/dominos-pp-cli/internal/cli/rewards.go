package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newRewardsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "rewards", Short: "Show loyalty points, rewards, and member deals"}
	cmd.AddCommand(newRewardsPointsCmd(flags), newRewardsListCmd(flags), newRewardsDealsCmd(flags))
	return cmd
}

func newRewardsPointsCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "points", Short: "Show current loyalty points balance", RunE: func(cmd *cobra.Command, args []string) error {
		data, err := rewardsGraphQL(flags, "LoyaltyPoints", nil)
		if err != nil {
			return err
		}
		row := map[string]any{"id": pickGraph(data, "id", "memberId", "loyaltyId"), "points": pickGraphInt(data, "points", "balance"), "pending_points": pickGraphInt(data, "pendingPoints"), "status": pickGraph(data, "status", "tier")}
		s, err := flags.openStore()
		if err != nil {
			return err
		}
		defer s.Close()
		if err := s.UpsertLoyalty(fmt.Sprint(row["id"]), row["points"].(int), row["pending_points"].(int), fmt.Sprint(row["status"])); err != nil {
			return err
		}
		if flags.asJSON {
			return flags.printJSON(cmd, row)
		}
		return flags.printTable(cmd, []string{"POINTS", "PENDING", "STATUS"}, [][]string{{fmt.Sprintf("%v", row["points"]), fmt.Sprintf("%v", row["pending_points"]), fmt.Sprintf("%v", row["status"])}})
	}}
}

func newRewardsListCmd(flags *rootFlags) *cobra.Command {
	var storeID string
	return &cobra.Command{Use: "list", Short: "Show available rewards by tier", RunE: func(cmd *cobra.Command, args []string) error {
		vars := map[string]any{}
		if storeID != "" {
			vars["storeId"] = storeID
		}
		data, err := rewardsGraphQL(flags, "LoyaltyRewards", vars)
		if err != nil {
			return err
		}
		items := rewardsItems(data, false, storeID)
		if flags.asJSON {
			return flags.printJSON(cmd, items)
		}
		rows := make([][]string, 0, len(items))
		for _, it := range items {
			rows = append(rows, []string{fmt.Sprintf("%v", it["tier"]), fmt.Sprintf("%v", it["name"]), fmt.Sprintf("%v", it["price"])})
		}
		return flags.printTable(cmd, []string{"TIER", "NAME", "PRICE"}, rows)
	}}
}

func newRewardsDealsCmd(flags *rootFlags) *cobra.Command {
	var storeID string
	cmd := &cobra.Command{Use: "deals", Short: "Show member-exclusive deals", RunE: func(cmd *cobra.Command, args []string) error {
		vars := map[string]any{}
		if storeID != "" {
			vars["storeId"] = storeID
		}
		data, err := rewardsGraphQL(flags, "LoyaltyDeals", vars)
		if err != nil {
			return err
		}
		items := rewardsItems(data, true, storeID)
		if storeID != "" {
			s, err := flags.openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			for _, it := range items {
				_ = s.UpsertDeal(fmt.Sprint(it["id"]), storeID, fmt.Sprint(it["code"]), fmt.Sprint(it["name"]), fmt.Sprint(it["description"]), fmt.Sprint(it["price"]), "", true)
			}
		}
		if flags.asJSON {
			return flags.printJSON(cmd, items)
		}
		rows := make([][]string, 0, len(items))
		for _, it := range items {
			rows = append(rows, []string{fmt.Sprintf("%v", it["code"]), fmt.Sprintf("%v", it["name"]), fmt.Sprintf("%v", it["price"])})
		}
		return flags.printTable(cmd, []string{"CODE", "NAME", "PRICE"}, rows)
	}}
	cmd.Flags().StringVar(&storeID, "store", "", "Store ID")
	return cmd
}

func rewardsGraphQL(flags *rootFlags, op string, vars map[string]any) (json.RawMessage, error) {
	c, err := flags.newClient()
	if err != nil {
		return nil, err
	}
	data, _, err := c.Post("/web-bff/graphql", map[string]any{"operationName": op, "variables": vars, "query": ""})
	if err != nil {
		return nil, classifyAPIError(err)
	}
	return data, nil
}

func rewardsItems(data json.RawMessage, memberOnly bool, storeID string) []map[string]any {
	out, seen := []map[string]any{}, map[string]bool{}
	for _, m := range findMaps(data) {
		name := pickS(m, "name", "title")
		if name == "" {
			continue
		}
		id := firstNonEmpty(pickS(m, "id", "rewardId", "dealId", "code"), storeID+":"+name)
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, map[string]any{"id": id, "tier": pickS(m, "tier", "points", "requiredPoints"), "code": pickS(m, "code", "dealCode"), "name": name, "description": pickS(m, "description", "desc"), "price": pickS(m, "price", "displayPrice", "value"), "member_only": memberOnly})
	}
	return out
}

func pickGraph(data json.RawMessage, keys ...string) string {
	for _, m := range findMaps(data) {
		if v := pickS(m, keys...); v != "" {
			return v
		}
	}
	return ""
}
func pickGraphInt(data json.RawMessage, keys ...string) int {
	for _, m := range findMaps(data) {
		for _, k := range keys {
			switch v := m[k].(type) {
			case float64:
				return int(v)
			case int:
				return v
			}
		}
	}
	return 0
}
