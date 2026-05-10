// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.
// PATCH: novel transcendence command (cross-campaign aggregation, hand-built on top of generator output)

package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newStaleCmd(flags *rootFlags) *cobra.Command {
	var days int
	var dbPath string

	cmd := &cobra.Command{
		Use:   "stale",
		Short: "Find leads not touched in N days across every campaign",
		Long: `Cross-campaign aggregation that surfaces leads whose last_event_at
is older than --days days. Reads from the local SQLite store; run
'smartlead-pp-cli sync --full' first.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		Example:     "  smartlead-pp-cli stale --days 30",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openLocalStore(flags, dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			cutoff := time.Now().AddDate(0, 0, -days)
			cutoffISO := cutoff.Format(time.RFC3339)

			rows, err := db.DB().Query(`
				SELECT id, campaigns_id, data
				FROM campaigns_leads
				WHERE COALESCE(
					json_extract(data,'$.last_event_at'),
					json_extract(data,'$.last_email_sent_time'),
					json_extract(data,'$.updated_at'),
					json_extract(data,'$.created_at'),
					''
				) < ?
				LIMIT 5000
			`, cutoffISO)
			if err != nil {
				return fmt.Errorf("query stale leads: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
			}
			defer rows.Close()

			type staleRow struct {
				LeadID     string `json:"lead_id"`
				CampaignID string `json:"campaign_id"`
				Email      string `json:"email,omitempty"`
				LastEvent  string `json:"last_event_at,omitempty"`
			}
			var out []staleRow
			for rows.Next() {
				var id, campID, raw string
				if err := rows.Scan(&id, &campID, &raw); err != nil {
					continue
				}
				var obj map[string]any
				_ = json.Unmarshal([]byte(raw), &obj)
				email, _ := obj["email"].(string)
				le, _ := obj["last_event_at"].(string)
				if le == "" {
					if v, ok := obj["last_email_sent_time"].(string); ok {
						le = v
					}
				}
				out = append(out, staleRow{LeadID: id, CampaignID: campID, Email: email, LastEvent: le})
			}

			result := map[string]any{
				"days":   days,
				"cutoff": cutoffISO,
				"count":  len(out),
				"leads":  out,
			}
			empty := fmt.Sprintf("No stale leads older than %d days. (synced data may be empty — run sync --full)", days)
			return emitTranscendOutput(cmd, flags, result, empty)
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Stale threshold in days")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	return cmd
}
