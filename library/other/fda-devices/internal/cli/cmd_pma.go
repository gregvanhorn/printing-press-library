// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

const pathPMA = "/device/pma.json"

func newPMACmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "pma",
		Short:       "PMA approvals — search, list, get",
		Long:        "Query openFDA Premarket Approval (PMA) records (/device/pma.json).",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(newPMASearchCmd(flags), newPMAGetCmd(flags), newPMAListCmd(flags))
	return cmd
}

func newPMASearchCmd(flags *rootFlags) *cobra.Command {
	var q openfda.Query
	var all bool
	cmd := &cobra.Command{
		Use:         "search",
		Short:       "Raw search against /device/pma.json",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpenFDAList(cmd, flags, pathPMA, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	return cmd
}

func newPMAGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "get <pma-number>",
		Short:       "Fetch one PMA by PMA number",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runOpenFDAGet(cmd, flags, pathPMA, "pma_number", args[0])
		},
	}
}

func newPMAListCmd(flags *rootFlags) *cobra.Command {
	var (
		productCode string
		applicant   string
		since       string
		q           openfda.Query
		all         bool
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List PMA approvals with curated filters",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			clauses := []string{}
			if productCode != "" {
				clauses = append(clauses, fmt.Sprintf("product_code:%q", productCode))
			}
			if applicant != "" {
				clauses = append(clauses, fmt.Sprintf("applicant:%q", applicant))
			}
			if since != "" {
				ymd, err := openfda.SinceToYYYYMMDD(since)
				if err != nil {
					return usageErr(err)
				}
				clauses = append(clauses, fmt.Sprintf("decision_date:[%s TO 99991231]", ymd))
			}
			if q.Search == "" {
				q.Search = openfda.BuildLuceneAND(clauses...)
			} else if len(clauses) > 0 {
				q.Search = openfda.BuildLuceneAND(append([]string{q.Search}, clauses...)...)
			}
			if q.Sort == "" {
				q.Sort = "decision_date:desc"
			}
			return runOpenFDAList(cmd, flags, pathPMA, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	cmd.Flags().StringVar(&productCode, "product-code", "", "FDA product code")
	cmd.Flags().StringVar(&applicant, "applicant", "", "Applicant company name")
	cmd.Flags().StringVar(&since, "since", "", "Approvals since e.g. 2y, 6m, 30d")
	return cmd
}
