# Meta Ads CLI

Meta Marketing API CLI with a local SQLite brain and a ROAS/frequency/utilization budget autopilot.

## Install

### Go

```
go install github.com/mvanhorn/printing-press-library/library/meta-ads/cmd/meta-ads-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Quick Start

### 1. Install

See [Install](#install) above.

### 2. Set Up Credentials

Generate a long-lived access token from
[Meta for Developers → Business → System Users](https://business.facebook.com/settings/system-users)
or from the
[Graph API Explorer](https://developers.facebook.com/tools/explorer/)
with the `ads_read`, `ads_management`, and `business_management` scopes.

```bash
export META_ACCESS_TOKEN="your-token-here"
```

Or store it in the config file via:

```bash
meta-ads-pp-cli auth set-token YOUR_TOKEN_HERE
```

### 3. Verify Setup

```bash
meta-ads-pp-cli doctor
```

This checks your configuration, credentials, and local store.

### 4. Pull Everything Into the Local SQLite Brain

```bash
meta-ads-pp-cli sync
```

This walks the hierarchy (accounts → campaigns → ad sets → ads → creatives) into
`~/.local/share/meta-ads-pp-cli/db.sqlite` and enables offline search, SQL, and
the recommendation engine.

### 5. Ask the Autopilot What to Do

```bash
meta-ads-pp-cli recommend --roas-target 3.0 --strategy moderate --json
```

You get INCREASE/DECREASE/HOLD/FLAG for every campaign with a reasoning envelope,
confidence level, and proposed dollar amount — agent-ready.

## Unique Features

These capabilities aren't available in any other tool for this API.

- **`recommend`** — Per-campaign INCREASE/DECREASE/HOLD/FLAG recommendations based on ROAS vs target, frequency saturation, and budget utilization. Three strategies (conservative/moderate/aggressive) with per-campaign caps and learning-period awareness.
- **`apply`** — Applies a batch of budget changes with a required CONFIRM token, logs each change with validation KPIs (expected ROAS lift, expected spend shift) and a follow-up date so you can audit outcomes later.
- **`verify roas`** — Detects when Meta returns both `omni_purchase` and `purchase` in the same response, which causes dashboards to double-count revenue. Reports the legacy-combined vs deduplicated ROAS delta per campaign.
- **`capacity`** — Per-campaign has-headroom/confidence derived from frequency, budget utilization, and delivery status. Defaults: saturation at freq>3.5, warning at freq≥2.5, low-utilization (<70%) means NOT budget-constrained.
- **`history`** — FTS5-indexed audit log of every budget decision and status change. `history search` finds decisions by reasoning text; `history due` lists applied decisions whose follow-up window has passed and have no analysis yet.
- **`decision-review`** — Attaches an outcome analysis to a past decision log entry: success/partial/failure/inconclusive, with KPI results, observation, and hypothesis for the next pass.
- **`fatigue`** — Finds ads where CPM is rising and frequency is rising and CTR is falling over a rolling window. Sorts by severity.
- **`pace`** — Hourly spend-rate vs daily_budget projection with ETA-to-cap per campaign.
- **`query`** — Run arbitrary SQL or one-shot queries against the local SQLite store. FTS5 tables available for names and decision reasoning.
- **`sync`** — `sync --since 4h` reads the account activities edge for changes since your last sync and refetches only touched entities. `sync --full` walks the whole hierarchy.

## Usage

<!-- HELP_OUTPUT -->

## Commands

### accounts

Ad accounts: listing, details, pages, activity log

- **`meta-ads-pp-cli accounts activities`** - List account-level change events (change log)
- **`meta-ads-pp-cli accounts get`** - Get ad account details
- **`meta-ads-pp-cli accounts list`** - List accessible ad accounts
- **`meta-ads-pp-cli accounts pages`** - List Facebook pages associated with this ad account
- **`meta-ads-pp-cli accounts users`** - List assigned users on the ad account

### ads

Ads: the creative-bearing entity. CRUD, insights, previews.

- **`meta-ads-pp-cli ads create`** - Create a new ad (PAUSED by default)
- **`meta-ads-pp-cli ads get`** - Get a single ad
- **`meta-ads-pp-cli ads insights`** - Ad-level insights with attribution dedup
- **`meta-ads-pp-cli ads list_by_account`** - List ads under an ad account
- **`meta-ads-pp-cli ads list_by_adset`** - List ads inside a specific ad set
- **`meta-ads-pp-cli ads previews`** - Render an ad across Meta placements
- **`meta-ads-pp-cli ads update`** - Update an ad

### adsets

Ad sets: targeting, budget, optimization — CRUD and insights

- **`meta-ads-pp-cli adsets create`** - Create a new ad set (PAUSED by default)
- **`meta-ads-pp-cli adsets delete`** - Delete an ad set
- **`meta-ads-pp-cli adsets get`** - Get a single ad set
- **`meta-ads-pp-cli adsets insights`** - Ad set insights with attribution dedup
- **`meta-ads-pp-cli adsets list_by_account`** - List ad sets in an ad account
- **`meta-ads-pp-cli adsets list_by_campaign`** - List ad sets inside a specific campaign
- **`meta-ads-pp-cli adsets update`** - Update an ad set

### audiences

Custom and lookalike audiences

- **`meta-ads-pp-cli audiences create`** - Create a custom audience
- **`meta-ads-pp-cli audiences create_lookalike`** - Create a lookalike audience from a source
- **`meta-ads-pp-cli audiences get`** - Get audience details and health status
- **`meta-ads-pp-cli audiences list`** - List custom audiences

### campaigns

Campaigns: list, CRUD, insights, budget schedules, diagnostics

- **`meta-ads-pp-cli campaigns budget_schedules`** - List scheduled budget increases for a campaign
- **`meta-ads-pp-cli campaigns create`** - Create a new campaign (PAUSED by default)
- **`meta-ads-pp-cli campaigns create_budget_schedule`** - Schedule a budget increase window
- **`meta-ads-pp-cli campaigns delete`** - Delete a campaign permanently
- **`meta-ads-pp-cli campaigns get`** - Get a single campaign
- **`meta-ads-pp-cli campaigns insights`** - Campaign-level insights with attribution dedup
- **`meta-ads-pp-cli campaigns list`** - List campaigns under an ad account
- **`meta-ads-pp-cli campaigns update`** - Update a campaign's name/budget/status

### creatives

Ad creatives: the headlines, images, CTAs, videos

- **`meta-ads-pp-cli creatives create`** - Create a new creative
- **`meta-ads-pp-cli creatives get`** - Get a single creative
- **`meta-ads-pp-cli creatives list_by_account`** - List all creatives for an ad account
- **`meta-ads-pp-cli creatives list_by_ad`** - List creatives attached to an ad
- **`meta-ads-pp-cli creatives list_images`** - List uploaded images in an ad account
- **`meta-ads-pp-cli creatives update`** - Update a creative's text fields
- **`meta-ads-pp-cli creatives upload_image`** - Upload an image and get an image_hash

### insights

Unified insights endpoint at the account level

- **`meta-ads-pp-cli insights account`** - Account-level performance insights

### targeting

Interest, behavior, geo, and demographic targeting search

- **`meta-ads-pp-cli targeting behaviors`** - List behavior targeting options
- **`meta-ads-pp-cli targeting demographics`** - List demographic targeting options
- **`meta-ads-pp-cli targeting geo`** - Search Meta's geographic targeting database
- **`meta-ads-pp-cli targeting interests`** - Search Meta's interest targeting catalog
- **`meta-ads-pp-cli targeting suggestions`** - Get interest suggestions based on seed interests
- **`meta-ads-pp-cli targeting validate`** - Validate interests by name or ID

### Autopilot & analysis

Hand-authored commands that reason over the local store — not thin wrappers around a single endpoint.

- **`meta-ads-pp-cli recommend`** - Budget autopilot: ROAS + frequency + utilization decisioning
- **`meta-ads-pp-cli apply`** - Batch-apply budget changes with validation KPIs and follow-up log
- **`meta-ads-pp-cli verify roas`** - Detect omni_purchase vs purchase double-counting in local insights
- **`meta-ads-pp-cli capacity`** - Per-campaign frequency capacity signal (headroom/confidence/details)
- **`meta-ads-pp-cli history list`** - List recent decisions (with filters for platform and entry type)
- **`meta-ads-pp-cli history search`** - Full-text search the decision log (FTS5)
- **`meta-ads-pp-cli history due`** - Applied decisions whose follow-up window has passed and have no analysis yet
- **`meta-ads-pp-cli decision-review`** - Attach a post-mortem analysis to a past decision (alias of `history review`)
- **`meta-ads-pp-cli query`** - Run SQL (SELECT/PRAGMA/EXPLAIN) against the local store
- **`meta-ads-pp-cli sync`** - Pull the hierarchy + insights into local SQLite

### Deferred transcendence (approved stubs)

These ship as labeled stubs this iteration and document the manual workaround in their `--help`.

- **`meta-ads-pp-cli fatigue`** - Detect creative fatigue (rising CPM × rising freq × falling CTR)
- **`meta-ads-pp-cli pace`** - Budget pacing monitor: hourly burn rate to ETA-to-cap
- **`meta-ads-pp-cli learning`** - Identify campaigns in the Meta Smart Bidding learning phase
- **`meta-ads-pp-cli overlap`** - Audience overlap analysis between two ad sets
- **`meta-ads-pp-cli alerts`** - Threshold watchers against local data (ROAS min, freq max)
- **`meta-ads-pp-cli rollup`** - Aggregate spend/ROAS/conversions across multiple ad accounts


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
meta-ads-pp-cli accounts list

# JSON for scripting and agents
meta-ads-pp-cli accounts list --json

# Filter to specific fields
meta-ads-pp-cli accounts list --json --select id,name,status

# Dry run — show the request without sending
meta-ads-pp-cli accounts list --dry-run

# Agent mode — JSON + compact + no prompts in one flag
meta-ads-pp-cli accounts list --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Retryable** - creates return "already exists" on retry, deletes return "already deleted"
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - `echo '{"key":"value"}' | meta-ads-pp-cli <resource> create --stdin`
- **Cacheable** - GET responses cached for 5 minutes, bypass with `--no-cache`
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set
- **Progress events** - paginated commands emit NDJSON events to stderr in default mode

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use as MCP Server

This CLI ships a companion MCP server for use with Claude Desktop, Cursor, and other MCP-compatible tools.

### Claude Code

```bash
claude mcp add meta-ads meta-ads-pp-mcp -e META_ACCESS_TOKEN=<your-token>
```

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "meta-ads": {
      "command": "meta-ads-pp-mcp",
      "env": {
        "META_ACCESS_TOKEN": "<your-key>"
      }
    }
  }
}
```

