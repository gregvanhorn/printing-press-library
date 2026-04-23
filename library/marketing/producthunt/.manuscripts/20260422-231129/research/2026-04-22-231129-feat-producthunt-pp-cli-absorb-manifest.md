# Product Hunt CLI — Absorb Manifest

## Tools Cataloged
| Tool | URL | Token? |
|------|-----|--------|
| jaipandya/producthunt-mcp-server | https://github.com/jaipandya/producthunt-mcp-server | yes |
| @yilin-jing/producthunt-mcp (npm) | https://www.npmjs.com/package/@yilin-jing/producthunt-mcp | yes |
| sunilkumarc/product-hunt-cli | https://github.com/sunilkumarc/product-hunt-cli | yes |
| sibis/producthunt-cli | https://github.com/sibis/producthunt-cli | yes |
| karan/Spear | https://github.com/karan/Spear | yes |
| sungwoncho/node-producthunt | https://github.com/sungwoncho/node-producthunt | yes |
| `producthunt` (PyPI) | https://pypi.org/project/producthunt/ | yes |
| fernandod1/ProductHunt-scraper | https://github.com/fernandod1/ProductHunt-scraper | no (HTML) |
| bennyblanco4/producthunt-scraper | https://github.com/bennyblanco4/producthunt-scraper | no (HTML) |
| shashankpolanki/Producthunt-Scraper | https://github.com/shashankpolanki/Producthunt-Scraper | no (HTML) |
| toughyear/producthunt-scraper | https://github.com/toughyear/producthunt-scraper | no (HTML) |
| yoanbernabeu/producthunt-skills | https://github.com/yoanbernabeu/producthunt-skills | n/a (prompts) |
| n8n / Pipedream connectors | https://n8n.io/integrations/product-hunt/ , https://pipedream.com/apps/product-hunt | yes |

**Reachability constraint:** Cloudflare Turnstile blocks all HTML routes (`/posts/<slug>`, `/leaderboard/...`, `/topics/...`, `/@<handle>`, `/collections`, newsletter archive) from any automated HTTP client, including a Chrome-UA curl. The public `/feed` (Atom) is CF-free and returns 50 featured entries. The official GraphQL API at `api.producthunt.com/v2/api/graphql` is Bearer-token-only and user explicitly declined the token path.

**Shipping scope envelope:**
- **In scope:** Every feature reachable through the public `/feed` Atom endpoint, plus features composed from locally-persisted snapshots (the bulk of the Absorbed + all Transcendence rows).
- **Stubs:** Features that would require HTML or the token API — shipped as explicit stubs that emit an honest "this surface requires browser clearance / token auth, run `--help` to see what works now" message.

---

## Absorbed (Tier A — /feed-backed, shippable today)

| # | Feature | Best Source | Our Implementation | Added Value | Status |
|---|---------|-------------|--------------------|-------------|--------|
| 1 | List today's featured products | `producthunt.get_daily()` (PyPI), `ph home` (sunilkumarc) | `today` — parse `/feed`, show 50 entries with rank/title/tagline/author/link | `--json`, `--select`, `--csv`, `--limit` | shipping |
| 2 | Get a product by slug | jaipandya `get_post_details` | `info <slug>` — return the `/feed` entry for a slug (from store or live) | Works offline after `sync`; agent-native flags | shipping |
| 3 | Open product page in browser | sunilkumarc `ph open` | `open <slug>` — launch default browser on canonical PH URL | `--url-only` for piping | shipping |
| 4 | Raw feed dump | scraper parity | `feed raw` — emit the raw Atom XML | `--validate` flag asserts Atom shape | shipping |
| 5 | Product link via PH redirect | scraper parity | `info <slug> --external` — print the external product URL (via `r/p/<id>?app_id=339`) | `--open` chains to system browser | shipping |

## Absorbed (Tier B — local snapshot store, compounds over time)

