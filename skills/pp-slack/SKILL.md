---
name: pp-slack
description: "Printing Press CLI for Slack. Send messages, search conversations, monitor channels, and manage your Slack workspace from the terminal Trigger phrases: 'install slack', 'use slack', 'run slack', 'Slack commands', 'setup slack'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
---

# Slack — Printing Press CLI

Send messages, search conversations, monitor channels, and manage your Slack workspace from the terminal

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `slack-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/slack/cmd/slack-pp-cli@latest
   ```
3. Verify: `slack-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.
5. Auth setup — set the API key and register it with the CLI:
   ```bash
   export SLACK_BOT_TOKEN="your-key-here"
   slack-pp-cli auth set-token
   ```
   Run `slack-pp-cli doctor` to verify credentials.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/slack/cmd/slack-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add -e SLACK_BOT_TOKEN=value slack-pp-mcp -- slack-pp-mcp
   ```
   Ask the user for actual values of required API keys before running.
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which slack-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `slack-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `slack-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   slack-pp-cli <command> [subcommand] [args] --agent
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
