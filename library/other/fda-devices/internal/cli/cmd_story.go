// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

type storyOut struct {
	KNumber      string            `json:"k_number"`
	Applicant    string            `json:"applicant"`
	DeviceName   string            `json:"device_name"`
	DecisionDate string            `json:"decision_date"`
	ProductCode  string            `json:"product_code"`
	DeviceClass  string            `json:"device_class"`
	RecallCount  int               `json:"recall_count"`
	Recalls      []json.RawMessage `json:"recalls"`
	Note         string            `json:"note"`
}

func newStoryCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "story <K-number>",
		Short:       "One-paragraph briefing about a 510(k) clearance",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return nil
			}
			kNumber := args[0]
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			rec, err := openfda.One(ctx, c, path510k, openfda.Query{
				Search: fmt.Sprintf("k_number:%q", kNumber),
				Limit:  1,
			})
			if err != nil {
				return classifyAPIError(err, flags)
			}
			if rec == nil {
				return notFoundErr(fmt.Errorf("no 510(k) record found for %s", kNumber))
			}
			var f struct {
				KNumber      string `json:"k_number"`
				Applicant    string `json:"applicant"`
				DeviceName   string `json:"device_name"`
				DecisionDate string `json:"decision_date"`
				ProductCode  string `json:"product_code"`
				DeviceClass  string `json:"device_class"`
			}
			_ = json.Unmarshal(rec, &f)
			recEnv, _ := openfda.Run(ctx, c, "/device/recall.json", openfda.Query{
				Search: fmt.Sprintf("k_numbers:%q", kNumber),
				Limit:  100,
			})
			recalls := []json.RawMessage{}
			if recEnv != nil {
				recalls = recEnv.Results
			}
			out := storyOut{
				KNumber: f.KNumber, Applicant: f.Applicant, DeviceName: f.DeviceName,
				DecisionDate: f.DecisionDate, ProductCode: f.ProductCode, DeviceClass: f.DeviceClass,
				RecallCount: len(recalls), Recalls: recalls,
				Note: "openFDA does not expose structured predicate references; predicate chain not included.",
			}

			if flags.asJSON || flags.agent {
				raw, err := json.Marshal(out)
				if err != nil {
					return err
				}
				return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"%s was cleared by %s on %s for %s (product code %s, class %s). %d recall(s) on record. %s\n",
				orDash(f.KNumber), orDash(f.Applicant), orDash(f.DecisionDate),
				orDash(f.DeviceName), orDash(f.ProductCode), orDash(f.DeviceClass),
				len(recalls), out.Note)
			return nil
		},
	}
	return cmd
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
