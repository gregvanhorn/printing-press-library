---
name: pp-producthunt
description: "Read-only Product Hunt CLI backed by the public Atom feed (`/feed`), optional Product Hunt GraphQL auth, and a local SQLite store. Use for today's featured launches, rank trajectory over time, per-day launch calendars, top-maker aggregates, tagline full-text search, GraphQL-seeded historical backfill, search enrichment, co-occurrence signals, and diff-since-last-sync alerts. Do NOT use for posting, upvoting, commenting, following, writing, arbitrary GraphQL exploration/mutations, post detail pages, historical leaderboards, user profiles, topics, or newsletter content (Cloudflare-gated; stubs exist but exit 3). Trigger phrases: `today's top product hunt`, `product hunt trend for <slug>`, `product hunt makers this week`, `what launched on product hunt today`, `product hunt tagline search`, `run producthunt-pp-cli`."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["producthunt-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/marketing/producthunt/cmd/producthunt-pp-cli@latest","bins":["producthunt-pp-cli"],"label":"Install via go install"}]}}'
---

# Product Hunt — Printing Press CLI

A self-warming read-only terminal view of Product Hunt. Three tiers, smallest first:

1. **Atom auto-sync (always on, no auth).** Any read command (`search`, `list`, `today`, etc.) checks `ph_meta.last_sync_at` and silently syncs the public `/feed` when the local store is >24h stale. No flags needed — integrators just call `search` and the CLI keeps itself fresh.
2. **`search --enrich` (opt-in, Product Hunt GraphQL auth required).** When local FTS returns fewer than `--enrich-threshold` results, fires one narrow read-only GraphQL query for the topic, upserts matches, and re-runs the search. Fail-soft.
3. **`backfill` (explicit, Product Hunt GraphQL auth required).** Paginates GraphQL over a window (default 30 days) and seeds the store in one shot. Budget-aware, resumable via `backfill resume`.

Authentication:

- The Atom runtime is token-free by construction — no setup needed.
- GraphQL auth (Tier 2 + 3) is opt-in. Run `auth setup` for instructions, then either:
- Create a Product Hunt application at https://www.producthunt.com/v2/oauth/applications with redirect URI `https://localhost/callback`, then run `auth register`.
- For agents/CI, set `PRODUCTHUNT_DEVELOPER_TOKEN` (or `PRODUCTHUNT_GRAPHQL_TOKEN`) and run `auth set-token --token-env PRODUCTHUNT_DEVELOPER_TOKEN`.

No write operations against Product Hunt. This CLI reads data and persists it locally; it does not post, upvote, comment, or follow.

## Decide Fast: Use This CLI When…

- The user wants **today's featured launches** (`today`, `recent`).
- The user asks **how a specific product trended on Product Hunt** over time — rank trajectory, first-seen, days on feed (`trend <slug>`).
- The user wants a **weekly calendar** of what launched each day (`calendar --week`).
- The user wants **top makers/hunters** across a window (`makers --since 30d`).
- The user wants **full-text search across launch taglines** (`tagline-grep`, `search`).
- The user wants a **cron-friendly diff** of new launches since the last run (`watch`).
- The user wants **CSV export** of synced posts (`list --csv --select ...`).

## Do NOT Use This CLI When…

- The user wants to **post a product, upvote, comment, follow** — this CLI has no write path. Decline or refer them to producthunt.com.
- The user needs **post detail pages, comments, historical daily leaderboards, topic feeds, user profiles, collections, or newsletter issues**. Those routes are Cloudflare-gated and ship as stubs that exit 3. Stub names below.
- The user wants **arbitrary GraphQL queries, mutations, schema exploration, or write-like actions**. The CLI only uses a curated read-only Product Hunt GraphQL path for `backfill` and `search --enrich`.

## Cloudflare-Gated Stubs (Exit 3)

These commands exist so stubs are discoverable, but every one emits a JSON explanation and exits 3 when called. **Do not route user requests to them.** If a user asks for something that maps to one of these, say the feature isn't reachable from an automated HTTP client and suggest the listed alternative.

