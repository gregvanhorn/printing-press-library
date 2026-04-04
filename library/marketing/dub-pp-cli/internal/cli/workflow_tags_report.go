package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mvanhorn/printing-press-library/library/marketing/dub-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newWorkflowTagsReportCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "tags-report",
		Short: "Show which tags drive the most clicks, leads, and sales",
		Long: `Aggregate link performance by tag to identify which campaigns and
categories drive the most traffic. Requires a prior sync of links and tags.`,
		Example: `  # Tag performance report
  dub-pp-cli workflow tags-report

  # As JSON
  dub-pp-cli workflow tags-report --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = defaultDBPath("dub-pp-cli")
			}
			s, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening store: %w", err)
			}
			defer s.Close()

			links, err := s.List("links", 5000)
			if err != nil {
				return fmt.Errorf("listing links: %w", err)
			}
			tags, err := s.List("tags", 1000)
			if err != nil {
				return fmt.Errorf("listing tags: %w", err)
			}

			// Build tag name lookup
			tagNames := make(map[string]string)
			for _, t := range tags {
				var obj map[string]any
				if err := json.Unmarshal(t, &obj); err != nil {
					continue
				}
				tagNames[strVal(obj, "id")] = strVal(obj, "name")
			}

			type tagStat struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Links  int    `json:"links"`
				Clicks int    `json:"clicks"`
				Leads  int    `json:"leads"`
				Sales  int    `json:"sales"`
			}

			tagStats := make(map[string]*tagStat)

			for _, item := range links {
				var obj map[string]any
				if err := json.Unmarshal(item, &obj); err != nil {
					continue
				}

				clicks := intVal(obj, "clicks")
				leads := intVal(obj, "leads")
				sales := intVal(obj, "sales")

				// Links can have tagId (single) or tags (array)
				var tagIDs []string
				if tagID := strVal(obj, "tagId"); tagID != "" && tagID != "<nil>" {
					tagIDs = append(tagIDs, tagID)
				}
				if rawTags, ok := obj["tags"]; ok {
					if arr, ok := rawTags.([]any); ok {
						for _, t := range arr {
							if m, ok := t.(map[string]any); ok {
								if id := strVal(m, "id"); id != "" {
									tagIDs = append(tagIDs, id)
								}
							}
						}
					}
				}

				for _, tagID := range tagIDs {
					ts, ok := tagStats[tagID]
					if !ok {
						name := tagNames[tagID]
						if name == "" {
							name = tagID
						}
						ts = &tagStat{ID: tagID, Name: name}
						tagStats[tagID] = ts
					}
					ts.Links++
					ts.Clicks += clicks
					ts.Leads += leads
					ts.Sales += sales
				}
			}

			var sorted []*tagStat
			for _, ts := range tagStats {
				sorted = append(sorted, ts)
			}
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Clicks > sorted[j].Clicks
			})

			if flags.asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(sorted)
			}

			if len(sorted) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No tagged links found. Create tags and assign them to links first.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-25s %6s %8s %6s %6s\n", "Tag", "Links", "Clicks", "Leads", "Sales")
			fmt.Fprintf(cmd.OutOrStdout(), "%-25s %6s %8s %6s %6s\n", "-------------------------", "------", "--------", "------", "------")
			for _, ts := range sorted {
				name := ts.Name
				if len(name) > 25 {
					name = name[:22] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-25s %6d %8d %6d %6d\n", name, ts.Links, ts.Clicks, ts.Leads, ts.Sales)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path")

	return cmd
}
