# Meta Ads CLI — Absorb Manifest

## Source Tools Cataloged

| Tool | Tools | Notable coverage |
|---|---|---|
| [pipeboard-co/meta-ads-mcp](https://github.com/pipeboard-co/meta-ads-mcp) | 29 | Full CRUD, creative upload, interest/geo/behavior/demographic search, budget schedules, unified search |
| [brijr/meta-mcp](https://github.com/brijr/meta-mcp) | 25 | Export JSON/CSV, lookalike audiences, diagnose_campaign_readiness, special ad categories, health_check |
| [hashcott/meta-ads-mcp-server](https://github.com/hashcott/meta-ads-mcp-server) | 24 | Activities log (change history), ad previews, batch adset fetch |
| [gomarble-ai/facebook-ads-mcp-server](https://github.com/gomarble-ai/facebook-ads-mcp-server) | 21 | Read-focused, pagination helper, activities log |
| [attainmentlabs/meta-ads-cli](https://github.com/attainmentlabs/meta-ads-cli) | 6 | YAML-declared campaigns, dry-run validation, PAUSED-by-default safety, interest search |
| [magoosh-founder](file:///Users/pejman/code/magoosh-founder) | — | **Production-validated.** ROAS decisioning, frequency capacity, attribution dedup, decision log, validation KPIs, follow-up timing |

## Absorbed (match or beat everything that exists)

Every row below ships in our CLI. Where a competitor has it, we match the feature and beat it with SQLite persistence, `--json/--select/--agent` output, `--dry-run`, typed exit codes, and offline query.

### Accounts & Pages

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---|---|---|---|
| 1 | List accessible ad accounts | pipeboard `get_ad_accounts`, hashcott `list_ad_accounts` | `accounts list` → `GET /me/adaccounts` | Cached locally; `--json`; typed exit codes |
| 2 | Ad account details | pipeboard `get_account_info` | `accounts get <id>` → `GET /act_{id}` | `--select` field narrowing |
| 3 | List associated pages | pipeboard `get_account_pages` | `accounts pages <id>` → `GET /act_{id}/promote_pages` | SQLite join with campaigns |
| 4 | Account activity / change log | hashcott `get_activities_by_adaccount`, gomarble `get_activities_by_adaccount` | `accounts activities <id> --days N` → `GET /act_{id}/activities` | Appends to local `activities` table; drives `sync --since` delta cursor |
| 5 | Assigned users | (none) | `accounts users <id>` → `GET /act_{id}/assigned_users` | New coverage |

### Campaigns

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---|---|---|---|
| 6 | List campaigns | pipeboard `get_campaigns`, magoosh-founder `ads campaigns` | `campaigns list --status <enabled\|paused\|all> --active-days N` | `active-days` filter from magoosh-founder; status filtering; SQLite-backed FTS |
| 7 | Campaign details | pipeboard `get_campaign_details`, hashcott `get_campaign_by_id` | `campaigns get <id>` | `--select`, `--json` |
| 8 | Create campaign | pipeboard `create_campaign`, brijr `create_campaign`, attainment `create` | `campaigns create --name --objective --status --budget --bid-strategy --special-categories` | `--dry-run`, PAUSED default (attainment pattern), `--from-yaml <file>` (attainment pattern) |
| 9 | Update campaign | brijr `update_campaign` | `campaigns update <id> --name --budget --status` | `--dry-run`, diff preview |
| 10 | Pause campaign | brijr `pause_campaign`, magoosh-founder `ads pause` | `campaigns pause <id> --reason <text> --confirmation CONFIRM` | Writes `status_change` decision-log entry |
| 11 | Resume / unpause | brijr `resume_campaign`, magoosh-founder `ads unpause` | `campaigns resume <id> --reason <text>` | Decision-log entry |
| 12 | Delete campaign | attainment `delete` | `campaigns delete <id> --confirmation CONFIRM` | `--dry-run` |
| 13 | Campaign-level insights | pipeboard `get_insights`, hashcott `get_campaign_insights` | `campaigns insights <id> --days N --breakdowns --attribution-windows` | Magoosh-founder field list, attribution dedup built in |
| 14 | Budget schedules | pipeboard `create_budget_schedule` | `campaigns budget-schedule list/create <id> --budget-value --time-start --time-end` | `--dry-run` |
| 15 | Diagnose campaign readiness | brijr `diagnose_campaign_readiness` | `campaigns diagnose <id>` | Local-data deltas (warns on freq>3.5, utilization<70%, etc.) |

### Ad Sets

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---|---|---|---|
| 16 | List ad sets (by account or campaign) | pipeboard `get_adsets`, hashcott `get_adsets_by_adaccount` | `adsets list --account <id> --campaign <id>` | SQLite join + FTS |
| 17 | Ad set details | pipeboard `get_adset_details`, hashcott `get_adset_by_id` | `adsets get <id>` | `--select`, `--json` |
| 18 | Batch fetch ad sets | hashcott `get_adsets_by_ids` | `adsets batch <id1,id2,...>` | `--json` array |
| 19 | Create ad set | pipeboard `create_adset`, brijr `create_ad_set` | `adsets create --campaign <id> --name --budget --targeting <json> --optimization-goal --bid-strategy` | `--dry-run`, `--from-yaml <file>` |
| 20 | Update ad set | pipeboard `update_adset` | `adsets update <id> --budget --frequency-caps --targeting --bid-strategy` | `--dry-run` |
| 21 | Delete / pause ad set | — | `adsets delete <id>`, `adsets pause <id>` | Decision-log entry |
| 22 | Ad set insights | hashcott `get_adset_insights` | `adsets insights <id> --days N` | Magoosh-founder field list |

### Ads & Creatives

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---|---|---|---|
| 23 | List ads (account/campaign/adset) | pipeboard `get_ads`, hashcott `get_ads_by_*` | `ads list --account <id> --campaign <id> --adset <id>` | SQLite join + FTS |
| 24 | Ad details | pipeboard `get_ad_details` | `ads get <id>` | `--select`, `--json` |
| 25 | Create ad | pipeboard `create_ad` | `ads create --adset <id> --creative <id> --name --status` | `--dry-run`, PAUSED default |
| 26 | Update ad | pipeboard `update_ad` | `ads update <id> --status --bid-amount` | `--dry-run` |
| 27 | Ad insights | hashcott `get_ad_insights` | `ads insights <id> --days N` | Attribution dedup built in |
| 28 | Ad previews | hashcott `get_ad_previews` | `ads preview <id> --format FEED\|STORY\|REELS...` | All placements, saves HTML |
| 29 | List creatives | pipeboard `get_ad_creatives`, hashcott `get_adcreatives_by_adaccount` | `creatives list --account <id> --ad <id>` | `--select`, `--json` |
| 30 | Create creative | pipeboard `create_ad_creative`, brijr `create_ad_creative` | `creatives create --account <id> --image-hash --page-id --link-url --message --headlines <csv> --descriptions <csv> --cta-type` | External URL support (brijr pattern); `--dry-run` |
| 31 | Update creative | pipeboard `update_ad_creative` | `creatives update <id> --headlines --descriptions --dynamic-spec` | `--dry-run` |
| 32 | Upload ad image | pipeboard `upload_ad_image` | `creatives upload --account <id> --image-path <path> --name` | `--dry-run` (returns image_hash it would yield), SHA256 de-dup check |
| 33 | List ad images | hashcott `get_ad_images` | `creatives images <account_id>` | `--select` |

### Audiences & Targeting

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---|---|---|---|
| 34 | List custom audiences | brijr `list_audiences` | `audiences list --account <id>` | SQLite cache |
| 35 | Audience details | brijr `get_audience_info` | `audiences get <id>` | `--select`, health status |
| 36 | Create custom audience | brijr `create_custom_audience` | `audiences create --account <id> --name --subtype --description --rule <json>` | `--dry-run` |
| 37 | Create lookalike | brijr `create_lookalike_audience` | `audiences lookalike --source <id> --name --country --ratio` | `--dry-run` |
| 38 | Search interests | pipeboard `search_interests`, attainment | `targeting interests <query> --limit` | SQLite cache of resolved IDs |
| 39 | Interest suggestions | pipeboard `get_interest_suggestions` | `targeting suggestions --interest-ids <csv>` | Cached |
| 40 | Validate interests | pipeboard `validate_interests` | `targeting validate --interest-ids <csv>` | `--json` |
| 41 | Search behaviors | pipeboard `search_behaviors` | `targeting behaviors --limit` | Cached |
| 42 | Search demographics | pipeboard `search_demographics` | `targeting demographics --class <class>` | Cached |
| 43 | Search geo locations | pipeboard `search_geo_locations` | `targeting geo <query> --location-types` | Cached |

### Insights & Analytics

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---|---|---|---|
| 44 | Account insights | hashcott `get_adaccount_insights` | `insights account <id> --days N --breakdowns --level` | Magoosh-founder fields + attribution dedup |
| 45 | Compare performance | brijr `compare_performance` | `insights compare <id1> <id2> --days N` | Diff vs target, change in ROAS, spend, freq |
| 46 | Export insights CSV/JSON | brijr `export_insights` | `insights export --level <account\|campaign\|adset\|ad> --format csv\|json --days N` | From SQLite, no re-fetch |
| 47 | Attribution window control | pipeboard `get_insights` param | `--attribution-windows 1d_click,7d_view,...` on all insights cmds | Default matches magoosh-founder |
| 48 | Breakdowns | pipeboard `get_insights` param | `--breakdowns age,gender,country,placement,...` on all insights cmds | CSV export cleanly |
| 49 | Pagination helper | hashcott, gomarble `fetch_pagination_url` | Built into every list command via internal pagination loop | User never handles pagination manually |

### Diagnostics & Health

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---|---|---|---|
| 50 | Health check | brijr `health_check`, pipeboard `get_token_info` | `doctor` | Token validity, scopes, account access, rate limit, attribution dedup sanity check, SQLite schema version |
| 51 | Auth status | magoosh-founder `auth status` | `auth status` | Token expiry, last-extended, scopes |
| 52 | Auth login | magoosh-founder `auth google/meta` | `auth login --chrome` (browser session reuse) or `--device` fallback | OAuth long-lived exchange (60d), store at `~/.config/meta-ads-pp-cli/meta_token.json` |
| 53 | Auth logout | — | `auth logout` | Deletes token file |
| 54 | Check account setup | brijr `check_account_setup` | `doctor account <id>` | Pixel presence, funding source, ad categories |
| 55 | Unified search | pipeboard `search` | `search <query> --type <account\|campaign\|adset\|ad\|creative\|audience>` | FTS5 over local store — works offline |

Total absorbed features: **55**.

Every row matches or beats a competitor feature. Every write command ships with `--dry-run` and typed exit codes (0/2/3/4/5/7). Every list command ships with `--json`, `--select <field-path,...>`, `--limit`, and pagination built in.

## Transcendence (only possible with our approach)

These are the commands nobody else can ship — they require local state, attribution dedup logic, a decision log, or offline-query capability.

| # | Feature | Command | Why Only We Can Do This | Score |
|---|---|---|---|---|
| T1 | **Budget autopilot: ROAS + frequency + utilization decisioning** | `recommend --days 14 --strategy moderate` | Requires joining campaign metadata + insights + decision history + frequency thresholds + per-campaign budget limits. MCPs just proxy the Graph API. | 10/10 |
| T2 | **Batch apply with validation KPIs and follow-up log** | `apply --changes <json> --confirmation CONFIRM` (reads from `recommend` output) | Requires writing structured audit entries with validation_kpis and follow_up_date so the next pass can analyze outcomes. | 9/10 |
| T3 | **Double-count attribution verification** | `verify roas --days 14` | Detects when `purchase` and `omni_purchase` are both present in insights. Requires domain knowledge the MCPs don't encode. Magoosh-founder invented this, we absorb it. | 9/10 |
| T4 | **Frequency capacity signal** | `capacity --days 7` | Per-campaign has-headroom/confidence derived from freq + utilization + delivery status. The capacity signal magoosh-founder ships, exposed as its own command. | 8/10 |
| T5 | **Decision log FTS search + due-for-followup** | `history search <text>`, `history due --platform meta` | Append-only JSONL → SQLite FTS. `due` returns applied decisions whose follow_up_date has passed and that have no analysis yet. | 8/10 |
| T6 | **Decision analysis (post-mortem)** | `decision-review <log-id> --outcome success\|partial\|failure --observation <text> --hypothesis <text>` | Appends a `decision_analysis` entry linking back to `original_log_id`. Future `recommend` runs learn from this. | 7/10 |
| T7 | **Creative fatigue detection** | `fatigue --days 14 --cpm-rise 0.15 --freq-threshold 3.0` | Needs rolling-window comparison of CPM + frequency + CTR across multiple days from local insights table. Not a single API call. | 8/10 |
| T8 | **Budget pacing monitor (ETA-to-cap)** | `pace --day today` | Hourly spend-rate vs daily_budget projection — requires hourly insights stored over time. | 7/10 |
| T9 | **Learning-phase tracker** | `learning --days 14` | Detects campaigns in learning (status_issues + recent change >20% within 14d). Requires history + status join. | 7/10 |
| T10 | **Audience overlap analysis** | `overlap <adset_a> <adset_b>` | Joins targeting specs from two adsets, computes overlap on interests/geos/age/behaviors. Local data only. | 7/10 |
| T11 | **Local SQL REPL / one-shot** | `query "SELECT campaign_name, roas FROM insights WHERE freq > 3.0"` | No MCP can ship a query interface — they proxy requests. SQLite makes arbitrary analysis one-liners cheap. | 8/10 |
| T12 | **Threshold alerts against local data** | `alerts watch --roas-min 1.5 --freq-max 3.5 --utilization-min 0.5` | Watches local store, prints offenders. Needs cron/one-shot mode. | 7/10 |
| T13 | **Cross-account rollup** | `rollup --accounts <id1,id2,...> --days 14` | Aggregate spend/ROAS/conversions across multiple ad accounts. Requires local store across accounts. | 6/10 |
| T14 | **Delta sync via activities cursor** | `sync --since 4h`, `sync --full` | Reads `activities` edge with `event_time > last_sync`, only refetches touched entities. Meta has no `updated_since` filter, so this is non-trivial. | 8/10 |
| T15 | **Campaign name convention parser + dimensions** | `campaigns list --dimension type=TOEFL` | Magoosh-founder has `parse_campaign_name` (type/exam/geo/variant). Expose as a filter + aggregation dimension. | 6/10 |
| T16 | **YAML-defined campaigns (beats attainment)** | `campaigns create --from-yaml campaign.yaml`, `adsets create --from-yaml ...`, `ads create --from-yaml ...` | Absorb attainment-labs's killer feature and extend to adsets/ads/creatives — and add `--dry-run` so a full campaign tree can be validated offline before a single API call. | 7/10 |

Total transcendence features: **16**, all scoring ≥6/10, **13 scoring ≥7/10**.

## Status

All absorbed and transcendence features above are **shipping scope**. No stubs.

Exceptions (explicit stub candidates — will be deferred if time constrains Phase 3):
- **None planned.** The data-layer-backed features all fall out of the sync + SQLite schema.

## Total

**55 absorbed + 16 transcendence = 71 features.** The closest competitor is pipeboard with 29 tools — our CLI ships **145% more surface** with a local brain, attribution dedup, decisioning engine, and decision audit log that no MCP has.
