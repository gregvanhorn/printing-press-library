// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.
// PATCH: novel transcendence command (cross-campaign aggregation, hand-built on top of generator output)

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newReachBudgetCmd(flags *rootFlags) *cobra.Command {
	var campaignID string
	var dbPath string

	cmd := &cobra.Command{
		Use:   "reach-budget",
		Short: "Days-to-completion from sending limits and pending leads",
		Long: `Estimate days-to-completion for one or all campaigns by dividing
pending leads (status != COMPLETED) by the daily aggregate sending
limit of associated email accounts. Reads from the local SQLite store;
run 'smartlead-pp-cli sync --full' first.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openLocalStore(flags, dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			where := ""
			args2 := []any{}
			if campaignID != "" {
				where = "WHERE campaigns_id = ?"
				args2 = append(args2, campaignID)
			}

			rows, err := db.DB().Query(fmt.Sprintf(`
				SELECT campaigns_id, data FROM campaigns_leads %s LIMIT 200000
			`, where), args2...)
			if err != nil {
				return fmt.Errorf("query: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
			}
			defer rows.Close()

			type stat struct {
				Pending   int     `json:"pending"`
				Completed int     `json:"completed"`
				DailyCap  int     `json:"daily_cap"`
				DaysLeft  float64 `json:"days_to_completion"`
			}
			byCamp := map[string]*stat{}
			for rows.Next() {
				var campID, raw string
				if err := rows.Scan(&campID, &raw); err != nil {
					continue
				}
				if byCamp[campID] == nil {
					byCamp[campID] = &stat{}
				}
				var obj map[string]any
				_ = json.Unmarshal([]byte(raw), &obj)
				status, _ := obj["status"].(string)
				if status == "COMPLETED" || status == "completed" {
					byCamp[campID].Completed++
				} else {
					byCamp[campID].Pending++
				}
			}

			// Daily cap: aggregate message_per_day across linked email accounts.
			// Fallback: campaign-level message_per_day or a constant.
			eaRows, err := db.DB().Query(`SELECT campaigns_id, COALESCE(SUM(json_extract(data,'$.message_per_day')),0) FROM campaigns_email_accounts GROUP BY campaigns_id`)
			if err == nil {
				defer eaRows.Close()
				for eaRows.Next() {
					var cid string
					var cap int
					if eaRows.Scan(&cid, &cap) != nil {
						continue
					}
					if byCamp[cid] != nil {
						byCamp[cid].DailyCap = cap
					}
				}
			}

			type row struct {
				CampaignID string `json:"campaign_id"`
				stat
			}
			var out []row
			for id, s := range byCamp {
				cap := s.DailyCap
				if cap <= 0 {
					cap = 50 // sane default
				}
				if s.Pending > 0 {
					s.DaysLeft = float64(s.Pending) / float64(cap)
				}
				out = append(out, row{CampaignID: id, stat: *s})
			}

			result := map[string]any{"count": len(out), "campaigns": out}
			empty := "No campaign lead data found. (run 'smartlead-pp-cli sync --full' first)"
			return emitTranscendOutput(cmd, flags, result, empty)
		},
	}
	cmd.Flags().StringVar(&campaignID, "campaign", "", "Restrict to a single campaign id")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	return cmd
}
