---
name: pp-flightgoat
description: "Printing Press CLI for flightgoat. Free Google Flights search, Kayak nonstop route explorer, and optional FlightAware live tracking in one CLI. No API key required for search. Trigger phrases: 'install flightgoat', 'use flightgoat', 'run flightgoat', 'flightgoat commands', 'setup flightgoat'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
---

# flightgoat — Printing Press CLI

Free Google Flights search, Kayak nonstop route explorer, and optional FlightAware live tracking in one CLI. No API key required for search.

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `flightgoat-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/flightgoat/cmd/flightgoat-pp-cli@latest
   ```
3. Verify: `flightgoat-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.
5. Auth setup — set the API key and register it with the CLI:
   ```bash
   export FLIGHTGOAT_API_KEY_AUTH="your-key-here"
   flightgoat-pp-cli auth set-token
   ```
   Run `flightgoat-pp-cli doctor` to verify credentials.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/flightgoat/cmd/flightgoat-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add -e FLIGHTGOAT_API_KEY_AUTH=value flightgoat-pp-mcp -- flightgoat-pp-mcp
   ```
   Ask the user for actual values of required API keys before running.
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which flightgoat-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `flightgoat-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `flightgoat-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   flightgoat-pp-cli <command> [subcommand] [args] --agent
   ```
5. The `--agent` flag sets `--json --compact --no-input --no-color --yes` for structured, token-efficient output.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
