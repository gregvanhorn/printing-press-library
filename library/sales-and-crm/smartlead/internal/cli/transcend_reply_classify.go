// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.
// PATCH: novel transcendence command (cross-campaign aggregation, hand-built on top of generator output)

package cli

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

var (
	rePositive  = regexp.MustCompile(`(?i)\b(interested|book|call|demo|sounds good|let'?s chat|happy to|works for me|yes\b)`)
	reObjection = regexp.MustCompile(`(?i)\b(not interested|remove me|no thanks|stop|don'?t contact|unsubscribe me)`)
	reOOO       = regexp.MustCompile(`(?i)\b(out of office|on vacation|away from|maternity|paternity|on leave|will be back)`)
	reUnsub     = regexp.MustCompile(`(?i)\b(unsubscribe|opt[- ]?out|remove from list)`)
)

func classifyReply(body string) string {
	if reUnsub.MatchString(body) {
		return "unsub"
	}
	if reOOO.MatchString(body) {
		return "ooo"
	}
	if reObjection.MatchString(body) {
		return "objection"
	}
	if rePositive.MatchString(body) {
		return "positive"
	}
	return "other"
}

func newReplyClassifyCmd(flags *rootFlags) *cobra.Command {
	var since string
	var dbPath string

	cmd := &cobra.Command{
		Use:   "reply-classify",
		Short: "Classify master_inbox replies (positive/objection/OOO/unsub)",
		Long: `Classify recent master_inbox replies into buckets via local
keyword matching. Reads from the local SQLite store; run
'smartlead-pp-cli sync --full' first.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		Example:     "  smartlead-pp-cli reply-classify --since 7d --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			cutoff, err := parseSinceDuration(since)
			if err != nil {
				return err
			}
			db, err := openLocalStore(flags, dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			rows, err := db.DB().Query(`SELECT data FROM master_inbox LIMIT 50000`)
			if err != nil {
				return fmt.Errorf("query master_inbox: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
			}
			defer rows.Close()

			counts := map[string]int{"positive": 0, "objection": 0, "ooo": 0, "unsub": 0, "other": 0}
			type sample struct {
				ID    string `json:"id,omitempty"`
				Class string `json:"class"`
				From  string `json:"from,omitempty"`
				Snip  string `json:"snippet,omitempty"`
			}
			var samples []sample
			scanned := 0

			for rows.Next() {
				var raw string
				if rows.Scan(&raw) != nil {
					continue
				}
				var obj map[string]any
				if json.Unmarshal([]byte(raw), &obj) != nil {
					continue
				}
				ts, _ := obj["received_time"].(string)
				if ts == "" {
					ts, _ = obj["created_at"].(string)
				}
				if ts != "" {
					if t, err := time.Parse(time.RFC3339, ts); err == nil && t.Before(cutoff) {
						continue
					}
				}
				body, _ := obj["body"].(string)
				if body == "" {
					body, _ = obj["snippet"].(string)
				}
				if body == "" {
					body, _ = obj["subject"].(string)
				}
				class := classifyReply(body)
				counts[class]++
				scanned++
				if len(samples) < 25 {
					id, _ := obj["id"].(string)
					from, _ := obj["from"].(string)
					snip := body
					if len(snip) > 120 {
						snip = snip[:120] + "…"
					}
					samples = append(samples, sample{ID: id, Class: class, From: from, Snip: snip})
				}
			}

			result := map[string]any{
				"since":   since,
				"scanned": scanned,
				"counts":  counts,
				"samples": samples,
			}
			empty := "No replies found in master_inbox. (run 'smartlead-pp-cli sync --full' first)"
			if scanned == 0 {
				return emitTranscendOutput(cmd, flags, result, empty)
			}
			return emitTranscendOutput(cmd, flags, result, "")
		},
	}
	cmd.Flags().StringVar(&since, "since", "7d", "Lookback window (e.g. 7d, 24h, 2w)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")
	return cmd
}
