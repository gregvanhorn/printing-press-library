---
name: pp-dub
description: "Printing Press CLI for Dub. Create short links, track analytics, manage domains, and run affiliate programs via the Dub API Trigger phrases: 'install dub', 'use dub', 'run dub', 'Dub commands', 'setup dub'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["dub-pp-cli"],"env":["DUB_TOKEN"]},"primaryEnv":"DUB_TOKEN","install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/marketing/dub/cmd/dub-pp-cli@latest","bins":["dub-pp-cli"],"label":"Install via go install"}]}}'
---

# Dub — Printing Press CLI

Create short links, track analytics, manage domains, and run affiliate programs via the Dub API

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `dub-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/marketing/dub/cmd/dub-pp-cli@latest
   ```
3. Verify: `dub-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.
5. Auth setup — set the API key and register it with the CLI:
   ```bash
   export DUB_TOKEN="your-key-here"
   dub-pp-cli auth set-token
   ```
   Run `dub-pp-cli doctor` to verify credentials.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/marketing/dub/cmd/dub-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add -e DUB_TOKEN=value dub-pp-mcp -- dub-pp-mcp
   ```
   Ask the user for actual values of required API keys before running.
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which dub-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `dub-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `dub-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   dub-pp-cli <command> [subcommand] [args] --agent
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