| Stub | What the user asked for | Alternative |
|------|--------------------------|-------------|
| `post <slug>` | Full post page | `get <slug>` (feed-level metadata) or `open <slug>` |
| `comments <slug>` | Post comments | `open <slug>` (browser) |
| `leaderboard {daily,weekly,monthly,yearly}` | Historical leaderboard | Run `sync` on a schedule, then `list --since` / `trend` |
| `topic <slug>` | Topic feed | None — `/feed?category=` is ignored server-side |
| `user <handle>` | Maker/hunter profile | `list --author "<Display Name>"` |
| `collection <slug>` | Curated list | None |
| `newsletter` | Newsletter archive | None |

## Command Reference

Every command supports `--json`, `--select`, `--csv`, `--agent`, `--quiet`, `--human-friendly`, and `--timeout`.

Self-warming flags (also global):

| Flag | Purpose |
|------|---------|
| `--no-auto-sync` | Skip the stale-store Atom sync that read commands run by default |
| `--caller <id>` | Identify the calling integrator (e.g. `last30days/3.0.1`) for logs |

Full flag list in `producthunt-pp-cli --help`.

### /feed reads + local store

| Command | Purpose |
|---------|---------|
| `today [--limit N] [--live]` | Top N featured launches (store first, live `/feed` fallback) |
| `recent [--limit N]` | Always live-fetch `/feed` |
| `sync [--dry-run-feed] [--db PATH]` | Fetch `/feed` and persist a ranked snapshot (auto-sync runs this implicitly when the store is stale) |
| `backfill [--days N \| --from DATE --to DATE] [--dry-run]` | Seed the store with a bulk 30-day history via GraphQL (requires GraphQL auth; run `auth setup`) |
| `backfill resume [--window-id ID]` | Continue an interrupted backfill from the saved cursor |
| `auth setup [--json]` | Show guided Product Hunt app setup, including redirect URI and agent-safe env var paths |
| `auth register [--client-id ID --client-secret SECRET] [--no-input]` | Exchange a Product Hunt OAuth app API key/secret for a client_credentials token |
| `auth set-token --token-env PRODUCTHUNT_DEVELOPER_TOKEN` | Persist an existing developer/API token without putting it in shell history |
| `list [--author NAME] [--since DUR] [--until DUR] [--sort FIELD] [--asc] [--limit N]` | Query the local store |
| `search <query> [--limit N] [--enrich] [--enrich-threshold N]` | FTS5 match across slug, title, tagline, author. `--enrich` tops up from GraphQL when local results are thin (GraphQL auth required). |
| `get <slug> [--live] [--external] [--url-only]` | Single post payload (`info` remains an alias for compatibility) |
| `open <slug> [--external] [--dry]` | Open in the default browser |
| `feed raw [--validate]` | Raw Atom XML to stdout |
| `feed refresh` | Alias for `sync` |

### Aggregates across snapshots

| Command | Purpose |
|---------|---------|
| `trend <slug> [--summary]` | Rank trajectory, first/last seen, days on feed |
| `watch [--no-write]` | Diff live `/feed` vs the last snapshot |
| `makers [--since DUR] [--top N]` | Top authors across snapshots in window |
| `calendar [--week W] [--days N] [--include-posts]` | Day-by-day breakdown |
| `outbound-diff [--since DUR] [--limit N]` | External URLs that changed across syncs |
| `tagline-grep <pattern> [--mode fts|regex] [--since DUR] [--limit N]` | Tagline search; auto-switches to regex for `.*+?()[]|\` patterns |
| `authors related --to <name> [--since DUR] [--limit N]` | Co-occurrence graph |

### Diagnostics & utility

| Command | Purpose |
|---------|---------|
| `doctor [--fail-on error] [--json]` | Probe `/feed`, parse Atom, verify schema, list CF-gated routes |
| `version` | Print CLI version |
| `which <capability>` | Resolve a natural-language capability query to a command name |
| `auth {setup,status,register,set-token,logout}` | Optional GraphQL credential setup/status. Atom commands do not require auth. |
| `profile {save,list,show,delete}` | Saved flag sets for reuse |
| `feedback "<note>"` / `feedback list` | Record surprises locally at `~/.producthunt-pp-cli/feedback.jsonl` |
| `agent-context` | Dump full command/flag inventory as JSON |
| `api` | Browse endpoints by internal handler name |

## Recipes

Every command accepts `--agent` (expands to `--json --compact --no-input --no-color --yes`).

### Daily briefing

```bash
producthunt-pp-cli sync && \
  producthunt-pp-cli today --limit 10 --agent --select 'slug,title,tagline,author,published'
