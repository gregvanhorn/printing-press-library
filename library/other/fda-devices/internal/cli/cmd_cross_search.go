// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

func newCrossSearchCmd(flags *rootFlags) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:         "search <query>",
		Short:       "Cross-entity Lucene search across all 6 openFDA device endpoints",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return nil
			}
			query := args[0]
			if limit <= 0 {
				limit = 5
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			endpoints := map[string]string{
				"510k":           "/device/510k.json",
				"pma":            "/device/pma.json",
				"recall":         "/device/recall.json",
				"maude":          "/device/event.json",
				"classification": "/device/classification.json",
				"establishment":  "/device/registrationlisting.json",
			}
			results := map[string][]json.RawMessage{}
			var mu sync.Mutex
			var wg sync.WaitGroup
			for key, path := range endpoints {
				wg.Add(1)
				go func(key, path string) {
					defer wg.Done()
					env, err := openfda.Run(ctx, c, path, openfda.Query{
						Search: query,
						Limit:  limit,
					})
					out := []json.RawMessage{}
					if err == nil && env != nil {
						out = env.Results
					}
					mu.Lock()
					results[key] = out
					mu.Unlock()
				}(key, path)
			}
			wg.Wait()
			// Ensure all keys present
			for key := range endpoints {
				if results[key] == nil {
					results[key] = []json.RawMessage{}
				}
			}
			raw, err := json.Marshal(results)
			if err != nil {
				return err
			}
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 5, "Max records per endpoint")
	return cmd
}
