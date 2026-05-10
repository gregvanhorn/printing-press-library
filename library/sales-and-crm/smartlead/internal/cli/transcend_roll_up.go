// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.
// PATCH: novel transcendence command (cross-campaign aggregation, hand-built on top of generator output)

package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mvanhorn/printing-press-library/library/sales-and-crm/smartlead/internal/client"
	"github.com/mvanhorn/printing-press-library/library/sales-and-crm/smartlead/internal/config"
	"github.com/spf13/cobra"
)

func newRollUpCmd(flags *rootFlags) *cobra.Command {
	var keysFile string

	cmd := &cobra.Command{
		Use:   "roll-up",
		Short: "Aggregate stats across multiple API keys",
		Long: `Fan out a basic /campaigns request across each API key listed in
--keys-from (one key per line, comments with #), and aggregate
campaign counts and basic stats per key plus portfolio-wide totals.`,
		Annotations: map[string]string{"mcp:read-only": "true", "pp:novel": "true"},
		Example:     "  smartlead-pp-cli roll-up --keys-from ./keys.txt --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if keysFile == "" {
				return fmt.Errorf("--keys-from <file> is required (one API key per line)")
			}
			f, err := os.Open(keysFile)
			if err != nil {
				return fmt.Errorf("opening keys file: %w", err)
			}
			defer f.Close()

			var keys []string
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				keys = append(keys, line)
			}
			if len(keys) == 0 {
				return fmt.Errorf("no keys found in %s", keysFile)
			}

			type perKey struct {
				KeyPrefix     string `json:"key_prefix"`
				CampaignCount int    `json:"campaign_count"`
				Error         string `json:"error,omitempty"`
			}
			var rows []perKey
			total := 0

			baseCfg, _ := config.Load(flags.configPath)
			for _, k := range keys {
				cfg := *baseCfg
				cfg.SmartleadApiKey = k
				c := client.New(&cfg, flags.timeout, flags.rateLimit)
				timeout := flags.timeout
				if timeout == 0 {
					timeout = 30 * time.Second
				}
				data, err := c.Get("/campaigns/", nil)
				prefix := k
				if len(prefix) > 6 {
					prefix = prefix[:6] + "…"
				}
				if err != nil {
					rows = append(rows, perKey{KeyPrefix: prefix, Error: err.Error()})
					continue
				}
				var arr []json.RawMessage
				if json.Unmarshal(data, &arr) != nil {
					var wrapped struct {
						Data []json.RawMessage `json:"data"`
					}
					_ = json.Unmarshal(data, &wrapped)
					arr = wrapped.Data
				}
				rows = append(rows, perKey{KeyPrefix: prefix, CampaignCount: len(arr)})
				total += len(arr)
			}

			result := map[string]any{
				"keys":           len(keys),
				"per_key":        rows,
				"total_campaigns": total,
			}
			return emitTranscendOutput(cmd, flags, result, "")
		},
	}
	cmd.Flags().StringVar(&keysFile, "keys-from", "", "Path to file containing one API key per line")
	return cmd
}
