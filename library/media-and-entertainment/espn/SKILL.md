---
name: pp-espn
description: "Use this skill whenever the user asks about live sports scores, standings, team stats, boxscores, NFL / NBA / MLB / NHL / NCAA / MLS / EPL / WNBA games, upcoming schedules, injuries, odds, or player leaderboards. ESPN sports CLI with live scores across 10 leagues, offline search, and head-to-head comparisons. No API key required. Triggers on natural phrasings like 'what's the score of the Lakers game', 'Patriots schedule this week', 'who's leading the NBA in scoring', 'NFL standings', 'compare Mahomes and Allen stats', 'any injuries for the Yankees'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["espn-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn/cmd/espn-pp-cli@latest","bins":["espn-pp-cli"],"label":"Install via go install"}]}}'
---

# ESPN — Printing Press CLI

ESPN from your terminal. Live scores, standings, stats, boxscores, play-by-play, injuries, odds, and search across 10 major leagues (NFL, NBA, MLB, NHL, NCAAF, NCAAM, NCAAW, MLS, EPL, WNBA). No API key needed — the spec was sniffed from live ESPN endpoints that back their own apps and website.

## When to Use This CLI

Reach for this when a user wants a quick sports lookup — current score, standings, upcoming schedule, team-vs-team comparison, or an injury check. Also good for aggregating stats across leagues (trending athletes, power rankings) without clicking through ESPN's site. Works offline once synced.

Don't reach for this if the user has a paid feed like Stats Perform or Sportradar that provides cleaner data, or if they need real-time websocket updates (ESPN's endpoints are polling-only).

## Unique Capabilities

Commands that only work because of local sync + cross-league tooling.

### Cross-league discovery

- **`trending`** — Most-followed athletes and teams across all leagues, ranked by current interest.

  _One scan of "what everyone is watching" without picking a sport first._

- **`watch`** — Currently-live or upcoming-today games across every league you track.

- **`dashboard`** — Your favorite teams' status at a glance (add teams via `watch` or config).

### Strength and scheduling intelligence

- **`sos <sport> <league>`** — Strength-of-schedule analysis across the league.

- **`h2h <team1> <team2>`** — Head-to-head series history with matchup stats.

- **`compare <athlete1> <athlete2>`** — Side-by-side athlete stat comparison.

- **`rankings <sport> <league>`** — Power rankings and coaches' polls (NCAAF/NCAAM especially).

### Deep game detail

- **`plays <game_id>`** — Play-by-play for a specific game.

- **`boxscore <game_id>`** — Full boxscore with per-player stats.

- **`preview <game_id>`** / **`recap <sport> <league>`** — Pre-game previews and post-game recaps.

### Depth and context

- **`leaders <sport> <league>`** — Statistical leaderboards (points, yards, WAR, etc.).

- **`injuries <sport> <league>`** — Current injury reports across a league.

- **`odds <sport> <league>`** — Betting odds feed.

- **`transactions <sport> <league>`** — Recent trades, signings, waivers.

## Command Reference

Live action:

- `espn-pp-cli scores <sport> <league>` — Current scores
- `espn-pp-cli watch` — Live + upcoming-today across leagues
- `espn-pp-cli schedule <sport> <league>` — Upcoming games
- `espn-pp-cli standings <sport> <league>` — League standings
- `espn-pp-cli calendar <sport> <league>` — Season calendar

Team/athlete:

- `espn-pp-cli team get <sport> <league> <team_id>` — Team detail
- `espn-pp-cli team list <sport> <league>` — All teams in a league
- `espn-pp-cli athlete <sport> <league>` — Athletes

Stats:

- `espn-pp-cli leaders <sport> <league>` — Stat leaders
- `espn-pp-cli rankings <sport> <league>` — Polls / power rankings
- `espn-pp-cli sos <sport> <league>` — Strength of schedule

Game:

- `espn-pp-cli boxscore <game_id>` — Full boxscore
- `espn-pp-cli plays <game_id>` — Play-by-play
- `espn-pp-cli preview <game_id>` / `recap` — Pre/post

Info:

- `espn-pp-cli news <sport> <league>` — Latest news
- `espn-pp-cli injuries <sport> <league>` — Injury reports
- `espn-pp-cli odds <sport> <league>` — Betting odds
- `espn-pp-cli transactions <sport> <league>` — Trades/signings

Discovery & local:

- `espn-pp-cli search "<query>"` — Full-text search across synced data
- `espn-pp-cli sync` — Pull full dataset for offline analysis
- `espn-pp-cli trending` — Cross-league interest scan
- `espn-pp-cli doctor` — Verify connectivity

Sport values: `football`, `basketball`, `baseball`, `hockey`, `soccer`.
League values: `nfl`, `nba`, `mlb`, `nhl`, `ncaaf`, `ncaam`, `ncaaw`, `mls`, `eng.1` (EPL), `wnba`.

## Recipes

### Morning sports scan

```bash
espn-pp-cli trending --agent                    # who's everyone watching
espn-pp-cli watch --agent                       # live now + today
espn-pp-cli scores football nfl --agent         # specific-league drilldown
```

One trending call, one watch call, one scores call — covers "what's happening in sports" with enough granularity for a morning briefing.

### Team-vs-team deep research

```bash
espn-pp-cli h2h chiefs eagles --agent           # head-to-head history
espn-pp-cli injuries football nfl --agent | jq 'select(.team=="Chiefs" or .team=="Eagles")'
espn-pp-cli odds football nfl --agent           # current lines
```

Combine historical matchup data, current injuries, and betting lines to build a complete pre-game view.

### Offline search after sync

```bash
espn-pp-cli sync --sport football --league nfl
espn-pp-cli search "Mahomes"                    # finds in local store
```

Useful for repeated lookups in poor-connectivity environments or when batch-analyzing historical data.

## Auth Setup

**None required.** ESPN's public endpoints don't require an API key. The `auth` command exists for consistency but is a no-op.

Optional config:
- `ESPN_CONFIG` — override config file path
- `ESPN_BASE_URL` — override base URL (for proxies or mirrors)
- `NO_COLOR` — standard no-color env var

## Agent Mode

Add `--agent` to any command. Expands to `--json --compact --no-input --no-color --yes`. Use `--select` for field cherry-picking, `--dry-run` to preview requests, `--no-cache` to bypass GET cache.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error |
| 3 | Not found (team, game, athlete) |
| 5 | API error |
| 7 | Rate limited |

## Installation

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn/cmd/espn-pp-cli@latest
espn-pp-cli doctor
```

### MCP Server

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn/cmd/espn-pp-mcp@latest
claude mcp add espn-pp-mcp -- espn-pp-mcp
```

## Argument Parsing

Given `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → run `espn-pp-cli --help`
2. **`install`** → CLI; **`install mcp`** → MCP
3. **Anything else** → resolve `<sport> <league>` from user intent (e.g., "Lakers" → `basketball nba`), check `which espn-pp-cli` (offer install if missing), run with `--agent`.
