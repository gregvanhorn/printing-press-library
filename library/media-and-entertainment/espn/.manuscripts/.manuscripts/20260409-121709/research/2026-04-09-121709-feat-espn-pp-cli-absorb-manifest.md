# ESPN CLI Absorb Manifest

## Source Tools Cataloged

| Tool | Type | Stars | Features Absorbed |
|------|------|-------|-------------------|
| Left-Coast-Tech/espn-mcp | MCP Server | 1 | 6 |
| KBThree13/mcp_espn_ff | MCP Server (Fantasy) | 30 | 6 |
| cwendt94/espn-api (PyPI) | Python Library | 890 | 8 |
| pseudo-r/sportly (PyPI) | Python SDK | 2 | 5 |
| pseudo-r/Public-ESPN-API | API Docs + Django | 443 | 4 |
| uberfastman/fantasy-football-metrics | CLI Report Gen | 223 | 3 |
| espn-fantasy-football-api (npm) | JS Library | — | 4 |
| n8n-nodes-espn-api (npm) | n8n Integration | — | 3 |

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Get scoreboard/scores | espn-mcp get_scoreboard | `scores <sport> <league>` | Formatted table, all sports not just 3, --json --select |
| 2 | Get standings | espn-mcp get_standings | `standings <sport> <league>` | Full stat display W/L/PCT/GB/DIFF/STRK, --season flag |
| 3 | Get team info | espn-mcp get_team | `teams get <sport> <league> <id>` | Full team detail with record, links, logos |
| 4 | Get team schedule | espn-mcp get_schedule | `teams schedule <sport> <league> <id>` | Past + future, --season filter |
| 5 | Get game detail | espn-mcp get_game | `recap <sport> <league> --event <id>` | Box score, leaders, scoring plays formatted |
| 6 | Get playoff picture | espn-mcp get_playoffs | `rankings <sport> <league>` | College polls + playoff picture |
| 7 | List teams | Public-ESPN-API | `teams list <sport> <league>` | Formatted table with abbreviations, colors |
| 8 | Get news | Public-ESPN-API | `news <sport> <league>` | Headlines, bylines, publish dates |
| 9 | Fantasy league info | mcp_espn_ff | Out of scope (requires auth) | — |
| 10 | Fantasy team rosters | mcp_espn_ff/espn-api | Out of scope | — |
| 11 | Fantasy player stats | mcp_espn_ff/espn-api | Out of scope | — |
| 12 | Fantasy standings | mcp_espn_ff/espn-api | Out of scope | — |
| 13 | Fantasy matchups | mcp_espn_ff/espn-api | Out of scope | — |
| 14 | Multi-source data (ESPN+NHL+FotMob) | sportly | Single-source (ESPN only) but deeper | SQLite persistence, FTS5 search |
| 15 | Data export to CSV/XLSX | ESPN_Extractor | `export --format jsonl` | JSONL for agent pipelines |
| 16 | Weekly reports | fantasy-football-metrics | Out of scope (fantasy-specific) | — |
| 17 | Health/diagnostics | Standard | `doctor` | API reachability, config validation |
| 18 | Sync to local database | ESPN-Fantasy-Data-Archive | `sync` | Domain-specific tables, incremental date-range sync |
| 19 | n8n workflow integration | n8n-nodes-espn-api | --json output for any pipeline | Agent-native, pipe to anything |
| 20 | Game summary with odds | Public-ESPN-API | `summary get <sport> <league> --event <id>` | Odds, win probability, injuries |

**Note:** Features 9-13 and 16 are fantasy-league-specific requiring ESPN auth cookies (SWID/S2). Deliberately out of scope for this CLI which targets the public unauthenticated API.

## Transcendence (only possible with our local data layer)

| # | Feature | Command | Score | Why Only We Can Do This |
|---|---------|---------|-------|------------------------|
| 1 | Cross-sport daily dashboard | `today` | 9/10 | Parallel fetch of 4+ sport scoreboards, grouped display. No tool aggregates across sports. |
| 2 | Full-text game search | `search "Lakers vs Celtics"` | 9/10 | FTS5 across event names, team names, venues. Requires local SQLite with domain columns. |
| 3 | Raw SQL on sports data | `sql "SELECT ..."` | 8/10 | Arbitrary read-only SQL against events/teams/news/standings. No other tool exposes this. |
| 4 | Live game monitor (tail) | `watch <sport> <league> --event <id>` | 7/10 | Polls scoreboard every 30s, shows score updates. CLI-native live monitoring. |
| 5 | Team win/loss streak | `streak <sport> <league> --team KC` | 7/10 | Queries local events table for sequential W/L results. Requires synced game history. |
| 6 | Head-to-head record | `rivals <sport> <league> --teams KC,BUF` | 7/10 | Cross-references local events for all matchups between two teams. |
| 7 | "When do my teams play?" | `calendar` | 6/10 | Aggregates next-game dates across configured favorite teams. |
| 8 | Date-range sync with incremental cursor | `sync --since 7d` | 8/10 | Uses ESPN's ?dates= param for incremental sync. No other tool does this. |

## Build Plan

### Priority 0 — Data Layer
- events table (26 domain columns + FTS5)
- teams_domain table (12 columns + FTS5)
- news_domain table (12 columns + FTS5)
- standings table (18 columns)
- sync_state_v2 for sport/league scoped cursors
- UpsertEvent/UpsertTeamDomain/UpsertNewsDomain methods
- SearchEvents/SearchNews FTS5 methods
- QueryRaw for arbitrary SQL

### Priority 1 — Absorbed Features (ALL of them)
- scoreboard get (with formatted table)
- teams list / teams get / teams schedule
- news list
- summary get (recap formatting)
- rankings get
- standings (web API endpoint)
- sync with ESPN-specific date-range pagination
- export/import
- doctor

### Priority 2 — Transcendence
- `today` — cross-sport daily dashboard
- `scores` — formatted live scoreboard
- `search` — FTS5 full-text search
- `sql` — raw SQL queries
- `watch` — live game polling
- `streak` — win/loss streak analysis
- `rivals` — head-to-head records
- `calendar` — next games for favorites

### Priority 3 — Polish
- README cookbook with ESPN-specific examples
- Realistic examples in --help (not "abc123")
- Flag description enrichment
