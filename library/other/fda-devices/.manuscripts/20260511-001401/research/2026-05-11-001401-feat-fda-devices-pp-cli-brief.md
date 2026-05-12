# FDA Devices CLI Brief

## API Identity
- Domain: openFDA medical device regulatory data (https://api.fda.gov/device/*)
- Users: medical device sales reps, R&D engineers, regulatory affairs professionals, investors, journalists
- Data profile: 510(k) clearances (~175k records), PMA approvals, MAUDE adverse events (~11.4M, quarterly partitioned), recalls, classifications (~7k), establishment registrations (~325k)

## Reachability Risk
- None. Official, documented JSON API. No auth required (optional API key raises daily cap from 1k to 120k requests).

## Top Workflows
1. **Competitor scanning** — "who else cleared a device in product code XYZ recently"
2. **Predicate-chain audit** — "trace this K-number's ancestor tree back to root"
3. **Recall + safety surveillance** — "what's the adverse-event volume in this device category"
4. **Applicant intelligence** — "what has Medtronic cleared in the past 5 years"
5. **Market-entry timing** — "new entrants in product code XYZ, ranked by time-to-clearance"

## Table Stakes (must absorb)
- Full Lucene-style search per endpoint (`search=field:term`, `.exact` suffix, AND/OR, range, _exists_, _missing_)
- Get-by-id, list, count aggregation, sort, limit, skip + deep paging via `search_after`
- API key support via env var; rate-limit aware backoff
- JSON / CSV / NDJSON / table output
- Bulk download with resume + checksum verify
- MCP server mode exposing every command as a tool
- Cross-reference between 510(k) ↔ recall ↔ MAUDE via `product_code` / `k_number` / `device_name`

## Data Layer (SQLite mirror)
- **Primary entities:** `clearances_510k`, `pma`, `recalls`, `maude_events`, `classifications`, `establishments`
- **Predicate graph table:** `predicates(child_k, parent_k, source, confidence)` — sparse, populated from PDF letter parsing (best-effort) when available
- **FTS5 over:** `device_name`, `applicant`, `product_code` text, recall reason, MAUDE event narrative
- **Sync cursor:** `sync_state(endpoint, last_export_date, last_synced_at, partition)` — openFDA exposes weekly `export_date` per bulk file

## Codebase Intelligence
- Source: openFDA official ecosystem + 14 community wrappers (rOpenHealth/openfda, coderxio/openfda, tsbischof/fda predicate analyzer, PyMAUDE, MAUDEMetrics, several MCP servers)
- Auth: optional `api_key` query param (env var `OPENFDA_API_KEY`)
- Rate limiting: 240/min always; 1k/day anon, 120k/day with key. Backoff on HTTP 429.
- Architecture: REST + JSONL bulk files at `https://download.open.fda.gov/device/<endpoint>/<file>.json.zip`. Predicate references live in clearance-letter PDFs, NOT structured response fields — predicate-chain command requires PDF scraping (documented limitation, ship best-effort).
- Pagination quirk: `skip` capped at 25k; deep paging requires `search_after` cursor.

## Product Thesis
- **Name:** `fda-devices-pp-cli`
- **Why it should exist:** The FDA markets these endpoints as "look up cleared devices." The deeper read is competitive intelligence, predicate auditing, market-entry radar, and safety pattern detection — sitting on free public data that's only been queryable one record at a time. The CLI surfaces those use cases as named commands.

## Neutrality stance
The CLI surfaces data. The predicate-chain walker shows recalled ancestors when they exist; it does not characterize them as proof of anything. `safety-pattern` shows adverse events per year on market; it does not claim devices are unsafe. Industry users and reform advocates should both be able to install it without feeling targeted. Documented explicitly in README "Methodology" section.

## Build Priorities
1. Data layer + bulk sync workflow (Priority 0)
2. Endpoint primitives: `510k`, `recall`, `maude`, `pma`, `classification`, `establishment` with `{get,search,list}` (Priority 1)
3. Compound commands: `competitors`, `predicate-chain`, `applicant-history`, `recall-link`, `market-entry`, `safety-pattern`, `watch`, `story` (Priority 2)
4. SQL passthrough, FTS search, MCP server (Priority 1)
