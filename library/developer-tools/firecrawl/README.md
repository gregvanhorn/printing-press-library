# Firecrawl CLI

API for interacting with Firecrawl services to perform web scraping and crawling tasks.

Learn more at [Firecrawl](https://firecrawl.dev/support).

## Install

### Go

```
go install github.com/mvanhorn/printing-press-library/library/developer-tools/firecrawl/cmd/firecrawl-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Quick Start

### 1. Install

See [Install](#install) above.

### 2. Set Up Credentials

Get your access token from your API provider's developer portal, then store it:

```bash
firecrawl-pp-cli auth set-token YOUR_TOKEN_HERE
```

Or set it via environment variable:

```bash
export FIRECRAWL_TOKEN="your-token-here"
```

### 3. Verify Setup

```bash
firecrawl-pp-cli doctor
```

This checks your configuration and credentials.

### 4. Try Your First Command

```bash
firecrawl-pp-cli batch cancel-scrape
```

## Usage

Run `firecrawl-pp-cli --help` for the full command reference and flag list.

## Commands

### batch

Manage batch

- **`firecrawl-pp-cli batch cancel-scrape`** - Cancel a batch scrape job
- **`firecrawl-pp-cli batch get-scrape-errors`** - Get the errors of a batch scrape job
- **`firecrawl-pp-cli batch get-scrape-status`** - Get the status of a batch scrape job
- **`firecrawl-pp-cli batch scrape-and-extract-from-urls`** - Scrape multiple URLs and optionally extract information using an LLM

### crawl

Manage crawl

- **`firecrawl-pp-cli crawl cancel`** - Cancel a crawl job
- **`firecrawl-pp-cli crawl get-active`** - Get all active crawls for the authenticated team
- **`firecrawl-pp-cli crawl get-status`** - Get the status of a crawl job
- **`firecrawl-pp-cli crawl urls`** - Crawl multiple URLs based on options

### deep-research

Manage deep research

- **`firecrawl-pp-cli deep-research get-status`** - Get the status and results of a deep research operation
- **`firecrawl-pp-cli deep-research start`** - Start a deep research operation on a query

### extract

Manage extract

- **`firecrawl-pp-cli extract data`** - Extract structured data from pages using LLMs
- **`firecrawl-pp-cli extract get-status`** - Get the status of an extract job

### llmstxt

Manage llmstxt

- **`firecrawl-pp-cli llmstxt generate-llms-txt`** - Generate LLMs.txt for a website
- **`firecrawl-pp-cli llmstxt get-llms-txt-status`** - Get the status and results of an LLMs.txt generation job

### map

Manage map

- **`firecrawl-pp-cli map urls`** - Map multiple URLs based on options

### scrape

Manage scrape

- **`firecrawl-pp-cli scrape and-extract-from-url`** - Scrape a single URL and optionally extract information using an LLM

### search

Manage search

- **`firecrawl-pp-cli search and-scrape`** - Search and optionally scrape search results

### team

Manage team

- **`firecrawl-pp-cli team get-credit-usage`** - Get remaining credits for the authenticated team
- **`firecrawl-pp-cli team get-token-usage`** - Get remaining tokens for the authenticated team (Extract only)


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
firecrawl-pp-cli batch cancel-scrape

# JSON for scripting and agents
firecrawl-pp-cli batch cancel-scrape --json

# Filter to specific fields
firecrawl-pp-cli batch cancel-scrape --json --select id,name,status

# Dry run — show the request without sending
firecrawl-pp-cli batch cancel-scrape --dry-run

# Agent mode — JSON + compact + no prompts in one flag
firecrawl-pp-cli batch cancel-scrape --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Retryable** - creates return "already exists" on retry, deletes return "already deleted"
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use as MCP Server

This CLI ships a companion MCP server for use with Claude Desktop, Cursor, and other MCP-compatible tools.

### Claude Code

```bash
claude mcp add firecrawl firecrawl-pp-mcp -e FIRECRAWL_TOKEN=<your-token>
```

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "firecrawl": {
      "command": "firecrawl-pp-mcp",
      "env": {
        "FIRECRAWL_TOKEN": "<your-key>"
      }
    }
  }
}
```

## Health Check

```bash
firecrawl-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Config file: `~/.config/firecrawl-pp-cli/config.toml`

Environment variables:
- `FIRECRAWL_TOKEN`

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `firecrawl-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $FIRECRAWL_TOKEN`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
