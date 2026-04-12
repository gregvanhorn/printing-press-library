---
name: pp-weather-goat
description: "Printing Press CLI for Open-Meteo + NWS. Weather forecasts, severe weather alerts, air quality, and GO/CAUTION/STOP activity verdicts for walk, bike, hike, commute, and drive Trigger phrases: 'install weather-goat', 'use weather-goat', 'run weather-goat', 'Open-Meteo + NWS commands', 'setup weather-goat'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["weather-goat-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/other/weather-goat/cmd/weather-goat-pp-cli@latest","bins":["weather-goat-pp-cli"],"label":"Install via go install"}]}}'
---

# Open-Meteo + NWS — Printing Press CLI

Weather forecasts, severe weather alerts, air quality, and GO/CAUTION/STOP activity verdicts for walk, bike, hike, commute, and drive

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `weather-goat-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/weather-goat/cmd/weather-goat-pp-cli@latest
   ```
3. Verify: `weather-goat-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/other/weather-goat/cmd/weather-goat-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add weather-goat-pp-mcp -- weather-goat-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which weather-goat-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `weather-goat-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `weather-goat-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   weather-goat-pp-cli <command> [subcommand] [args] --agent
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
