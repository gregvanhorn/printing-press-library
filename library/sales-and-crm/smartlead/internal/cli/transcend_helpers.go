// Copyright 2026 gregvanhorn. Licensed under Apache-2.0. See LICENSE.
// PATCH: novel transcendence command (cross-campaign aggregation, hand-built on top of generator output)
// Transcendence command helpers.

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mvanhorn/printing-press-library/library/sales-and-crm/smartlead/internal/store"
	"github.com/spf13/cobra"
)

// openLocalStore opens the local SQLite store read-only and returns a friendly
// error suggesting `sync` if it can't be opened.
func openLocalStore(flags *rootFlags, dbPath string) (*store.Store, error) {
	if dbPath == "" {
		dbPath = defaultDBPath("smartlead-pp-cli")
	}
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("local database not found at %s: run 'smartlead-pp-cli sync --full' first", dbPath)
	}
	db, err := store.OpenReadOnly(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening local database: %w\nhint: run 'smartlead-pp-cli sync --full' first", err)
	}
	return db, nil
}

// emitTranscendOutput renders a result as JSON (when --json) or a friendly
// fallback. The result is filtered through --select when provided.
func emitTranscendOutput(cmd *cobra.Command, flags *rootFlags, result any, emptyMsg string) error {
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if flags.selectFields != "" {
		raw = filterFields(raw, flags.selectFields)
	}

	w := cmd.OutOrStdout()
	if flags.asJSON {
		var pretty json.RawMessage = raw
		out, _ := json.MarshalIndent(pretty, "", "  ")
		fmt.Fprintln(w, string(out))
		return nil
	}

	// Plain mode: pretty-print JSON with header (or empty message)
	if isEmptyResult(result) && emptyMsg != "" {
		fmt.Fprintln(w, emptyMsg)
		return nil
	}
	out, _ := json.MarshalIndent(json.RawMessage(raw), "", "  ")
	fmt.Fprintln(w, string(out))
	return nil
}

func isEmptyResult(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case []any:
		return len(x) == 0
	}
	// Marshal-and-check fallback
	b, err := json.Marshal(v)
	if err != nil {
		return false
	}
	s := string(b)
	return s == "[]" || s == "null" || s == "{}"
}
