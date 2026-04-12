---
name: pp-archive-is
description: "Printing Press CLI for Archive.today. Bypass paywalls and look up web archives via archive.today. Hero command: find or create an archive for any URL with lookup-before-submit, Wayback Machine fallback, and agent-hints on stderr when called non-interactively. Trigger phrases: 'install archive-is', 'use archive-is', 'run archive-is', 'Archive.today commands', 'setup archive-is'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["archive-is-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/archive-is/cmd/archive-is-pp-cli@latest","bins":["archive-is-pp-cli"],"label":"Install via go install"}]}}'
---

# Archive.today — Printing Press CLI

Bypass paywalls and look up web archives via archive.today. Hero command: find or create an archive for any URL with lookup-before-submit, Wayback Machine fallback, and agent-hints on stderr when called non-interactively.

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `archive-is-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/archive-is/cmd/archive-is-pp-cli@latest
   ```
3. Verify: `archive-is-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/archive-is/cmd/archive-is-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add archive-is-pp-mcp -- archive-is-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which archive-is-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `archive-is-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `archive-is-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   archive-is-pp-cli <command> [subcommand] [args] --agent
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
