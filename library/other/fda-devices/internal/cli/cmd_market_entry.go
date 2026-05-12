// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

type marketEntryRow struct {
	KNumber         string `json:"k_number"`
	Applicant       string `json:"applicant"`
	DecisionDate    string `json:"decision_date"`
	DateReceived    string `json:"date_received"`
	DaysToClearance int    `json:"days_to_clearance"`
}

func newMarketEntryCmd(flags *rootFlags) *cobra.Command {
	var (
		productCode string
		since       string
	)
	cmd := &cobra.Command{
		Use:         "market-entry",
		Short:       "Rank recent entrants in a product code by time-to-clearance",
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
			q := openfda.Query{
				Search: joinAND(clauses...),
				Sort:   "decision_date:desc",
				Limit:  100,
			}
			env, err := openfda.Run(ctx, c, path510k, q)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			rows := make([]marketEntryRow, 0, len(env.Results))
			for _, r := range env.Results {
				var f struct {
					KNumber      string `json:"k_number"`
					Applicant    string `json:"applicant"`
					DecisionDate string `json:"decision_date"`
					DateReceived string `json:"date_received"`
				}
				if err := json.Unmarshal(r, &f); err != nil {
					continue
				}
				days := computeDays(f.DateReceived, f.DecisionDate)
				rows = append(rows, marketEntryRow{
					KNumber: f.KNumber, Applicant: f.Applicant,
					DecisionDate: f.DecisionDate, DateReceived: f.DateReceived,
					DaysToClearance: days,
				})
			}
			sort.SliceStable(rows, func(i, j int) bool {
				if rows[i].DaysToClearance < 0 {
					return false
				}
				if rows[j].DaysToClearance < 0 {
					return true
				}
				return rows[i].DaysToClearance < rows[j].DaysToClearance
			})
			raw, err := json.Marshal(rows)
			if err != nil {
				return err
			}
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().StringVar(&productCode, "product-code", "", "3-letter FDA product code (e.g. DQA)")
	cmd.Flags().StringVar(&since, "since", "1y", "Filter clearances since e.g. 2y, 6m, 30d")
	return cmd
}

// computeDays parses two YYYY-MM-DD or YYYYMMDD strings and returns the number
// of days between them. Returns -1 if either is unparseable.
func computeDays(start, end string) int {
	s, err := parseFDADate(start)
	if err != nil {
		return -1
	}
	e, err := parseFDADate(end)
	if err != nil {
		return -1
	}
	return int(e.Sub(s).Hours() / 24)
}

func parseFDADate(s string) (time.Time, error) {
	for _, layout := range []string{"2006-01-02", "20060102"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable date: %q", s)
}
