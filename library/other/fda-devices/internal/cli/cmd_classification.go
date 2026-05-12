// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

const pathClassification = "/device/classification.json"

// newClassificationCmd returns the classification command tree. We surface it
// twice — once as `classification` (preserves the API name) and once as
// `product-code` (preferred user-facing alias because records are keyed by
// 3-letter product codes).
func newClassificationCmd(flags *rootFlags, use string) *cobra.Command {
	short := "Device classifications by product code — search, list, get"
	if use == "product-code" {
		short = "Product-code classifications — search, list, get (alias for classification)"
	}
	cmd := &cobra.Command{
		Use:         use,
		Short:       short,
		Long:        "Query openFDA device classifications (/device/classification.json). Records are keyed by 3-letter product codes.",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(
		newClassificationSearchCmd(flags),
		newClassificationGetCmd(flags),
		newClassificationListCmd(flags),
	)
	return cmd
}

func newClassificationSearchCmd(flags *rootFlags) *cobra.Command {
	var q openfda.Query
	var all bool
	cmd := &cobra.Command{
		Use:         "search",
		Short:       "Raw search against /device/classification.json",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpenFDAList(cmd, flags, pathClassification, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	return cmd
}

func newClassificationGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "get <product-code>",
		Short:       "Fetch one classification by 3-letter product code",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return runOpenFDAGet(cmd, flags, pathClassification, "product_code", args[0])
		},
	}
}

func newClassificationListCmd(flags *rootFlags) *cobra.Command {
	var (
		deviceClass      string
		medicalSpecialty string
		q                openfda.Query
		all              bool
	)
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List classifications with curated filters",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			clauses := []string{}
			if deviceClass != "" {
				clauses = append(clauses, fmt.Sprintf("device_class:%q", deviceClass))
			}
			if medicalSpecialty != "" {
				clauses = append(clauses, fmt.Sprintf("medical_specialty_description:%q", medicalSpecialty))
			}
			if q.Search == "" {
				q.Search = openfda.BuildLuceneAND(clauses...)
			} else if len(clauses) > 0 {
				q.Search = openfda.BuildLuceneAND(append([]string{q.Search}, clauses...)...)
			}
			return runOpenFDAList(cmd, flags, pathClassification, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	cmd.Flags().StringVar(&deviceClass, "device-class", "", "Device class: 1 | 2 | 3")
	cmd.Flags().StringVar(&medicalSpecialty, "medical-specialty", "", "Medical specialty description")
	return cmd
}
