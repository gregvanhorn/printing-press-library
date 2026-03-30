// Copyright 2026 trevin-chow. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newOpenCmd(flags *rootFlags) *cobra.Command {
	var flagDryRun bool

	cmd := &cobra.Command{
		Use:   "open <query>",
		Short: "Search and open the first result in your browser",
		Long:  "Searches the Postman API Network and opens the top result's URL in your default browser (macOS).",
		Args:  cobra.MinimumNArgs(1),
		Example: `  # Open the top result for "stripe"
  postman-explore-pp-cli open stripe

  # Preview the URL without opening
  postman-explore-pp-cli open stripe --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			queryText := strings.Join(args, " ")

			body := map[string]any{
				"queryText":         queryText,
				"queryIndices":      []string{"runtime.collection"},
				"mergeEntities":     true,
				"nested":            false,
				"nonNestedRequests": true,
				"domain":            "public",
				"filter":            map[string]any{},
				"from":              0,
				"size":              1,
			}

			data, _, err := c.Post("/search-all", body)
			if err != nil {
				return classifyAPIError(err)
			}

			url := extractFirstURL(data)
			if url == "" {
				return fmt.Errorf("no results found for %q", queryText)
			}

			if flagDryRun {
				fmt.Fprintln(cmd.OutOrStdout(), url)
				return nil
			}

			fmt.Fprintf(os.Stderr, "Opening %s\n", url)
			return exec.Command("open", url).Run()
		},
	}
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the URL instead of opening it")

	_ = io.ReadAll // ensure import
	return cmd
}

// extractFirstURL tries to find a URL from the first search result.
func extractFirstURL(data json.RawMessage) string {
	// Try top-level object with data array
	var obj map[string]json.RawMessage
	if json.Unmarshal(data, &obj) == nil {
		for _, field := range []string{"data", "items", "results"} {
			if arr, ok := obj[field]; ok {
				if u := urlFromArray(arr); u != "" {
					return u
				}
			}
		}
	}
	// Try direct array
	if u := urlFromArray(data); u != "" {
		return u
	}
	return ""
}

func urlFromArray(data json.RawMessage) string {
	var items []map[string]any
	if json.Unmarshal(data, &items) != nil || len(items) == 0 {
		return ""
	}
	item := items[0]
	// Check common URL fields
	for _, key := range []string{"url", "publicUrl", "publisherUrl", "webUrl", "href"} {
		if v, ok := item[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	// Try to construct from slug/id
	if slug, ok := item["slug"].(string); ok && slug != "" {
		return "https://www.postman.com/explore/" + slug
	}
	if id, ok := item["id"].(string); ok && id != "" {
		return "https://www.postman.com/explore/" + id
	}
	return ""
}