```

Sync first (the store starts empty), then narrow the payload to only the fields you'll read.

### Weekly maker recap

```bash
producthunt-pp-cli makers --since 7d --top 10 --agent
```

Only meaningful after the CLI has run `sync` for several days. A single snapshot contains 50 authors.

### Tagline trend check

```bash
# FTS mode (fast, no regex chars)
producthunt-pp-cli tagline-grep agent --since 90d --agent

# Regex mode (auto-switched by dotted patterns)
producthunt-pp-cli tagline-grep 'ai.*agent' --since 90d --agent --select 'slug,title,tagline,published'
```

### Scraper-parity CSV export

```bash
producthunt-pp-cli list --since 30d --csv \
  --select 'id,slug,title,tagline,author,published,discussion_url,external_url' > ph.csv
```

Reproduces the `fernandod1/ProductHunt-scraper` column set from the local store.

### Cron-friendly new-launch watcher

```bash
producthunt-pp-cli watch --agent --compact
```

Records a new snapshot and prints only entries that weren't in the previous one. Idempotent at steady state.

### Rank trajectory for a product

```bash
# Replace <slug> with a real slug from `today` output.
producthunt-pp-cli trend <slug> --summary --agent \
  --select 'slug,title,best_rank,worst_rank,appearance_count,days_on_feed'
```

Returns empty when the slug hasn't appeared in any synced snapshot; prompt the user to run `sync` or try a different slug.

### Finding the right command

When the agent isn't sure which command maps to a user's ask:

```bash
producthunt-pp-cli which "<capability in the user's own words>"
```

`which` returns scored matches from a curated feature index. Exit 0 means at least one confident match; exit 2 means no confident match — fall back to `--help`.

## Agent Mode Details

`--agent` on any command expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr.
- **Filterable with dotted paths.** `--select` keeps only the named fields. Dotted paths descend into nested objects and arrays. Real examples for this CLI:

  ```bash
  # Bare array of posts — slice each element:
  producthunt-pp-cli today --agent --select 'slug,title,tagline,author'

  # Trend response has an `appearances` array — descend into it:
  producthunt-pp-cli trend <slug> --agent --select 'slug,best_rank,appearances.rank,appearances.taken_at'

  # Calendar response has a `days` array of day objects:
  producthunt-pp-cli calendar --days 7 --agent --select 'days.date,days.count'
  ```

- **Response shape:** most commands emit a bare JSON array or object. Commands that include auth/enrichment hints can emit `{"results":[...],"_meta":{...}}`; `--select slug,title` applies to `results` while preserving `_meta`.

### Flags that exist but don't behave as they might elsewhere

- `--dry-run` is a persistent flag inherited from the generator. On this CLI's read commands it is a no-op. Use `sync --dry-run-feed` to actually preview a sync without writing.
- `--no-cache` exists but has no effect: the Atom fetch path doesn't go through the generator's cache layer.
- `--data-source {auto,live,local}` controls store-vs-live preference on `today` and `get`; most commands ignore it.

## Exit Codes (Actually Used)

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (missing arg, bad flag, invalid sort field) |
| 3 | Not found — slug absent from store, or a Cloudflare-gated stub |
| 5 | API / parse error (feed unreachable, Atom body malformed) |
| 10 | Config error (store path, config file) |

Codes 4 (auth) and 7 (rate limit) are defined in the generator but not raised by any command in this CLI.

## Auth Setup

No auth is required for Atom-backed commands. Use auth only when the user needs immediate historical backfill or GraphQL search enrichment beyond snapshots already accumulated locally.

Interactive setup:

```bash
producthunt-pp-cli auth setup
```

Create the Product Hunt app at https://www.producthunt.com/v2/oauth/applications with:

| Field | Value |
|-------|-------|
| Name | `producthunt-pp-cli` (or any recognizable local name) |
| Redirect URI | `https://localhost/callback` |