## Cookbook

Common workflows and recipes:

```bash
# List ad accounts you can access, keeping only the fields you actually use
meta-ads-pp-cli accounts list --json --select id,name,account_status,currency

# Campaigns under an account, narrowed to scheduling + budget
meta-ads-pp-cli campaigns list --account-id act_1234567890 \
    --json --select id,name,objective,daily_budget,effective_status

# Campaign-level insights with attribution dedup (avoids omni_purchase double-count)
meta-ads-pp-cli campaigns insights 120123456789 --date-preset last_30d --json

# Pull everything into local SQLite (used by recommend/verify/history/query)
meta-ads-pp-cli sync --full

# Recommend budget changes — the flagship autopilot
meta-ads-pp-cli recommend --days 14 --strategy moderate --roas-target 3.0 --json \
    > recommendations.json

# Preview the API calls the autopilot would make (no writes)
meta-ads-pp-cli apply --from-recommendations recommendations.json \
    --confirmation CONFIRM --dry-run

# Apply for real and audit-log each change with a follow-up date
meta-ads-pp-cli apply --from-recommendations recommendations.json \
    --confirmation CONFIRM

# Detect omni_purchase vs purchase double-counting in local insights
meta-ads-pp-cli verify roas --json \
    --select rows.campaign_name,rows.roas_deduplicated,rows.roas_legacy,rows.legacy_delta_pct

# Applied decisions whose follow-up window has passed and have no analysis yet
meta-ads-pp-cli history due --platform meta_ads --json

# Attach a post-mortem to a decision so the next run learns from outcomes
meta-ads-pp-cli decision-review <log-id> --outcome partial \
    --summary "ROAS lifted 8% but spend lagged" \
    --observation "Audience too narrow; utilization stuck at 60%" \
    --hypothesis "Expand lookalike ratio from 1% to 2%"

# Offline SQL over the local store — ad hoc analytics without hitting the API
meta-ads-pp-cli query "SELECT json_extract(data,'\$.campaign_name') AS name, \
    CAST(json_extract(data,'\$.spend') AS REAL) AS spend \
    FROM insights ORDER BY spend DESC LIMIT 10" --json

# Per-campaign frequency capacity: headroom, confidence, reasoning details
meta-ads-pp-cli capacity --days 7 --agent

# Full-text search decision history
meta-ads-pp-cli history search "saturation" --limit 20

# Export the local SQLite snapshot for backup or cross-host analysis
meta-ads-pp-cli export --format jsonl > backup.jsonl
```

