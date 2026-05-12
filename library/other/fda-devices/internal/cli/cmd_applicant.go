// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

func newApplicantCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "applicant",
		Short:       "Look up 510(k) applicant companies",
		Long:        "Search and summarize 510(k) applicants (manufacturers/submitters).",
		Annotations: readOnlyAnnotations,
	}
	cmd.AddCommand(newApplicantGetCmd(flags), newApplicantHistoryCmd(flags))
	return cmd
}

func newApplicantGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:         "get <name>",
		Short:       "Find an applicant by name and report total 510(k) clearance count",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
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
			q := openfda.Query{
				Search: fmt.Sprintf("applicant:%q", name),
				Limit:  1,
			}
			if dryRunOK(flags) {
				c.DryRun = true
				_, _ = c.Get(path510k, q.Params())
				return nil
			}
			env, err := openfda.Run(ctx, c, path510k, q)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			if len(env.Results) == 0 {
				return notFoundErr(fmt.Errorf("no 510(k) records found for applicant %q", name))
			}
			total := 0
			if t, ok := env.Meta.Results["total"]; ok {
				if f, ok := t.(float64); ok {
					total = int(f)
				}
			}
			summary := map[string]any{
				"applicant_query": name,
				"clearance_count": total,
				"sample":          env.Results[0],
			}
			raw, _ := json.Marshal(summary)
			return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
		},
	}
}

func newApplicantHistoryCmd(flags *rootFlags) *cobra.Command {
	var (
		q   openfda.Query
		all bool
	)
	cmd := &cobra.Command{
		Use:         "history <name>",
		Short:       "List every 510(k) clearance issued to an applicant",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			name := args[0]
			clause := fmt.Sprintf("applicant:%q", name)
			if q.Search == "" {
				q.Search = clause
			} else {
				q.Search = openfda.BuildLuceneAND(q.Search, clause)
			}
			if q.Sort == "" {
				q.Sort = "decision_date:desc"
			}
			return runOpenFDAList(cmd, flags, path510k, q, all)
		},
	}
	addSearchFlags(cmd, &q, &all)
	return cmd
}
