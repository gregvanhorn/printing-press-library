// Copyright 2026 trevin-chow. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCategoriesCmd(flags *rootFlags) *cobra.Command {
	var flagSort string

	cmd := &cobra.Command{
		Use:     "categories",
		Aliases: []string{"cats"},
		Short:   "List all API categories on the network",
		Example: `  # List categories
  postman-explore-pp-cli categories

  # Get JSON output
  postman-explore-pp-cli categories --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			path := "/v2/api/category"
			params := map[string]string{}
			if flagSort != "" {
				params["sort"] = flagSort
			}
			data, err := c.Get(path, params)
			if err != nil {
				return classifyAPIError(err)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				items := extractEntityItems(data)
				if len(items) > 0 {
					headers := []string{"ID", "SLUG", "NAME"}
					rows := make([][]string, 0, len(items))
					for _, item := range items {
						rows = append(rows, []string{
							numField(item, "id"),
							strField(item, "slug"),
							strField(item, "name"),
						})
					}
					tw := newTabWriter(cmd.OutOrStdout())
					fmt.Fprintln(tw, strings.Join(headers, "\t"))
					for _, row := range rows {
						fmt.Fprintln(tw, strings.Join(row, "\t"))
					}
					return tw.Flush()
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().StringVar(&flagSort, "sort", "spotlighted", "Sort order")

	return cmd
}
