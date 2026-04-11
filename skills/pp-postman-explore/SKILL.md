---
name: pp-postman-explore
description: "Printing Press CLI for Postman Explore. Search and browse the Postman API Network Trigger phrases: 'install postman-explore', 'use postman-explore', 'run postman-explore', 'Postman Explore commands', 'setup postman-explore'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
---

# Postman Explore — Printing Press CLI

Search and browse the Postman API Network

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `postman-explore-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/developer-tools/postman-explore/cmd/postman-explore-pp-cli@latest
   ```
3. Verify: `postman-explore-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/developer-tools/postman-explore/cmd/postman-explore-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add postman-explore-pp-mcp -- postman-explore-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which postman-explore-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `postman-explore-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `postman-explore-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   postman-explore-pp-cli <command> [subcommand] [args] --agent
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
