# ESPN CLI Build Log

## What Was Built

### Priority 0: Data Layer
- events table (26 domain columns + 6 indexes + FTS5)
- teams_domain table (12 columns + FTS5)
- news_domain table (12 columns + FTS5)
- standings table (18 columns)
- sync_state_v2 for sport/league scoped sync
- 11 domain methods: UpsertEvent, UpsertTeamDomain, UpsertNewsDomain, SearchEvents, SearchNews, ListEvents, QueryRaw, EventCount, TeamCount, NewsCount, GetSyncStateV2/SaveSyncStateV2
- sync.go wired: 12 ESPN resources, domain-specific upsert routing

### Priority 1: Absorbed Features (6 new commands)
- scores.go — live scoreboard with formatted table
- today.go — cross-sport daily dashboard (parallel NFL/NBA/MLB/NHL)
- standings_cmd.go — league standings from web API
- recap.go — game box score with leaders
- search_cmd.go — FTS5 search across events + news
- sql_cmd.go — raw SQL against local database

### Priority 2: Transcendence (3 new commands)
- watch.go — live game polling with interval
- streak.go — win/loss streak analysis from local data
- rivals.go — head-to-head records from local data

## Total: 9 new hand-built commands + 24 generated = 33 commands

## What Was Deferred
- calendar command (next games for favorites) — lower priority, can be added later
- Fantasy league features — requires ESPN auth (SWID/S2 cookies), deliberately out of scope

## Skipped Body Fields
None — ESPN API is read-only (all GET), no POST/PUT bodies.

## Generator Limitations Found
None — ESPN's simple REST structure mapped cleanly to the generator's internal YAML format.
