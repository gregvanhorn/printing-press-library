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

type safetyPatternRow struct {
	KNumber       string  `json:"k_number"`
	Applicant     string  `json:"applicant"`
	DecisionDate  string  `json:"decision_date"`
	Events        int     `json:"events"`
	YearsOnMarket float64 `json:"years_on_market"`
	EventsPerYear float64 `json:"events_per_year"`
}

type safetyPatternOut struct {
	Results []safetyPatternRow `json:"results"`
}

func newSafetyPatternCmd(flags *rootFlags) *cobra.Command {
	var productCode string
	cmd := &cobra.Command{
		Use:         "safety-pattern",
		Short:       "Rank devices in a product code by adverse-event rate per year on market",
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
			env, err := openfda.Run(ctx, c, path510k, openfda.Query{
				Search: fmt.Sprintf("product_code:%q", productCode),
				Sort:   "decision_date:desc",
				Limit:  100,
			})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			type kInfo struct {
				Applicant    string
				DecisionDate string
			}
			devices := map[string]kInfo{}
			for _, r := range env.Results {
				var f struct {
					KNumber      string `json:"k_number"`
					Applicant    string `json:"applicant"`
					DecisionDate string `json:"decision_date"`
				}
				if err := json.Unmarshal(r, &f); err != nil {
					continue
				}
				if f.KNumber == "" {
					continue
				}
				devices[f.KNumber] = kInfo{Applicant: f.Applicant, DecisionDate: f.DecisionDate}
			}

			counts := map[string]int{}
			countEnv, cerr := openfda.Run(ctx, c, "/device/event.json", openfda.Query{
				Search: fmt.Sprintf("device.device_report_product_code:%q", productCode),
				Count:  "device.510_k_number.exact",
				Limit:  1000,
			})
			if cerr == nil && countEnv != nil {
				for _, r := range countEnv.Results {
					var f struct {
						Term  string `json:"term"`
						Count int    `json:"count"`
					}
					if err := json.Unmarshal(r, &f); err == nil {
						counts[f.Term] = f.Count
					}
				}
			}

			now := time.Now().UTC()
			rows := make([]safetyPatternRow, 0, len(devices))
			for k, info := range devices {
				events := counts[k]
				years := 0.0
				if t, err := parseFDADate(info.DecisionDate); err == nil {
					years = now.Sub(t).Hours() / (24 * 365.25)
				}
				epy := 0.0
				if years > 0 {
					epy = float64(events) / years
				}
				rows = append(rows, safetyPatternRow{
					KNumber: k, Applicant: info.Applicant, DecisionDate: info.DecisionDate,
					Events: events, YearsOnMarket: roundFloat(years, 2), EventsPerYear: roundFloat(epy, 2),
				})
			}
			sort.SliceStable(rows, func(i, j int) bool {
				return rows[i].EventsPerYear > rows[j].EventsPerYear
			})
			raw, err := json.Marshal(safetyPatternOut{Results: rows})
			if err != nil {
				return err
			}
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().StringVar(&productCode, "product-code", "", "3-letter FDA product code (e.g. DQA)")
	return cmd
}

func roundFloat(v float64, places int) float64 {
	pow := 1.0
	for i := 0; i < places; i++ {
		pow *= 10
	}
	return float64(int64(v*pow+0.5)) / pow
}
