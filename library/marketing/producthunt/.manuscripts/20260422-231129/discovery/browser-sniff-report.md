# Product Hunt Browser-Sniff Report

## User Goal Flow
- **Primary goal:** Discover and browse Product Hunt anonymously — today's top, historical leaderboard, product detail with comments, topic feed, user profile, newsletter archive, search.
- **Steps attempted:**
  1. Open https://www.producthunt.com/ in Playwright Chromium via `browser-use open`.
  2. Install fetch/XHR interceptors to catch any BFF or GraphQL traffic.
  3. Scroll homepage to trigger lazy-loaded content.
  4. Planned: walk to `/leaderboard/daily/2026/4/21`, `/topics/artificial-intelligence`, a `/posts/<slug>` detail, `/@<handle>` profile, `/collections`, newsletter archive.
- **Steps completed:** only step 1 and the scroll of step 3. Playwright Chromium was served a Cloudflare "Just a moment…" challenge page and never reached the real homepage even after 28+ seconds of waiting.
- **Secondary flows attempted:** none.
- **Coverage:** 1 of 7 planned interactive steps completed; effectively 0 meaningful capture of the HTML surface.

## Pages & Interactions
- `https://www.producthunt.com/` — Playwright opened the URL; Cloudflare interposed with the challenge page. No product anchors, no SSR HTML, `document.title` stuck at "Just a moment…", `body.innerText.length=267`. Two scrolls attempted; no XHR activity because no real page had loaded.

## Browser-Sniff Configuration
- Backend: browser-use 0.12.5 (CLI mode, no LLM key needed).
- Anonymous sniff — `AUTH_SESSION_AVAILABLE=false`; no Chrome profile loaded.
- Pacing: 1s default; irrelevant because interceptors never fired.
- Proxy pattern detection: **not run** (no captured traffic to classify).

## Endpoints Discovered
| Method | Path | Status | Content-Type | Reachable without browser? |
|--------|------|--------|--------------|----------------------------|
| GET | https://www.producthunt.com/feed | 200 | application/atom+xml; charset=utf-8 | **Yes** (confirmed via curl + Chrome UA, 44 KB, 50 entries) |
| GET | https://www.producthunt.com/ | 403 | text/html | No (CF-blocked) |
| GET | https://www.producthunt.com/leaderboard/daily/{YYYY}/{M}/{D} | 403 | text/html | No (CF-blocked) |
| GET | https://www.producthunt.com/topics/{slug} | 403 | text/html | No (CF-blocked) |
| GET | https://www.producthunt.com/topics/{slug}.atom | 403 | text/html | No (CF-blocked) |
| GET | https://www.producthunt.com/topics/{slug}/feed | 403 | text/html | No (CF-blocked) |
| GET | https://www.producthunt.com/newest | 403 | text/html | No (CF-blocked) |
| GET | https://www.producthunt.com/discussions.rss | 403 | text/html | No (CF-blocked) |
| POST | https://api.producthunt.com/v2/api/graphql | N/A (GET probe 404) | application/json | Optional / token-gated, not used by default |

Phase 1 research (WebFetch, different TLS/UA fingerprint) did return 200 for the HTML paths. WebFetch is a separate Anthropic-side fetcher with different characteristics; a Go HTTP client running from a user's machine will not replicate that bypass.

## Traffic Analysis
- **Protocols observed:** `atom_feed` (public, 50 entries) with canonical product slugs as `<link>` targets.
- **Protection signals:** Cloudflare Turnstile interactive challenge on every HTML route. `cdn-cgi/challenge-platform` detected. No CAPTCHA solved.
- **Auth signals:** None required for `/feed`. Official GraphQL API would require `Authorization: Bearer <PRODUCT_HUNT_TOKEN>`; explicit non-goal per the user's argument.
- **Reachability mode:** `atom_primary` (custom label — see `traffic-analysis.json`). Equivalent to "public feed is the only replayable surface without browser clearance".
- **Warnings:**
  - Cloudflare-challenge on HTML pages means deep commands (post detail, comments, leaderboard, topic feed, user profile) are not reachable from a standard Go HTTP client.
  - A future enhancement path exists: `auth login --chrome` could import `cf_clearance` + browser-matched TLS fingerprint (e.g., via `utls`) to unlock HTML routes. Out of scope for this run.
