package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mvanhorn/printing-press-library/library/marketing/dub-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newWorkflowPartnersLeaderboardCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var sortBy string
	var limit int

	cmd := &cobra.Command{
		Use:   "partners-leaderboard",
		Short: "Rank partners by revenue, clicks, or conversions",
		Long: `Show a leaderboard of partners ranked by total commission earnings,
click counts, or conversion counts. Requires a prior sync of partners and commissions.`,
		Example: `  # Partner leaderboard by earnings
  dub-pp-cli workflow partners-leaderboard

  # Top 5 by clicks
  dub-pp-cli workflow partners-leaderboard --by clicks --limit 5 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("dub-pp-cli")
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			partners, err := s.List("partners", 1000)
			if err != nil {
				return fmt.Errorf("listing partners: %w", err)
			}
			commissions, err := s.List("commissions", 5000)
			if err != nil {
				return fmt.Errorf("listing commissions: %w", err)
			}

			type partnerStat struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Email       string  `json:"email"`
				Clicks      int     `json:"clicks"`
				Leads       int     `json:"leads"`
				Sales       int     `json:"sales"`
				Earnings    float64 `json:"earnings"`
				Commissions int     `json:"commissions"`
			}

			pMap := make(map[string]*partnerStat)
			for _, p := range partners {
				var obj map[string]any
				if err := json.Unmarshal(p, &obj); err != nil {
					continue
				}
				id := strVal(obj, "id")
				pMap[id] = &partnerStat{
					ID:     id,
					Name:   strVal(obj, "name"),
					Email:  strVal(obj, "email"),
					Clicks: intVal(obj, "clicks"),
					Leads:  intVal(obj, "leads"),
					Sales:  intVal(obj, "sales"),
				}
			}

			// Aggregate commissions
			for _, c := range commissions {
				var obj map[string]any
				if err := json.Unmarshal(c, &obj); err != nil {
					continue
				}
				partnerID := strVal(obj, "partnerId")
				ps, ok := pMap[partnerID]
				if !ok {
					continue
				}
				ps.Commissions++
				if amount, ok := obj["amount"]; ok {
					if n, ok := amount.(float64); ok {
						ps.Earnings += n / 100.0 // amounts are typically in cents
					}
				}
			}

			var sorted []*partnerStat
			for _, ps := range pMap {
				sorted = append(sorted, ps)
			}
			sort.Slice(sorted, func(i, j int) bool {
				switch sortBy {
				case "clicks":
					return sorted[i].Clicks > sorted[j].Clicks
				case "conversions":
					return sorted[i].Sales > sorted[j].Sales
				default: // revenue
					return sorted[i].Earnings > sorted[j].Earnings
				}
			})

			if limit > 0 && limit < len(sorted) {
				sorted = sorted[:limit]
			}

			if flags.asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(sorted)
			}

			if len(sorted) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No partners found. Run 'sync' first.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-4s %-25s %8s %6s %6s %10s\n", "Rank", "Partner", "Clicks", "Leads", "Sales", "Earnings")
			fmt.Fprintf(cmd.OutOrStdout(), "%-4s %-25s %8s %6s %6s %10s\n", "----", "-------------------------", "--------", "------", "------", "----------")
			for i, ps := range sorted {
				name := ps.Name
				if name == "" {
					name = ps.Email
				}
				if len(name) > 25 {
					name = name[:22] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-4d %-25s %8d %6d %6d %10.2f\n", i+1, name, ps.Clicks, ps.Leads, ps.Sales, ps.Earnings)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().StringVar(&sortBy, "by", "revenue", "Sort by: revenue, clicks, or conversions")
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of partners to show")

	return cmd
}
