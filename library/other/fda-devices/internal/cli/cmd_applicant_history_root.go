// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

type applicantHistoryOut struct {
	ApplicantQuery string                       `json:"applicant_query"`
	Total          int                          `json:"total"`
	ByProductCode  map[string][]json.RawMessage `json:"by_product_code"`
}

// newApplicantHistoryRootCmd registers the top-level `applicant-history`
// command (distinct from the `applicant history` subcommand) — it groups
// every clearance for a company by product code.
func newApplicantHistoryRootCmd(flags *rootFlags) *cobra.Command {
	var (
		limit int
		all   bool
	)
	cmd := &cobra.Command{
		Use:         "applicant-history <company>",
		Short:       "Every 510(k) clearance for a company, grouped by product code",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return nil
			}
			name := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if limit <= 0 {
				limit = 100
			}
			if limit > 1000 {
				limit = 1000
			}
			q := openfda.Query{
				Search: fmt.Sprintf("applicant:%q", name),
				Sort:   "decision_date:desc",
				Limit:  limit,
			}
			var results []json.RawMessage
			if all {
				results, err = openfda.AllPages(ctx, c, path510k, q, 25000)
			} else {
				env, e := openfda.Run(ctx, c, path510k, q)
				if e == nil {
					results = env.Results
				}
				err = e
			}
			if err != nil {
				return classifyAPIError(err, flags)
			}
			grouped := map[string][]json.RawMessage{}
			for _, r := range results {
				var f struct {
					ProductCode string `json:"product_code"`
				}
				_ = json.Unmarshal(r, &f)
				key := f.ProductCode
				if key == "" {
					key = "_unknown"
				}
				grouped[key] = append(grouped[key], r)
			}
			out := applicantHistoryOut{
				ApplicantQuery: name,
				Total:          len(results),
				ByProductCode:  grouped,
			}
			raw, err := json.Marshal(out)
			if err != nil {
				return err
			}
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max records (cap 1000)")
	cmd.Flags().BoolVar(&all, "all", false, "Paginate up to 25k records")
	return cmd
}
