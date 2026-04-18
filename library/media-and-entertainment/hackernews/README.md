# Hacker News CLI

Hacker News from your terminal — browse, search, and analyze stories with pipe-friendly output

Learn more at [Hacker News](https://news.ycombinator.com).

## Install

### Go

```
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/hackernews/cmd/hackernews-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Quick Start

```bash
# Check that everything works
hackernews-pp-cli doctor

# What's on the front page right now
hackernews-pp-cli stories

# What hit HN while you were away
hackernews-pp-cli since 8h

# Sync stories locally for fast search
hackernews-pp-cli sync

# Search synced data
hackernews-pp-cli search "rust"
```

## Unique Features

These capabilities aren't available in any other tool for this API.

- **`tldr`** — Condense a massive comment thread into key arguments, consensus, and dissenting views
- **`pulse`** — See what HN is saying about any topic this week — frequency, points, and velocity
- **`hiring-stats`** — See the most requested languages, remote %, and salary mentions across months of Who's Hiring threads
- **`my`** — Track your submissions — which got traction, which died, your average score, and best posting time

## Commands

### Browsing

| Command | Description |
|---------|-------------|
| `stories` | Top and best stories on Hacker News |
| `stories top` | Current top stories |
| `stories new` | Newest stories |
| `stories get <id>` | Details for a specific item |
| `ask` | Latest Ask HN posts |
| `show` | Latest Show HN posts |
| `jobs` | Latest job postings |
| `comments <id>` | Read comment threads with indentation and filtering |

### Discovery

| Command | Description |
|---------|-------------|
| `since <duration>` | Stories from the last N hours/days (e.g., `2h`, `7d`) |
| `controversial` | Stories with high comment-to-point ratios |
| `pulse <topic>` | Activity timeline and stats for a topic |
| `repost <url>` | Check if a URL has been posted before |
| `hiring` | Filter the latest Who is Hiring thread |

### Analysis

| Command | Description |
|---------|-------------|
| `tldr <id>` | Thread digest — key takes, active commenters, controversy score |
| `hiring-stats` | Aggregate hiring trends across Who's Hiring threads |
| `my <username>` | Track a user's submissions and posting stats |

### Search

| Command | Description |
|---------|-------------|
| `search <query>` | Full-text search across synced data or live API |
| `search query <query>` | Search via Algolia ranked by relevance |
| `search by-date <query>` | Search via Algolia sorted by date |

### Data & Utilities

| Command | Description |
|---------|-------------|
| `sync` | Sync data to local SQLite for offline search |
| `export` | Export data to JSONL or JSON |
| `import` | Import data from JSONL file |
| `users <username>` | Look up a user's profile and karma |
| `api` | Browse all raw API endpoints |
| `auth` | Manage authentication tokens |
| `doctor` | Check CLI health |
| `workflow archive` | Sync all resources for offline access |
| `workflow status` | Show local sync state |

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
hackernews-pp-cli stories

# JSON for scripting and agents
hackernews-pp-cli stories --json

# Filter to specific fields
hackernews-pp-cli stories --json --select title,url,score

# CSV for spreadsheets
hackernews-pp-cli stories --csv

# Compact mode for minimal token usage
hackernews-pp-cli stories --compact

# Dry run — show the request without sending
hackernews-pp-cli stories --dry-run

# Agent mode — JSON + compact + no prompts in one flag
hackernews-pp-cli stories --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** — never prompts, every input is a flag
- **Pipeable** — `--json` output to stdout, errors to stderr
- **Filterable** — `--select title,url,score` returns only fields you need
- **Previewable** — `--dry-run` shows the request without sending
- **Confirmable** — `--yes` for explicit confirmation of destructive actions
- **Cacheable** — GET responses cached for 5 minutes, bypass with `--no-cache`
- **Agent-safe by default** — no colors or formatting unless `--human-friendly` is set
- **Progress events** — paginated commands emit NDJSON events to stderr in default mode

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use as MCP Server

This CLI ships a companion MCP server for use with Claude Desktop, Cursor, and other MCP-compatible tools.

### Claude Code

```bash
claude mcp add hackernews hackernews-pp-mcp
```

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "hackernews": {
      "command": "hackernews-pp-mcp"
    }
  }
}
```

## Cookbook

```bash
# What's trending right now — top 10 stories
hackernews-pp-cli stories --limit 10

# Stories from the last 2 hours sorted by velocity
hackernews-pp-cli since 2h --sort velocity

# High-signal stories from the last week
hackernews-pp-cli since 7d --min-points 200 --limit 20

# Most controversial discussions today
hackernews-pp-cli controversial --since 24h --min-comments 50

# Search HN by date for recent Go articles
hackernews-pp-cli search by-date "golang" --tags story

# Who's hiring — filter for remote Rust jobs
hackernews-pp-cli hiring --pattern "rust" --remote

# Hiring trends across months
hackernews-pp-cli hiring-stats --months 6

# Thread digest for a big discussion
hackernews-pp-cli tldr 12345678

# Topic activity over the past week
hackernews-pp-cli pulse "AI agents"

# Check if your blog post was already shared
hackernews-pp-cli repost "https://example.com/my-post"

# Your posting stats
hackernews-pp-cli my dang

# Sync and search offline
hackernews-pp-cli sync && hackernews-pp-cli search "database"

# Export stories for analysis
hackernews-pp-cli export --format jsonl > hn-stories.jsonl

# Pipe to jq for custom filtering
hackernews-pp-cli stories --json | jq '.results[] | select(.score > 100)'
```

## Health Check

```bash
hackernews-pp-cli doctor
```

```
  OK Config: ok
  WARN Auth: not required
  OK API: reachable
  config_path: /Users/you/.config/hackernews-pp-cli/config.toml
  base_url: https://hacker-news.firebaseio.com/v0
  version: 1.0.0
```

## Configuration

Config file: `~/.config/hackernews-pp-cli/config.toml`

Environment variables:

| Variable | Description |
|----------|-------------|
| `HACKERNEWS_CONFIG` | Override config file path |
| `HACKERNEWS_BASE_URL` | Override API base URL (default: `https://hacker-news.firebaseio.com/v0`) |

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `hackernews-pp-cli doctor` to check credentials

**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run `stories` or `search` to find valid item IDs

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- If persistent, wait a few minutes and try again

**API errors (exit code 5)**
- Check `hackernews-pp-cli doctor` for connectivity
- Use `--data-source local` to fall back to synced data

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**circumflex**](https://github.com/bensadeh/circumflex) — Go
- [**haxor-news**](https://github.com/donnemartin/haxor-news) — Python
- [**mcp-hacker-news**](https://github.com/paabloLC/mcp-hacker-news) — TypeScript
- [**hackernews-mcp**](https://github.com/Malayke/hackernews-mcp) — Python

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)

<!-- pr-218-features -->
## Agent workflow features

This CLI was patched to add these agent-workflow capabilities (see [`printing-press patch`](https://github.com/mvanhorn/cli-printing-press/pull/221)):

- **Named profiles** — save a set of flags under a name and reuse them: `hackernews-pp-cli profile save <name> --<flag> <value>`, then `hackernews-pp-cli --profile <name> <command>`. Flag precedence: explicit flag > env var > profile > default.
- **`--deliver`** — route command output to a sink other than stdout. Values: `file:<path>` writes atomically via tmp+rename; `webhook:<url>` POSTs as JSON (or NDJSON with `--compact`).
- **`feedback`** — record in-band feedback about the CLI. Entries append as JSON lines to `~/.hackernews-pp-cli/feedback.jsonl`. When `HACKERNEWS_FEEDBACK_ENDPOINT` is set and either `--send` is passed or `HACKERNEWS_FEEDBACK_AUTO_SEND=true`, the entry is also POSTed upstream.
