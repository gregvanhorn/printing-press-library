package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSearchCmd(flags *rootFlags) *cobra.Command {
	var limit int
	var dbPath string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search the local post store",
		Long: `Run an FTS5 match against every post ever synced. The index covers
slug, title, tagline, and author. Works offline — empty store returns [].

FTS5 supports quoted phrases, boolean operators (AND, OR, NOT), and column
filters via the column:value shorthand. See SQLite FTS5 docs for the full
query grammar.`,
		Example: `  # Simple keyword
  producthunt-pp-cli search agent

  # Phrase + column filter
  producthunt-pp-cli search '"ai agent" author:hoover'

  # Agent-friendly narrow payload
  producthunt-pp-cli search "cli tool" --agent --select 'slug,title,tagline'`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(strings.Join(args, " "))
			if query == "" {
				return usageErr(fmt.Errorf("search query is required"))
			}
			db, err := openStore(dbPath)
			if err != nil {
				return configErr(err)
			}
			defer db.Close()
			posts, err := db.SearchPostsFTS(query, limit)
			if err != nil {
				return apiErr(err)
			}
			return printOutputWithFlags(cmd.OutOrStdout(), postsToJSON(posts), flags)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Max results to return")
	cmd.Flags().StringVar(&dbPath, "db", "", "Path to local SQLite store")
	return cmd
}