- **Candidate commands:** derived from `/feed` only — today, recent, list, search, export, watch (diff vs last sync), trend (rank/vote over snapshots), info <slug>, open <slug>.

## Coverage Analysis
- **Exercised:** `/feed` only (via curl after Playwright failure; 50 entries).
- **Missed but reachable via /feed:** most common "today's top" queries — each entry exposes slug, title, tagline, author, published/updated timestamps, canonical product URL.
- **Missed entirely:** historical leaderboards (date-keyed HTML pages), topic feeds (HTML only), user profile pages, post detail pages with comments, collections, newsletter archive, search.
- **Compared to Phase 1 brief:** the brief imagined a broader feature set (see absorb manifest). The Atom-first scope covers the table-stakes "today's top" workflow reliably and uses local snapshots to compose rank-trajectory views. Deep-post commands (comments, leaderboard history, profiles) are stubbed in the first cut and re-approach-ready in a future pass with proper browser-clearance support.

## Response Samples

### GET /feed (200, application/atom+xml)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom">
  <id>tag:www.producthunt.com,2005:/feed</id>
  <link rel="alternate" type="text/html" href="https://www.producthunt.com"/>
  <link rel="self" type="application/atom+xml" href="https://www.producthunt.com/feed"/>
  <title>Product Hunt — Latest</title>
  <entry>
    <id>tag:www.producthunt.com,2005:Post/1129094</id>
    <published>2026-04-21T09:02:49-07:00</published>
    <updated>2026-04-22T23:02:14-07:00</updated>
    <link rel="alternate" type="text/html" href="https://www.producthunt.com/products/seeknal"/>
    <title>Seeknal</title>
    <content type="html">
      &lt;p&gt;Data &amp; AI/ML CLI for pipelines and NL queries&lt;/p&gt;
      &lt;p&gt;&lt;a href="https://www.producthunt.com/products/seeknal?...&quot;&gt;Discussion&lt;/a&gt; | &lt;a href="https://www.producthunt.com/r/p/1129094?app_id=339"&gt;Link&lt;/a&gt;&lt;/p&gt;
    </content>
    <author><name>Fitra Kacamarga</name></author>
  </entry>
  ...50 entries total
</feed>
```

Per-entry fields extractable (all confirmed present in the live feed):
- `id` → `tag:www.producthunt.com,2005:Post/<numeric>` (numeric Post ID usable as a stable key)
- `published` / `updated` — ISO 8601 with TZ
- `link[rel=alternate]` → canonical product URL (slug-derived)
- `title` — product name
- `content` — HTML-wrapped tagline + a Discussion link (to PH) and an external Link (the product's website, via PH's redirect endpoint)
- `author/name` — maker/hunter display name

### All other HTML paths (403)

Cloudflare Turnstile HTML challenge. Not a real response.

## Rate Limiting Events
None observed. `/feed` returned instantly with no rate-limit headers. No 429s during probes.

## Authentication Context
- No authenticated session used.
- `AUTH_SESSION_AVAILABLE=false` — user declined the "I'm logged in" option at briefing time.
- Session state file (`session-state.json`) not written.
- Cookie auth validation (Step 2d): **skipped** — browser-sniff was anonymous.

## Bundle Extraction
Not attempted. The CF challenge page blocked access to the real JS bundle. Given the user's decision to scope to Atom-first, bundle extraction would not have altered the runtime surface.

## Verdict
- **Replayable surface:** `GET /feed` only.
- **Runtime mode:** `atom_primary` — standard HTTP GET against one endpoint, Atom XML parsing, local store for history.
- **No resident browser.** No `chromedp`/Playwright sidecar. No CF clearance cookie import in this version.
- **Future enhancement note:** `auth login --chrome` + browser-matched TLS fingerprint (uTLS) could unlock HTML routes; documented in the generated README as a Known Gap with a clear upgrade path.
