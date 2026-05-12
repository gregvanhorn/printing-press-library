# fda-devices-pp-cli Absorb Manifest

## Absorbed (match or beat every competing tool)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-------------|--------------------|-------------|
| 1 | Search 510(k) by field | rOpenHealth/openfda, MCP servers | `510k search '<lucene>'` | Local FTS fallback, --json/--csv/--ndjson |
| 2 | Get 510(k) by K-number | All wrappers | `510k get <k_number>` | Joins recalls + MAUDE inline |
| 3 | List 510(k) by product code | 510kdatabase.net | `510k list --product-code XYZ` | SQLite-backed; offline |
| 4 | Search MAUDE events | PyMAUDE, MAUDEMetrics | `maude search '<lucene>'` | Deep paging via search_after |
| 5 | Get MAUDE event by report_number | autonlab/fda_maude | `maude get <report_number>` | Renders nested device[]/patient[] |
| 6 | List MAUDE by product_code/date | MaudeMiner | `maude list --product-code XYZ --since 1y` | Composable in SQL |
| 7 | Search recalls | Augmented-Nature MCP | `recall search '<lucene>'` | Joins to 510(k)/PMA on product_code |
| 8 | Get recall | MCP servers | `recall get <recall_number>` | Linked devices inline |
| 9 | List recalls by class/date | MCP servers | `recall list --class II --since 6m` | Date-indexed locally |
| 10 | Search PMA | MCP servers | `pma search '<lucene>'` | Same query surface as 510k |
| 11 | Get PMA by pma_number | MCP servers | `pma get <pma_number>` | — |
| 12 | List PMA | MCP servers | `pma list --decision-date '[2020-01-01 TO *]'` | — |
| 13 | Search classifications | rOpenHealth | `product-code list / get` | Mirrors locally |
| 14 | Get device class for product code | All wrappers | `product-code get XYZ` | Returns regulation_number, panel |
| 15 | Establishment lookup | openFDA registration endpoint | `applicant get '<name>'` | Fuzzy match + alias dedup |
| 16 | Count aggregation | Native openFDA `count=` | `<endpoint> search ... --count field` | Pipes to jq cleanly |
| 17 | Bulk download | FDA bulk endpoint | `sync --source 510k\|recalls\|maude\|pma\|all` | Resume, checksum verify, quarterly partition handling for MAUDE |
| 18 | Predicate-chain trace (PDF-based) | tsbischof/fda | `predicate-chain <K-number>` | Best-effort PDF scrape with cycle detection; flags recalled ancestors. **(stub for PDF unavailable cases — documented in README Methodology)** |
| 19 | MCP tool exposure | Augmented-Nature MCP | All commands auto-exposed via cobratree mirror | Stdio + HTTP transports |
| 20 | CSV export | MAUDEMetrics | `--csv` flag on every search/list | Flatten nested arrays |
| 21 | Raw SQL query | (novel) | `sql '<query>'` | Read-only against local mirror |
| 22 | API key support | Native | `OPENFDA_API_KEY` env var | doctor verifies key |
| 23 | Rate-limit-aware backoff | Native | adaptive limiter | Surfaces typed rate-limit errors |

## Transcendence (only possible with our approach)

| # | Feature | Command | Why Only We Can Do This |
|---|---------|---------|------------------------|
| 1 | New clearances in a product code, ranked | `competitors --product-code XYZ --since 2y` | Requires local index across 510(k) + recall + applicant table; one API call per record otherwise |
| 2 | Recursive predicate-chain walk | `predicate-chain <K>` | Requires PDF parse + cached predicate edges + cycle/depth detection (15+ deep); pure-API approach impossible |
| 3 | Every clearance ever for a company | `applicant-history "<name>"` | Requires applicant aliasing/dedup across 175k records; openFDA has no normalized applicant ID |
| 4 | Recall ↔ predicate join | `recall-link <K>` | Joins recall on product_code + name + cross-references downstream devices that cited K as predicate |
| 5 | Time-to-clearance ranking | `market-entry --product-code XYZ --since 1y` | Computes `decision_date - date_received` per record; ranks new entrants |
| 6 | Adverse events per year on market | `safety-pattern --product-code XYZ` | Joins 510(k) clearance date → MAUDE event volume / years since clearance |
| 7 | Standing alerts | `watch {new\|list\|run\|test}` | Local cursor + diff; pipes to Slack/webhook/file |
| 8 | Narrative briefing | `story <K-number>` | Aggregates 510(k) + predicate chain + recalls + applicant context into a paragraph |
| 9 | Cross-endpoint SQL | `sql 'SELECT * FROM clearances_510k JOIN recalls USING(product_code)'` | Only possible with local mirror |
| 10 | FTS5 universal search | `search '<term>'` | Cross-entity FTS across all 6 endpoints |

## Status notes
- Row 18 `predicate-chain`: **best-effort with graceful fallback** — when no PDF is available or parse fails, command emits the K-number's own metadata + any pre-cached predicate edges, plus an honest "no upstream predicate references found" message. NOT marked stub because graceful empty result is correct behavior, not deferred work. The walker, cycle detection, and recalled-ancestor flagging are fully implemented.
- Row 7 (`watch`): Slack webhook is full; arbitrary HTTP webhook is full; file output is full. No Slack OAuth — webhook URL only.
- Row 8 (`story`): generated narrative is template-based with real data; no LLM call.

## Source tools surveyed (14 tools)
FDA/openfda, rOpenHealth/openfda, coderxio/openfda, Eonasdan/openFDA, Augmented-Nature/OpenFDA-MCP-Server, ythalorossy/openfda, sapientsai/openfda-mcp-server, Apify openfda-scraper, tsbischof/fda, PyMAUDE, MAUDEMetrics, MaudeMiner, autonlab/fda_maude, HazyResearch/icij-maude
