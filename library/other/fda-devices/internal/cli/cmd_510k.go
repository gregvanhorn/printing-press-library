// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

const path510k = "/device/510k.json"

func new510kCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "510k",
		Short:       "510(k) clearances — search, list, get",
		Long:        "Query openFDA 510(k) premarket notification clearances (/device/510k.json).",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(new510kSearchCmd(flags), new510kGetCmd(flags), new510kListCmd(flags))
	return cmd
}

func new510kSearchCmd(flags *rootFlags) *cobra.Command {
	var q openfda.Query
	var all bool
	cmd := &cobra.Command{
		Use:         "search",
		Short:       "Raw search against /device/510k.json",
		Annotations: readOnlyAnnotations,
		Example: `  # Raw Lucene search for clearances mentioning "da Vinci"
  fda-devices-pp-cli 510k search --search 'device_name:"da Vinci"'

  # Fetch every page (paginates via skip up to 25k records)
  fda-devices-pp-cli 510k search --search 'applicant:"Intuitive Surgical"' --all --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpenFDAList(cmd, flags, path510k, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	return cmd
}

func new510kGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "get <k-number>",
		Short:       "Fetch one 510(k) clearance by K-number",
		Annotations: readOnlyAnnotations,
		Example: `  # Fetch a specific 510(k) clearance by K-number
  fda-devices-pp-cli 510k get K990144

  # Get just the headline fields as JSON
  fda-devices-pp-cli 510k get K171651 --agent --select k_number,applicant,decision_date,device_name`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runOpenFDAGet(cmd, flags, path510k, "k_number", args[0])
		},
	}
}

func new510kListCmd(flags *rootFlags) *cobra.Command {
	var (
		productCode  string
		applicant    string
		since        string
		decisionDate string
		q            openfda.Query
		all          bool
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List 510(k) clearances with curated filters",
		Annotations: readOnlyAnnotations,
		Example: `  # Recent clearances in a product code
  fda-devices-pp-cli 510k list --product-code DQA --since 2y

  # All clearances for an applicant, JSON output
  fda-devices-pp-cli 510k list --applicant "Intuitive Surgical" --agent

  # Custom decision-date range
  fda-devices-pp-cli 510k list --decision-date '[20240101 TO 20241231]' --limit 100`,
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
			if decisionDate != "" {
				clauses = append(clauses, "decision_date:"+decisionDate)
			}
			if q.Search == "" {
				q.Search = openfda.BuildLuceneAND(clauses...)
			} else if len(clauses) > 0 {
				q.Search = openfda.BuildLuceneAND(append([]string{q.Search}, clauses...)...)
			}
			if q.Sort == "" {
				q.Sort = "decision_date:desc"
			}
			return runOpenFDAList(cmd, flags, path510k, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	cmd.Flags().StringVar(&productCode, "product-code", "", "3-letter FDA product code (e.g. DQA)")
	cmd.Flags().StringVar(&applicant, "applicant", "", "Applicant company name (exact-ish match)")
	cmd.Flags().StringVar(&since, "since", "", "Clearances since e.g. 2y, 6m, 30d, 12w")
	cmd.Flags().StringVar(&decisionDate, "decision-date", "", "Raw Lucene range expression e.g. '[20200101 TO 20231231]'")
	return cmd
}
