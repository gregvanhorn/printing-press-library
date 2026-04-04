# Steam Web API CLI Brief

## API Identity
- **Domain:** Gaming platform — player profiles, game libraries, achievements, friends, stats, store, community guides
- **Users:** Game developers, data analysts, community tool builders, bot developers, gaming stats enthusiasts
- **Data profile:** 120M+ MAU, 70K+ games, player stats, achievements, market data, community guides
- **Auth:** API key required (free at steamcommunity.com/dev/apikey), passed as `?key=` query param
- **Protocol:** REST, base URL `https://api.steampowered.com/<interface>/<method>/v<version>/`
- **Store API:** `https://store.steampowered.com/api/` (no auth needed for some endpoints)

## Reachability Risk
- **Low** — API confirmed reachable (HTTP 200 with key). Known issues: 403 for both rate-limiting AND private profiles (ambiguous error codes). 429 for exceeded rate limits with possible IP blocks. Rate limit thresholds are undocumented but generally lenient for read operations.

## Spec Sources
- **Zuplo/Steam-OpenAPI** — 118 public endpoints (JSON, curated, includes public + undocumented)
- **ceva24/openapi-steamworks-web-api** — Auto-generated from Valve's GetSupportedAPIList, Swagger UI
- **xPaw/SteamWebAPIDocumentation** — 170+ interfaces reference (HTML, not OpenAPI)
- **Live GetSupportedAPIList** — Confirmed 32+ interfaces with key, public subset: ISteamUser, IPlayerService, ISteamUserStats, ISteamApps, ISteamNews, ISteamWebAPIUtil

## Top Workflows
1. **Player lookup** — "Show me this Steam user's profile, games, and recent activity"
2. **Game library browsing** — "What games does this user own? Sort by playtime"
3. **Achievement tracking** — "Show achievement progress for a specific game"
4. **Friend network** — "Who are this user's friends? What are they playing?"
5. **Game stats** — "Global achievement percentages, current player count trends"

## Table Stakes
- Resolve vanity URLs to Steam IDs
- Get player summaries (avatar, status, last logoff, profile URL)
- List owned games with playtime
- List recently played games
- Get friend lists with online status
- Get player achievements per game
- Get global achievement percentages
- Get game news
- Get full app list (70K+ games)
- Check player bans (VAC, game bans)
- Get Steam level and badges
- Get app details from Store API
- Get user stats for game
- Get game schema (stats + achievements definitions)
- Get current player count
- Get user groups
- Search Steam Community guides

## Competitor Analysis
| Tool | Type | Status | Endpoints | Key Limitation |
|------|------|--------|-----------|----------------|
| algorhythmic/steam-mcp | MCP | Active | 10 tools | Online-only, no persistence |
| matheusslg/steam-mcp | MCP | Active | 20 tools, 6 modules | Online-only, no offline |
| Fllugel/steam-mcp-server | MCP | Active | 3 tools | Achievements + guides only |
| dsp/mcp-server-steam | MCP | Active | Gaming context | Library info only |
| fenxer/steam-review-mcp | MCP | Active | Reviews | Reviews only |
| steamapi npm (23K/wk) | SDK | Active | 25+ methods | Library, no CLI |
| steam PyPI (ValvePython) | SDK | Active | Comprehensive | Library, no CLI |
| UberEnt/SteamCLI | CLI | Abandoned | Generic caller | No commands, C#, dead |
| steamgames-exporter | CLI | Narrow | 1 (export) | Export only, Ruby |
| steam-hours-tool | CLI | Narrow | 1 (playtime) | Playtime only |

No comprehensive, actively maintained CLI exists.

## Data Layer
- **Primary entities:** Player, Game (App), Achievement, Friend, NewsItem, Badge, PlayerStats
- **Sync cursor:** No native cursor — paginate by SteamID lists or app lists, store snapshots
- **FTS/search:** Full-text search across game names, achievement names, news titles
- **Store cache:** App details, pricing (from Store API)

## Product Thesis
- **Name:** steam-pp-cli
- **Why it should exist:** Steam has 120M+ MAU but zero actively maintained CLI for querying player data, game libraries, achievements, or friend networks from the terminal. The best existing tool (matheusslg MCP, 20 tools) is online-only with no persistence. A CLI with offline sync, cross-entity SQLite queries (player+games+achievements+friends), and agent-native output fills a real gap for game developers, data analysts, and community tool builders.

## Build Priorities
1. **Player lookup** — `player <steamid-or-vanity>` with profile, games, level
2. **Games** — `games <steamid>` with playtime, sort, filter
3. **Achievements** — `achievements <steamid> <appid>` with progress
4. **Friends** — `friends <steamid>` with online status
5. **App search** — `search "game name"` across synced app list
6. **Stats** — Global achievement percentages, user stats, player bans
7. **Sync** — Persist app list, player data, achievements to SQLite
8. **News** — `news <appid>` for game news feeds
9. **Store** — `app <appid>` for store details, pricing
10. **Badges** — `badges <steamid>` for badge collection

## Known Public API Interfaces
ISteamUser, IPlayerService, ISteamUserStats, ISteamApps, ISteamNews, ISteamWebAPIUtil,
IStoreService, ISteamRemoteStorage, IEconService, IAuthenticationService + Store API
