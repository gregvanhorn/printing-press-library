# Smartlead CLI Brief

## API Identity
- **Domain**: Cold-email outreach platform (campaigns + leads + sender mailboxes + sequences + master inbox + analytics + deliverability tooling)
- **Users**: SDR teams, agencies (whitelabel/multi-client), founders running outbound, ops engineers building reply-detection pipelines
- **Data profile**: Hierarchical with high-cardinality joins. `client → campaign → (sequence_step, schedule, email_account_link, lead) → message/event → analytics_rollup`. Time-series stats are append-only (warmup, deliverability, replies). Agencies run multiple API keys.
- **Base URLs**: `https://server.smartlead.ai/api/v1` (main), `https://smartdelivery.smartlead.ai/api/v1` (spam-test)
- **Auth**: `?api_key=<key>` query string. No header auth, no OAuth.
- **Rate limit**: 10 req / 2 sec ≈ 60 req/min per key. 429 on overflow. Per-client keys (whitelabel) each get their own 60/min budget.

## Reachability Risk
- **None**. Live verified with user's admin key — HTTP 200, returns campaign list including "Roofing + SupplyHawk".

## Top Workflows
1. **Daily campaign analytics roll-up** — pull all campaigns + per-campaign reply/bounce/open rates, push to BI/sheet
2. **Bulk lead loading** — ingest CSV → push to campaign in 400-lead chunks (API cap), tag, dedupe
3. **Reply triage** — pull master-inbox unread replies, classify (positive/objection/OOO/unsub), reply or push to subsequence
4. **Mailbox health monitoring** — check warmup-stats and sending-limits across all email accounts; pause underperforming senders
5. **Deliverability spot-checks** — run spam-test placement, check DKIM/SPF/DMARC/blacklists per sender domain
6. **Agency client roll-up** — aggregate stats across all client API keys for portfolio-level reporting

## Table Stakes (must absorb everything LeadMagic + bcharleson have)
- LeadMagic MCP: 113 tools across 9 modules
- bcharleson CLI: 142 commands
- Combined: **~155 unique endpoints** across **13 resource groups**: campaigns, email_accounts, leads, lead_lists, sequences, analytics, statistics, webhooks, clients, crm, master_inbox, smart_senders, smart_delivery (+ the smart-delivery DNS/blacklist checker subgroup)

## Data Layer (local SQLite)
- **Primary entities**: campaigns, leads, email_accounts, sequences, webhooks, clients, master_inbox_messages
- **Time-series rollups** (append-only, snapshot daily): campaign_analytics_daily, warmup_stats_daily, deliverability_snapshots
- **Sync cursor**: `updated_at` per entity; for analytics/stats, snapshot-by-date
- **FTS5 search**: leads (email + name + custom_fields), master_inbox_messages (subject + body), campaigns (name + tags)
- **High-cardinality joins**: lead↔campaign (n:m), email_account↔campaign (n:m, rotation pool), lead↔reply (1:n), client↔campaign (1:n)

## User Vision
User is the admin on a live Smartlead account ("Roofing + SupplyHawk" + others visible). API key is admin-tier. No specific vision provided beyond "build a CLI for Smartlead." Scope inferred from full admin access.

## Codebase Intelligence
- **Source absorbed**: LeadMagic/smartlead-mcp-server (TS, 18★) + bcharleson/smartlead-cli (TS, 2★) + jzakirov/smartlead-cli (Py)
- **Auth pattern from source**: `?api_key=` appended to every request URL. Single env var: `SMARTLEAD_API_KEY`.
- **Path inconsistency**: source-of-truth disagreements between LeadMagic and bcharleson on a few paths (`/client` vs `/client/`, `/leads/email` vs `/leads/by-email`, PUT vs PATCH on `/webhooks/{id}`, singular `/sequence` vs plural `/sequences`). Resolution strategy: try both as a fallback chain in the client; prefer llms.txt canonical when available.
- **Rate limit handling**: per-source-rate-limiting MUST use `cliutil.AdaptiveLimiter` and surface `*cliutil.RateLimitError` (60/min budget).
- **Pagination**: most list endpoints take `?offset` + `?limit`. `/campaigns/` does NOT paginate (returns full list — large agencies pull thousands).
- **Lead chunk cap**: `POST /campaigns/{id}/leads` is hard-capped at 400 leads/request. CLI must auto-chunk.

## Product Thesis
- **Name**: `smartlead-pp-cli`
- **Why it should exist**:
  - Two open-source MCPs exist but are MCP-only (agent-mounted, not human-friendly).
  - bcharleson CLI is closest competitor but: no local SQLite cache, no FTS, no `--json --select`, no offline analytics, no cross-campaign aggregation, no auto-chunking on lead push, no rate-limit-aware retry.
  - Agencies need cross-client + cross-campaign roll-ups the API doesn't expose — only possible with local sync.
- **Differentiators**:
  - Every Smartlead feature, agent-native (`--json --select --csv --quiet`)
  - Local SQLite + FTS5 → sub-second search across leads/replies/campaigns offline
  - Auto-chunking lead push (handles >400 transparently)
  - Adaptive rate limiter respects 60/min, surfaces typed `RateLimitError`
  - Cross-campaign aggregations (stale-leads, overlap-leads, mailbox-burnout, reply-velocity) — only possible with local sync
  - Multi-key (agency) support — query across all client API keys with one command

## Build Priorities
1. **Foundation**: Internal YAML spec from inventory (155 endpoints, 13 resources), data layer covering 7 primary entities + 3 time-series tables, FTS5 over leads + master_inbox.
2. **Absorb**: All 155 endpoints as typed commands. `--json`, `--select`, `--csv`, `--dry-run`, `--limit` on every list. Auto-chunking on lead-add. Adaptive rate limiter.
3. **Transcend**: 7 novel commands (stale-leads, overlap-leads, mailbox-burnout, reply-velocity, deliverability-drift, reach-budget, agency-roll-up).
4. **Polish**: README, SKILL.md, MCP server exposing the Cobra tree.
