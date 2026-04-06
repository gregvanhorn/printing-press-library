package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mvanhorn/printing-press-library/library/productivity/cal-com-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newStatsCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var period string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Booking analytics: volume, busiest hours, cancellation rates",
		Long: `Analyze booking patterns from synced data. Shows volume over time,
busiest days and hours, cancellation and no-show rates, and average duration.`,
		Example: `  # Overall booking stats
  cal-com-pp-cli stats

  # Stats for the last 30 days
  cal-com-pp-cli stats --period 30d

  # JSON output
  cal-com-pp-cli stats --json`,
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

			if len(bookings) == 0 {
				fmt.Println("No bookings found. Run 'cal-com-pp-cli sync' first.")
				return nil
			}

			type statsResult struct {
				Total          int            `json:"total"`
				Accepted       int            `json:"accepted"`
				Cancelled      int            `json:"cancelled"`
				Pending        int            `json:"pending"`
				CancelRate     float64        `json:"cancel_rate_pct"`
				ByDayOfWeek    map[string]int `json:"by_day_of_week"`
				ByHour         map[int]int    `json:"by_hour"`
				BusiestDay     string         `json:"busiest_day"`
				BusiestHour    int            `json:"busiest_hour"`
				AvgDurationMin float64        `json:"avg_duration_min"`
			}

			result := statsResult{
				ByDayOfWeek: make(map[string]int),
				ByHour:      make(map[int]int),
			}

			var totalDuration float64
			var durationCount int

			for _, raw := range bookings {
				var b map[string]any
				if err := json.Unmarshal(raw, &b); err != nil {
					continue
				}

				result.Total++
				status := getString(b, "status")
				switch status {
				case "accepted":
					result.Accepted++
				case "cancelled":
					result.Cancelled++
				case "pending":
					result.Pending++
				}

				if start, err := time.Parse(time.RFC3339, getString(b, "start")); err == nil {
					day := start.Weekday().String()
					result.ByDayOfWeek[day]++
					result.ByHour[start.Hour()]++

					if endStr := getString(b, "end"); endStr != "" {
						if end, err := time.Parse(time.RFC3339, endStr); err == nil {
							totalDuration += end.Sub(start).Minutes()
							durationCount++
						}
					}
				}
			}

			if result.Total > 0 {
				result.CancelRate = float64(result.Cancelled) / float64(result.Total) * 100
			}
			if durationCount > 0 {
				result.AvgDurationMin = totalDuration / float64(durationCount)
			}

			maxDay := 0
			for day, count := range result.ByDayOfWeek {
				if count > maxDay {
					maxDay = count
					result.BusiestDay = day
				}
			}
			maxHour := 0
			for hour, count := range result.ByHour {
				if count > maxHour {
					maxHour = count
					result.BusiestHour = hour
				}
			}

			if flags.asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			fmt.Println("Booking Statistics")
			fmt.Println("==================")
			fmt.Printf("Total:       %d\n", result.Total)
			fmt.Printf("Accepted:    %d\n", result.Accepted)
			fmt.Printf("Cancelled:   %d (%.1f%%)\n", result.Cancelled, result.CancelRate)
			fmt.Printf("Pending:     %d\n", result.Pending)
			fmt.Printf("Avg Duration: %.0f min\n", result.AvgDurationMin)
			fmt.Printf("Busiest Day:  %s (%d bookings)\n", result.BusiestDay, maxDay)
			fmt.Printf("Busiest Hour: %d:00 (%d bookings)\n", result.BusiestHour, maxHour)

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().StringVar(&period, "period", "", "Analysis period (e.g. 30d, 12w)")

	return cmd
}
