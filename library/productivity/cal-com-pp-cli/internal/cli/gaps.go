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

func newGapsCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var days int
	var minGapMin int

	cmd := &cobra.Command{
		Use:   "gaps",
		Short: "Find unbooked availability windows in your schedule",
		Long: `Analyze schedule availability vs actual bookings to find time slots that are
available but chronically unbooked. Helps identify underutilized schedule windows.`,
		Example: `  # Show availability gaps for the next 7 days
  cal-com-pp-cli gaps

  # Show gaps of at least 60 minutes
  cal-com-pp-cli gaps --min-gap 60

  # JSON output
  cal-com-pp-cli gaps --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("cal-com-pp-cli")
			}

			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening local database: %w\nRun 'cal-com-pp-cli sync' first.", err)
			}
			defer db.Close()

			now := time.Now()
			end := now.Add(time.Duration(days) * 24 * time.Hour)

			bookings, err := db.QueryJSON("bookings", fmt.Sprintf(
				`json_extract(data, '$.status') = 'accepted' AND json_extract(data, '$.start') >= '%s' AND json_extract(data, '$.start') < '%s'`,
				now.Format(time.RFC3339), end.Format(time.RFC3339)))
			if err != nil {
				return fmt.Errorf("querying bookings: %w", err)
			}

			type timeSlot struct {
				Start time.Time
				End   time.Time
			}

			var booked []timeSlot
			for _, raw := range bookings {
				var b map[string]any
				if err := json.Unmarshal(raw, &b); err != nil {
					continue
				}
				start, err1 := time.Parse(time.RFC3339, getString(b, "start"))
				endT, err2 := time.Parse(time.RFC3339, getString(b, "end"))
				if err1 == nil && err2 == nil {
					booked = append(booked, timeSlot{start, endT})
				}
			}

			sort.Slice(booked, func(i, j int) bool {
				return booked[i].Start.Before(booked[j].Start)
			})

			type gap struct {
				Date        string `json:"date"`
				StartTime   string `json:"start_time"`
				EndTime     string `json:"end_time"`
				DurationMin int    `json:"duration_min"`
			}

			var gaps []gap

			// Find gaps between consecutive bookings on the same day
			for i := 0; i < len(booked)-1; i++ {
				if booked[i].End.YearDay() == booked[i+1].Start.YearDay() &&
					booked[i].End.Year() == booked[i+1].Start.Year() {
					gapDuration := int(booked[i+1].Start.Sub(booked[i].End).Minutes())
					if gapDuration >= minGapMin {
						gaps = append(gaps, gap{
							Date:        booked[i].End.Local().Format("2006-01-02"),
							StartTime:   booked[i].End.Local().Format("3:04 PM"),
							EndTime:     booked[i+1].Start.Local().Format("3:04 PM"),
							DurationMin: gapDuration,
						})
					}
				}
			}

			if flags.asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"gaps":     gaps,
					"total":    len(gaps),
					"days":     days,
					"min_gap":  minGapMin,
					"bookings": len(booked),
				})
			}

			if len(gaps) == 0 {
				fmt.Printf("No gaps >= %d min found between bookings in the next %d days\n", minGapMin, days)
				return nil
			}

			fmt.Printf("Found %d gap(s) >= %d min in the next %d days:\n\n", len(gaps), minGapMin, days)
			for _, g := range gaps {
				fmt.Printf("  %s  %s - %s  (%d min)\n", g.Date, g.StartTime, g.EndTime, g.DurationMin)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days ahead to analyze")
	cmd.Flags().IntVar(&minGapMin, "min-gap", 30, "Minimum gap duration in minutes")

	return cmd
}
