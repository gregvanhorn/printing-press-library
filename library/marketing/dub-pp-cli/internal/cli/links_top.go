package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mvanhorn/printing-press-library/library/marketing/dub-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newLinksTopCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var limit int
	var sortBy string

	cmd := &cobra.Command{
		Use:   "top",
		Short: "Rank links by performance (clicks, leads, sales) from local analytics data",
		Long: `Rank your links by click count, lead conversions, or sale conversions.
Requires a prior sync to populate the local store with link and analytics data.`,
		Example: `  # Top 10 links by clicks
  dub-pp-cli links top

  # Top 20 links by clicks as JSON
  dub-pp-cli links top --limit 20 --json

  # Top links sorted by leads
  dub-pp-cli links top --by leads`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("dub-pp-cli")
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			items, err := s.List("links", 1000)
			if err != nil {
				return fmt.Errorf("listing links: %w", err)
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "No links in local store. Run 'sync' first.")
				return nil
			}

			type linkStat struct {
				ID     string `json:"id"`
				Key    string `json:"key"`
				Domain string `json:"domain"`
				URL    string `json:"url"`
				Clicks int    `json:"clicks"`
				Leads  int    `json:"leads"`
				Sales  int    `json:"sales"`
			}

			var stats []linkStat
			for _, item := range items {
				var obj map[string]any
				if err := json.Unmarshal(item, &obj); err != nil {
					continue
				}
				ls := linkStat{
					ID:     strVal(obj, "id"),
					Key:    strVal(obj, "key"),
					Domain: strVal(obj, "domain"),
					URL:    strVal(obj, "url"),
					Clicks: intVal(obj, "clicks"),
					Leads:  intVal(obj, "leads"),
					Sales:  intVal(obj, "sales"),
				}
				stats = append(stats, ls)
			}

			sort.Slice(stats, func(i, j int) bool {
				switch sortBy {
				case "leads":
					return stats[i].Leads > stats[j].Leads
				case "sales":
					return stats[i].Sales > stats[j].Sales
				default:
					return stats[i].Clicks > stats[j].Clicks
				}
			})

			if limit > 0 && limit < len(stats) {
				stats = stats[:limit]
			}

			if flags.asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-4s %-30s %-20s %8s %6s %6s\n", "Rank", "Short Link", "Domain", "Clicks", "Leads", "Sales")
			fmt.Fprintf(cmd.OutOrStdout(), "%-4s %-30s %-20s %8s %6s %6s\n", "----", "------------------------------", "--------------------", "--------", "------", "------")
			for i, ls := range stats {
				shortLink := ls.Key
				if len(shortLink) > 30 {
					shortLink = shortLink[:27] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-4d %-30s %-20s %8d %6d %6d\n", i+1, shortLink, ls.Domain, ls.Clicks, ls.Leads, ls.Sales)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().IntVar(&limit, "limit", 10, "Number of top links to show")
	cmd.Flags().StringVar(&sortBy, "by", "clicks", "Sort by: clicks, leads, or sales")

	return cmd
}

func strVal(obj map[string]any, key string) string {
	if v, ok := obj[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func intVal(obj map[string]any, key string) int {
	if v, ok := obj[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}
