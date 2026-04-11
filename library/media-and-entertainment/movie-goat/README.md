# Movie Goat CLI

The only movie CLI with multi-source ratings (TMDb + IMDb + Rotten Tomatoes + Metacritic), streaming availability, and cross-taste recommendations. Powered by TMDb and OMDb.

Look up any movie and see ratings from four sources in one view. Find where to stream it. Get recommendations by telling it movies you like. Plan a franchise marathon. Discover hidden gems by genre and decade. Compare two movies head-to-head.

## Install

### Go

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Authentication

Get a free TMDb API key at [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api), then set it:

```bash
export TMDB_API_KEY="your-api-key"
```

Or store it persistently:

```bash
movie-goat-pp-cli auth set-token YOUR_API_KEY
```

### OMDb (optional — enables Rotten Tomatoes & Metacritic)

OMDb provides Rotten Tomatoes and Metacritic ratings that TMDb doesn't have. Without it, everything works — you just see TMDb ratings only. With it, `movies get`, `versus`, and other commands show all four rating sources.

Get a free key (1,000 requests/day) at [omdbapi.com/apikey.aspx](http://www.omdbapi.com/apikey.aspx). For heavier use, [Patreon supporters](https://www.patreon.com/join/omdb) get 100,000 requests/day starting at $1/month.

```bash
export OMDB_API_KEY="your-omdb-key"
```

## Quick Start

```bash
# Check your setup
movie-goat-pp-cli doctor

# What should I watch tonight?
movie-goat-pp-cli tonight

# Random surprise pick
movie-goat-pp-cli blind --genre 878 --min-rating 7.5

# Compare two movies head-to-head
movie-goat-pp-cli versus "The Dark Knight" "Inception"

# Where can I stream this?
movie-goat-pp-cli watch "Breaking Bad" --type tv
```

## Unique Features

These capabilities aren't available in any other tool for this API.

- **`movies get`** -- See TMDb, IMDb, Rotten Tomatoes, and Metacritic ratings for any movie in one view
- **`watch`** -- Find every streaming platform, rental option, and purchase option for any movie or show
- **`career`** -- Explore any actor or director's complete filmography with ratings and career trajectory
- **`tonight`** -- Get a smart recommendation for what to watch tonight based on trending and your streaming services
- **`versus`** -- Compare two movies or shows side-by-side across every metric: ratings, cast, box office, streaming
- **`marathon`** -- Plan a franchise movie marathon with watch order, total runtime, and suggested breaks

## Commands

### Discovery and Recommendations

| Command | Description |
|---------|-------------|
| `tonight` | Smart "what should I watch tonight?" recommendation |
| `blind` | Random high-quality movie pick with genre, decade, and streaming filters |
| `recommend-for-me` | Consensus recommendations based on movies you love |
| `discover` | Discover movies by genre, year, rating, certification, cast, crew, streaming provider |
| `discover tv` | Discover TV shows by genre, year, rating, network, streaming provider |
| `trending` | Trending movies, TV shows, and people |

### Movies

| Command | Description |
|---------|-------------|
| `movies get <id>` | Full movie detail with cast, ratings, streaming, and recommendations |
| `movies search <query>` | Search movies by title |
| `movies now-playing` | Movies currently in theaters |
| `movies upcoming` | Movies coming soon |
| `movies top-rated` | Highest rated movies |

### TV Shows

| Command | Description |
|---------|-------------|
| `tv get <id>` | Full TV show detail |
| `tv search <query>` | Search TV shows by title |
| `tv airing-today` | Shows with episodes airing today |
| `tv on-the-air` | Shows currently on the air |
| `tv top-rated` | Highest rated TV shows |

### People

| Command | Description |
|---------|-------------|
| `people` | Popular people in entertainment |
| `people get <id>` | Person detail with filmography |
| `people search <query>` | Search people by name |
| `career <name-or-id>` | Full career filmography with stats and highlights |

### Comparison and Planning

| Command | Description |
|---------|-------------|
| `versus <title1> <title2>` | Side-by-side comparison of two movies |
| `marathon <collection>` | Franchise marathon planner with watch order and runtime |
| `watch <title-or-id>` | Streaming availability by region |

### Data and Utilities

| Command | Description |
|---------|-------------|
| `search <query>` | Full-text search across synced data or live API |
| `sync` | Sync API data to local SQLite for offline search |
| `analytics` | Analyze synced data with count, group-by, and summary |
| `export` | Export data to JSONL or JSON |
| `import` | Import data from JSONL file |
| `genres` | Genre lists for movies and TV |
| `doctor` | Check CLI health, auth, and API connectivity |
| `auth` | Manage authentication tokens |

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
movie-goat-pp-cli trending movies

# JSON for scripting and agents
movie-goat-pp-cli trending movies --json

# Filter to specific fields
movie-goat-pp-cli trending movies --json --select title,vote_average

# CSV output
movie-goat-pp-cli trending movies --csv

# Compact output (key fields only, minimal tokens)
movie-goat-pp-cli trending movies --compact

# Dry run -- show the request without sending
movie-goat-pp-cli movies get 550 --dry-run

# Agent mode -- JSON + compact + no prompts in one flag
movie-goat-pp-cli tonight --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** -- never prompts, every input is a flag
- **Pipeable** -- `--json` output to stdout, errors to stderr
- **Filterable** -- `--select id,name` returns only fields you need
- **Previewable** -- `--dry-run` shows the request without sending
- **Confirmable** -- `--yes` for explicit confirmation of destructive actions
- **Cacheable** -- GET responses cached for 5 minutes, bypass with `--no-cache`
- **Agent-safe by default** -- no colors or formatting unless `--human-friendly` is set
- **Progress events** -- paginated commands emit NDJSON events to stderr

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use as MCP Server

This CLI ships a companion MCP server for use with Claude Desktop, Cursor, and other MCP-compatible tools.

### Claude Code

```bash
claude mcp add movie movie-goat-pp-mcp -e TMDB_API_KEY=<your-token>
```

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "movie": {
      "command": "movie-goat-pp-mcp",
      "env": {
        "TMDB_API_KEY": "<your-key>"
      }
    }
  }
}
```

## Cookbook

```bash
# What should I watch tonight?
movie-goat-pp-cli tonight

