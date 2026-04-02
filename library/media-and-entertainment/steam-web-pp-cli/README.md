# steam-web-pp-cli

Query Steam player profiles, game libraries, achievements, and friends from the terminal.

Get your API key from [here](https://steamcommunity.com/dev/apikey)

## Install

### Homebrew

```
brew install trevin-chow/tap/steam-web-pp-cli
```

### Go

```
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/cmd/steam-web-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Quick Start

### 1. Set your API key

```bash
export STEAM_API_KEY=your_key_here
```

Or add to `~/.config/steam-web-pp-cli/config.toml`:

```toml
api_key = "your_key_here"
```

### 2. Verify setup

```bash
steam-web-pp-cli doctor
```

### 3. Try a command

```bash
# Look up a player by vanity URL or SteamID64
steam-web-pp-cli player gabelogannewell

# See their game library sorted by playtime
steam-web-pp-cli games gabelogannewell --sort playtime --limit 10

# Check current player count for CS2
steam-web-pp-cli players-count 730
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** -- never prompts, every input is a flag or positional arg
- **Pipeable** -- `--json` output to stdout, errors to stderr
- **Filterable** -- `--select id,name` returns only fields you need
- **Previewable** -- `--dry-run` shows the request without sending
- **Retryable** -- creates return "already exists" on retry, deletes return "already deleted"
- **Confirmable** -- `--yes` for explicit confirmation of destructive actions
- **Cacheable** -- GET responses cached for 5 minutes, bypass with `--no-cache`
- **Agent-safe by default** -- no colors or formatting unless `--human-friendly` is set
- **Smart defaults** -- `--agent` sets `--json --compact --no-input --no-color --yes` in one flag

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

### Vanity URL resolution

All player commands accept either a 17-digit SteamID64 or a vanity URL name. The CLI resolves vanity names automatically via the Steam API.

```bash
# These are equivalent:
steam-web-pp-cli player 76561198006409530
steam-web-pp-cli player gabelogannewell
```

## Health Check

```bash
steam-web-pp-cli doctor
```

Checks config file, API key, API reachability, and credential validity.

```bash
steam-web-pp-cli doctor --json   # Machine-readable output
```

## Troubleshooting

**Authentication errors (exit code 4)**
- Verify your API key: `echo $STEAM_API_KEY`
- Run `steam-web-pp-cli doctor` to check credentials
- Get a key from https://steamcommunity.com/dev/apikey

**Not found errors (exit code 3)**
- Check the SteamID64 or vanity name is correct
- Use `steam-web-pp-cli resolve <name>` to verify resolution

**Private profile errors (403)**
- Some endpoints require the target profile to be public
- `completionist` and `achievements` skip private profiles gracefully
- `profile` detects private profiles and returns available data

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- Use `--rate-limit 1` to throttle requests to 1/sec
- If persistent, wait a few minutes and try again

**Empty results**
- Some games have no achievements (use `schema <appid>` to check)
- Free-to-play games may not appear without `include_played_free_games`

## Cookbook

### Player lookup and profiles

```bash
# Quick player summary
steam-web-pp-cli player gabelogannewell

# Full profile with level, badges, game count, recent activity
steam-web-pp-cli profile gabelogannewell

# Resolve vanity URL to SteamID64
steam-web-pp-cli resolve gabelogannewell

# Get Steam level
steam-web-pp-cli level gabelogannewell
```

### Game library analysis

```bash
# All games sorted by playtime, top 20
steam-web-pp-cli games gabelogannewell --sort playtime --limit 20

# Playtime statistics and distribution
steam-web-pp-cli playtime gabelogannewell

# Unplayed games (the backlog of shame)
steam-web-pp-cli backlog gabelogannewell

# Games with < 2 hours playtime
steam-web-pp-cli backlog gabelogannewell --min-playtime 2
```

### Achievements and stats

```bash
# Player achievements for TF2
steam-web-pp-cli achievements gabelogannewell 440

# Rarest achievements a player has earned
steam-web-pp-cli rare gabelogannewell 440

# Achievement completion rates across library (top 20 most-played)
steam-web-pp-cli completionist gabelogannewell --limit 20

# Only show games where >80% complete
steam-web-pp-cli completionist gabelogannewell --min-pct 80

# Global achievement percentages for CS2
steam-web-pp-cli global-achievements 730 --rare --limit 10

# Full game schema (available achievements and stats)
steam-web-pp-cli schema 440

# Player stats for a specific game
steam-web-pp-cli stats gabelogannewell 730
```

### Social and multiplayer

```bash
# Friends list with profile summaries
steam-web-pp-cli friends gabelogannewell

# Compare game libraries between two players
steam-web-pp-cli compare player1 player2

# Find games all players share (pick a game to play together)
steam-web-pp-cli overlap player1 player2 player3

# Check bans for one or more players
steam-web-pp-cli bans 76561198006409530

# Player's Steam groups
steam-web-pp-cli groups gabelogannewell
```

### Game info

```bash
# Current player count for a game
steam-web-pp-cli players-count 730

# Also works with the "players" alias
steam-web-pp-cli players 440

# Recent news for a game
steam-web-pp-cli news 730 --count 5

# Recently played games
steam-web-pp-cli recent gabelogannewell --count 10

# Player badges with XP data
steam-web-pp-cli badges gabelogannewell
```

### Scripting and pipelines

```bash
# JSON output for scripting
steam-web-pp-cli games gabelogannewell --sort playtime --limit 5 --json

# Compact output for agents (strips verbose fields)
steam-web-pp-cli friends gabelogannewell --agent

# Select specific fields
steam-web-pp-cli games gabelogannewell --json --select appid,name,playtime_forever

# Dry run to preview the request
steam-web-pp-cli player gabelogannewell --dry-run

# CSV output for spreadsheets
steam-web-pp-cli games gabelogannewell --csv
```

## Commands

### Wrapper Commands

| Command | Description |
|---------|-------------|
| `player <id>` | Get player profile summary |
| `games <id>` | List owned games with `--sort` and `--limit` |
| `friends <id>` | List friends with profile summaries (batched) |
| `achievements <id> <appid>` | Player achievements for a game |
| `news <appid>` | Recent news for a game with `--count` |
| `bans <id>...` | VAC and community ban status |
| `recent <id>` | Recently played games with `--count` |
| `badges <id>` | Player badges with XP |
| `players-count <appid>` | Current online player count (alias: `players`) |
| `resolve <vanity>` | Resolve vanity URL to SteamID64 |

### Transcendence Commands

| Command | Description |
|---------|-------------|
| `completionist <id>` | Achievement completion rates across library |
| `compare <id1> <id2>` | Compare game libraries between two players |
| `rare <id> <appid>` | Rarest achievements a player has earned |
| `backlog <id>` | Unplayed/barely-played games |
| `overlap <id>...` | Games owned by all listed players |
| `profile <id>` | Full aggregated profile (level, badges, games, recent) |

### Insight Commands

| Command | Description |
|---------|-------------|
| `playtime <id>` | Playtime statistics and distribution |
| `stats <id> <appid>` | Player in-game stats for a game |
| `global-achievements <appid>` | Global achievement percentages with `--rare` / `--limit` |
| `schema <appid>` | Achievement and stat schema for a game |
| `groups <id>` | Player's Steam groups |
| `level <id>` | Player's Steam level |

### Raw API Commands

The full Steam Web API is also available via interface commands (e.g. `isteam-user get-player-summaries`). Use `steam-web-pp-cli --help` to see all available interface commands.

## Output Formats

```bash
# Human-readable table (default in terminal)
steam-web-pp-cli games gabelogannewell --sort playtime --limit 10

# JSON for scripting and agents (default when piped)
steam-web-pp-cli games gabelogannewell --json

# Filter specific fields
steam-web-pp-cli games gabelogannewell --json --select appid,name

# CSV for spreadsheets
steam-web-pp-cli games gabelogannewell --csv

# Plain tab-separated for piping
steam-web-pp-cli games gabelogannewell --plain

# Compact (agent-optimized, strips verbose fields)
steam-web-pp-cli games gabelogannewell --compact

# Dry run (show request without sending)
steam-web-pp-cli player gabelogannewell --dry-run
```

## Configuration

Config file: `~/.config/steam-web-pp-cli/config.toml`

```toml
api_key = "your_steam_api_key"
base_url = "https://api.steampowered.com"   # optional, default
```

Environment variables (override config file):
- `STEAM_API_KEY` -- Steam Web API key
- `STEAM_KEY` -- Alternative env var for the API key
- `STEAM_WEB_CONFIG` -- Custom config file path
- `STEAM_WEB_BASE_URL` -- Override base URL

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
