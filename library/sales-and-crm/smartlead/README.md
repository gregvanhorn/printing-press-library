# Smartlead CLI

**Every Smartlead feature, plus offline sync, FTS5 search, auto-chunking, and cross-campaign aggregations no other Smartlead tool has.**

smartlead-pp-cli mirrors all 155 Smartlead endpoints as agent-native commands and adds a local SQLite store so cross-campaign queries (stale, overlap, mailbox-burnout, deliv-drift, agency roll-up) run in milliseconds without re-paginating. Every command supports --json, --select, --csv, and a typed exit-code palette so it composes with jq and shell pipelines.

Learn more at [Smartlead](https://smartlead.ai).

Printed by [@gregvanhorn](https://github.com/gregvanhorn) (Greg Van Horn).

## Install

The recommended path installs both the `smartlead-pp-cli` binary and the `pp-smartlead` agent skill in one shot:

```bash
npx -y @mvanhorn/printing-press install smartlead
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press install smartlead --cli-only
```


### Without Node

The generated install path is category-agnostic until this CLI is published. If `npx` is not available before publish, install Node or use the category-specific Go fallback from the public-library entry after publish.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/smartlead-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-smartlead --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-smartlead --force
```

## Install for OpenClaw

Tell your OpenClaw agent (copy this):

```
Install the pp-smartlead skill from https://github.com/mvanhorn/printing-press-library/tree/main/cli-skills/pp-smartlead. The skill defines how its required CLI can be installed.
```

## Authentication

Smartlead authenticates by query string. Set `SMARTLEAD_API_KEY=<your-admin-or-client-key>` in your environment, or run `smartlead-pp-cli auth set-token` to store it. The CLI honors Smartlead's 60 requests/minute budget per key with an adaptive limiter that surfaces a typed RateLimitError when exhausted instead of returning empty results.

## Quick Start

```bash
# Confirm auth + reachability before anything else.
smartlead-pp-cli doctor


# List every campaign on the account.
smartlead-pp-cli campaigns list --json --select id,name,status


# Hydrate the local SQLite store so transcendence commands can run offline.
smartlead-pp-cli sync --full


# Find leads neglected for two weeks across every campaign at once.
smartlead-pp-cli stale --days 14 --json


# Rank senders by burnout risk so you can pause the at-risk ones.
smartlead-pp-cli mailbox-burnout --json

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Local state that compounds

- **`stale`** — Find leads not touched in N days across every campaign at once.

  _Use when an agent needs to sweep neglected leads across an entire account without paginating every campaign individually._

  ```bash
  smartlead-pp-cli stale --days 14 --json --select email,campaign_name,last_event_at
  ```
- **`overlap`** — Surface leads that exist in multiple campaigns (duplicate outreach risk).

  _Use before adding a new lead list to confirm no contact is already mid-sequence somewhere else._

  ```bash
  smartlead-pp-cli overlap --json --select email,campaign_count,campaign_names
  ```
- **`mailbox-burnout`** — Rank email accounts approaching warmup decline or sending-limit ceiling.

  _Use to pause at-risk senders before deliverability craters or 60/min rate-limits trigger 429s._

  ```bash
  smartlead-pp-cli mailbox-burnout --json --select email,warmup_score,utilization_pct
  ```
- **`reply-velocity`** — Time-to-first-reply per campaign with historical baseline.

  _Use to spot which campaigns are speeding up or slowing down vs their own baseline._

  ```bash
  smartlead-pp-cli reply-velocity --weeks 4 --json
  ```
- **`deliv-drift`** — Compare today vs 7d vs 30d spam-test placement per sender domain.

  _Use to catch deliverability decline before it becomes a crisis._

  ```bash
  smartlead-pp-cli deliv-drift --domain acme.com --json
  ```
- **`reach-budget`** — Days-to-completion per campaign from current sending-limits and pending leads.

  _Use to set realistic delivery expectations for a sales team or to size how many senders a campaign needs._

  ```bash
  smartlead-pp-cli reach-budget --campaign 3217809 --json
  ```

### Agency-scale

- **`roll-up`** — Aggregate stats across all client API keys, per-client and portfolio-wide.

  _Use when a portfolio owner wants a single view across every client account._

  ```bash
  smartlead-pp-cli roll-up --keys-from ~/.smartlead/keys.txt --json
  ```

### Agent-native plumbing

- **`reply-classify`** — Classify master-inbox replies (positive / objection / OOO / unsub) using local FTS + keywords.

  _Use to triage hundreds of replies into action buckets without reading each one._

  ```bash
  smartlead-pp-cli reply-classify --since 7d --json --select email,classification,confidence
  ```

## Usage

Run `smartlead-pp-cli --help` for the full command reference and flag list.

## Commands

### campaigns

Manage campaigns

- **`smartlead-pp-cli campaigns compare-statistics`** - Compare campaign statistics
- **`smartlead-pp-cli campaigns create`** - Create campaign
- **`smartlead-pp-cli campaigns delete`** - Delete campaign
- **`smartlead-pp-cli campaigns get`** - Get campaign by ID
- **`smartlead-pp-cli campaigns list`** - List all campaigns
- **`smartlead-pp-cli campaigns list-all-leads-activities`** - All leads activities
- **`smartlead-pp-cli campaigns list-by-lead`** - Campaigns containing a lead
- **`smartlead-pp-cli campaigns list-with-analytics`** - Campaigns with analytics
- **`smartlead-pp-cli campaigns update-lead-email-account`** - Reassign lead's sending account

### client

Manage client

- **`smartlead-pp-cli client create`** - Create client
- **`smartlead-pp-cli client create-api-key`** - Create API key
- **`smartlead-pp-cli client delete-api-key`** - Delete API key
- **`smartlead-pp-cli client list`** - List clients
- **`smartlead-pp-cli client list-api-keys`** - List API keys
- **`smartlead-pp-cli client reset-api-key`** - Reset API key

### crm

Manage crm

- **`smartlead-pp-cli crm create-lead-note`** - Create lead note
- **`smartlead-pp-cli crm create-lead-tag`** - Tag lead
- **`smartlead-pp-cli crm create-lead-task`** - Create lead task
- **`smartlead-pp-cli crm create-tag`** - Create CRM tag
- **`smartlead-pp-cli crm delete-lead-tag`** - Remove lead tag mapping
- **`smartlead-pp-cli crm list-lead-notes`** - List lead notes
- **`smartlead-pp-cli crm list-lead-tags`** - List lead tag mappings
- **`smartlead-pp-cli crm list-lead-tasks`** - List lead tasks
- **`smartlead-pp-cli crm list-tags`** - List CRM tags
- **`smartlead-pp-cli crm update-lead-task`** - Update lead task

### download-statistics

Manage download statistics

- **`smartlead-pp-cli download-statistics download_statistics`** - Download statistics

### email-accounts

Manage email accounts

- **`smartlead-pp-cli email-accounts bulk-update`** - Bulk update
- **`smartlead-pp-cli email-accounts create`** - Add SMTP/IMAP account
- **`smartlead-pp-cli email-accounts create-tag-mapping`** - Map tags to emails
- **`smartlead-pp-cli email-accounts delete`** - Delete account
- **`smartlead-pp-cli email-accounts delete-tag-mapping`** - Remove tag mapping
- **`smartlead-pp-cli email-accounts get`** - Get email account
- **`smartlead-pp-cli email-accounts list`** - List email accounts
- **`smartlead-pp-cli email-accounts reconnect`** - Reconnect
- **`smartlead-pp-cli email-accounts suspend`** - Suspend account
- **`smartlead-pp-cli email-accounts unsuspend`** - Unsuspend account
- **`smartlead-pp-cli email-accounts update`** - Update email account

### email-campaigns

Manage email campaigns

- **`smartlead-pp-cli email-campaigns forward-reply-email`** - Forward reply email

### lead-list

Manage lead list

- **`smartlead-pp-cli lead-list assign-tags`** - Assign tags to lead list
- **`smartlead-pp-cli lead-list create`** - Create lead list
- **`smartlead-pp-cli lead-list delete`** - Delete lead list
- **`smartlead-pp-cli lead-list get`** - Get lead list
- **`smartlead-pp-cli lead-list list`** - List lead lists
- **`smartlead-pp-cli lead-list update`** - Update lead list

### leads

Manage leads

- **`smartlead-pp-cli leads add-domain-blocklist`** - Block domain
- **`smartlead-pp-cli leads add-to-global-blocklist`** - Add to global blocklist
- **`smartlead-pp-cli leads delete`** - Delete lead
- **`smartlead-pp-cli leads get-by-email`** - Search lead by email
- **`smartlead-pp-cli leads get-global-blocklist`** - Get global blocklist
- **`smartlead-pp-cli leads list`** - List leads
- **`smartlead-pp-cli leads list-categories`** - Lead categories
- **`smartlead-pp-cli leads push-between-lists`** - Move leads between lists
- **`smartlead-pp-cli leads push-to-campaign`** - Push leads to campaign
- **`smartlead-pp-cli leads update`** - Update lead

### master-inbox

Manage master inbox

- **`smartlead-pp-cli master-inbox change-read-status`** - Change read status
- **`smartlead-pp-cli master-inbox create-note`** - Create inbox note
- **`smartlead-pp-cli master-inbox create-task`** - Create inbox task
- **`smartlead-pp-cli master-inbox get-item`** - Get inbox item
- **`smartlead-pp-cli master-inbox list-archived`** - List archived
- **`smartlead-pp-cli master-inbox list-important`** - List important
- **`smartlead-pp-cli master-inbox list-reminders`** - List reminders
- **`smartlead-pp-cli master-inbox list-replies`** - List inbox replies
- **`smartlead-pp-cli master-inbox list-scheduled`** - List scheduled
- **`smartlead-pp-cli master-inbox list-sent`** - List sent
- **`smartlead-pp-cli master-inbox list-snoozed`** - List snoozed
- **`smartlead-pp-cli master-inbox list-unread-replies`** - List unread replies
- **`smartlead-pp-cli master-inbox push-to-subsequence`** - Push to subsequence
- **`smartlead-pp-cli master-inbox resume-lead`** - Resume lead from inbox
- **`smartlead-pp-cli master-inbox set-reminder`** - Set reminder
- **`smartlead-pp-cli master-inbox update-category`** - Update inbox category
- **`smartlead-pp-cli master-inbox update-revenue`** - Update revenue
- **`smartlead-pp-cli master-inbox update-team-member`** - Assign team member

### smart-senders

Manage smart senders

- **`smartlead-pp-cli smart-senders auto-generate-mailboxes`** - Auto-generate mailboxes
- **`smartlead-pp-cli smart-senders get-analytics-dashboard`** - Smart sender analytics dashboard
- **`smartlead-pp-cli smart-senders get-health-monitoring`** - Smart sender health monitoring
- **`smartlead-pp-cli smart-senders get-performance-metrics`** - Smart sender performance metrics
- **`smartlead-pp-cli smart-senders get-reputation-scores`** - Smart sender reputation scores
- **`smartlead-pp-cli smart-senders list-domains`** - List smart sender domains
- **`smartlead-pp-cli smart-senders list-vendors`** - List smart sender vendors
- **`smartlead-pp-cli smart-senders place-order`** - Place smart sender order
- **`smartlead-pp-cli smart-senders search-domain`** - Search smart sender domains

### smartlead-analytics

Manage smartlead analytics

- **`smartlead-pp-cli smartlead-analytics get-campaign-follow-up-reply-rate`** - Follow-up reply rate
- **`smartlead-pp-cli smartlead-analytics get-campaign-lead-to-reply-time`** - Lead to reply time
- **`smartlead-pp-cli smartlead-analytics get-campaign-leads-take-for-first-reply`** - Leads take for first reply
- **`smartlead-pp-cli smartlead-analytics get-campaign-list`** - Analytics campaigns list
- **`smartlead-pp-cli smartlead-analytics get-campaign-overall-stats`** - Campaign overall stats
- **`smartlead-pp-cli smartlead-analytics get-campaign-response-stats`** - Campaign response stats
- **`smartlead-pp-cli smartlead-analytics get-campaign-status-stats`** - Campaign status stats
- **`smartlead-pp-cli smartlead-analytics get-client-list`** - Analytics clients list
- **`smartlead-pp-cli smartlead-analytics get-client-month-wise-count`** - Client month-wise count
- **`smartlead-pp-cli smartlead-analytics get-client-overall-stats`** - Client overall stats
- **`smartlead-pp-cli smartlead-analytics get-day-wise-overall-stats`** - Day-wise overall stats
- **`smartlead-pp-cli smartlead-analytics get-day-wise-positive-reply-stats`** - Day-wise positive reply stats
- **`smartlead-pp-cli smartlead-analytics get-lead-category-wise-response`** - Lead category-wise response
- **`smartlead-pp-cli smartlead-analytics get-lead-overall-stats`** - Lead overall stats
- **`smartlead-pp-cli smartlead-analytics get-mailbox-domain-wise-health-metrics`** - Mailbox domain-wise health metrics
- **`smartlead-pp-cli smartlead-analytics get-mailbox-name-wise-health-metrics`** - Mailbox name-wise health metrics
- **`smartlead-pp-cli smartlead-analytics get-mailbox-overall-stats`** - Mailbox overall stats
- **`smartlead-pp-cli smartlead-analytics get-mailbox-provider-wise-overall-performance`** - Mailbox provider-wise performance
- **`smartlead-pp-cli smartlead-analytics get-overall-stats-v2`** - Overall stats v2
- **`smartlead-pp-cli smartlead-analytics get-overview`** - Analytics overview
- **`smartlead-pp-cli smartlead-analytics get-team-board-overall-stats`** - Team-board overall stats

### spam-test

Manage spam test

- **`smartlead-pp-cli spam-test bulk-delete`** - Bulk delete spam tests
- **`smartlead-pp-cli spam-test create-automated`** - Create automated spam test
- **`smartlead-pp-cli spam-test create-folder`** - Create spam-test folder
- **`smartlead-pp-cli spam-test create-manual`** - Create manual spam test
- **`smartlead-pp-cli spam-test delete-folder`** - Delete spam-test folder
- **`smartlead-pp-cli spam-test get`** - Get spam test
- **`smartlead-pp-cli spam-test get-folder`** - Get spam-test folder
- **`smartlead-pp-cli spam-test get-geo-wise-report`** - Spam test geo-wise report
- **`smartlead-pp-cli spam-test get-provider-wise-results`** - Spam test provider-wise results
- **`smartlead-pp-cli spam-test get-report`** - Spam test report
- **`smartlead-pp-cli spam-test get-results`** - Spam test results
- **`smartlead-pp-cli spam-test get-sender-account-wise-report`** - Spam test sender-account-wise report
- **`smartlead-pp-cli spam-test list`** - List spam tests
- **`smartlead-pp-cli spam-test list-folders`** - List spam-test folders
- **`smartlead-pp-cli spam-test list-seed-providers`** - List spam-test seed providers

### webhooks

Manage webhooks

- **`smartlead-pp-cli webhooks create`** - Create webhook
- **`smartlead-pp-cli webhooks delete`** - Delete webhook
- **`smartlead-pp-cli webhooks get`** - Get webhook
- **`smartlead-pp-cli webhooks get-publish-summary`** - Webhook publish summary
- **`smartlead-pp-cli webhooks list`** - List webhooks
- **`smartlead-pp-cli webhooks list-event-types`** - List webhook event types
- **`smartlead-pp-cli webhooks retrigger-failed-events`** - Retrigger failed webhook events
- **`smartlead-pp-cli webhooks update`** - Update webhook


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
smartlead-pp-cli campaigns list

# JSON for scripting and agents
smartlead-pp-cli campaigns list --json

# Filter to specific fields
smartlead-pp-cli campaigns list --json --select id,name,status

# Dry run — show the request without sending
smartlead-pp-cli campaigns list --dry-run

# Agent mode — JSON + compact + no prompts in one flag
smartlead-pp-cli campaigns list --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries and `--ignore-missing` to delete retries when a no-op success is acceptable
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use with Claude Code

Install the focused skill — it auto-installs the CLI on first invocation:

```bash
npx skills add mvanhorn/printing-press-library/cli-skills/pp-smartlead -g
```

Then invoke `/pp-smartlead <query>` in Claude Code. The skill is the most efficient path — Claude Code drives the CLI directly without an MCP server in the middle.

<details>
<summary>Use as an MCP server in Claude Code (advanced)</summary>

If you'd rather register this CLI as an MCP server in Claude Code, install the MCP binary first:


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Then register it:

```bash
claude mcp add smartlead smartlead-pp-mcp -e SMARTLEAD_API_KEY=<your-key>
```

</details>

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/smartlead-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `SMARTLEAD_API_KEY` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "smartlead": {
      "command": "smartlead-pp-mcp",
      "env": {
        "SMARTLEAD_API_KEY": "<your-key>"
      }
    }
  }
}
```

</details>

## Health Check

```bash
smartlead-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Config file: `~/.config/smartlead-pp-cli/config.toml`

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `SMARTLEAD_API_KEY` | per_call | Yes | Set to your API credential. |

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `smartlead-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $SMARTLEAD_API_KEY`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific

- **HTTP 429 on bulk operations** — The CLI already throttles to 60/min; if you're running multiple shells, serialize them or split workload across client API keys.
- **Lead push fails on >400 leads** — The CLI auto-chunks leads in batches of 400; if you see this, you're on an older build — rerun `go install`.
- **auth: missing SMARTLEAD_API_KEY** — Run `export SMARTLEAD_API_KEY=<key>` or `smartlead-pp-cli auth set-token`. Check `smartlead-pp-cli doctor` for resolution.

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**LeadMagic/smartlead-mcp-server**](https://github.com/LeadMagic/smartlead-mcp-server) — TypeScript (18 stars)
- [**jean-technologies/smartlead-mcp-server-local**](https://github.com/jean-technologies/smartlead-mcp-server-local) — TypeScript (17 stars)
- [**bcharleson/smartlead-cli**](https://github.com/bcharleson/smartlead-cli) — TypeScript (2 stars)
- [**jzakirov/smartlead-cli**](https://github.com/jzakirov/smartlead-cli) — Python
- [**smartlead-ai/API-Python-Library**](https://github.com/smartlead-ai/API-Python-Library) — Python

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
