// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

// joinAND joins Lucene clauses with literal " AND " separators. The openFDA
// API needs spaces (or '+' as URL-space) between clauses; the standard Go
// url.Values.Encode percent-encodes a literal '+', so we use real spaces
// and let url-encoding turn them into %20.
func joinAND(clauses ...string) string {
	out := make([]string, 0, len(clauses))
	for _, c := range clauses {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	return strings.Join(out, " AND ")
}

type competitorRow struct {
	KNumber      string `json:"k_number"`
	Applicant    string `json:"applicant"`
	DeviceName   string `json:"device_name"`
	DecisionDate string `json:"decision_date"`
	ProductCode  string `json:"product_code"`
}

func newCompetitorsCmd(flags *rootFlags) *cobra.Command {
	var (
		productCode string
		since       string
		limit       int
	)
	cmd := &cobra.Command{
		Use:         "competitors",
		Short:       "List every recent 510(k) clearance in a product code, ranked by recency",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if productCode == "" {
				return usageErr(fmt.Errorf("--product-code is required"))
			}
			if dryRunOK(flags) {
				return nil
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			clauses := []string{fmt.Sprintf("product_code:%q", productCode)}
			if since != "" {
				ymd, err := openfda.SinceToYYYYMMDD(since)
				if err != nil {
					return usageErr(err)
				}
				clauses = append(clauses, fmt.Sprintf("decision_date:[%s TO 99991231]", ymd))
			}
			if limit <= 0 {
				limit = 50
			}
			if limit > 1000 {
				limit = 1000
			}
			q := openfda.Query{
				Search: joinAND(clauses...),
				Sort:   "decision_date:desc",
				Limit:  limit,
			}
			env, err := openfda.Run(ctx, c, path510k, q)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			rows := make([]competitorRow, 0, len(env.Results))
			for _, r := range env.Results {
				var row competitorRow
				_ = json.Unmarshal(r, &row)
				rows = append(rows, row)
			}
			raw, err := json.Marshal(rows)
			if err != nil {
				return err
			}
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().StringVar(&productCode, "product-code", "", "3-letter FDA product code (e.g. DQA)")
	cmd.Flags().StringVar(&since, "since", "2y", "Clearances since e.g. 2y, 6m, 30d, 12w")
	cmd.Flags().IntVar(&limit, "limit", 50, "Max records (cap 1000)")
	return cmd
}
