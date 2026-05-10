// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.
// PATCH: novel transcendence command (cross-campaign aggregation, hand-built on top of generator output)

package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newOverlapCmd(flags *rootFlags) *cobra.Command {
	var dbPath string
	var minCampaigns int

	cmd := &cobra.Command{
		Use:   "overlap",
		Short: "Find leads (by email) appearing in multiple campaigns",
		Long: `Surfaces duplicate-outreach risk: leads whose email appears in
two or more campaigns. Reads from the local SQLite store; run
'smartlead-pp-cli sync --full' first.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		Example:     "  smartlead-pp-cli overlap --min 2 --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openLocalStore(flags, dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			rows, err := db.DB().Query(`SELECT campaigns_id, data FROM campaigns_leads LIMIT 100000`)
			if err != nil {
				return fmt.Errorf("query campaigns_leads: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
			}
			defer rows.Close()

			byEmail := make(map[string]map[string]bool)
			for rows.Next() {
				var campID, raw string
				if err := rows.Scan(&campID, &raw); err != nil {
					continue
				}
				var obj map[string]any
				if json.Unmarshal([]byte(raw), &obj) != nil {
					continue
				}
				email, _ := obj["email"].(string)
				if email == "" {
					continue
				}
				if byEmail[email] == nil {
					byEmail[email] = map[string]bool{}
				}
				byEmail[email][campID] = true
			}

			type entry struct {
				Email     string   `json:"email"`
				Campaigns []string `json:"campaign_ids"`
				Count     int      `json:"count"`
			}
			var out []entry
			for email, set := range byEmail {
				if len(set) < minCampaigns {
					continue
				}
				ids := make([]string, 0, len(set))
				for k := range set {
					ids = append(ids, k)
				}
				sort.Strings(ids)
				out = append(out, entry{Email: email, Campaigns: ids, Count: len(ids)})
			}
			sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })

			result := map[string]any{
				"min_campaigns": minCampaigns,
				"count":         len(out),
				"overlaps":      out,
			}
			empty := "No overlapping leads found. (synced data may be empty — run sync --full)"
			return emitTranscendOutput(cmd, flags, result, empty)
		},
	}

	cmd.Flags().IntVar(&minCampaigns, "min", 2, "Minimum number of campaigns a lead must appear in")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	return cmd
}
