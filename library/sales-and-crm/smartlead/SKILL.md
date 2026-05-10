---
name: pp-smartlead
description: "Every Smartlead feature, plus offline sync, FTS5 search, auto-chunking, and cross-campaign aggregations no other... Trigger phrases: `find stale leads in smartlead`, `check smartlead mailbox health`, `smartlead reply triage`, `smartlead campaign analytics`, `use smartlead`, `run smartlead`."
author: "Greg Van Horn"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - smartlead-pp-cli
---

# Smartlead — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `smartlead-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer:
   ```bash
   npx -y @mvanhorn/printing-press install smartlead --cli-only
   ```
2. Verify: `smartlead-pp-cli --version`
3. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.3 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/sales-and-crm/smartlead/cmd/smartlead-pp-cli@latest
```

If `--version` reports "command not found" after install, the install step did not put the binary on `$PATH`. Do not proceed with skill commands until verification succeeds.

This CLI mirrors all 155 Smartlead endpoints as agent-native commands and adds a local SQLite store so cross-campaign queries (stale, overlap, mailbox-burnout, deliv-drift, agency roll-up) run in milliseconds without re-paginating. Every command supports --json, --select, --csv, and a typed exit-code palette so it composes with jq and shell pipelines.

## When to Use This CLI

Use this CLI when an agent needs to operate across many Smartlead campaigns or clients at once, when the workflow demands cross-campaign aggregation that the Smartlead UI and API don't expose, or when you need fast offline search over leads and replies. It also fits any Smartlead automation pipeline that benefits from typed exit codes, structured JSON, and adaptive rate-limit handling.

## Unique Capabilities

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

## Command Reference

**campaigns** — Manage campaigns

- `smartlead-pp-cli campaigns compare-statistics` — Compare campaign statistics
- `smartlead-pp-cli campaigns create` — Create campaign
- `smartlead-pp-cli campaigns delete` — Delete campaign
- `smartlead-pp-cli campaigns get` — Get campaign by ID
- `smartlead-pp-cli campaigns list` — List all campaigns
- `smartlead-pp-cli campaigns list-all-leads-activities` — All leads activities
- `smartlead-pp-cli campaigns list-by-lead` — Campaigns containing a lead
- `smartlead-pp-cli campaigns list-with-analytics` — Campaigns with analytics
- `smartlead-pp-cli campaigns update-lead-email-account` — Reassign lead's sending account

**client** — Manage client

- `smartlead-pp-cli client create` — Create client
- `smartlead-pp-cli client create-api-key` — Create API key
- `smartlead-pp-cli client delete-api-key` — Delete API key
- `smartlead-pp-cli client list` — List clients
- `smartlead-pp-cli client list-api-keys` — List API keys
- `smartlead-pp-cli client reset-api-key` — Reset API key

**crm** — Manage crm

- `smartlead-pp-cli crm create-lead-note` — Create lead note
- `smartlead-pp-cli crm create-lead-tag` — Tag lead
- `smartlead-pp-cli crm create-lead-task` — Create lead task
- `smartlead-pp-cli crm create-tag` — Create CRM tag
- `smartlead-pp-cli crm delete-lead-tag` — Remove lead tag mapping
- `smartlead-pp-cli crm list-lead-notes` — List lead notes
- `smartlead-pp-cli crm list-lead-tags` — List lead tag mappings
- `smartlead-pp-cli crm list-lead-tasks` — List lead tasks
- `smartlead-pp-cli crm list-tags` — List CRM tags
- `smartlead-pp-cli crm update-lead-task` — Update lead task

**download-statistics** — Manage download statistics

- `smartlead-pp-cli download-statistics` — Download statistics

**email-accounts** — Manage email accounts

- `smartlead-pp-cli email-accounts bulk-update` — Bulk update
- `smartlead-pp-cli email-accounts create` — Add SMTP/IMAP account
- `smartlead-pp-cli email-accounts create-tag-mapping` — Map tags to emails
- `smartlead-pp-cli email-accounts delete` — Delete account
- `smartlead-pp-cli email-accounts delete-tag-mapping` — Remove tag mapping
- `smartlead-pp-cli email-accounts get` — Get email account
- `smartlead-pp-cli email-accounts list` — List email accounts
- `smartlead-pp-cli email-accounts reconnect` — Reconnect
- `smartlead-pp-cli email-accounts suspend` — Suspend account
- `smartlead-pp-cli email-accounts unsuspend` — Unsuspend account
- `smartlead-pp-cli email-accounts update` — Update email account

**email-campaigns** — Manage email campaigns

- `smartlead-pp-cli email-campaigns` — Forward reply email

**lead-list** — Manage lead list

- `smartlead-pp-cli lead-list assign-tags` — Assign tags to lead list
- `smartlead-pp-cli lead-list create` — Create lead list
- `smartlead-pp-cli lead-list delete` — Delete lead list
- `smartlead-pp-cli lead-list get` — Get lead list
- `smartlead-pp-cli lead-list list` — List lead lists
- `smartlead-pp-cli lead-list update` — Update lead list

**leads** — Manage leads

- `smartlead-pp-cli leads add-domain-blocklist` — Block domain
- `smartlead-pp-cli leads add-to-global-blocklist` — Add to global blocklist
- `smartlead-pp-cli leads delete` — Delete lead
- `smartlead-pp-cli leads get-by-email` — Search lead by email
- `smartlead-pp-cli leads get-global-blocklist` — Get global blocklist
- `smartlead-pp-cli leads list` — List leads
- `smartlead-pp-cli leads list-categories` — Lead categories
- `smartlead-pp-cli leads push-between-lists` — Move leads between lists
- `smartlead-pp-cli leads push-to-campaign` — Push leads to campaign
- `smartlead-pp-cli leads update` — Update lead

**master-inbox** — Manage master inbox

- `smartlead-pp-cli master-inbox change-read-status` — Change read status
- `smartlead-pp-cli master-inbox create-note` — Create inbox note
- `smartlead-pp-cli master-inbox create-task` — Create inbox task
- `smartlead-pp-cli master-inbox get-item` — Get inbox item
- `smartlead-pp-cli master-inbox list-archived` — List archived
- `smartlead-pp-cli master-inbox list-important` — List important
- `smartlead-pp-cli master-inbox list-reminders` — List reminders
- `smartlead-pp-cli master-inbox list-replies` — List inbox replies
- `smartlead-pp-cli master-inbox list-scheduled` — List scheduled
- `smartlead-pp-cli master-inbox list-sent` — List sent
- `smartlead-pp-cli master-inbox list-snoozed` — List snoozed
- `smartlead-pp-cli master-inbox list-unread-replies` — List unread replies
- `smartlead-pp-cli master-inbox push-to-subsequence` — Push to subsequence
- `smartlead-pp-cli master-inbox resume-lead` — Resume lead from inbox
- `smartlead-pp-cli master-inbox set-reminder` — Set reminder
- `smartlead-pp-cli master-inbox update-category` — Update inbox category
- `smartlead-pp-cli master-inbox update-revenue` — Update revenue
- `smartlead-pp-cli master-inbox update-team-member` — Assign team member

**smart-senders** — Manage smart senders

- `smartlead-pp-cli smart-senders auto-generate-mailboxes` — Auto-generate mailboxes
- `smartlead-pp-cli smart-senders get-analytics-dashboard` — Smart sender analytics dashboard
- `smartlead-pp-cli smart-senders get-health-monitoring` — Smart sender health monitoring
- `smartlead-pp-cli smart-senders get-performance-metrics` — Smart sender performance metrics
- `smartlead-pp-cli smart-senders get-reputation-scores` — Smart sender reputation scores
- `smartlead-pp-cli smart-senders list-domains` — List smart sender domains
- `smartlead-pp-cli smart-senders list-vendors` — List smart sender vendors
- `smartlead-pp-cli smart-senders place-order` — Place smart sender order
- `smartlead-pp-cli smart-senders search-domain` — Search smart sender domains

**smartlead-analytics** — Manage smartlead analytics

- `smartlead-pp-cli smartlead-analytics get-campaign-follow-up-reply-rate` — Follow-up reply rate
- `smartlead-pp-cli smartlead-analytics get-campaign-lead-to-reply-time` — Lead to reply time
- `smartlead-pp-cli smartlead-analytics get-campaign-leads-take-for-first-reply` — Leads take for first reply
- `smartlead-pp-cli smartlead-analytics get-campaign-list` — Analytics campaigns list
- `smartlead-pp-cli smartlead-analytics get-campaign-overall-stats` — Campaign overall stats
- `smartlead-pp-cli smartlead-analytics get-campaign-response-stats` — Campaign response stats
- `smartlead-pp-cli smartlead-analytics get-campaign-status-stats` — Campaign status stats
- `smartlead-pp-cli smartlead-analytics get-client-list` — Analytics clients list
- `smartlead-pp-cli smartlead-analytics get-client-month-wise-count` — Client month-wise count
- `smartlead-pp-cli smartlead-analytics get-client-overall-stats` — Client overall stats
- `smartlead-pp-cli smartlead-analytics get-day-wise-overall-stats` — Day-wise overall stats
- `smartlead-pp-cli smartlead-analytics get-day-wise-positive-reply-stats` — Day-wise positive reply stats
- `smartlead-pp-cli smartlead-analytics get-lead-category-wise-response` — Lead category-wise response
- `smartlead-pp-cli smartlead-analytics get-lead-overall-stats` — Lead overall stats
- `smartlead-pp-cli smartlead-analytics get-mailbox-domain-wise-health-metrics` — Mailbox domain-wise health metrics
- `smartlead-pp-cli smartlead-analytics get-mailbox-name-wise-health-metrics` — Mailbox name-wise health metrics
- `smartlead-pp-cli smartlead-analytics get-mailbox-overall-stats` — Mailbox overall stats
- `smartlead-pp-cli smartlead-analytics get-mailbox-provider-wise-overall-performance` — Mailbox provider-wise performance
- `smartlead-pp-cli smartlead-analytics get-overall-stats-v2` — Overall stats v2
- `smartlead-pp-cli smartlead-analytics get-overview` — Analytics overview
- `smartlead-pp-cli smartlead-analytics get-team-board-overall-stats` — Team-board overall stats

**spam-test** — Manage spam test

- `smartlead-pp-cli spam-test bulk-delete` — Bulk delete spam tests
- `smartlead-pp-cli spam-test create-automated` — Create automated spam test
- `smartlead-pp-cli spam-test create-folder` — Create spam-test folder
- `smartlead-pp-cli spam-test create-manual` — Create manual spam test
- `smartlead-pp-cli spam-test delete-folder` — Delete spam-test folder
- `smartlead-pp-cli spam-test get` — Get spam test
- `smartlead-pp-cli spam-test get-folder` — Get spam-test folder
- `smartlead-pp-cli spam-test get-geo-wise-report` — Spam test geo-wise report
- `smartlead-pp-cli spam-test get-provider-wise-results` — Spam test provider-wise results
- `smartlead-pp-cli spam-test get-report` — Spam test report
- `smartlead-pp-cli spam-test get-results` — Spam test results
- `smartlead-pp-cli spam-test get-sender-account-wise-report` — Spam test sender-account-wise report
- `smartlead-pp-cli spam-test list` — List spam tests
- `smartlead-pp-cli spam-test list-folders` — List spam-test folders
- `smartlead-pp-cli spam-test list-seed-providers` — List spam-test seed providers

**webhooks** — Manage webhooks

- `smartlead-pp-cli webhooks create` — Create webhook
- `smartlead-pp-cli webhooks delete` — Delete webhook
- `smartlead-pp-cli webhooks get` — Get webhook
- `smartlead-pp-cli webhooks get-publish-summary` — Webhook publish summary
- `smartlead-pp-cli webhooks list` — List webhooks
- `smartlead-pp-cli webhooks list-event-types` — List webhook event types
- `smartlead-pp-cli webhooks retrigger-failed-events` — Retrigger failed webhook events
- `smartlead-pp-cli webhooks update` — Update webhook


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
smartlead-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes


### Find stale leads with selected fields

```bash
smartlead-pp-cli stale --days 14 --agent --select email,campaign_name,last_event_at
```

Project only the columns the agent needs; --agent emits compact JSON without ANSI.

### Trace duplicate outreach

```bash
smartlead-pp-cli overlap --json --select email,campaign_names
```

Surface every lead present in multiple campaigns so you can dedupe before importing more.

### Triage replies into action buckets

```bash
smartlead-pp-cli reply-classify --since 7d --json --select email,classification,confidence
```

Run a local FTS+keyword pass over the master-inbox to cluster replies.

### Bulk lead import without 400-cap pain

```bash
smartlead-pp-cli campaigns leads add-campaign 3217809 --stdin < leads.json
```

The CLI auto-chunks anything over 400 transparently.

## Auth Setup

Smartlead authenticates by query string. Set `SMARTLEAD_API_KEY=<your-admin-or-client-key>` in your environment, or run `smartlead-pp-cli auth set-token` to store it. The CLI honors Smartlead's 60 requests/minute budget per key with an adaptive limiter that surfaces a typed RateLimitError when exhausted instead of returning empty results.

Run `smartlead-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  smartlead-pp-cli campaigns list --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Explicit retries** — use `--idempotent` only when an already-existing create should count as success, and `--ignore-missing` only when a missing delete target should count as success

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal — piped/agent consumers get pure JSON on stdout.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
smartlead-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
smartlead-pp-cli feedback --stdin < notes.txt
smartlead-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.smartlead-pp-cli/feedback.jsonl`. They are never POSTed unless `SMARTLEAD_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `SMARTLEAD_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
smartlead-pp-cli profile save briefing --json
smartlead-pp-cli --profile briefing campaigns list
smartlead-pp-cli profile list --json
smartlead-pp-cli profile show briefing
smartlead-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `smartlead-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

Install the MCP binary from this CLI's published public-library entry or pre-built release, then register it:

```bash
claude mcp add smartlead-pp-mcp -- smartlead-pp-mcp
```

Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which smartlead-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   smartlead-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `smartlead-pp-cli <command> --help`.
