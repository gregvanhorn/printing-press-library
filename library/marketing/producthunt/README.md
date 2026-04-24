# Product Hunt CLI

**A self-warming terminal view of Product Hunt.**

[Product Hunt](https://www.producthunt.com) is the daily launch board for new products — makers post, hunters submit, and the community votes and comments.

This CLI ships three tiers of access, picking the lightest one automatically:

1. **Atom auto-sync (always on, no auth).** Every read command checks whether the local SQLite store is stale (>24h since last sync) and silently fetches the public `/feed` before serving the query. Integrators just call `search`/`list`/etc. — the CLI keeps itself fresh. This builds history from the moment you start syncing.
2. **`search --enrich` (opt-in, authenticated).** When local results are thin for a topic, fires one narrow GraphQL query for that topic over the last 30 days. Fail-soft: if Product Hunt GraphQL auth isn't configured or the budget is close to the floor, local results are returned with structured metadata explaining the skipped enrichment.
3. **`backfill` (explicit, authenticated).** Paginates the Product Hunt GraphQL `posts` field over a window (default 30 days) and seeds the store in one shot. Budget-aware, resumable via `backfill resume`. Used for cold-starts, gap recovery after offline periods, or deliberate historical windows.

Every command speaks JSON, filters fields with `--select`, and exits with typed codes for scripts and agents.

## Install

### Go

```
go install github.com/mvanhorn/printing-press-library/library/marketing/producthunt/cmd/producthunt-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Authentication

The Atom runtime is token-free by construction. All the daily-driver commands (`sync`, `today`, `recent`, `list`, `search`, `trend`, `calendar`, `makers`, `tagline-grep`, `watch`, `outbound-diff`) read the public `/feed` and need no credentials.

Product Hunt GraphQL auth is opt-in and unlocks two capabilities that the anonymous Atom feed cannot provide:

- `search --enrich` — top up thin local search results from the GraphQL API
- `backfill` — bulk-seed historical posts in one call

Anonymous mode cannot retroactively backfill the past because `/feed` exposes only the current feed window. It can build durable long-term history by running `sync` or relying on auto-sync from now forward.

For guided setup:

```bash
producthunt-pp-cli auth setup
```

OAuth app path:

```bash
# Visit https://www.producthunt.com/v2/oauth/applications
# Create an app:
#   Name: producthunt-pp-cli
#   Redirect URI: https://localhost/callback
producthunt-pp-cli auth register
```

Agent/CI path:

```bash
PRODUCTHUNT_CLIENT_ID=... PRODUCTHUNT_CLIENT_SECRET=... \
  producthunt-pp-cli auth register --no-input

PRODUCTHUNT_DEVELOPER_TOKEN=... \
  producthunt-pp-cli auth set-token --token-env PRODUCTHUNT_DEVELOPER_TOKEN
```

The CLI saves credentials to `~/.config/producthunt-pp-cli/config.toml` with `0600` perms. Revoke with `auth logout`. `auth status --json`, `doctor --json`, and `agent-context` expose whether GraphQL-powered features are currently available.

Product Hunt's HTML pages (post detail, user profiles, topic pages, historical leaderboards, newsletter archive) are gated by Cloudflare against automated HTTP clients. Commands that would need those routes (`post`, `comments`, `leaderboard`, `topic`, `user`, `collection`, `newsletter`) ship as explicit stubs that emit a structured JSON explanation and exit with code 3. They are named in the command reference below so agents and scripts can discover the gap without hitting opaque timeouts.

## Quick Start

```bash
# First run: no setup needed — auto-sync fires when the store is stale.
producthunt-pp-cli search "ai agent"

# Or kick off a manual sync explicitly.
producthunt-pp-cli sync

# Optional Tier 2/3: configure Product Hunt GraphQL, then bulk-seed 30 days.
producthunt-pp-cli auth setup
producthunt-pp-cli auth register
producthunt-pp-cli backfill --days 30 --dry-run  # estimate first
producthunt-pp-cli backfill --days 30

# Agent-friendly shortlist with only the fields you want.
producthunt-pp-cli today --limit 10 --json --select 'slug,title,tagline,author'

# Rank trajectory for a slug across every snapshot you've synced.
# (Replace 'seeknal' with any real slug from 'today' output.)
producthunt-pp-cli trend seeknal --json

# Diff the live feed against your last sync; idempotent, cron-safe.
producthunt-pp-cli watch --agent

# Regex search across every tagline ever synced.
producthunt-pp-cli tagline-grep 'ai.*agent' --since 90d
```

## Unique Features

These capabilities aren't available in any other Product Hunt tool.

### Local state that compounds

- **`trend`** — See when a product first appeared on the feed, how many days it lingered, and its best/worst rank across every snapshot.

  _Reach for this when you want to judge a product's momentum without a token or the web UI._

  ```bash
  producthunt-pp-cli trend seeknal --json
  ```

- **`calendar`** — Week-at-a-glance showing which products were featured each day. Zero-count days are included so the window is always complete.

  _Use for retrospectives, competitive scans, or weekly maker newsletters._

  ```bash
  producthunt-pp-cli calendar --week 2026-W16 --agent
  ```

- **`makers`** — Top authors (makers and hunters) aggregated across every snapshot in a time window.

  _Pick this when scouting prolific makers or writing a monthly recap._

  ```bash
  producthunt-pp-cli makers --since 30d --top 10 --agent
  ```

- **`outbound-diff`** — Products whose external landing URL changed between sync cycles. Detects beta→launch domain moves and link swaps.

  _Only meaningful after at least two syncs with different URLs; a single snapshot returns `[]` honestly._

  ```bash
  producthunt-pp-cli outbound-diff --since 30d --json
  ```

- **`tagline-grep`** — FTS5 or regex search across every tagline the CLI has ever synced. Auto-switches to regex mode when the pattern contains `.*+?()[]|\` so `ai.*agent` just works.

  ```bash
  producthunt-pp-cli tagline-grep 'ai.*agent' --since 90d --json --select 'slug,title,tagline,published'
  ```

- **`authors related`** — Authors who repeatedly appear alongside a given author in the same feed snapshots — a rough social signal from pure /feed data.

  ```bash
  producthunt-pp-cli authors related --to 'Ryan Hoover' --since 90d --json
  ```

### Agent-native plumbing

- **`watch`** — New entries since the last sync, nothing else. Idempotent: back-to-back runs at the same rate return `new_count: 0` when the feed hasn't changed.

  ```bash
  producthunt-pp-cli watch --agent --compact
  ```

### Atom-first diagnostics

- **`doctor`** — Probes `/feed`, parses the Atom body, reports entry count and fetch latency, verifies the local SQLite schema, and names every Cloudflare-gated route so you know why certain commands are stubbed.

  ```bash
  producthunt-pp-cli doctor --json
  ```

## Commands

### Reads (backed by /feed + local store)

| Command | What it does |
|---------|--------------|
| `today` | Top N featured launches (store first, live `/feed` fallback) |
| `recent` | Live-fetch `/feed`, bypass the store |
| `sync` | Fetch `/feed` and persist a ranked snapshot |
| `list` | Filter the local store by author, date range, or sort field |
| `search <query>` | FTS5 match across titles, taglines, authors, and slugs |
| `get <slug>` | One post's full `/feed` payload (`info` remains an alias) |
| `open <slug>` | Launch the Product Hunt page in your default browser |
| `feed raw` | Dump the raw Atom XML to stdout |
| `feed refresh` | Alias for `sync` |

### Aggregates (only possible because we keep history)

| Command | What it does |
|---------|--------------|
| `trend <slug>` | Rank trajectory, first/last seen, days on feed, appearances |
| `watch` | Diff-since-last-sync; optionally records a fresh snapshot |
| `makers --since <window>` | Top authors across snapshots in a window |
| `calendar --week` or `--days` | Calendar view from daily snapshots |
| `outbound-diff` | Products whose external URL changed across syncs |
| `tagline-grep <pattern>` | FTS or regex search across every tagline seen |
| `authors related --to <name>` | Co-occurrence graph from snapshots |

### Cloudflare-gated stubs (exit 3, structured JSON)

Cloudflare blocks Product Hunt's HTML routes for automated HTTP clients. These commands exist so scripts and agents discover the gap instead of hitting opaque timeouts. Each emits JSON with `cf_gated: true`, names the alternative, and exits with code 3.

`post <slug>` · `comments <slug>` · `leaderboard {daily,weekly,monthly,yearly}` · `topic <slug>` · `user <handle>` · `collection <slug>` · `newsletter`

### Utility

`doctor` · `version` · `auth {setup,status,register,set-token,logout}` · `profile` · `which` · `feedback` · `agent-context` · `api` · `workflow` · `export` · `import`

Auth is optional: Atom-backed commands need no credentials, while `backfill` and `search --enrich` use configured Product Hunt GraphQL credentials when present.

## Output Formats

```bash
# Human-readable table (default in terminal; JSON when piped).
producthunt-pp-cli today

# JSON for scripting and agents.
producthunt-pp-cli today --json

# Narrow to exactly the fields you want (dotted paths descend into nested structures).
producthunt-pp-cli today --json --select 'slug,title,tagline,author,published'

# CSV for spreadsheets.
producthunt-pp-cli list --since 30d --csv --select 'id,slug,title,author,published'

# Show what sync would do without writing.
producthunt-pp-cli sync --dry-run-feed

# Agent mode: JSON + compact + no prompts + no color in one flag.
producthunt-pp-cli today --agent
```

## Cookbook

```bash
# Daily agent briefing — sync, then narrow payload for the model.
producthunt-pp-cli sync && \
  producthunt-pp-cli today --limit 10 --agent --select 'slug,title,tagline,author,published'

# Weekly maker recap — aggregate authors from the last 7 days of snapshots.
producthunt-pp-cli makers --since 7d --top 10 --agent

# Tagline trend check — regex across 90 days of taglines.
producthunt-pp-cli tagline-grep 'agent|copilot' --since 90d --json \
  --select 'slug,title,tagline,published'

# Scraper-parity CSV export — matches the column set of fernandod1/ProductHunt-scraper.
producthunt-pp-cli list --since 30d --csv \
  --select 'id,slug,title,tagline,author,published,discussion_url,external_url' > ph.csv

# Cron-friendly new-launch watcher.
producthunt-pp-cli watch --agent --compact

# Rank trajectory for a specific slug.
producthunt-pp-cli trend <slug> --json --select 'slug,title,best_rank,appearance_count,days_on_feed'

# Open a product page in your browser from a slug.
producthunt-pp-cli open <slug>

# Full-text search the local store.
producthunt-pp-cli search 'ai agent' --limit 5 --json --select 'slug,title,tagline'
```

## Agent Usage

This CLI was designed for non-interactive use by AI agents, scripts, and CI.

- **Non-interactive** — never prompts; every input is a flag or positional argument.
- **Pipeable** — `--json` writes to stdout, errors to stderr.
- **Filterable** — `--select slug,title,author` returns only the fields you name. Dotted paths descend into nested objects and arrays (`--select 'appearances.rank,appearances.taken_at'` on `trend` output).
- **Compact** — `--compact` returns only high-gravity fields.
- **Previewable** — `sync --dry-run-feed` fetches and parses without writing.
- **Agent-safe by default** — no colors or terminal formatting unless `--human-friendly` is set.
- **Read-only** — no writes against Product Hunt. Commands that would need write auth are not shipped.

**Agent mode:** pass `--agent` to any command for the bundle: `--json --compact --no-input --no-color --yes`.

**Exit codes used by this CLI:**

| Code | Meaning |
|------|---------|
| `0` | Success |
| `2` | Usage error (bad flag, missing required arg) |
| `3` | Not found, or Cloudflare-gated stub command |
| `5` | API / parse error |
| `10` | Config error |

## Use as MCP Server

The repo also ships a companion MCP server (`producthunt-pp-mcp`) for Claude Desktop, Cursor, and other MCP-compatible tools. **The MCP surface is deliberately narrow** — it exposes the `/feed` read operation as a single tool. For the richer transcendence commands (`trend`, `calendar`, `makers`, `outbound-diff`, `tagline-grep`, `authors related`), use the CLI directly; they depend on a local snapshot store that an in-process MCP tool cannot maintain meaningfully.

### Claude Code

```bash
claude mcp add producthunt producthunt-pp-mcp
```

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "producthunt": {
      "command": "producthunt-pp-mcp"
    }
  }
}
```

## Health Check

```bash
producthunt-pp-cli doctor --json
```

Reports:
- `/feed` reachability (HTTP status + fetch latency) and Atom parse status.
- Entry count in the most recent fetch.
- Local SQLite schema version and last sync timestamp.
- The full list of Cloudflare-gated routes the runtime intentionally skips.

`doctor --fail-on error` exits non-zero when any section reports an error, useful in monitoring wrappers.

## Configuration

Config file (optional): `~/.config/producthunt-pp-cli/config.toml`

Local data store: `~/.local/share/producthunt-pp-cli/data.db` (SQLite, created on first `sync`).

Override the store path with `--db <path>` on any command that reads or writes it.

## Troubleshooting

- **Empty results after install** — run `producthunt-pp-cli sync` first. The store starts empty and the CLI refuses to fabricate data.
- **`post <slug>`, `leaderboard daily`, etc. exit with code 3** — expected. Those commands are Cloudflare-gated stubs. Use `get <slug>` for `/feed`-level metadata and `open <slug>` to view the real page in your browser.
- **`/feed` parse fails** — run `producthunt-pp-cli doctor --json` to confirm the feed is still serving Atom XML. Product Hunt occasionally returns a 503 during deploys.
- **`search` returns nothing** — FTS5 only indexes what has been synced. One snapshot is 50 entries; run `sync` on a schedule to build a meaningful index.
- **`trend <slug>` says "not in store"** — the slug was not in any snapshot the CLI has taken. Run `sync` when the product is currently featured, or wait for the next sync to pick it up.
- **`tagline-grep` with a dotted pattern** — the pattern auto-switches to regex mode when it sees `.*+?()[]|\`. To force FTS5 literal phrase matching, use `--mode fts` and quote the phrase.

## HTTP Transport

All HTTP requests go out with a Chrome-compatible `User-Agent` and `Accept` headers so the public Atom endpoint responds identically whether the CLI is running from a laptop or CI. There is no resident browser, no Playwright, no Chromium sidecar — plain `net/http` against `https://www.producthunt.com/feed`.

Cloudflare's challenge on the HTML routes does not trigger for `/feed`; no clearance cookie is needed for the surface this CLI uses.

## Discovery Signals

This CLI was generated from live traffic analysis of `https://www.producthunt.com`.

- **Reachability:** `atom_primary` (confidence 0.95) — `/feed` is the single replayable surface.
- **Protocols observed:** `atom_feed` (50 entries per snapshot).
- **Auth signals:** none.
- **Protection signals:** Cloudflare Turnstile on all HTML routes.
- **Generation hints:** `emit_browser_compatible_ua`, `emit_atom_parser`, `emit_store_snapshots`, `atom_primary_runtime`.
- **Verified during discovery:** the `/feed?category=<slug>` query parameter is ignored server-side (response is byte-equivalent to the unfiltered feed). No per-topic Atom endpoint exists.

---

## Sources & Inspiration

Built by studying every Product Hunt client that reached public release:

- [**jaipandya/producthunt-mcp-server**](https://github.com/jaipandya/producthunt-mcp-server) — Python, 43 stars. Canonical MCP mapping of the GraphQL API; taught us the resource shape.
- [**sunilkumarc/product-hunt-cli**](https://github.com/sunilkumarc/product-hunt-cli) — JavaScript. Original "what should a PH CLI do" reference.
- [**fernandod1/ProductHunt-scraper**](https://github.com/fernandod1/ProductHunt-scraper) — Python. Source of the scraper-parity CSV column set.
- [**shashankpolanki/Producthunt-Scraper**](https://github.com/shashankpolanki/Producthunt-Scraper) — Python. Daily-leaderboard Scrapy cron that inspired `sync` + `trend`.
- [**sungwoncho/node-producthunt**](https://github.com/sungwoncho/node-producthunt) — JavaScript.
- [**sibis/producthunt-cli**](https://github.com/sibis/producthunt-cli) — JavaScript.

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press).