## Health Check

```bash
meta-ads-pp-cli doctor
```

<!-- DOCTOR_OUTPUT -->

## Configuration

Config file: `~/.config/meta-ads-pp-cli/config.toml` (overridable with `--config <path>` or the `META_ADS_CONFIG` env var)

Local store: `~/.local/share/meta-ads-pp-cli/db.sqlite`

Response cache: `~/.cache/meta-ads-pp-cli/` (5-minute TTL, bypass with `--no-cache`)

Environment variables:

| Variable | Purpose |
|----------|---------|
| `META_ACCESS_TOKEN` | Long-lived user or system-user access token (required) |
| `META_ADS_CONFIG`   | Override the config file path |
| `META_ADS_BASE_URL` | Override the Graph API base URL (used by tests / mock servers) |

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `meta-ads-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $META_ACCESS_TOKEN`

**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- If persistent, wait a few minutes and try again

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**pipeboard-co/meta-ads-mcp**](https://github.com/pipeboard-co/meta-ads-mcp) — Python
- [**brijr/meta-mcp**](https://github.com/brijr/meta-mcp) — TypeScript
- [**hashcott/meta-ads-mcp-server**](https://github.com/hashcott/meta-ads-mcp-server) — TypeScript
- [**gomarble-ai/facebook-ads-mcp-server**](https://github.com/gomarble-ai/facebook-ads-mcp-server) — Python
- [**attainmentlabs/meta-ads-cli**](https://github.com/attainmentlabs/meta-ads-cli) — Python

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
