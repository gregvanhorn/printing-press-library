// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newDelivDriftCmd(flags *rootFlags) *cobra.Command {
	var domain string
	var dbPath string

	cmd := &cobra.Command{
		Use:   "deliv-drift",
		Short: "Compare today vs 7d vs 30d spam-test placement per sender domain",
		Long: `Compare inbox/spam placement for a sender domain across three
windows (today, last 7d, last 30d) using locally-synced spam_test
results. Reads from the local SQLite store; run
'smartlead-pp-cli sync --full' first.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openLocalStore(flags, dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			rows, err := db.DB().Query(`SELECT data FROM spam_test LIMIT 5000`)
			if err != nil {
				return fmt.Errorf("query spam_test: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
			}
			defer rows.Close()

			type bucket struct {
				Inbox int `json:"inbox"`
				Spam  int `json:"spam"`
				Other int `json:"other"`
				N     int `json:"n"`
			}
			today, last7, last30 := bucket{}, bucket{}, bucket{}
			now := time.Now()
			cutoff1 := now.AddDate(0, 0, -1)
			cutoff7 := now.AddDate(0, 0, -7)
			cutoff30 := now.AddDate(0, 0, -30)

			matched := 0
			for rows.Next() {
				var raw string
				if rows.Scan(&raw) != nil {
					continue
				}
				var obj map[string]any
				if json.Unmarshal([]byte(raw), &obj) != nil {
					continue
				}
				if domain != "" {
					fromEmail, _ := obj["from_email"].(string)
					sender, _ := obj["sender_domain"].(string)
					if !strings.Contains(strings.ToLower(fromEmail), strings.ToLower(domain)) &&
						!strings.Contains(strings.ToLower(sender), strings.ToLower(domain)) {
						continue
					}
				}
				matched++
				createdStr, _ := obj["created_at"].(string)
				if createdStr == "" {
					createdStr, _ = obj["test_started_at"].(string)
				}
				ts, _ := time.Parse(time.RFC3339, createdStr)

				placement := strings.ToLower(fmt.Sprintf("%v", obj["placement"]))
				inbox, _ := obj["inbox"].(float64)
				spam, _ := obj["spam"].(float64)

				add := func(b *bucket) {
					b.N++
					if placement == "inbox" || inbox > 0 {
						b.Inbox += int(max1(inbox, 1))
					} else if placement == "spam" || spam > 0 {
						b.Spam += int(max1(spam, 1))
					} else {
						b.Other++
					}
				}
				if !ts.IsZero() {
					if ts.After(cutoff1) {
						add(&today)
					}
					if ts.After(cutoff7) {
						add(&last7)
					}
					if ts.After(cutoff30) {
						add(&last30)
					}
				} else {
					add(&last30)
				}
			}

			result := map[string]any{
				"domain":  domain,
				"matched": matched,
				"today":   today,
				"last_7d": last7,
				"last_30d": last30,
			}
			empty := "No spam-test data found for that domain. (run 'smartlead-pp-cli sync --full' first)"
			if matched == 0 {
				return emitTranscendOutput(cmd, flags, result, empty)
			}
			return emitTranscendOutput(cmd, flags, result, "")
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "Sender domain filter (substring match)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	return cmd
}

func max1(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
