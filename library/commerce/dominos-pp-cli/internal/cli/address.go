package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/spf13/cobra"
)

func newAddressCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{Use: "address", Short: "Manage saved addresses"}
	cmd.AddCommand(newAddressAddCmd(flags), newAddressListCmd(flags), newAddressRemoveCmd(flags), newAddressDefaultCmd(flags))
	return cmd
}

func newAddressAddCmd(flags *rootFlags) *cobra.Command {
	var city, state, zip, label string
	var isDefault bool
	cmd := &cobra.Command{
		Use:   "add <street> --city <city> --state <state> --zip <zip> [--label <label>] [--default]",
		Short: "Save an address to local storage",
		Example: "  dominos-pp-cli address add \"421 N 63rd St\" --city Seattle --state WA --zip 98103 --label home --default\n" +
			"  dominos-pp-cli address add \"500 Howard St\" --city \"San Francisco\" --state CA --zip 94105 --label work",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := newUUID()
			if err != nil {
				return err
			}
			rec := map[string]any{"action": "address.add", "id": id, "label": label, "street": args[0], "city": city, "state": state, "zip": zip, "is_default": isDefault}
			if flags.dryRun {
				return renderAction(cmd, flags, "updated", rec)
			}
			s, err := flags.openStore()
			if err != nil {
				return err
			}
			defer s.Close()
			if err := s.UpsertAddress(id, label, args[0], city, state, zip, isDefault); err != nil {
				return err
			}
			return renderAction(cmd, flags, "updated", rec)
		},
	}
	cmd.Flags().StringVar(&city, "city", "", "City")
	cmd.Flags().StringVar(&state, "state", "", "State")
	cmd.Flags().StringVar(&zip, "zip", "", "ZIP code")
	cmd.Flags().StringVar(&label, "label", "", "Address label such as home, work, or custom")
	cmd.Flags().BoolVar(&isDefault, "default", false, "Set as the default address")
	_ = cmd.MarkFlagRequired("city")
	_ = cmd.MarkFlagRequired("state")
	_ = cmd.MarkFlagRequired("zip")
	return cmd
}

func newAddressListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List saved addresses", Example: "  dominos-pp-cli address list\n  dominos-pp-cli address list --json", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := flags.openStore()
		if err != nil {
			return err
		}
		defer s.Close()
		rows, err := s.ListAddresses()
		if err != nil {
			return err
		}
		if flags.asJSON {
			return flags.printJSON(cmd, rows)
		}
		items := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			var obj map[string]any
			if json.Unmarshal(row, &obj) == nil {
				items = append(items, obj)
			}
		}
		rowsOut := make([][]string, 0, len(items))
		for _, item := range items {
			rowsOut = append(rowsOut, []string{fmt.Sprintf("%v", item["id"]), fmt.Sprintf("%v", item["label"]), fmt.Sprintf("%v", item["street"]), fmt.Sprintf("%v", item["city"]), fmt.Sprintf("%v", item["state"]), fmt.Sprintf("%v", item["zip"]), fmt.Sprintf("%v", item["is_default"])})
		}
		return flags.printTable(cmd, []string{"ID", "LABEL", "STREET", "CITY", "STATE", "ZIP", "DEFAULT"}, rowsOut)
	}}
}

func newAddressRemoveCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "remove <label-or-id>", Short: "Remove a saved address", Example: "  dominos-pp-cli address remove home --dry-run\n  dominos-pp-cli address remove 9c4ef5d4-1111-2222-3333-444455556666", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		out := map[string]any{"action": "address.remove", "target": args[0]}
		if flags.dryRun {
			return renderAction(cmd, flags, "updated", out)
		}
		db, err := openAddressDB()
		if err != nil {
			return err
		}
		defer db.Close()
		res, err := db.Exec(`DELETE FROM addresses WHERE id = ? OR label = ?`, args[0], args[0])
		if err != nil {
			return fmt.Errorf("remove address: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return notFoundErr(fmt.Errorf("address %q not found", args[0]))
		}
		return renderAction(cmd, flags, "updated", out)
	}}
}

func newAddressDefaultCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{Use: "default <label-or-id>", Short: "Set the default address", Example: "  dominos-pp-cli address default home\n  dominos-pp-cli address default 9c4ef5d4-1111-2222-3333-444455556666 --dry-run", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		out := map[string]any{"action": "address.default", "target": args[0]}
		if flags.dryRun {
			return renderAction(cmd, flags, "updated", out)
		}
		db, err := openAddressDB()
		if err != nil {
			return err
		}
		defer db.Close()
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("set default address: %w", err)
		}
		defer tx.Rollback()
		if _, err := tx.Exec(`UPDATE addresses SET is_default = FALSE WHERE is_default = TRUE`); err != nil {
			return fmt.Errorf("set default address: %w", err)
		}
		res, err := tx.Exec(`UPDATE addresses SET is_default = TRUE WHERE id = ? OR label = ?`, args[0], args[0])
		if err != nil {
			return fmt.Errorf("set default address: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return notFoundErr(fmt.Errorf("address %q not found", args[0]))
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("set default address: %w", err)
		}
		return renderAction(cmd, flags, "updated", out)
	}}
}

func openAddressDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", defaultDBPath("dominos-pp-cli")+"?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000&_foreign_keys=ON&_temp_store=MEMORY&_mmap_size=268435456")
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}
	return db, nil
}

func renderAction(cmd *cobra.Command, flags *rootFlags, done string, v map[string]any) error {
	if flags.dryRun {
		v["dry_run"] = true
	}
	if flags.asJSON {
		return flags.printJSON(cmd, v)
	}
	status := done
	if flags.dryRun {
		status = "dry-run"
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", status, v["action"])
	return nil
}
