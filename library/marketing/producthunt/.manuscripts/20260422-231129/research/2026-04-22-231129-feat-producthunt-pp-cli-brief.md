# Product Hunt CLI Brief

## API Identity
- **Domain:** Product Hunt (producthunt.com) — community-driven product launch platform. Daily leaderboard, upvotes, comments, topics, collections, maker/hunter profiles, newsletters.
- **Users:** makers/indie founders (launching), hunters (submitting), VCs/scouts (deal flow), tech enthusiasts (discovery), PMs/marketers (competitor tracking).
- **Data profile:** highly relational — Post ↔ Maker/Hunter ↔ Topic ↔ Collection ↔ Comment; time-series value in tracking rank/upvotes across days; editorial content (newsletters, stories) is searchable long-tail.
- **Surface choice:** public website (SSR HTML + `/feed` Atom). **Not** the official GraphQL API at `api.producthunt.com/v2/api/graphql`, which requires OAuth registration and imposes a 6,250 complexity-pts / 15-min budget that chokes bulk reads.

## Reachability Risk
- **Low.** Direct WebFetch probes to `/`, `/leaderboard/daily/YYYY/M/D`, `/feed`, `/posts/<slug>`, `/topics/<slug>`, `/@<handle>`, and `/collections` all return 200 with real content. No Cloudflare/DataDome challenge page observed.
- Third-party scraping guides note PH is JS-heavy — Browser-Sniff will confirm whether SSR HTML + `/feed` covers our needs or whether an internal `/frontend/graphql` persisted-query path exists.
- Evidence: first research pass, 2026-04-22, 7/7 endpoint probes returned 200 HTML/Atom without challenge.

## Top Workflows
1. **Today's top launches** — `ph today` / `ph top` with rank, upvotes, comments, topics. One command that replaces opening the browser each morning.
2. **Historical leaderboard** — `ph leaderboard daily --date YYYY-MM-DD` and weekly/monthly variants. PH's own UI surfaces the page but buries the navigation; a CLI backed by a local store keeps the archive queryable offline.
3. **Topic watchlist** — `ph topic artificial-intelligence --since 7d --min-votes 200` for VCs and PMs tracking a vertical.
4. **Post-launch triage** — `ph comments <slug> --json | jq` so a maker can export and review 200+ comments on launch day without OAuth setup.
5. **Maker/hunter history** — `ph user <handle>` shows everything a person has launched, hunted, or curated.
6. **Newsletter archive** — `ph newsletter latest` and full-text search across the daily/weekly editorial archive.

## Table Stakes
All existing tools (MCP, npm SDKs, PyPI clients) require OAuth tokens. Features they collectively cover:
- List posts with filters (date, topic, featured)
- Get post detail by slug or ID
- List comments for a post
- Get user profile by handle
- List topics / categories
- List collections
- Vote / get vote counts (read-only in our CLI — no write)
- Search products
- Rate-limit status

Scrapers (token-free) additionally cover:
- Daily leaderboard scraping with historical retention
- Topic page scraping with pagination
- Newsletter archive scraping

**Our bar:** match every read-side feature of the official token-required SDKs **without** a token, plus everything the scrapers do, plus offline search, `--json`, `--select`, agent-native error paths, typed exit codes, and a local SQLite store.

## Data Layer
- **Primary entities:** `post`, `user`, `topic`, `collection`, `comment`, `newsletter_issue`, `leaderboard_snapshot`.
- **Time-series tables:** `post_snapshot(post_id, ts, upvotes, comments_count, day_rank)`, `leaderboard_snapshot(date, post_id, rank, score, comments_count)`. These unlock trend commands no token-gated tool provides — PH's own API doesn't expose historical rank trajectory.
- **Sync cursor:** per-day leaderboard (a date is the cursor); per-post (post slug + last-seen upvote count).
- **FTS:** `posts_fts` (name + tagline + description + topics), `newsletters_fts` (full body), `comments_fts` (for maker comment triage).

## User Vision
User elected default intelligence ("No, let's go"). No extra constraints beyond the argument itself, which specified: website-backed, NOT the official GraphQL API, full fresh redo with no artifact reuse.

## Product Thesis
- **Name:** `producthunt-pp-cli`.
- **Why it should exist:** every current Product Hunt CLI/SDK/MCP requires OAuth app registration and a per-user access token, and imposes a stingy complexity budget. A Printed CLI built on the public SSR website surface plus the `/feed` Atom feed gives token-free read access at CLI speeds, adds a local SQLite data layer for trend/history views PH itself doesn't expose, and is agent-native by construction.
- **"I need this" moment:**
  - Maker at 8am the day after launch: `producthunt-pp-cli comments my-product --json | jq '.comments | map(select(.body | test("bug|broken"; "i")))'` — triage feedback without an OAuth app.
  - VC scout nightly cron: `producthunt-pp-cli topic ai-agents --since 7d --min-votes 200 --json` — rate-limited token would burn out; SSR scrape + local cache doesn't.
  - Any maker: `producthunt-pp-cli trend <slug>` — intraday rank trajectory from local snapshots; PH's site hides this entirely.

## Build Priorities
1. **Data layer first.** Posts, users, topics, collections, comments, newsletter issues, and the two snapshot tables. Without this, transcendence features can't exist.
2. **Primary discovery commands.** `today`, `leaderboard daily|weekly|monthly`, `post <slug>`, `user <handle>`, `topic <slug>`, `collection <slug>`, `newsletter latest|archive`, `search <query>`. All powered by SSR HTML parsing plus `/feed` for live incrementals.
3. **Transcendence commands.** `trend <slug>` (rank/upvote trajectory over days), `topic watch --min-votes N --since` (composable historical view), `maker-history <handle>` joining posts + comments + collections, `newsletter grep <term>` (full-text search across editorial archive), `compare <slug1> <slug2>` (rank trajectory side-by-side).
4. **Shipcheck polish.** `--json`, `--select`, `--csv`, `--compact`, typed exit codes, `doctor`, `auth` (no-op / diagnostic only — there is no account to log into for our read-only surface; document this).

## Known Follow-Ups
- Phase 1.7 Browser-Sniff will confirm whether an internal BFF/GraphQL path gives cleaner JSON than HTML scraping for high-volume commands. If it does, prefer it; if not, commit to SSR + Atom + JSON-LD parsers.
- Phase 1.5a ecosystem search (running in parallel) will produce the full absorb catalog.
