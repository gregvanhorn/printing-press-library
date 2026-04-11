---
name: pp-kalshi
description: "Printing Press CLI for Kalshi. Trade prediction markets, track portfolios, and analyze odds on Kalshi from the command line Capabilities include: account, analytics, api-keys, communications, events, exchange, fcm, historical, incentive-programs, live-data, markets, milestones, multivariate-event-collections, portfolio, search, series, structured-targets, tail. Trigger phrases: 'install kalshi', 'use kalshi', 'run kalshi', 'Kalshi commands', 'setup kalshi'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
---

# Kalshi — Printing Press CLI

Trade prediction markets, track portfolios, and analyze odds on Kalshi from the command line

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `kalshi-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/payments/kalshi/cmd/kalshi-pp-cli@latest
   ```
3. Verify: `kalshi-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.
5. Auth setup — set the API key and register it with the CLI:
   ```bash
   export KALSHI_API_KEY="your-key-here"
   kalshi-pp-cli auth set-token
   ```
   Run `kalshi-pp-cli doctor` to verify credentials.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/payments/kalshi/cmd/kalshi-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add -e KALSHI_API_KEY=value kalshi-pp-mcp -- kalshi-pp-mcp
   ```
   Ask the user for actual values of required API keys before running.
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which kalshi-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `kalshi-pp-cli --help`
   Key commands:
   - `account` — Get Account API Limits
   - `analytics` — Run analytics queries on locally synced data
   - `api-keys` — Get API Keys
   - `communications` — Get Communications ID
   - `events` — Get Events
   - `exchange` — Get Exchange Announcements
   - `fcm` — Get FCM Orders
   - `historical` — Get Historical Trades
   - `incentive-programs` — Get Incentives
   - `live-data` — Get Multiple Live Data
   - `markets` — Get Multiple Market Orderbooks
   - `milestones` — Get Milestones
   - `multivariate-event-collections` — Get Multivariate Event Collections
   - `portfolio` — Get Positions
   - `search` — Full-text search across synced data or live API
   - `series` — Get Series List
   - `structured-targets` — Get Structured Targets
   - `tail` — Stream live changes by polling the API at regular intervals
3. Match the user query to the best command. Drill into subcommand help if needed: `kalshi-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   kalshi-pp-cli <command> [subcommand] [args] --agent
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
