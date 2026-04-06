package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// printJSON marshals v to JSON and prints to stdout.
func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printTable prints rows as a tab-aligned table with headers.
func printTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

// errorf returns a formatted error. Cobra prints it to stderr.
func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// infof prints a formatted info message to stderr (progress/status).
func infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// resolveOutput determines the output file path from either an -o/--output flag
// or a positional argument. It enforces mutual exclusivity: if both are provided,
// it returns an error. If neither is provided and no fallback is given, it errors.
//
// flagVal: the value of the --output/-o flag (empty string if not set)
// args: the cobra args slice
// argIndex: which positional arg holds the output path (e.g., 0 for screenshot, 1 for convert)
// fallback: optional default when neither flag nor arg is provided (empty string = no default, error instead)
func resolveOutput(flagVal string, args []string, argIndex int, fallback string) (string, error) {
	hasFlag := flagVal != ""
	hasArg := argIndex < len(args)

	switch {
	case hasFlag && hasArg:
		return "", errorf("specify output as either a positional argument or --output/-o, not both")
	case hasFlag:
		return flagVal, nil
	case hasArg:
		return args[argIndex], nil
	case fallback != "":
		return fallback, nil
	default:
		return "", errorf("output file path required (positional arg or --output/-o)")
	}
}
