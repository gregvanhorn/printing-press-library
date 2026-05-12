// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

type recallLinkOut struct {
	KNumber string            `json:"k_number"`
	Device  json.RawMessage   `json:"device"`
	Recalls []json.RawMessage `json:"recalls"`
}

func newRecallLinkCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "recall-link <K-number>",
		Short:       "Find recalls associated with a 510(k) clearance",
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
			out := recallLinkOut{KNumber: kNumber, Recalls: []json.RawMessage{}}
			if rec != nil {
				out.Device = rec
			}
			recEnv, rerr := openfda.Run(ctx, c, "/device/recall.json", openfda.Query{
				Search: fmt.Sprintf("k_numbers:%q", kNumber),
				Limit:  100,
			})
			if rerr == nil && recEnv != nil {
				out.Recalls = recEnv.Results
			}
			raw, err := json.Marshal(out)
			if err != nil {
				return err
			}
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	return cmd
}
