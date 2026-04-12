---
name: pp-instacart
description: "Printing Press CLI for Instacart. Natural-language Instacart CLI that talks directly to the web GraphQL API. Add items to your cart, search products, and manage carts across retailers without browser automation. Trigger phrases: 'install instacart', 'use instacart', 'run instacart', 'Instacart commands', 'setup instacart'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["instacart-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/commerce/instacart/cmd/instacart-pp-cli@latest","bins":["instacart-pp-cli"],"label":"Install via go install"}]}}'
---

# Instacart — Printing Press CLI

Natural-language Instacart CLI that talks directly to the web GraphQL API. Add items to your cart, search products, and manage carts across retailers without browser automation.

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `instacart-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/commerce/instacart/cmd/instacart-pp-cli@latest
   ```
3. Verify: `instacart-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## Direct Use

1. Check if installed: `which instacart-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `instacart-pp-cli --help`
3. Match the user query to the best command. Drill into subcommand help if needed: `instacart-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   instacart-pp-cli <command> [subcommand] [args] --agent
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
