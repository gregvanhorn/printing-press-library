---
name: pp-hackernews
description: "Printing Press CLI for Hacker News. Browse, search, and analyze Hacker News — front page, Show HN, Ask HN, Who is Hiring, topic pulse, and pipe-friendly output Trigger phrases: 'install hackernews', 'use hackernews', 'run hackernews', 'Hacker News commands', 'setup hackernews'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["hackernews-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/hackernews/cmd/hackernews-pp-cli@latest","bins":["hackernews-pp-cli"],"label":"Install via go install"}]}}'
---

# Hacker News — Printing Press CLI

Browse, search, and analyze Hacker News — front page, Show HN, Ask HN, Who is Hiring, topic pulse, and pipe-friendly output

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `hackernews-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/hackernews/cmd/hackernews-pp-cli@latest
   ```
3. Verify: `hackernews-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/hackernews/cmd/hackernews-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add hackernews-pp-mcp -- hackernews-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which hackernews-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `hackernews-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `hackernews-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   hackernews-pp-cli <command> [subcommand] [args] --agent
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
