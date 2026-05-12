// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

const pathRecall = "/device/recall.json"

func newRecallCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "recall",
		Short:       "Device recalls — search, list, get",
		Long:        "Query openFDA device recall events (/device/recall.json).",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(newRecallSearchCmd(flags), newRecallGetCmd(flags), newRecallListCmd(flags))
	return cmd
}

func newRecallSearchCmd(flags *rootFlags) *cobra.Command {
	var q openfda.Query
	var all bool
	cmd := &cobra.Command{
		Use:         "search",
		Short:       "Raw search against /device/recall.json",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpenFDAList(cmd, flags, pathRecall, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	return cmd
}

func newRecallGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "get <recall-number>",
		Short:       "Fetch one recall by recall number",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runOpenFDAGet(cmd, flags, pathRecall, "recall_number", args[0])
		},
	}
}

func newRecallListCmd(flags *rootFlags) *cobra.Command {
	var (
		classFlag   string
		productCode string
		since       string
		q           openfda.Query
		all         bool
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List device recalls with curated filters",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			clauses := []string{}
			if classFlag != "" {
				cls := strings.ToUpper(strings.TrimSpace(classFlag))
				clauses = append(clauses, fmt.Sprintf("classification:%q", "Class "+cls))
			}
			if productCode != "" {
				clauses = append(clauses, fmt.Sprintf("product_code:%q", productCode))
			}
			if since != "" {
				ymd, err := openfda.SinceToYYYYMMDD(since)
				if err != nil {
					return usageErr(err)
				}
				clauses = append(clauses, fmt.Sprintf("event_date_initiated:[%s TO 99991231]", ymd))
			}
			if q.Search == "" {
				q.Search = openfda.BuildLuceneAND(clauses...)
			} else if len(clauses) > 0 {
				q.Search = openfda.BuildLuceneAND(append([]string{q.Search}, clauses...)...)
			}
			if q.Sort == "" {
				q.Sort = "event_date_initiated:desc"
			}
			return runOpenFDAList(cmd, flags, pathRecall, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	cmd.Flags().StringVar(&classFlag, "class", "", "Recall classification: I, II, or III")
	cmd.Flags().StringVar(&productCode, "product-code", "", "FDA product code")
	cmd.Flags().StringVar(&since, "since", "", "Recalls since e.g. 2y, 6m, 30d")
	return cmd
}
