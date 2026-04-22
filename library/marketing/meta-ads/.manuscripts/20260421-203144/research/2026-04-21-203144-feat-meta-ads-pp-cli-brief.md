# Meta Ads CLI Brief

## API Identity
- **Domain:** Meta Marketing API (Graph API v22.0 / v23.0 — Facebook + Instagram ad management).
- **Users:** Performance marketers, growth engineers, agencies, founder-led SMBs running their own paid social. Power users frustrated with the Ads Manager UI's click depth.
- **Data profile:** Hierarchical — account → campaign → adset → ad → creative. Insights are time-series metrics queryable at every level. High-gravity entities: campaigns, ad sets, insights, custom audiences, creatives. Meta exposes per-request change history via the `activities` edge.

## Reachability Risk
- **None.** Meta Graph API is one of the most stable commercial APIs. The magoosh-founder CLI is running against live production without reachability issues. Rate limits apply (9000 calls / 5 min for standard tier, 60 for dev tier) but are well-documented and predictable.

## Top Workflows
1. **Budget autopilot on ROAS signal** — Pull campaign performance over a 14d window, evaluate ROAS vs target + frequency saturation + budget utilization, recommend INCREASE/DECREASE/HOLD/FLAG per campaign, apply the batch with confirmation.
2. **Daily performance review** — Sync recent insights to a local store, run a delta since yesterday, spot movers (big spend changes, ROAS swings, frequency climbing past 3.5).
3. **Attribution double-check** — Verify `omni_purchase` vs `purchase` vs `offsite_conversion.fb_pixel_purchase` selection per campaign so dashboards don't overcount revenue.
4. **Creative fatigue triage** — Find ads with rising CPM + rising frequency + declining CTR that need refresh.
5. **Audience operations** — Discover interests, build custom/lookalike audiences, inspect targeting overlap between ad sets.

## Table Stakes (features every competitor CLI/MCP has — must match)
- Account discovery (`me/adaccounts`), campaign/adset/ad CRUD, creative and image upload, insights at every level with time_range and breakdowns, interest/geo/behavior search, custom audience creation, lookalike audiences, activity log, ad previews across placements, pause/resume, budget schedules.
- Structured output for AI agents (JSON, narrow field projection), pagination handling, proper auth error mapping.

## Data Layer (what deserves local SQLite)
- **Primary entities:** `accounts`, `campaigns`, `adsets`, `ads`, `creatives`, `insights` (time-keyed by date + level), `audiences`, `interests` (cached lookup), `decisions` (the audit log), `activities` (change events).
- **Sync cursor:** Meta has no global `updated_since` filter, but `activities` edge gives change events with `event_time`. Use `MAX(event_time)` from the local activities table as the delta cursor. Fall back to full-sync of hierarchy when stale.
- **FTS/search:** Campaign names, adset names, ad names, creative bodies/headlines, audience names — all benefit from FTS5. Decision log reasoning text should also be FTS-indexed.

## Codebase Intelligence (from magoosh-founder deep-read)
- **Source:** `/Users/pejman/code/magoosh-founder/src/founder_agent/platforms/meta_ads.py` (1104 lines, production-validated).
- **Auth:** OAuth2 long-lived token (60d). App ID/secret in env, token stored in JSON at credentials-dir/meta_token.json. Scopes: `ads_read`, `ads_management`, `business_management`. Redirect URI must be HTTPS self-signed at fixed port `127.0.0.1:8477` (whitelisted in app settings).
- **Ground-truth field selections:**
  - Campaigns: `id, name, status, effective_status, daily_budget, lifetime_budget, objective, bid_strategy`
  - Insights: `campaign_id, campaign_name, spend, conversions, purchase_roas, actions, action_values, clicks, impressions, reach, frequency, cpm`
- **Purchase attribution priority (de-dup to prevent double-count):**
  `omni_purchase` > `purchase` > `offsite_conversion.fb_pixel_purchase` > `offsite_conversion.purchase`. Pick the first present in `actions`/`action_values`/`purchase_roas` maps; never sum across types.