Then run:

```bash
producthunt-pp-cli auth register
producthunt-pp-cli auth status --json
```

Agent/CI setup without prompts:

```bash
PRODUCTHUNT_CLIENT_ID=... PRODUCTHUNT_CLIENT_SECRET=... \
  producthunt-pp-cli auth register --no-input --json
```

If you already have a Product Hunt developer/API token:

```bash
PRODUCTHUNT_DEVELOPER_TOKEN=... \
  producthunt-pp-cli auth set-token --token-env PRODUCTHUNT_DEVELOPER_TOKEN --json
```

`auth status` and `doctor --json` report configured capabilities, but they do not live-validate the token. Use `producthunt-pp-cli backfill --days 30 --dry-run` to verify API access before a real backfill.

## Agent Feedback

When the CLI surprises you, record it:

```bash
producthunt-pp-cli feedback "trend <slug> returned no data even after sync — expected at least the current snapshot"
producthunt-pp-cli feedback --stdin < notes.txt
producthunt-pp-cli feedback list --json --limit 10
```

Entries land at `~/.producthunt-pp-cli/feedback.jsonl`. They are local-only unless `PRODUCTHUNT_FEEDBACK_ENDPOINT` is set AND `--send` is passed (or `PRODUCTHUNT_FEEDBACK_AUTO_SEND=true`).

Write what *surprised* you, not a bug report. Short, specific, one line.

## Output Delivery

Every command accepts `--deliver <sink>`:

| Sink | Effect |
|------|--------|
| `stdout` | Default |
| `file:<path>` | Atomically write to path (tmp + rename) |
| `webhook:<url>` | POST body to URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes refuse with a structured error. Webhook HTTP failures return non-zero and log to stderr.

## Named Profiles

Save a set of flags and reuse them across invocations:

```bash
producthunt-pp-cli profile save briefing --limit 10 --agent --select 'slug,title,tagline,author,published'
producthunt-pp-cli --profile briefing today
producthunt-pp-cli profile list --json
producthunt-pp-cli profile show briefing
producthunt-pp-cli profile delete briefing --yes
```

Explicit flags override profile values; profile values override defaults. `agent-context` lists all saved profiles under `available_profiles`.

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `producthunt-pp-cli --help`.
2. **Starts with `install`**:
   - ends with `mcp` → run MCP Installation.
   - otherwise → run CLI Installation.
3. **Anything else** → Direct Use: match the request to a command from Command Reference, execute with `--agent`.

## CLI Installation

1. Check Go ≥ 1.25: `go version`.
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/marketing/producthunt/cmd/producthunt-pp-cli@latest
   ```
3. Verify: `producthunt-pp-cli --version`.
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

The MCP server surface is **deliberately narrow** — it exposes the `/feed` read as a single tool. The transcendence commands (`trend`, `calendar`, `makers`, `outbound-diff`, `tagline-grep`, `authors related`) depend on a local snapshot store and are not in the MCP surface; use the CLI for those.

1. Install the MCP binary:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/marketing/producthunt/cmd/producthunt-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add producthunt-pp-mcp -- producthunt-pp-mcp
   ```
3. Verify: `claude mcp list`.

## Direct Use

1. Check if installed: `which producthunt-pp-cli`. If missing, offer CLI Installation.
2. Match the user's request to a command from the Command Reference, respecting the Do-Not-Use list and stubs.
3. For store-backed commands (`trend`, `makers`, `calendar`, `outbound-diff`, `tagline-grep`, `authors related`), ensure the store is populated: run `sync` first if the user hasn't synced yet, or propose it.
4. Execute with `--agent` to keep output narrow and script-friendly:
   ```bash
   producthunt-pp-cli <command> [args] --agent --select '<fields>'
   ```
5. If the result is empty, run `producthunt-pp-cli doctor --json` before assuming the feature is broken — the most common cause is "store not synced" or a transient `/feed` 503.
