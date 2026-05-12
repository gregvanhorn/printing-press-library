// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

const pathMaude = "/device/event.json"

func newMaudeCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "maude",
		Short:       "MAUDE adverse events — search, list, get",
		Long:        "Query openFDA MAUDE (Manufacturer and User Facility Device Experience) reports (/device/event.json).",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(newMaudeSearchCmd(flags), newMaudeGetCmd(flags), newMaudeListCmd(flags))
	return cmd
}

func newMaudeSearchCmd(flags *rootFlags) *cobra.Command {
	var q openfda.Query
	var all bool
	cmd := &cobra.Command{
		Use:         "search",
		Short:       "Raw search against /device/event.json",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpenFDAList(cmd, flags, pathMaude, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	return cmd
}

func newMaudeGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "get <report-number>",
		Short:       "Fetch one MAUDE report by report number",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runOpenFDAGet(cmd, flags, pathMaude, "report_number", args[0])
		},
	}
}

func newMaudeListCmd(flags *rootFlags) *cobra.Command {
	var (
		productCode string
		eventType   string
		since       string
		q           openfda.Query
		all         bool
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List MAUDE events with curated filters",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			clauses := []string{}
			if productCode != "" {
				clauses = append(clauses, fmt.Sprintf("device.device_report_product_code:%q", productCode))
			}
			if eventType != "" {
				clauses = append(clauses, fmt.Sprintf("event_type:%q", eventType))
			}
			if since != "" {
				ymd, err := openfda.SinceToYYYYMMDD(since)
				if err != nil {
					return usageErr(err)
				}
				clauses = append(clauses, fmt.Sprintf("date_received:[%s TO 99991231]", ymd))
			}
			if q.Search == "" {
				q.Search = openfda.BuildLuceneAND(clauses...)
			} else if len(clauses) > 0 {
				q.Search = openfda.BuildLuceneAND(append([]string{q.Search}, clauses...)...)
			}
			if q.Sort == "" {
				q.Sort = "date_received:desc"
			}
			return runOpenFDAList(cmd, flags, pathMaude, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	cmd.Flags().StringVar(&productCode, "product-code", "", "FDA product code")
	cmd.Flags().StringVar(&eventType, "event-type", "", "Event type: Death | Injury | Malfunction")
	cmd.Flags().StringVar(&since, "since", "", "Events since e.g. 2y, 6m, 30d")
	return cmd
}
