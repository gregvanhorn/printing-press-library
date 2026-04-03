// Copyright 2026 trevin-chow. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var _ = strings.ReplaceAll // ensure import
var _ = fmt.Sprintf        // ensure import
var _ = io.ReadAll         // ensure import
var _ = os.Stdin           // ensure import
var _ json.RawMessage      // ensure import

// queryIndicesMap maps friendly type names to the Postman search queryIndices values.
var queryIndicesMap = map[string][]string{
	"collection": {"runtime.collection"},
	"workspace":  {"collaboration.workspace"},
	"request":    {"runtime.request"},
	"flow":       {"flow.flow"},
	"team":       {"apinetwork.team"},
	"all":        {"runtime.collection", "collaboration.workspace", "runtime.request", "flow.flow", "apinetwork.team"},
}

func newSearchAllSearchAllCmd(flags *rootFlags) *cobra.Command {
	var flagType string
	var bodyFrom int
	var bodySize int
	var stdinBody bool

	cmd := &cobra.Command{
		Use:     "search <query>",
		Aliases: []string{"find"},
		Short:   "Full-text search across the public API network",
		Args:    cobra.MinimumNArgs(1),
		Example: `  # Search for Stripe collections
  postman-explore-pp-cli search stripe

  # Search workspaces
  postman-explore-pp-cli search "github api" --type workspace

  # Search all entity types
  postman-explore-pp-cli search kubernetes --type all

  # Search with JSON output for scripting
  postman-explore-pp-cli search oauth --json --select data`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			path := "/search-all"
			var body map[string]any
			if stdinBody {
				stdinData, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				var jsonBody map[string]any
				if err := json.Unmarshal(stdinData, &jsonBody); err != nil {
					return fmt.Errorf("parsing stdin JSON: %w", err)
				}
				body = jsonBody
			} else {
				queryText := strings.Join(args, " ")

				indices, ok := queryIndicesMap[flagType]
				if !ok {
					valid := make([]string, 0, len(queryIndicesMap))
					for k := range queryIndicesMap {
						valid = append(valid, k)
					}
					return usageErr(fmt.Errorf("invalid --type %q; valid types: %s", flagType, strings.Join(valid, ", ")))
				}

				body = map[string]any{
					"queryText":          queryText,
					"queryIndices":       indices,
					"mergeEntities":      true,
					"nested":             false,
					"nonNestedRequests":  true,
					"domain":             "public",
					"filter":             map[string]any{},
					"from":               bodyFrom,
					"size":               bodySize,
				}
			}
			data, statusCode, err := c.Post(path, body)
			if err != nil {
				return classifyAPIError(err)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printSearchTable(cmd, data)
			}
			if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				if flags.quiet {
					return nil
				}
				// Apply --compact and --select to the API response before wrapping
				filtered := data
				if flags.compact {
					filtered = compactFields(filtered)
				}
				if flags.selectFields != "" {
					filtered = filterFields(filtered, flags.selectFields)
				}
				envelope := map[string]any{
					"action":   "post",
					"resource": "search",
					"path":     path,
					"status":   statusCode,
					"success":  statusCode >= 200 && statusCode < 300,
				}
				if flags.dryRun {
					envelope["dry_run"] = true
					envelope["status"] = 0
					envelope["success"] = false
				}
				if len(filtered) > 0 {
					var parsed any
					if err := json.Unmarshal(filtered, &parsed); err == nil {
						envelope["data"] = parsed
					}
				}
				envelopeJSON, err := json.Marshal(envelope)
				if err != nil {
					return err
				}
				return printOutput(cmd.OutOrStdout(), json.RawMessage(envelopeJSON), true)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().StringVar(&flagType, "type", "collection", "Entity type to search: collection, workspace, request, flow, team, all")
	cmd.Flags().IntVar(&bodyFrom, "from", 0, "Pagination offset")
	cmd.Flags().IntVar(&bodySize, "size", 10, "Number of results to return")
	cmd.Flags().BoolVar(&stdinBody, "stdin", false, "Read request body as JSON from stdin")

	return cmd
}

// printSearchTable renders search results as a human-friendly table.
func printSearchTable(cmd *cobra.Command, data json.RawMessage) error {
	// Search response: {"data": [{"score":N, "document":{...}}, ...]}
	var resp struct {
		Data []struct {
			Document struct {
				Name                string `json:"name"`
				PublisherName       string `json:"publisherName"`
				ForkCount           int64  `json:"forkCount"`
				Views               int64  `json:"views"`
				IsPublisherVerified bool   `json:"isPublisherVerified"`
				EntityType          string `json:"entityType"`
			} `json:"document"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return printOutput(cmd.OutOrStdout(), data, false)
	}
	if len(resp.Data) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No results.")
		return nil
	}

	tw := newTabWriter(cmd.OutOrStdout())
	fmt.Fprintln(tw, strings.Join([]string{bold("NAME"), bold("PUBLISHER"), bold("FORKS"), bold("VIEWS"), bold("VERIFIED")}, "\t"))
	for _, item := range resp.Data {
		d := item.Document
		verified := "-"
		if d.IsPublisherVerified {
			verified = "yes"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			truncate(d.Name, 40),
			truncate(d.PublisherName, 25),
			formatCompact(d.ForkCount),
			formatCompact(d.Views),
			verified,
		)
	}
	return tw.Flush()
}
