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

func newNoshowCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "noshow",
		Short: "Analyze no-show patterns by event type, day, and time",
		Long: `Identify which event types, days of the week, and time slots have the highest
no-show rates. Helps optimize your schedule by avoiding high-risk slots.`,
		Example: `  # Show no-show analysis
  cal-com-pp-cli noshow

  # JSON output
  cal-com-pp-cli noshow --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("cal-com-pp-cli")
			}

			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening local database: %w\nRun 'cal-com-pp-cli sync' first.", err)
			}
			defer db.Close()

			bookings, err := db.QueryJSON("bookings", "json_extract(data, '$.status') IN ('accepted', 'no_show')")
			if err != nil {
				// Fallback: get all completed bookings and check noShowHost field
				bookings, err = db.QueryJSON("bookings", "1=1")
				if err != nil {
					return fmt.Errorf("querying bookings: %w", err)
				}
			}

			type noshowStats struct {
				ByEventType map[string]struct {
					Total  int `json:"total"`
					NoShow int `json:"no_show"`
				} `json:"by_event_type"`
				ByDay map[string]struct {
					Total  int `json:"total"`
					NoShow int `json:"no_show"`
				} `json:"by_day"`
				ByHour map[int]struct {
					Total  int `json:"total"`
					NoShow int `json:"no_show"`
				} `json:"by_hour"`
				TotalBookings int     `json:"total_bookings"`
				TotalNoShows  int     `json:"total_no_shows"`
				OverallRate   float64 `json:"overall_rate_pct"`
			}

			stats := noshowStats{
				ByEventType: make(map[string]struct {
					Total  int `json:"total"`
					NoShow int `json:"no_show"`
				}),
				ByDay: make(map[string]struct {
					Total  int `json:"total"`
					NoShow int `json:"no_show"`
				}),
				ByHour: make(map[int]struct {
					Total  int `json:"total"`
					NoShow int `json:"no_show"`
				}),
			}

			for _, raw := range bookings {
				var b map[string]any
				if err := json.Unmarshal(raw, &b); err != nil {
					continue
				}

				isNoShow := getString(b, "status") == "no_show"
				if !isNoShow {
					if noShowHost, ok := b["noShowHost"].(bool); ok && noShowHost {
						isNoShow = true
					}
				}

				stats.TotalBookings++
				if isNoShow {
					stats.TotalNoShows++
				}

				eventType := "unknown"
				if et, ok := b["eventType"].(map[string]any); ok {
					if title := getString(et, "title"); title != "" {
						eventType = title
					}
				}

				entry := stats.ByEventType[eventType]
				entry.Total++
				if isNoShow {
					entry.NoShow++
				}
				stats.ByEventType[eventType] = entry

				if start, err := time.Parse(time.RFC3339, getString(b, "start")); err == nil {
					day := start.Weekday().String()
					dayEntry := stats.ByDay[day]
					dayEntry.Total++
					if isNoShow {
						dayEntry.NoShow++
					}
					stats.ByDay[day] = dayEntry

					hour := start.Hour()
					hourEntry := stats.ByHour[hour]
					hourEntry.Total++
					if isNoShow {
						hourEntry.NoShow++
					}
					stats.ByHour[hour] = hourEntry
				}
			}

			if stats.TotalBookings > 0 {
				stats.OverallRate = float64(stats.TotalNoShows) / float64(stats.TotalBookings) * 100
			}

			if flags.asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(stats)
			}

			fmt.Println("No-Show Analysis")
			fmt.Println("================")
			fmt.Printf("Total: %d bookings, %d no-shows (%.1f%%)\n\n", stats.TotalBookings, stats.TotalNoShows, stats.OverallRate)

			if len(stats.ByEventType) > 0 {
				fmt.Println("By Event Type:")
				type etRate struct {
					Name string
					Rate float64
					N    int
				}
				var rates []etRate
				for name, s := range stats.ByEventType {
					rate := 0.0
					if s.Total > 0 {
						rate = float64(s.NoShow) / float64(s.Total) * 100
					}
					rates = append(rates, etRate{name, rate, s.Total})
				}
				sort.Slice(rates, func(i, j int) bool { return rates[i].Rate > rates[j].Rate })
				for _, r := range rates {
					fmt.Printf("  %-30s %.1f%% (%d bookings)\n", r.Name, r.Rate, r.N)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")

	return cmd
}
