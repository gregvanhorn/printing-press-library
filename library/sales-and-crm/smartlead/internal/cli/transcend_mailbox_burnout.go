// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.
// PATCH: novel transcendence command (cross-campaign aggregation, hand-built on top of generator output)

package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newMailboxBurnoutCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var top int

	cmd := &cobra.Command{
		Use:   "mailbox-burnout",
		Short: "Rank email accounts by warmup score / sending limit utilization",
		Long: `Rank locally-synced email_accounts by a composite burnout score
(warmup score, sending limits, recent reputation). Reads from the
local SQLite store; run 'smartlead-pp-cli sync --full' first.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openLocalStore(flags, dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			rows, err := db.DB().Query(`SELECT id, from_email, message_per_day, data FROM email_accounts LIMIT 5000`)
			if err != nil {
				return fmt.Errorf("query email_accounts: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
			}
			defer rows.Close()

			type entry struct {
				ID            string  `json:"id"`
				Email         string  `json:"email"`
				MessagePerDay int     `json:"message_per_day"`
				WarmupScore   float64 `json:"warmup_score,omitempty"`
				BurnoutScore  float64 `json:"burnout_score"`
				Reason        string  `json:"reason,omitempty"`
			}
			var out []entry
			for rows.Next() {
				var id, email, raw string
				var mpd int
				if err := rows.Scan(&id, &email, &mpd, &raw); err != nil {
					continue
				}
				var obj map[string]any
				_ = json.Unmarshal([]byte(raw), &obj)

				warmup := 0.0
				if w, ok := obj["warmup_details"].(map[string]any); ok {
					if s, ok := w["warmup_reputation"].(float64); ok {
						warmup = s
					}
				}
				if s, ok := obj["warmup_score"].(float64); ok && warmup == 0 {
					warmup = s
				}

				// Burnout = (100 - warmup) + (mpd > 50 ? extra : 0)
				burnout := 100.0 - warmup
				reason := ""
				if warmup > 0 && warmup < 80 {
					reason = "low_warmup"
				}
				if mpd > 100 {
					burnout += 10
					if reason != "" {
						reason += "+"
					}
					reason += "high_volume"
				}
				out = append(out, entry{ID: id, Email: email, MessagePerDay: mpd, WarmupScore: warmup, BurnoutScore: burnout, Reason: reason})
			}
			sort.Slice(out, func(i, j int) bool { return out[i].BurnoutScore > out[j].BurnoutScore })
			if top > 0 && len(out) > top {
				out = out[:top]
			}

			result := map[string]any{"count": len(out), "mailboxes": out}
			empty := "No email accounts found. (synced data may be empty — run sync --full)"
			return emitTranscendOutput(cmd, flags, result, empty)
		},
	}
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	cmd.Flags().IntVar(&top, "top", 25, "Show top N mailboxes by burnout score")
	return cmd
}