| # | Feature | Best Source | Our Implementation | Added Value | Status |
|---|---------|-------------|--------------------|-------------|--------|
| 6 | Sync feed to local store | scraper crons (shashankpolanki daily cron) | `sync` — fetch `/feed`, upsert entries + capture a snapshot row | Incremental; preserves history, CSV-ready | shipping |
| 7 | Search products | MCP `search` (token) | `search <query>` — SQLite FTS5 over locally-persisted titles/taglines/authors | Works without network; regex-capable | shipping |
| 8 | List products by date range | scraper parity | `list --since 24h`, `--from YYYY-MM-DD`, `--to YYYY-MM-DD` | Historical slice of what /feed has featured | shipping |
| 9 | List products by author/maker | MCP `UserPosts` (token) | `list --author "Name"` | Covers maker or hunter name as seen in Atom | shipping |
| 10 | Export scraper-parity CSV | fernandod1 column set | `export --format csv` with columns: `id, slug, title, tagline, author, published, updated, discussion_url, external_url` | Match scraper output verbatim | shipping |
| 11 | Export JSON | any | `--json` on every list command + `export --format json` | Structured, pipe-friendly | shipping |
| 12 | `recent` / `new` live view | scraper parity | `recent --limit N` — live fetch, pure read (no store write) | Fresh every call; skip sync loop | shipping |
| 13 | Watch for new launches | Pipedream "new product" trigger | `watch` — sync, diff against previous snapshot, print new entries | Persists seen set; idempotent | shipping |
| 14 | Newest-first ordering | every tool | Default sort on `published` desc, configurable via `--sort` | Can pivot to `updated`, `author`, `title` | shipping |
| 15 | "Since last sync" summary | n8n/Pipedream trigger | `watch --since last-sync` | Prints count + titles of new-since-sync entries | shipping |
| 16 | Pagination | MCP pageInfo | `--limit`, `--offset` on all list commands | Typed exits if offset exceeds set | shipping |
| 17 | Doctor / health | every CLI | `doctor` — probe `/feed`, parse Atom, check store schema | `--json` exit code 2 on feed breakage | shipping |
| 18 | Deep link builder | scraper parity | `info <slug> --url` | Construct canonical URL from slug | shipping |
| 19 | Agent-native output | MCP parity | `--agent` bundle + `--compact` for bounded payloads | SKILL recipes show dotted `--select` patterns | shipping |
| 20 | Typed exit codes | GOAT convention | 0 ok, 2 invalid args, 3 not found, 4 rate limit, 5 upstream error, 7 needs-sync | Documented in `--help` | shipping |

## Absorbed (Tier C — CF-gated or token-gated, shipped as honest stubs)

Each stub emits a one-paragraph explanation naming the CF gate and pointing at the `/feed` alternative. Not a silent no-op; not a fake success; not a fake dataset.

| # | Feature | Best Source | Status | Stub Message Tells User |
|---|---------|-------------|--------|--------------------------|
| 21 | Post detail with full description + media | MCP `Post` | (stub — CF-gated) | "The product detail page is blocked by Cloudflare for non-browser clients. `info <slug>` gives you the /feed-level metadata; open <slug> sends you to the real page." |
| 22 | Comments on a post | MCP `PostComments` | (stub — CF-gated) | "Comments live on the HTML page, which Cloudflare blocks for automated clients. Future `auth login --chrome` will unlock this by importing your browser clearance cookie." |
| 23 | Historical daily/weekly leaderboard | scrapers | (stub — CF-gated) | "Leaderboard pages are CF-gated. `/feed` is live-only; run `sync` on a schedule to build your own history, then `list --from/--to`." |
| 24 | Topic feed | MCP `Topic`, scrapers | (stub — CF-gated) | "Topic pages are CF-gated. /feed doesn't accept a working category filter (verified — the `?category=` query is ignored by PH)." |
| 25 | User / maker profile | MCP `User` / `UserPosts` | (stub — CF-gated) | "Profile pages are CF-gated. `list --author '<name>'` searches by author name across everything you've synced." |
| 26 | Collections | MCP `Collection` / `Collections` | (stub — CF-gated) | "Collection pages are CF-gated. No Atom alternative." |
| 27 | Newsletter archive | scraper parity | (stub — CF-gated) | "Newsletter archive is CF-gated. No Atom alternative." |
| 28 | Search products by author (name text) | scraper parity | **shipping** as `list --author` — not stubbed | — |
| 29 | Upvote / comment / follow (writes) | MCP `Vote` / `Comment` mutations | (stub — token required) | "Write actions require a PRODUCT_HUNT_TOKEN. Explicit non-goal per this CLI's scope. Use the Product Hunt website directly." |

