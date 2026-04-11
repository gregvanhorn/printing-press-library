---
name: pp-espn
description: "Printing Press CLI for ESPN. Live scores, standings, news, and game history across 17 sports from ESPN Capabilities include: boxscore, news, rankings, recap, rivals, scoreboard, scores, search, standings, streak, summary, teams, today, watch. Trigger phrases: 'install espn', 'use espn', 'run espn', 'ESPN commands', 'setup espn'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
---

# ESPN — Printing Press CLI

Live scores, standings, news, and game history across 17 sports from ESPN

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `espn-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn/cmd/espn-pp-cli@latest
   ```
3. Verify: `espn-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn/cmd/espn-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add espn-pp-mcp -- espn-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which espn-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Discover commands: `espn-pp-cli --help`
   Key commands:
   - `boxscore` — Box score with player stats for a game
   - `news` — Get latest news articles for a sport and league
   - `rankings` — Get current AP, Coaches, and CFP poll rankings for college sports
   - `recap` — Game recap with box score and leaders
   - `rivals` — Head-to-head record between two teams from synced data
   - `scoreboard` — Get scoreboard for a sport and league with optional date filtering
   - `scores` — Live scores and results for a sport and league
   - `search` — Full-text search across synced events and news
   - `standings` — Conference/division standings for a sport and league
   - `streak` — Current win/loss streak for a team from synced data
   - `summary` — Get detailed game summary including box score, leaders, scoring plays, odds, and win probability
   - `teams` — Get past and upcoming schedule for a specific team
   - `today` — Today's scores across all major sports
   - `watch` — Live score updates for a game (polls every 30s)
3. Match the user query to the best command. Drill into subcommand help if needed: `espn-pp-cli <command> --help`
4. Execute with the `--agent` flag:
   ```bash
   espn-pp-cli <command> [subcommand] [args] --agent
   ```
5. The `--agent` flag sets `--json --compact --no-input --no-color --yes` for structured, token-efficient output.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
