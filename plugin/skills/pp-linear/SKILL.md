---
name: pp-linear
description: "Printing Press CLI for Linear. Offline-capable, agent-native CLI for the Linear API with SQLite-backed sync, search, and cross-entity queries. Trigger phrases: 'install linear', 'use linear', 'run linear', 'Linear commands', 'setup linear'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["linear-pp-cli"],"env":["LINEAR_API_KEY"]},"primaryEnv":"LINEAR_API_KEY","install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/project-management/linear/cmd/linear-pp-cli@latest","bins":["linear-pp-cli"],"label":"Install via go install"}]}}'
---

# Linear — Printing Press CLI

Offline-capable, agent-native CLI for the Linear API with SQLite-backed sync, search, and cross-entity queries.

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `linear-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/project-management/linear/cmd/linear-pp-cli@latest
   ```
3. Verify: `linear-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.
5. Auth setup — set the API key and register it with the CLI:
   ```bash
   export LINEAR_API_KEY="your-key-here"
   linear-pp-cli auth set-token
   ```
   Run `linear-pp-cli doctor` to verify credentials.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/project-management/linear/cmd/linear-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add -e LINEAR_API_KEY=value linear-pp-mcp -- linear-pp-mcp
   ```
   Ask the user for actual values of required API keys before running.
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which linear-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `linear-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `linear-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   linear-pp-cli <command> [subcommand] [args] --agent
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