# Tonight picks filtered to sci-fi
movie-goat-pp-cli tonight --genre 878

# Random surprise movie
movie-goat-pp-cli blind --min-rating 7.5

# Random 90s comedy on Netflix
movie-goat-pp-cli blind --genre 35 --decade 1990s --streaming 8

# Compare two movies side-by-side
movie-goat-pp-cli versus "Barbie" "Oppenheimer"

# Plan a Lord of the Rings marathon
movie-goat-pp-cli marathon "Lord of the Rings"

# Where to stream a specific movie
movie-goat-pp-cli watch "Inception" --region US

# Full career of a director, sorted by rating
movie-goat-pp-cli career "Denis Villeneuve" --sort rating

# Get recommendations based on your favorites
movie-goat-pp-cli recommend-for-me "Parasite" "Oldboy" "The Handmaiden"

# Discover highly-rated action movies from 2024
movie-goat-pp-cli discover --with-genres 28 --primary-release-year 2024 --vote-average-gte 7.0

# Find all R-rated horror
movie-goat-pp-cli discover --with-genres 27 --certification R --certification-country US

# Sync data for offline search
movie-goat-pp-cli sync

# Search synced data
movie-goat-pp-cli search "time travel"

# Export trending movies as JSONL for scripting
movie-goat-pp-cli trending movies --json | jq -c '.[]'
```

## Health Check

```bash
$ movie-goat-pp-cli doctor
  OK Config: ok
  FAIL Auth: not configured
  OK API: reachable
  config_path: /Users/you/.config/movie-goat-pp-cli/config.toml
  base_url: https://api.themoviedb.org/3
  version: 1.0.0
  hint: export TMDB_API_KEY=<your-key>
  Get a key at: https://www.themoviedb.org/settings/api
```

## Configuration

Config file: `~/.config/movie-goat-pp-cli/config.toml`

Environment variables:

| Variable | Description |
|----------|-------------|
| `TMDB_API_KEY` | TMDb API key (required) |
| `OMDB_API_KEY` | OMDb API key (optional, enables multi-source ratings) |
| `MOVIE_GOAT_BASE_URL` | Override API base URL (for self-hosted or testing) |
| `MOVIE_GOAT_CONFIG` | Override config file path |

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `movie-goat-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $TMDB_API_KEY`
- Get a key at [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api)

**Not found errors (exit code 3)**
- Check the resource ID is correct
- Try searching by name instead of ID: `movie-goat-pp-cli movies search "title"`

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- Use `--rate-limit 2` to throttle requests
- If persistent, wait a few minutes and try again

**Config errors (exit code 10)**
- Check config file syntax: `cat ~/.config/movie-goat-pp-cli/config.toml`
- Reset with `movie-goat-pp-cli auth set-token YOUR_KEY`

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**tmdb-mcp**](https://github.com/xdwanj/tmdb-mcp) -- Go
- [**imdb-mcp-server**](https://github.com/uzaysozen/imdb-mcp-server) -- Python
- [**mediascore**](https://github.com/dkorunic/mediascore) -- Go
- [**TMDB_CLI**](https://github.com/illegalbyte/TMDB_CLI) -- Python
- [**tmdb-cli (bhantsi)**](https://github.com/bhantsi/tmdb-cli) -- JavaScript
- [**imdb-rename**](https://github.com/BurntSushi/imdb-rename) -- Rust

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
