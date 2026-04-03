// Copyright 2026 trevin-chow. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newTrendingCmd(flags *rootFlags) *cobra.Command {
	var flagPeriod string
	var flagLimit int
	var flagAll bool

	cmd := &cobra.Command{
		Use:   "trending",
		Short: "Show trending collections by fork count",
		Long:  "Browse collections sorted by fork activity. Use --period to choose weekly or monthly trending.",
		Example: `  # Weekly trending (default)
  postman-explore-pp-cli trending

  # Monthly trending
  postman-explore-pp-cli trending --period month

  # Top 50 trending
  postman-explore-pp-cli trending --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			sortField := "weekForkCount"
			if flagPeriod == "month" {
				sortField = "monthForkCount"
			}

			path := "/v1/api/networkentity"
			data, err := paginatedGet(c, path, map[string]string{
				"entityType": "collection",
				"limit":      fmt.Sprintf("%v", flagLimit),
				"sort":       sortField,
			}, flagAll, "offset", "", "")
			if err != nil {
				return classifyAPIError(err)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				items := extractEntityItems(data)
				if len(items) > 0 {
					headers := []string{"NAME", "PUBLISHER", "FORKS", "VIEWS", "CATEGORY"}
					rows := make([][]string, 0, len(items))
					for _, item := range items {
						rows = append(rows, []string{
							strField(item, "name"),
							strField(item, "publisherName"),
							numField(item, "forkCount"),
							numField(item, "viewCount"),
							strField(item, "category"),
						})
					}
					tw := newTabWriter(cmd.OutOrStdout())
					fmt.Fprintln(tw, strings.Join(headers, "\t"))
					for _, row := range rows {
						fmt.Fprintln(tw, strings.Join(row, "\t"))
					}
					tw.Flush()
					if len(items) >= 25 {
						fmt.Fprintf(os.Stderr, "\nShowing %d results. To narrow: add --limit, --json --select, or filter flags.\n", len(items))
					}
					return nil
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().StringVar(&flagPeriod, "period", "week", "Trending period: week, month")
	cmd.Flags().IntVar(&flagLimit, "limit", 20, "Number of results")
	cmd.Flags().BoolVar(&flagAll, "all", false, "Fetch all pages")

	return cmd
}
