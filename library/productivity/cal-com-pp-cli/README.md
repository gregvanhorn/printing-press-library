# Cal.com CLI

Manage bookings, event types, schedules, and availability via the Cal.com API.

## Install

### Go

```bash
go install github.com/mvanhorn/printing-press-library/cal-com-pp-cli/cmd/cal-com-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Authentication

Get your API key from [Cal.com Settings > Developer > API Keys](https://app.cal.com/settings/developer/api-keys).

```bash
# Option 1: Environment variable
export CAL_COM_TOKEN="cal_live_abc123..."

# Option 2: Save to config file
cal-com-pp-cli auth set-token cal_live_abc123...
```

For self-hosted Cal.com instances, override the base URL:

```bash
export CAL_COM_BASE_URL="https://cal.example.com"
```

Config file location: `~/.config/cal-com-pp-cli/config.toml`

## Quick Start

```bash
# 1. Verify your setup
cal-com-pp-cli doctor

# 2. Sync your data locally for offline search and analytics
cal-com-pp-cli sync

# 3. See today's schedule with attendee details and meeting links
cal-com-pp-cli today

# 4. Check for double-bookings in the next 7 days
cal-com-pp-cli conflicts

# 5. Search across all your bookings
cal-com-pp-cli search "design review"
```

## Unique Features

These capabilities aren't available in any other Cal.com tool.

- **`conflicts`** — Find double-bookings and overlapping events across all your calendars
- **`today`** — See today's bookings with attendee details, conferencing links, and prep notes
- **`stats`** — Track booking volume, busiest hours, and cancellation rates over time
- **`noshow`** — Identify which event types and time slots have the highest no-show rates
- **`search`** — Instant full-text search across all bookings, even offline
- **`gaps`** — Find unbooked availability windows to optimize your schedule
- **`workload`** — See booking distribution across team members to tune round-robin weights
- **`stale`** — Find event types with no recent bookings to clean up unused scheduling pages

## Commands

### Scheduling

| Command | Description |
|---------|-------------|
| `bookings` | List, create, cancel, reschedule, and manage bookings |
| `event-types` | Create, update, and delete event types |
| `slots` | Check available time slots, reserve and manage slot holds |
| `schedules` | Create and manage availability schedules |
| `calendars` | Connect calendars, list events, check free/busy times |
| `conferencing` | List and manage conferencing app integrations |

### Insights & Analytics

| Command | Description |
|---------|-------------|
| `today` | Today's bookings with attendee details and conferencing links |
| `conflicts` | Detect double-bookings and overlapping events |
| `stats` | Booking volume, busiest hours, cancellation rates over time |
| `noshow` | No-show patterns by event type, day, and time slot |
| `gaps` | Find unbooked availability windows in your schedule |
| `workload` | Booking distribution across team members |
| `stale` | Event types with no recent bookings |
| `analytics` | Custom queries on locally synced data |

### Data & Sync

| Command | Description |
|---------|-------------|
| `sync` | Sync API data to local SQLite for offline search and analysis |
| `search` | Full-text search across synced bookings, event types, and more |
| `export` | Export data to JSONL or JSON for backup or migration |
| `import` | Import data from JSONL file via API create calls |
| `tail` | Stream live changes by polling the API |

### Account & Config

| Command | Description |
|---------|-------------|
| `me` | View and update your profile |
| `teams` | Create, update, and manage teams |
| `webhooks` | Create, update, and manage webhook subscriptions |
| `stripe` | Check Stripe connection, save credentials |
| `verified-resources` | Manage verified emails and phone numbers |
| `oauth-clients` | Create and manage OAuth clients |
| `organizations` | Organization-level resources (teams, members, roles, attributes) |

### Utilities

| Command | Description |
|---------|-------------|
| `doctor` | Check CLI health, auth, and API connectivity |
| `auth` | Set token, check status, logout |
| `api` | Browse all API endpoints by resource |
| `workflow` | Compound workflows (archive, status) |

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
cal-com-pp-cli bookings

# JSON for scripting and agents
cal-com-pp-cli bookings --json

# Filter to specific fields
cal-com-pp-cli bookings --json --select id,status,start

# CSV for spreadsheets
cal-com-pp-cli event-types --csv

# Compact mode (key fields only, minimal tokens)
cal-com-pp-cli bookings --compact

# Dry run — show the request without sending
cal-com-pp-cli bookings create --dry-run

# Agent mode — JSON + compact + no prompts in one flag
cal-com-pp-cli today --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** -- never prompts, every input is a flag
- **Pipeable** -- `--json` output to stdout, errors to stderr
- **Filterable** -- `--select id,name` returns only fields you need
- **Previewable** -- `--dry-run` shows the request without sending
- **Confirmable** -- `--yes` for explicit confirmation of destructive actions
- **Cacheable** -- GET responses cached for 5 minutes, bypass with `--no-cache`
- **Agent-safe by default** -- no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Cookbook

```bash
# See your full schedule for today
cal-com-pp-cli today

