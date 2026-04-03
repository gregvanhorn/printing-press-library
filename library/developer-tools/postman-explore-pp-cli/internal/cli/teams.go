// Copyright 2026 trevin-chow. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newTeamsCmd(flags *rootFlags) *cobra.Command {
	var flagLimit int
	var flagSort string
	var flagAll bool

	cmd := &cobra.Command{
		Use:     "teams",
		Aliases: []string{"publishers"},
		Short:   "List publisher teams on the API network",
		Example: `  # List popular teams
  postman-explore-pp-cli teams

  # List recent teams
  postman-explore-pp-cli teams --sort recent

  # Fetch all teams
  postman-explore-pp-cli teams --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			path := "/v1/api/team"
			data, err := paginatedGet(c, path, map[string]string{
				"limit": fmt.Sprintf("%v", flagLimit),
				"sort":  flagSort,
			}, flagAll, "", "", "")
			if err != nil {
				return classifyAPIError(err)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				var items []map[string]any
				if json.Unmarshal(data, &items) == nil && len(items) > 0 {
					if err := printAutoTable(cmd.OutOrStdout(), items); err != nil {
						return err
					}
					if len(items) >= 25 {
						fmt.Fprintf(os.Stderr, "\nShowing %d results. To narrow: add --limit, --json --select, or filter flags.\n", len(items))
					}
					return nil
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().IntVar(&flagLimit, "limit", 10, "Number of results")
	cmd.Flags().StringVar(&flagSort, "sort", "popular", "Sort order: popular, recent")
	cmd.Flags().BoolVar(&flagAll, "all", false, "Fetch all pages")

	_ = strings.Join // ensure import
	return cmd
}
