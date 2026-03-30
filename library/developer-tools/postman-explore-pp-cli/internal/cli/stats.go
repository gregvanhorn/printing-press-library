// Copyright 2026 trevin-chow. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newStatsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Aliases: []string{"counts"},
		Short:   "Show total entity counts on the API network",
		Example: `  # Show network stats
  postman-explore-pp-cli stats

  # Get JSON output
  postman-explore-pp-cli stats --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			path := "/v1/api/networkentity/count"
			params := map[string]string{
				"flattenAPIVersions": "true",
			}
			data, err := c.Get(path, params)
			if err != nil {
				return classifyAPIError(err)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				var obj map[string]any
				if json.Unmarshal(data, &obj) == nil {
					w := cmd.OutOrStdout()
					fmt.Fprintln(w, "Postman API Network Stats")
					fmt.Fprintln(w, "-------------------------")
					for key, val := range obj {
						switch v := val.(type) {
						case float64:
							fmt.Fprintf(w, "  %-20s %s\n", key, commaFormat(int64(v)))
						default:
							fmt.Fprintf(w, "  %-20s %v\n", key, val)
						}
					}
					return nil
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}

	return cmd
}

// commaFormat inserts commas into large numbers for readability (e.g. 1234567 -> "1,234,567").
func commaFormat(n int64) string {
	if n < 0 {
		return "-" + commaFormat(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