# Show today's schedule for a specific date
cal-com-pp-cli today --date 2026-04-10

# Check for scheduling conflicts in the next 30 days
cal-com-pp-cli conflicts --days 30

# Find available slots for an event type
cal-com-pp-cli slots --event-type-id 12345 --start 2026-04-07T09:00:00Z --end 2026-04-07T17:00:00Z

# List upcoming bookings filtered by status
cal-com-pp-cli bookings --status upcoming --take 20

# Get booking stats for the last 30 days
cal-com-pp-cli stats --period 30d

# Find event types nobody has booked in 60 days
cal-com-pp-cli stale --days 60

# Analyze no-show patterns to optimize your schedule
cal-com-pp-cli noshow --json

# See team workload distribution
cal-com-pp-cli workload --period 30d

# Find availability gaps of at least 60 minutes
cal-com-pp-cli gaps --min-gap 60 --days 14

# Sync data, then search offline
cal-com-pp-cli sync
cal-com-pp-cli search "onboarding call" --data-source local

# Export all bookings for backup
cal-com-pp-cli export bookings --format jsonl -o bookings.jsonl

# Stream new bookings as NDJSON
cal-com-pp-cli tail bookings --interval 30s | jq '.status'

# Create an event type (with dry-run preview)
cal-com-pp-cli event-types create --title "30min Discovery" --slug discovery-30 --length 30 --dry-run
```

## Health Check

```bash
$ cal-com-pp-cli doctor
  OK Config: ok
  OK Auth: configured
  OK API: reachable
  OK Credentials: valid
  config_path: ~/.config/cal-com-pp-cli/config.toml
  base_url: https://api.cal.com
  auth_source: env:CAL_COM_TOKEN
  version: 1.0.0
```

## Configuration

Config file: `~/.config/cal-com-pp-cli/config.toml`

Environment variables:

| Variable | Description |
|----------|-------------|
| `CAL_COM_TOKEN` | API key (Bearer token) from Cal.com Settings > Developer > API Keys |
| `CAL_COM_BASE_URL` | Override API base URL (default: `https://api.cal.com`). Use for self-hosted instances |
| `CAL_COM_CONFIG` | Override config file path (default: `~/.config/cal-com-pp-cli/config.toml`) |

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `cal-com-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $CAL_COM_TOKEN`
- Ensure the key has not expired in Cal.com Settings > Developer > API Keys

**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the list command to see available items (e.g., `cal-com-pp-cli bookings`)

**Rate limit errors (exit code 7)**
- The CLI auto-retries with adaptive backoff
- Use `--rate-limit 2` to cap requests per second
- If persistent, wait a few minutes and try again

**Sync or search errors**
- Run `cal-com-pp-cli sync --full` to rebuild the local database
- Check disk space at `~/.local/share/cal-com-pp-cli/data.db`

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**@calcom/cal-mcp**](https://github.com/calcom/cal-mcp) -- TypeScript
- [**Composio Cal Toolkit**](https://github.com/ComposioHQ/composio) -- Python
- [**calendly-cli**](https://github.com/iloveitaly/calendly-cli) -- Python
- [**@calcom/sdk**](https://github.com/calcom/cal.com) -- TypeScript
- [**@modelcontext/cal-com-api-v2**](https://github.com/anthropics/claude-plugins-official) -- TypeScript

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
