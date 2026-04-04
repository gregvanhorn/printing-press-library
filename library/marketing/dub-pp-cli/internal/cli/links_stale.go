package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mvanhorn/printing-press-library/library/marketing/dub-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newLinksStaleCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var days int
	var limit int

	cmd := &cobra.Command{
		Use:   "stale",
		Short: "Find links with zero clicks in the last N days",
		Long: `Identify links that haven't received any clicks recently.
Useful for cleaning up unused links or identifying underperforming campaigns.
Requires a prior sync.`,
		Example: `  # Find links with zero clicks in 30 days
  dub-pp-cli links stale --days 30

  # Show as JSON
  dub-pp-cli links stale --days 7 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("dub-pp-cli")
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			items, err := s.List("links", 5000)
			if err != nil {
				return fmt.Errorf("listing links: %w", err)
			}
			if len(items) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "No links in local store. Run 'sync' first.")
				return nil
			}

			cutoff := time.Now().AddDate(0, 0, -days)

			type staleLink struct {
				ID        string `json:"id"`
				Key       string `json:"key"`
				Domain    string `json:"domain"`
				URL       string `json:"url"`
				Clicks    int    `json:"clicks"`
				CreatedAt string `json:"createdAt"`
				LastClick string `json:"lastClicked,omitempty"`
			}

			var stale []staleLink
			for _, item := range items {
				var obj map[string]any
				if err := json.Unmarshal(item, &obj); err != nil {
					continue
				}

				clicks := intVal(obj, "clicks")
				lastClicked := strVal(obj, "lastClicked")
				createdAt := strVal(obj, "createdAt")

				isStale := false
				if clicks == 0 {
					isStale = true
				} else if lastClicked != "" {
					if t, err := time.Parse(time.RFC3339, lastClicked); err == nil {
						if t.Before(cutoff) {
							isStale = true
						}
					}
				}

				if isStale {
					stale = append(stale, staleLink{
						ID:        strVal(obj, "id"),
						Key:       strVal(obj, "key"),
						Domain:    strVal(obj, "domain"),
						URL:       strVal(obj, "url"),
						Clicks:    clicks,
						CreatedAt: createdAt,
						LastClick: lastClicked,
					})
				}
			}

			if limit > 0 && limit < len(stale) {
				stale = stale[:limit]
			}

			if flags.asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(stale)
			}

			if len(stale) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No stale links found (all links have clicks in the last %d days).\n", days)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Found %d stale links (no clicks in %d days):\n\n", len(stale), days)
			fmt.Fprintf(cmd.OutOrStdout(), "%-25s %-20s %8s %s\n", "Short Link", "Domain", "Clicks", "Last Click")
			fmt.Fprintf(cmd.OutOrStdout(), "%-25s %-20s %8s %s\n", "-------------------------", "--------------------", "--------", "----------")
			for _, sl := range stale {
				lastClick := sl.LastClick
				if lastClick == "" {
					lastClick = "never"
				} else if t, err := time.Parse(time.RFC3339, lastClick); err == nil {
					lastClick = t.Format("2006-01-02")
				}
				key := sl.Key
				if len(key) > 25 {
					key = key[:22] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-25s %-20s %8d %s\n", key, sl.Domain, sl.Clicks, lastClick)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().IntVar(&days, "days", 30, "Number of days to consider a link stale")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of stale links to show")

	return cmd
}
