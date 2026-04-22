// Hand-authored: `query` and `decision-review` top-level commands.
// Not generated.

package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mvanhorn/printing-press-library/library/marketing/meta-ads/internal/store"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// newQueryCmd runs an arbitrary SQL query against the local SQLite store.
// Read-only: rejects anything that isn't SELECT / PRAGMA / EXPLAIN / WITH.
func newQueryCmd(flags *rootFlags) *cobra.Command {
	var limitRows int
	cmd := &cobra.Command{
		Use:   "query <sql>",
		Short: "Run SQL (SELECT/PRAGMA/EXPLAIN) against the local store",
		Long: `Run a read-only SQL query against the local SQLite store.

Rejects anything that isn't a SELECT, PRAGMA, EXPLAIN, or WITH statement.
Columns are printed as JSON objects (one per row) when --json is set,
or a tab-separated table otherwise.

The local store contains one resources table per entity: accounts, campaigns,
adsets, ads, creatives, insights, audiences, plus the decisions table the
audit log writes to. Resources are stored as JSON blobs in a 'data' column;
use json_extract(data, '$.field') to pull specific fields.`,
		Example: `  # Top-spending campaigns by frequency (reads the insights table)
  meta-ads-pp-cli query "SELECT json_extract(data,'$.campaign_name') AS name, CAST(json_extract(data,'$.spend') AS REAL) AS spend, CAST(json_extract(data,'$.frequency') AS REAL) AS freq FROM insights ORDER BY spend DESC LIMIT 10" --json

  # Decisions overdue for follow-up
  meta-ads-pp-cli query "SELECT log_id, campaign_name, action, follow_up_date FROM decisions WHERE was_applied = 1 AND follow_up_date <= datetime('now')"

  # What resource types are in the store?
  meta-ads-pp-cli query "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			q := strings.TrimSpace(strings.Join(args, " "))
			if !isReadOnlySQL(q) {
				// Not a SELECT/PRAGMA/EXPLAIN/WITH: show help so the user sees
				// the supported statement list instead of a bare error.
				fmt.Fprintln(cmd.ErrOrStderr(), "only SELECT / PRAGMA / EXPLAIN / WITH queries are allowed")
				return cmd.Help()
			}

			dbPath := defaultDBPath("meta-ads-pp-cli")
			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("opening local store at %s: %w", dbPath, err)
			}
			defer db.Close()

			rows, err := db.DB().Query(q)
			if err != nil {
				return fmt.Errorf("query failed: %w", err)
			}
			defer rows.Close()

			cols, err := rows.Columns()
			if err != nil {
				return err
			}

			all := []map[string]any{}
			for rows.Next() {
				if limitRows > 0 && len(all) >= limitRows {
					break
				}
				vals := make([]any, len(cols))
				ptrs := make([]any, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					return err
				}
				row := make(map[string]any, len(cols))
				for i, c := range cols {
					row[c] = coerceSQLValue(vals[i])
				}
				all = append(all, row)
			}
			if err := rows.Err(); err != nil {
				return err
			}

			if flags.asJSON {
				return flags.printJSON(cmd, map[string]any{
					"columns": cols,
					"rows":    all,
					"count":   len(all),
				})
			}
			if len(all) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no rows)")
				return nil
			}
			w := cmd.OutOrStdout()
			fmt.Fprintln(w, strings.Join(cols, "\t"))
			for _, row := range all {
				parts := make([]string, len(cols))
				for i, c := range cols {
					parts[i] = truncate(fmt.Sprintf("%v", row[c]), 80)
				}
				fmt.Fprintln(w, strings.Join(parts, "\t"))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limitRows, "limit", 1000, "Maximum rows to return (0 = no limit)")
	return cmd
}

// newDecisionReviewCmd is a top-level alias for `history review` — same flags,
// same behavior. Declared as a fresh struct literal (not a runtime reassignment
// of the history review cmd) so static tooling sees the top-level Use string.
func newDecisionReviewCmd(flags *rootFlags) *cobra.Command {
	inner := newHistoryReviewCmd(flags)
	innerRun := inner.RunE
	cmd := &cobra.Command{
		Use:     "decision-review <log-id>",
		Short:   "Attach a post-mortem analysis to a past decision (alias of 'history review')",
		Long:    inner.Long,
		Example: inner.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return innerRun(cmd, args)
		},
	}
	inner.Flags().VisitAll(func(f *pflag.Flag) {
		cmd.Flags().AddFlag(f)
	})
	return cmd
}

func isReadOnlySQL(q string) bool {
	trimmed := strings.TrimLeft(q, " \t\n(")
	u := strings.ToUpper(trimmed)
	return strings.HasPrefix(u, "SELECT") ||
		strings.HasPrefix(u, "PRAGMA") ||
		strings.HasPrefix(u, "EXPLAIN") ||
		strings.HasPrefix(u, "WITH")
}

func coerceSQLValue(v any) any {
	switch x := v.(type) {
	case []byte:
		// SQLite often returns TEXT as []byte. Try to decode JSON; fall back to string.
		s := string(x)
		var parsed any
		if json.Unmarshal(x, &parsed) == nil {
			switch parsed.(type) {
			case map[string]any, []any:
				return parsed
			}
		}
		return s
	case nil:
		return nil
	default:
		return x
	}
}

// Force sql import (used above via *sql.DB through store.DB()).
var _ = (*sql.DB)(nil)
