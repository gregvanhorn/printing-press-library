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

func newStaleCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var days int

	cmd := &cobra.Command{
		Use:   "stale",
		Short: "Find event types with no recent bookings",
		Long: `Identify event types that haven't received a booking in N days.
Helps clean up unused scheduling pages.`,
		Example: `  # Event types with no bookings in 30 days
  cal-com-pp-cli stale --days 30

  # JSON output
  cal-com-pp-cli stale --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("cal-com-pp-cli")
			}

			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening local database: %w\nRun 'cal-com-pp-cli sync' first.", err)
			}
			defer db.Close()

			eventTypes, err := db.QueryJSON("event-types", "1=1")
			if err != nil {
				return fmt.Errorf("querying event types: %w", err)
			}

			bookings, err := db.QueryJSON("bookings", "1=1")
			if err != nil {
				return fmt.Errorf("querying bookings: %w", err)
			}

			// Build map of event type ID -> last booking date
			lastBooking := make(map[string]time.Time)
			bookingCount := make(map[string]int)
			for _, raw := range bookings {
				var b map[string]any
				if err := json.Unmarshal(raw, &b); err != nil {
					continue
				}
				etID := ""
				if et, ok := b["eventType"].(map[string]any); ok {
					if id, ok := et["id"].(float64); ok {
						etID = fmt.Sprintf("%.0f", id)
					}
				}
				if etID == "" {
					if id, ok := b["eventTypeId"].(float64); ok {
						etID = fmt.Sprintf("%.0f", id)
					}
				}
				if etID == "" {
					continue
				}
				bookingCount[etID]++
				if start, err := time.Parse(time.RFC3339, getString(b, "start")); err == nil {
					if existing, ok := lastBooking[etID]; !ok || start.After(existing) {
						lastBooking[etID] = start
					}
				}
			}

			cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

			type staleET struct {
				ID         string `json:"id"`
				Title      string `json:"title"`
				Slug       string `json:"slug"`
				LastBooked string `json:"last_booked,omitempty"`
				DaysSince  int    `json:"days_since_last_booking"`
				Total      int    `json:"total_bookings"`
			}

			var staleList []staleET
			for _, raw := range eventTypes {
				var et map[string]any
				if err := json.Unmarshal(raw, &et); err != nil {
					continue
				}
				id := ""
				if idNum, ok := et["id"].(float64); ok {
					id = fmt.Sprintf("%.0f", idNum)
				} else if idStr, ok := et["id"].(string); ok {
					id = idStr
				}

				last, hasBooking := lastBooking[id]
				if hasBooking && last.After(cutoff) {
					continue // Not stale
				}

				entry := staleET{
					ID:    id,
					Title: getString(et, "title"),
					Slug:  getString(et, "slug"),
					Total: bookingCount[id],
				}
				if hasBooking {
					entry.LastBooked = last.Format("2006-01-02")
					entry.DaysSince = int(time.Since(last).Hours() / 24)
				} else {
					entry.DaysSince = -1 // Never booked
				}

				staleList = append(staleList, entry)
			}

			sort.Slice(staleList, func(i, j int) bool {
				return staleList[i].DaysSince > staleList[j].DaysSince
			})

			if flags.asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(staleList)
			}

			if len(staleList) == 0 {
				fmt.Printf("All event types have been booked in the last %d days\n", days)
				return nil
			}

			fmt.Printf("Stale Event Types (no bookings in %d+ days):\n\n", days)
			for _, s := range staleList {
				if s.DaysSince == -1 {
					fmt.Printf("  %-30s  NEVER BOOKED  (slug: %s)\n", s.Title, s.Slug)
				} else {
					fmt.Printf("  %-30s  %d days ago  (%d total, slug: %s)\n", s.Title, s.DaysSince, s.Total, s.Slug)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().IntVar(&days, "days", 30, "Number of days without bookings to consider stale")

	return cmd
}
