// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mvanhorn/printing-press-library/library/other/fda-devices/internal/openfda"
)

// runOpenFDAList runs a search query against an openFDA path and renders the
// results array (not the full envelope) through the standard output pipeline.
// On --all it skip-paginates up to 25k records.
func runOpenFDAList(cmd *cobra.Command, flags *rootFlags, path string, q openfda.Query, all bool) error {
	c, err := flags.newClient()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if dryRunOK(flags) {
		c.DryRun = true
		_, _ = c.Get(path, q.Params())
		return nil
	}

	var results []json.RawMessage
	if all {
		results, err = openfda.AllPages(ctx, c, path, q, 25000)
		if err != nil {
			return classifyAPIError(err, flags)
		}
	} else {
		env, err := openfda.Run(ctx, c, path, q)
		if err != nil {
			return classifyAPIError(err, flags)
		}
		results = env.Results
	}

	raw, err := json.Marshal(results)
	if err != nil {
		return err
	}
	return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
}

// runOpenFDAGet fetches a single record by exact-match ID field and prints it.
// Returns notFoundErr when no record matches.
func runOpenFDAGet(cmd *cobra.Command, flags *rootFlags, path, idField, id string) error {
	c, err := flags.newClient()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	q := openfda.Query{
		Search: fmt.Sprintf("%s:%q", idField, id),
		Limit:  1,
	}
	if dryRunOK(flags) {
		c.DryRun = true
		_, _ = c.Get(path, q.Params())
		return nil
	}
	one, err := openfda.One(ctx, c, path, q)
	if err != nil {
		return classifyAPIError(err, flags)
	}
	if one == nil {
		return notFoundErr(fmt.Errorf("no %s record found for %s=%q", path, idField, id))
	}
	return printOutputWithFlags(cmd.OutOrStdout(), one, flags)
}

// addSearchFlags wires the common --search/--sort/--count/--limit/--skip/--all
// flags onto a search subcommand.
func addSearchFlags(cmd *cobra.Command, q *openfda.Query, all *bool) {
	cmd.Flags().StringVar(&q.Search, "search", "", "Lucene search expression")
	cmd.Flags().StringVar(&q.Sort, "sort", "", "Sort spec, e.g. decision_date:desc")
	cmd.Flags().StringVar(&q.Count, "count", "", "Aggregate count on a field (e.g. product_code.exact)")
	cmd.Flags().IntVar(&q.Limit, "limit", 10, "Max records per page (openFDA max 1000)")
	cmd.Flags().IntVar(&q.Skip, "skip", 0, "Skip N records (openFDA max 25000)")
	cmd.Flags().BoolVar(all, "all", false, "Paginate via skip up to 25k records")
}

var readOnlyAnnotations = map[string]string{"mcp:read-only": "true"}
