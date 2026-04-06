package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/productivity/cal-com-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newTodayCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var dateStr string

	cmd := &cobra.Command{
		Use:   "today",
		Short: "Show today's schedule with attendee details and conferencing links",
		Long: `Display all bookings for today (or a specified date) with attendee names,
event type, conferencing links, and time until next meeting. Requires synced data.`,
		Example: `  # Show today's schedule
  cal-com-pp-cli today

  # Show schedule for a specific date
  cal-com-pp-cli today --date 2026-04-10

  # JSON output for agents
  cal-com-pp-cli today --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("cal-com-pp-cli")
			}

			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening local database: %w\nRun 'cal-com-pp-cli sync' first.", err)
			}
			defer db.Close()

			targetDate := time.Now()
			if dateStr != "" {
				t, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					return fmt.Errorf("invalid date %q (expected YYYY-MM-DD): %w", dateStr, err)
				}
				targetDate = t
			}

			dayStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
			dayEnd := dayStart.Add(24 * time.Hour)

			bookings, err := db.QueryJSON("bookings", fmt.Sprintf(
				`json_extract(data, '$.start') >= '%s' AND json_extract(data, '$.start') < '%s'`,
				dayStart.Format(time.RFC3339), dayEnd.Format(time.RFC3339)))
			if err != nil {
				return fmt.Errorf("querying bookings: %w", err)
			}

			type todayEntry struct {
				Start       string   `json:"start"`
				End         string   `json:"end"`
				Title       string   `json:"title"`
				Status      string   `json:"status"`
				Attendees   []string `json:"attendees"`
				MeetingURL  string   `json:"meeting_url,omitempty"`
				EventType   string   `json:"event_type,omitempty"`
				UID         string   `json:"uid"`
				MinutesAway int      `json:"minutes_away,omitempty"`
			}

			var entries []todayEntry
			now := time.Now()

			for _, raw := range bookings {
				var b map[string]any
				if err := json.Unmarshal(raw, &b); err != nil {
					continue
				}

				entry := todayEntry{
					Title:  getString(b, "title"),
					Status: getString(b, "status"),
					UID:    getString(b, "uid"),
					Start:  getString(b, "start"),
					End:    getString(b, "end"),
				}

				if start, err := time.Parse(time.RFC3339, entry.Start); err == nil {
					if start.After(now) {
						entry.MinutesAway = int(time.Until(start).Minutes())
					}
				}

				if attendees, ok := b["attendees"].([]any); ok {
					for _, a := range attendees {
						if am, ok := a.(map[string]any); ok {
							name := getString(am, "name")
							email := getString(am, "email")
							if name != "" {
								entry.Attendees = append(entry.Attendees, name)
							} else if email != "" {
								entry.Attendees = append(entry.Attendees, email)
							}
						}
					}
				}

				if meetingURL := getString(b, "meetingUrl"); meetingURL != "" {
					entry.MeetingURL = meetingURL
				}

				if et, ok := b["eventType"].(map[string]any); ok {
					entry.EventType = getString(et, "title")
				}

				entries = append(entries, entry)
			}

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Start < entries[j].Start
			})

			if flags.asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(entries)
			}

			if len(entries) == 0 {
				fmt.Printf("No bookings for %s\n", dayStart.Format("Mon, Jan 2 2006"))
				return nil
			}

			fmt.Printf("Schedule for %s (%d bookings)\n\n", dayStart.Format("Mon, Jan 2 2006"), len(entries))

			for _, e := range entries {
				startTime := ""
				endTime := ""
				if t, err := time.Parse(time.RFC3339, e.Start); err == nil {
					startTime = t.Local().Format("3:04 PM")
				}
				if t, err := time.Parse(time.RFC3339, e.End); err == nil {
					endTime = t.Local().Format("3:04 PM")
				}

				statusIcon := "  "
				switch e.Status {
				case "accepted":
					statusIcon = "OK"
				case "pending":
					statusIcon = "??"
				case "cancelled":
					statusIcon = "XX"
				}

				fmt.Printf("[%s] %s - %s  %s\n", statusIcon, startTime, endTime, e.Title)
				if len(e.Attendees) > 0 {
					fmt.Printf("     Attendees: %s\n", strings.Join(e.Attendees, ", "))
				}
				if e.MeetingURL != "" {
					fmt.Printf("     Link: %s\n", e.MeetingURL)
				}
				if e.MinutesAway > 0 {
					if e.MinutesAway < 60 {
						fmt.Printf("     Starts in %d min\n", e.MinutesAway)
					} else {
						fmt.Printf("     Starts in %dh %dm\n", e.MinutesAway/60, e.MinutesAway%60)
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().StringVar(&dateStr, "date", "", "Date to show (YYYY-MM-DD, default: today)")

	return cmd
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
