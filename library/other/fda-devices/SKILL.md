---
name: pp-fda-devices
description: "Regulatory intelligence for FDA medical devices — predicate-chain auditing, competitive scanning, and safety... Trigger phrases: `FDA 510(k)`, `medical device predicate`, `MAUDE adverse event`, `device recall`, `openFDA`, `use fda-devices`."
author: "Greg Van Horn"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - fda-devices-pp-cli
---

# FDA Devices — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `fda-devices-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer:
   ```bash
   npx -y @mvanhorn/printing-press install fda-devices --cli-only
   ```
2. Verify: `fda-devices-pp-cli --version`
3. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

If the `npx` install fails before this CLI has a public-library category, install Node or use the category-specific Go fallback after publish.

If `--version` reports "command not found" after install, the install step did not put the binary on `$PATH`. Do not proceed with skill commands until verification succeeds.

This CLI keeps a local copy of openFDA's 510(k), PMA, recall, MAUDE, classification, and establishment data into a local SQLite database, then surfaces it through agent-native commands like predicate-chain, competitors, and safety-pattern. Free public data, made queryable in compound ways the upstream API can't.

## When to Use This CLI

Reach for fda-devices-pp-cli when an agent needs structured answers about FDA medical device clearances, predicates, recalls, or adverse events. Best fit: questions that compound across endpoints ('safety record per year on market', 'who cleared this and what did they cite as predicate'), surveillance ('alert me when a new entrant clears in this product code'), or audit ('walk this device's ancestor tree'). Not the right tool for one-off lookups already easy on open.fda.gov.

## When Not to Use This CLI

Do not activate this CLI for requests that require creating, updating, deleting, publishing, commenting, upvoting, inviting, ordering, sending messages, booking, purchasing, or changing remote state. This printed CLI exposes read-only commands for inspection, export, sync, and analysis.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Competitive intelligence
- **`competitors`** — Every new 510(k) clearance in a product code, ranked by recency, with applicant info.

  _When an agent is asked 'who else cleared this kind of device recently', this gives a ranked answer in one call._

  ```bash
  fda-devices-pp-cli competitors --product-code DQA --since 2y --json
  ```
- **`applicant-history`** — Every clearance ever issued to a company, grouped by clinical area with go-to predicates.

  _Use this when asked about a vendor's regulatory footprint._

  ```bash
  fda-devices-pp-cli applicant-history 'Medtronic' --json
  ```
- **`market-entry`** — New entrants in a product code, ranked by time-to-clearance.

  _Use when asked about competitive timing or market dynamics in a device category._

  ```bash
  fda-devices-pp-cli market-entry --product-code DQA --since 1y --json
  ```

### Predicate audit
- **`predicate-chain`** — Recursive ancestor tree from a K-number, flagging any predicate that's been recalled.

  _Answers 'what is this device descended from' and 'has any ancestor been recalled' in one call._

  ```bash
  fda-devices-pp-cli predicate-chain K203456 --depth 10 --json
  ```

### Safety surveillance
- **`recall-link`** — Every recall on a device plus devices that used it as predicate.

  _Surfaces downstream exposure when a device or its ancestors get recalled._

  ```bash
  fda-devices-pp-cli recall-link K203456 --json
  ```
- **`safety-pattern`** — Devices in a category ranked by adverse-event count per year on market.

  _Quantitative answer to 'which devices in this category have the worst safety record per year on market'. Neutral statistic._

  ```bash
  fda-devices-pp-cli safety-pattern --product-code DQA --json
  ```

### Operational
- **`watch`** — Subscribe to product codes, applicants, predicate devices, or recall classes; pipes diff to file/Slack/webhook.

  _Use to set up ongoing surveillance for a category or vendor._

  ```bash
  fda-devices-pp-cli watch new --product-code DQA --notify slack
  ```

### Narrative
- **`story`** — One-paragraph briefing about a K-number: who cleared it, predicate chain, related recalls.

  _Quick context dump when an agent needs prose about a clearance._

  ```bash
  fda-devices-pp-cli story K203456
  ```

### Local data
- **`sql`** — Read-only SQL against the local mirror; join 510(k), recalls, MAUDE, classifications freely.

  _Escape hatch for arbitrary analytical questions an agent can compose on the fly._

  ```bash
  fda-devices-pp-cli sql 'SELECT applicant, COUNT(*) FROM clearances_510k WHERE product_code="DQA" GROUP BY applicant ORDER BY 2 DESC LIMIT 10'
  ```

## Command Reference

**device** — Manage device

- `fda-devices-pp-cli device list-classification` — Search device product-code classifications
- `fda-devices-pp-cli device list-establishment` — Search establishment registrations
- `fda-devices-pp-cli device list-maude` — Search MAUDE adverse events
- `fda-devices-pp-cli device list-pma` — Search PMA approvals
- `fda-devices-pp-cli device list-recall` — Search device recalls
- `fda-devices-pp-cli device list510k` — Search 510(k) clearances


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
fda-devices-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes


### Find competitors in a product code

```bash
fda-devices-pp-cli competitors --product-code DQA --since 2y --json --select results.k_number,results.applicant,results.decision_date
```

Ranks recent clearances. Use --select to extract just the fields you want; the full payload is large.

### Audit a predicate chain

```bash
fda-devices-pp-cli predicate-chain K203456 --depth 10 --json
```

Returns the full ancestor tree with clearance dates and recall flags. Cycles are detected and broken.

### Cross-endpoint SQL

```bash
fda-devices-pp-cli sql 'SELECT product_code, COUNT(*) as n FROM recalls WHERE recall_initiation_date >= "2024-01-01" GROUP BY product_code ORDER BY n DESC LIMIT 20' --json
```

Top product codes by recall count this year. Local mirror only.

### Set up an alert

```bash
fda-devices-pp-cli watch new --product-code DQA --notify file:/tmp/dqa-clearances.log
```

Diff-based; pipes new records on each sync to the notify target.

## Auth Setup

No authentication required.

Run `fda-devices-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  fda-devices-pp-cli device list-classification --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

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
fda-devices-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
fda-devices-pp-cli feedback --stdin < notes.txt
fda-devices-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.fda-devices-pp-cli/feedback.jsonl`. They are never POSTed unless `FDA_DEVICES_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `FDA_DEVICES_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

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
fda-devices-pp-cli profile save briefing --json
fda-devices-pp-cli --profile briefing device list-classification
fda-devices-pp-cli profile list --json
fda-devices-pp-cli profile show briefing
fda-devices-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `fda-devices-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

Install the MCP binary from this CLI's published public-library entry or pre-built release, then register it:

```bash
claude mcp add fda-devices-pp-mcp -- fda-devices-pp-mcp
```

Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which fda-devices-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   fda-devices-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `fda-devices-pp-cli <command> --help`.