Note: rows 21-27 are shown in `--help` with their stub message. The generator's novel-feature tracker will see they are not actually wired to a real surface; we declare them explicitly so agent users aren't surprised by "missing" commands.

---

## Transcendence (Tier D — only possible because we persist /feed over time)

| # | Feature | Command | Why Only We Can Do This | Score |
|---|---------|---------|-------------------------|-------|
| 1 | **First-seen trajectory for a slug** | `trend <slug>` | PH's own site shows current day rank but buries history. /feed gives a live ranked list; persisting daily snapshots lets us reconstruct "when did this product first appear on the feed?" and "how many days did it stay?" — data no token-based tool computes. | 9/10 |
| 2 | **Launch-calendar view** | `calendar --week` | With daily snapshots, compose a week-at-a-glance showing which slugs were featured which days, grouped by first-seen day. Turns ephemeral /feed into a durable calendar. | 8/10 |
| 3 | **Maker burn chart** | `makers --top --since 30d` | Aggregate the `author` field across all snapshots in a window. Shows who's been most active on PH in the last N days — cannot be derived from a single /feed snapshot. | 7/10 |
| 4 | **Outbound-URL drift detector** | `outbound-diff <slug>` | Each /feed entry carries a product's external URL (through the `/r/p/<id>` redirect). Compare across snapshots; flag slugs whose external URL changes between sync cycles (common during beta→launch transitions or domain moves). | 7/10 |
| 5 | **Tagline grep** | `tagline-grep <pattern>` | `/feed` taglines are compact, high-signal domain descriptions. FTS + regex over the full local tagline archive gives a semantic filter no other PH tool has ("show me every AI-agent tagline from the last 90 days"). | 8/10 |
| 6 | **New-since-last-sync report** | `watch` | Computes the delta between the current /feed and the most recent snapshot. Bounded, idempotent, cron-friendly. | 8/10 |
| 7 | **Author-co-occurrence graph** | `authors related --to <name>` | With many snapshots, find authors who repeatedly appear alongside a given author in feed batches — a rough social-signal graph from purely public data. | 6/10 |
| 8 | **Dead-slug reaper** | `reap --older-than 90d` | Store hygiene — flag entries that haven't been re-seen in N days. Keeps local SQLite lean while preserving the "first seen" record. | 5/10 |

All Transcendence rows (minimum 5 required by the skill) score ≥5/10 and surface in the Phase Gate 1.5 showcase.

---

## Narrative hooks for README / SKILL

**Headline:** "Product Hunt without the OAuth dance — watch today's launches, keep your own history, compose views PH itself doesn't expose."

**Quick Start:**
1. `producthunt-pp-cli sync` — pull today's /feed into your local store.
2. `producthunt-pp-cli today --limit 10 --json --select 'id,slug,title,tagline,author'` — agent-friendly shortlist.
3. `producthunt-pp-cli trend seeknal` — see when a slug first appeared and how long it stayed.

**Auth narrative:** None. The CLI runs against the public `/feed`. An optional `PRODUCT_HUNT_TOKEN` env var is documented only as a possible future enrichment — not wired in this build.

**Trigger phrases:** "use producthunt", "today's top on product hunt", "what launched on product hunt", "product hunt trend for <slug>", "run producthunt-pp-cli".

**Known Gap (README-level):** CF-gated HTML routes (post detail, comments, leaderboards, topic/user pages, newsletter archive, collections) are shipped as explicit stubs. A future `auth login --chrome` pass could import Cloudflare clearance from the user's Chrome profile and unlock them; not in this version.
