// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	_ "modernc.org/sqlite"
)

func newSQLCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "sql <query>",
		Short:       "Read-only SQL against the local mirror DB",
		Annotations: readOnlyAnnotations,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			query := strings.TrimSpace(strings.Join(args, " "))
			query = strings.Trim(query, "'\"")
			query = strings.TrimSpace(query)
			upper := strings.ToUpper(query)
			if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "WITH") {
				return usageErr(fmt.Errorf("only SELECT/WITH queries allowed in sql command"))
			}
			if dryRunOK(flags) {
				return nil
			}
			path := defaultDBPath("fda-devices-pp-cli")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				raw, _ := json.Marshal([]map[string]any{})
				return printOutputWithFlags(cmd.OutOrStdout(), raw, flags)
			}
			dsn := fmt.Sprintf("file:%s?mode=ro", url.QueryEscape(path))
			db, err := sql.Open("sqlite", dsn)
			if err != nil {
				return err
			}
			defer db.Close()
			rows, err := db.QueryContext(cmd.Context(), query)
			if err != nil {
				return apiErr(err)
			}
			defer rows.Close()
			cols, err := rows.Columns()
			if err != nil {
				return err
			}
			out := []map[string]any{}
			for rows.Next() {
				vals := make([]any, len(cols))
				ptrs := make([]any, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				if err := rows.Scan(ptrs...); err != nil {
					return err
				}
				row := map[string]any{}
				for i, c := range cols {
					v := vals[i]
					if b, ok := v.([]byte); ok {
						v = string(b)
					}
					row[c] = v
				}
				out = append(out, row)
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
