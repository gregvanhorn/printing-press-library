// Copyright 2026 trevin-chow. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// normalizeEntityType maps plural/casual names to the canonical API entity type.
func normalizeEntityType(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "collections", "collection":
		return "collection"
	case "workspaces", "workspace":
		return "workspace"
	case "apis", "api":
		return "api"
	case "flows", "flow":
		return "flow"
	case "teams", "team":
		return "team"
	default:
		return s
	}
}

func newBrowseCmd(flags *rootFlags) *cobra.Command {
	var flagSort string
	var flagLimit int
	var flagOffset int
	var flagCategoryID int
	var flagAll bool

	cmd := &cobra.Command{
		Use:   "browse [entity-type]",
		Short: "Browse public entities on the API network",
		Long:  "Browse collections, workspaces, APIs, flows, and teams on the Postman API Network. Defaults to collections.",
		Args:  cobra.MaximumNArgs(1),
		Example: `  # Browse popular collections
  postman-explore-pp-cli browse

  # Browse workspaces
  postman-explore-pp-cli browse workspaces

  # Browse APIs in a specific category
  postman-explore-pp-cli browse apis --categoryid 5

  # Browse all pages of teams
  postman-explore-pp-cli browse teams --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			entityType := "collection"
			if len(args) > 0 {
				entityType = normalizeEntityType(args[0])
			}

			path := "/v1/api/networkentity"
			data, err := paginatedGet(c, path, map[string]string{
				"entityType": entityType,
				"limit":      fmt.Sprintf("%v", flagLimit),
				"offset":     fmt.Sprintf("%v", flagOffset),
				"sort":       flagSort,
				"categoryId": fmt.Sprintf("%v", flagCategoryID),
			}, flagAll, "offset", "", "")
			if err != nil {
				return classifyAPIError(err)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				var raw json.RawMessage
				if err := json.Unmarshal(data, &raw); err == nil {
					items := extractEntityItems(data)
					if len(items) > 0 {
						headers := []string{"NAME", "PUBLISHER", "FORKS", "VIEWS", "CATEGORY"}
						rows := make([][]string, 0, len(items))
						for _, item := range items {
							rows = append(rows, []string{
								strField(item, "name"),
								strField(item, "publisherName"),
								numField(item, "forkCount"),
								numField(item, "viewCount"),
								strField(item, "category"),
							})
						}
						tw := newTabWriter(cmd.OutOrStdout())
						fmt.Fprintln(tw, strings.Join(headers, "\t"))
						for _, row := range rows {
							fmt.Fprintln(tw, strings.Join(row, "\t"))
						}
						tw.Flush()
						if len(items) >= 25 {
							fmt.Fprintf(os.Stderr, "\nShowing %d results. To narrow: add --limit, --json --select, or filter flags.\n", len(items))
						}
						return nil
					}
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().StringVar(&flagSort, "sort", "popular", "Sort order: popular, recent, upvotes")
	cmd.Flags().IntVar(&flagLimit, "limit", 20, "Number of results per page")
	cmd.Flags().IntVar(&flagOffset, "offset", 0, "Pagination offset")
	cmd.Flags().IntVar(&flagCategoryID, "categoryid", 0, "Filter by category ID (from categories command)")
	cmd.Flags().BoolVar(&flagAll, "all", false, "Fetch all pages")

	return cmd
}

// extractEntityItems tries to extract an array from the response data.
func extractEntityItems(data json.RawMessage) []map[string]any {
	// Direct array
	var items []map[string]any
	if json.Unmarshal(data, &items) == nil {
		return items
	}
	// Object with common data fields
	var obj map[string]json.RawMessage
	if json.Unmarshal(data, &obj) == nil {
		for _, field := range []string{"data", "items", "results", "entities"} {
			if arr, ok := obj[field]; ok {
				if json.Unmarshal(arr, &items) == nil {
					return items
				}
			}
		}
	}
	return nil
}

// strField safely extracts a string value from a map.
func strField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return truncate(val, 40)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// numField safely extracts a numeric value and formats it compactly.
func numField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return "0"
	}
	switch val := v.(type) {
	case float64:
		return formatCompact(int64(val))
	case int64:
		return formatCompact(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
