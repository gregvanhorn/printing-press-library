# ESPN CLI Brief

## API Identity
- **Domain:** Sports data (scores, standings, stats, news, schedules, play-by-play)
- **Users:** Sports fans who live in the terminal, developers building sports integrations, fantasy managers, data analysts
- **Data profile:** Append-only events, HIGH volume (thousands/season across sports), no auth required, REST polling for live scores, HIGH search need
- **Base URL:** `https://site.api.espn.com/apis/site/v2/sports/{sport}/{league}/{resource}`
- **Auth:** None. Public unauthenticated JSON API.

## Reachability Risk
- **None.** HTTP 200 confirmed. ESPN's public endpoints have been stable for years. Community docs (443 stars) prove sustained access. No 403/blocking issues reported.

## Top Workflows
1. **Live scores** — "What's happening right now?" across NFL/NBA/MLB/NHL (daily)
2. **Today's games** — Cross-sport schedule for today (daily)
3. **Standings** — Current league standings with W/L/PCT/GB (weekly)
4. **Game recap** — Box score, leaders, scoring plays for a specific game (per-game)
5. **Historical search** — "Lakers vs Celtics 2025" across synced local data (weekly)

## Table Stakes (from competitors)
- Live scoreboard (Left-Coast-Tech espn-mcp: 6 tools for NFL/NBA/NHL)
- Team info and standings (espn-mcp: get_standings, get_team)
- Fantasy league data (KBThree13/mcp_espn_ff: 6 tools, 30 stars)
- News headlines (Public-ESPN-API docs: news endpoint)
- Multi-sport coverage (Public-ESPN-API: 17 sports, 139 leagues, 449 endpoints)
- Schedule by team (espn-mcp: get_schedule)
- Game detail/summary (espn-mcp: get_game)

## Data Layer
- **Primary entities:** Events (gravity 15/12), News (10), Teams (9), Standings (8)
- **Sync cursor:** `?dates=YYYYMMDD-YYYYMMDD` on scoreboard endpoint (validated: returns 64 events for NFL Sept 2025)
- **FTS/search:** events_fts (name, short_name, home/away team names, venue), news_fts (headline, description, byline), teams_fts (display_name, location, abbreviation)
- **Schema:** Domain-specific columns per entity (26 cols for events, not JSON blobs)

## Product Thesis
- **Name:** espn-pp-cli (binary), ESPN CLI (product)
- **Why it should exist:** No CLI exists for ESPN's full 17-sport public API. The Python ecosystem has libraries (espn-api: 890 stars) and MCP servers (3 repos covering subsets), but zero terminal-native tools. Our CLI covers ALL sports with 20+ commands, adds SQLite persistence with FTS5 search, and works from any terminal with no auth.

## Build Priorities
1. Data layer: domain SQLite tables for events/teams/news/standings + FTS5 + sync
2. Absorb: every feature from every ESPN MCP + npm package + Python library
3. Transcend: cross-sport queries, historical search, streak analysis, head-to-head — only possible with local SQLite
