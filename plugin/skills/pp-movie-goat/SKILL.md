---
name: pp-movie-goat
description: "Printing Press CLI for TMDb + OMDb. Multi-source movie ratings (TMDb + OMDb), streaming availability, and cross-taste recommendations Trigger phrases: 'install movie-goat', 'use movie-goat', 'run movie-goat', 'TMDb + OMDb commands', 'setup movie-goat'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["movie-goat-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-cli@latest","bins":["movie-goat-pp-cli"],"label":"Install via go install"}]}}'
---

# TMDb + OMDb — Printing Press CLI

Multi-source movie ratings (TMDb + OMDb), streaming availability, and cross-taste recommendations

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `movie-goat-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-cli@latest
   ```
3. Verify: `movie-goat-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add -e TMDB_API_KEY=value movie-goat-pp-mcp -- movie-goat-pp-mcp
   ```
   Ask the user for actual values of required API keys before running.
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which movie-goat-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `movie-goat-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `movie-goat-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   movie-goat-pp-cli <command> [subcommand] [args] --agent
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
