// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func newReplyVelocityCmd(flags *rootFlags) *cobra.Command {
	var weeks int
	var dbPath string

	cmd := &cobra.Command{
		Use:   "reply-velocity",
		Short: "Time-to-first-reply per campaign with baseline",
		Long: `Compute median time-to-first-reply per campaign over the last
--weeks weeks, with an all-time baseline for comparison. Reads from
the local SQLite store; run 'smartlead-pp-cli sync --full' first.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openLocalStore(flags, dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			cutoff := time.Now().AddDate(0, 0, -7*weeks)

			rows, err := db.DB().Query(`SELECT campaigns_id, data FROM campaigns_leads LIMIT 100000`)
			if err != nil {
				return fmt.Errorf("query: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
			}
			defer rows.Close()

			type bucket struct{ recent, all []float64 }
			byCamp := map[string]*bucket{}

			for rows.Next() {
				var campID, raw string
				if err := rows.Scan(&campID, &raw); err != nil {
					continue
				}
				var obj map[string]any
				if json.Unmarshal([]byte(raw), &obj) != nil {
					continue
				}
				sentStr, _ := obj["sent_time"].(string)
				replyStr, _ := obj["reply_time"].(string)
				if sentStr == "" || replyStr == "" {
					continue
				}
				sent, err1 := time.Parse(time.RFC3339, sentStr)
				reply, err2 := time.Parse(time.RFC3339, replyStr)
				if err1 != nil || err2 != nil {
					continue
				}
				hrs := reply.Sub(sent).Hours()
				if hrs < 0 {
					continue
				}
				if byCamp[campID] == nil {
					byCamp[campID] = &bucket{}
				}
				byCamp[campID].all = append(byCamp[campID].all, hrs)
				if reply.After(cutoff) {
					byCamp[campID].recent = append(byCamp[campID].recent, hrs)
				}
			}

			type row struct {
				CampaignID    string  `json:"campaign_id"`
				RecentMedian  float64 `json:"recent_median_hours"`
				BaselineMedian float64 `json:"baseline_median_hours"`
				RecentN       int     `json:"recent_n"`
				BaselineN     int     `json:"baseline_n"`
			}
			var out []row
			for id, b := range byCamp {
				out = append(out, row{
					CampaignID:     id,
					RecentMedian:   median(b.recent),
					BaselineMedian: median(b.all),
					RecentN:        len(b.recent),
					BaselineN:      len(b.all),
				})
			}
			sort.Slice(out, func(i, j int) bool { return out[i].RecentN > out[j].RecentN })

			result := map[string]any{"weeks": weeks, "count": len(out), "campaigns": out}
			empty := "No reply data found. (sync may not include reply_time fields — run sync --full)"
			return emitTranscendOutput(cmd, flags, result, empty)
		},
	}
	cmd.Flags().IntVar(&weeks, "weeks", 4, "Lookback window in weeks for the recent bucket")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	return cmd
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	sort.Float64s(v)
	n := len(v)
	if n%2 == 1 {
		return v[n/2]
	}
	return (v[n/2-1] + v[n/2]) / 2
}
