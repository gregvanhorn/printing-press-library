package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/mvanhorn/printing-press-library/library/productivity/cal-com-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newWorkloadCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var period string

	cmd := &cobra.Command{
		Use:   "workload",
		Short: "Show booking distribution across team members",
		Long: `Analyze booking load across team members to identify who is overloaded or
underutilized. Useful for tuning round-robin weights and scheduling policies.`,
		Example: `  # Show team workload
  cal-com-pp-cli workload

  # Last 30 days
  cal-com-pp-cli workload --period 30d

  # JSON output
  cal-com-pp-cli workload --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("cal-com-pp-cli")
			}

			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening local database: %w\nRun 'cal-com-pp-cli sync' first.", err)
			}
			defer db.Close()

			filter := "1=1"
			if period != "" {
				ts, err := parseSinceDuration(period)
				if err != nil {
					return fmt.Errorf("invalid --period: %w", err)
				}
				filter = fmt.Sprintf("json_extract(data, '$.start') >= '%s'", ts.Format(time.RFC3339))
			}

			bookings, err := db.QueryJSON("bookings", filter)
			if err != nil {
				return fmt.Errorf("querying bookings: %w", err)
			}

			type memberLoad struct {
				Name           string  `json:"name"`
				Email          string  `json:"email"`
				BookingCount   int     `json:"booking_count"`
				TotalMinutes   float64 `json:"total_minutes"`
				CancelledCount int     `json:"cancelled_count"`
			}

			members := make(map[string]*memberLoad)

			for _, raw := range bookings {
				var b map[string]any
				if err := json.Unmarshal(raw, &b); err != nil {
					continue
				}

				// Extract hosts/organizer
				hosts := []map[string]any{}
				if h, ok := b["hosts"].([]any); ok {
					for _, host := range h {
						if hm, ok := host.(map[string]any); ok {
							hosts = append(hosts, hm)
						}
					}
				}
				// Fallback to user field
				if len(hosts) == 0 {
					if user, ok := b["user"].(map[string]any); ok {
						hosts = append(hosts, user)
					}
				}

				status := getString(b, "status")
				var duration float64
				if start, err := time.Parse(time.RFC3339, getString(b, "start")); err == nil {
					if end, err := time.Parse(time.RFC3339, getString(b, "end")); err == nil {
						duration = end.Sub(start).Minutes()
					}
				}

				for _, host := range hosts {
					email := getString(host, "email")
					if email == "" {
						continue
					}
					name := getString(host, "name")
					if _, ok := members[email]; !ok {
						members[email] = &memberLoad{Name: name, Email: email}
					}
					m := members[email]
					m.BookingCount++
					m.TotalMinutes += duration
					if status == "cancelled" {
						m.CancelledCount++
					}
				}
			}

			var sorted []*memberLoad
			for _, m := range members {
				sorted = append(sorted, m)
			}
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].BookingCount > sorted[j].BookingCount
			})

			if flags.asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(sorted)
			}

			if len(sorted) == 0 {
				fmt.Println("No host/organizer data found in bookings.")
				return nil
			}

			fmt.Println("Team Workload")
			fmt.Println("=============")
			for _, m := range sorted {
				name := m.Name
				if name == "" {
					name = m.Email
				}
				fmt.Printf("  %-25s %3d bookings  %5.0f min  %d cancelled\n",
					name, m.BookingCount, m.TotalMinutes, m.CancelledCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().StringVar(&period, "period", "", "Analysis period (e.g. 30d, 12w)")

	return cmd
}
