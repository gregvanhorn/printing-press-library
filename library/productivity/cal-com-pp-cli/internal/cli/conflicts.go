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

func newConflictsCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var days int

	cmd := &cobra.Command{
		Use:   "conflicts",
		Short: "Detect double-bookings and schedule overlaps",
		Long: `Scan synced bookings for time overlaps — double-bookings where two or more
accepted bookings share the same time window. Shows which bookings conflict
and by how many minutes.`,
		Example: `  # Check for conflicts in the next 7 days
  cal-com-pp-cli conflicts

  # Check next 30 days
  cal-com-pp-cli conflicts --days 30

  # JSON output
  cal-com-pp-cli conflicts --json`,
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

			type booking struct {
				UID   string    `json:"uid"`
				Title string    `json:"title"`
				Start time.Time `json:"start"`
				End   time.Time `json:"end"`
			}

			var parsed []booking
			for _, raw := range bookings {
				var b map[string]any
				if err := json.Unmarshal(raw, &b); err != nil {
					continue
				}
				start, err1 := time.Parse(time.RFC3339, getString(b, "start"))
				end, err2 := time.Parse(time.RFC3339, getString(b, "end"))
				if err1 != nil || err2 != nil {
					continue
				}
				parsed = append(parsed, booking{
					UID:   getString(b, "uid"),
					Title: getString(b, "title"),
					Start: start,
					End:   end,
				})
			}

			sort.Slice(parsed, func(i, j int) bool {
				return parsed[i].Start.Before(parsed[j].Start)
			})

			type conflict struct {
				BookingA       string `json:"booking_a"`
				TitleA         string `json:"title_a"`
				BookingB       string `json:"booking_b"`
				TitleB         string `json:"title_b"`
				OverlapMinutes int    `json:"overlap_minutes"`
				Date           string `json:"date"`
			}

			var conflicts []conflict
			for i := 0; i < len(parsed); i++ {
				for j := i + 1; j < len(parsed); j++ {
					if parsed[j].Start.After(parsed[i].End) || parsed[j].Start.Equal(parsed[i].End) {
						break
					}
					overlapEnd := parsed[i].End
					if parsed[j].End.Before(overlapEnd) {
						overlapEnd = parsed[j].End
					}
					overlapStart := parsed[j].Start
					overlap := int(overlapEnd.Sub(overlapStart).Minutes())
					if overlap > 0 {
						conflicts = append(conflicts, conflict{
							BookingA:       parsed[i].UID,
							TitleA:         parsed[i].Title,
							BookingB:       parsed[j].UID,
							TitleB:         parsed[j].Title,
							OverlapMinutes: overlap,
							Date:           parsed[i].Start.Local().Format("2006-01-02"),
						})
					}
				}
			}

			if flags.asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{
					"conflicts": conflicts,
					"total":     len(conflicts),
					"scanned":   len(parsed),
					"days":      days,
				})
			}

			if len(conflicts) == 0 {
				fmt.Printf("No conflicts found in the next %d days (%d bookings scanned)\n", days, len(parsed))
				return nil
			}

			fmt.Printf("Found %d conflict(s) in the next %d days:\n\n", len(conflicts), days)
			for _, c := range conflicts {
				fmt.Printf("  CONFLICT on %s (%d min overlap)\n", c.Date, c.OverlapMinutes)
				fmt.Printf("    A: %s (%s)\n", c.TitleA, c.BookingA[:8])
				fmt.Printf("    B: %s (%s)\n", c.TitleB, c.BookingB[:8])
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days ahead to scan for conflicts")

	return cmd
}