- **Frequency capacity heuristic:** saturation=3.5, warning=2.5. Decision tree in order: no freq→unknown → spending_limited flag→headroom-high → paused flag→no-headroom-low → utilization<70%→no-headroom-low (not budget-constrained) → freq>3.5→no-headroom-high (saturated) → freq≥2.5→headroom-medium → else headroom-high.
- **Units:** Meta stores budgets in cents. Convert via `cents/100.0` on read, `int(dollars*100)` on write.
- **Date window:** `--days N` → start = yesterday - (N-1), end = yesterday (completed days, matches Ads Manager).
- **Error code map:** 190/102/463/467→auth-failed, 10/200/294→permission-denied, 4/17/32/613→rate-limited, else platform-error.
- **Meta-specific gotcha:** `activities` edge capped at 30 days — magoosh-founder uses 29 days to stay inside.
- **Decision log:** JSONL append-only at `<logs>/decisions.jsonl`. Three entry types: budget decision, status_change, decision_analysis (post-mortem linked by `original_log_id`). Each budget decision carries `validation_kpis`, `follow_up_date`, `expected_outcome` — the next-pass engine reads the log to surface stale decisions.
- **Recommendations engine:** 4-way (INCREASE/DECREASE/HOLD/FLAG) × 3 strategies (conservative/moderate/aggressive). ROAS target diff_pct gates efficiency; utilization<70% overrides capacity ("KEY INSIGHT: not budget-constrained if not using the budget"); learning period (status_issues + changes>20% within 14d) dampens confidence.

## Product Thesis
- **Name:** `meta-ads-pp-cli` (binary), brand it as "the Meta Ads CLI for founders" in the README.
- **Headline:** Every Meta Ads tool's feature set, plus a local SQLite brain and a budget autopilot that reasons about ROAS + frequency + utilization the way a senior performance marketer would.
- **Why it should exist:** Five competing MCPs already cover CRUD and insights. None of them ship:
  1. A local SQLite store you can query with SQL or FTS.
  2. A structured recommendations engine with per-campaign reasoning, validation KPIs, and a follow-up audit log.
  3. Double-count detection against `omni_purchase` vs `purchase` confusion.
  4. Frequency-based capacity signals that drive real budget decisions (not just raw ROAS).
  5. Agent-native output with `--json`, `--select`, `--agent`, typed exit codes, and `--dry-run` on every mutation.
- **magoosh-founder's soul** (budget decisioning based on ROAS + capacity) runs in this CLI's heart. Everything else wraps around it.

## Build Priorities
1. **Priority 0 (data layer):** SQLite schema for `accounts`, `campaigns`, `adsets`, `ads`, `creatives`, `insights`, `audiences`, `activities`, `decisions`. FTS5 virtual tables for names/bodies/reasoning. `sync` command that walks the hierarchy + pulls insights + appends to activities delta.
2. **Priority 1 (absorb — match every competitor feature):**
   - Account: list, details, activities, pages.
   - Campaigns: list/get/create/update/pause/resume/delete, insights.
   - Ad sets: list/get/create/update/delete, insights, targeting inspection.
   - Ads: list/get/create/update, insights, previews.
   - Creatives: list/get/create, image upload, preview.
   - Audiences: list/get, custom audience create, lookalike create, interest/geo/behavior/demographic search.
   - Budget schedules: list/create.
   - Health/diagnostic: `doctor`, `auth status`, `auth login --chrome` or device flow fallback, `diagnose-campaign` (readiness), `check-account-setup`.
   - Insights at every level with `--breakdowns`, `--time-range`, `--attribution-windows`, `--level`, CSV/JSON export.
3. **Priority 2 (transcendence — features nobody else has):**
   - `recommend` — ROAS/frequency/utilization decisioning, outputs the full reasoning struct.
   - `apply` — batch budget change with confirmation + validation KPIs logged.
   - `verify roas` — double-count detection across purchase action types.
   - `fatigue` — rising CPM × rising freq × declining CTR over a rolling window.
   - `pace` — hourly budget burn rate vs target with ETA-to-cap.
   - `learning` — campaigns in/out of learning phase by recent change size.
   - `overlap` — audience overlap analysis between two adsets.
   - `history` — decision log search with FTS + follow-ups-due filter.
   - `alerts` — local-data threshold watchers (freq>3.5, ROAS<target, spend spike).
   - `query` — raw SQL REPL / one-shot against the local store.
   - `rollup` — aggregate across multiple ad accounts.
   - `decision-review` — read a decision log entry, compare against current metrics, write a `decision_analysis` outcome back.

## User Vision
User operates magoosh-founder in production. Wants a Go rewrite that (a) keeps the decisioning brain, (b) absorbs the wider Meta Marketing API surface the Python CLI didn't touch (ad sets, ads, audiences, creatives, activities), (c) adopts the printing-press conventions (SQLite, `--json/--select/--agent`, agent-native, typed exit codes, `--dry-run` everywhere). The CLI should be a plausibly-better replacement for magoosh-founder's Meta half, not a downgrade.
