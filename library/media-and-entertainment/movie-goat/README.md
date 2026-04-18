# Movie Goat CLI

**Find the best thing to watch with one CLI for multi-source ratings, streaming availability, franchise planning, and taste-bridging recommendations.**

Movie GOAT combines TMDb discovery, search, credits, collections, and watch-provider data with optional OMDb enrichment so one tool can answer practical movie-night questions fast:

- "What should I watch tonight?"
- "Which movie actually wins when I compare them side by side?"
- "Where can I stream this in the US or another region?"
- "What bridges the gap between my partner's taste and mine?"
- "How long would a full franchise marathon take?"

It also keeps a local SQLite archive for offline search and analytics, so the CLI is useful both as an interactive tool and as a reusable data source for scripts and agents.

## Install

### Go

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Required: TMDb API key

**All core commands require TMDb credentials.**

### What the key unlocks

TMDb powers the main CLI experience:

- `movies get`, `movies search`, `tv get`, `tv search`
- `watch` for region-specific streaming, rental, and purchase options
- `tonight`, `blind`, `discover`, and `trending`
- `recommend-for-me`, `versus`, `career`, and `marathon`
- `sync`, `search`, and `analytics` against a local archive

### Get a key

1. Create a free account at https://www.themoviedb.org/.
2. Request an API key or read token at https://www.themoviedb.org/settings/api.
3. Export it:

```bash
export TMDB_API_KEY="<your-key>"
```

Or persist credentials in the CLI config:

```bash
movie-goat-pp-cli auth set-token <your-key>
```

Verify with:

```bash
movie-goat-pp-cli doctor
```

`Auth` should move from `not configured` to `configured`.

## Optional: OMDb API key

**Movie GOAT works without OMDb.** OMDb adds the extra ratings and metadata that TMDb does not provide directly.

### What the key unlocks

When `OMDB_API_KEY` is set, Movie GOAT can enrich results with:

- IMDb rating
- Rotten Tomatoes score
- Metacritic score
- awards
- box office

This matters most for:

- `movies get`
- `blind`
- `versus`

### Get a key

Get a free key at http://www.omdbapi.com/apikey.aspx.

```bash
export OMDB_API_KEY="<your-omdb-key>"
```

If you do not set it, the CLI still works. You just get TMDb-only results where OMDb enrichment would otherwise appear.

## Quick Start

```bash
# Check config, auth, and API reachability
movie-goat-pp-cli doctor

# What should I watch tonight?
movie-goat-pp-cli tonight

# Random sci-fi pick above 7.5/10
movie-goat-pp-cli blind --genre 878 --min-rating 7.5

# Compare two movies head-to-head
movie-goat-pp-cli versus "The Dark Knight" "Inception"

# Find where to stream a show
movie-goat-pp-cli watch "Breaking Bad" --type tv --region US

# Get recommendations from multiple taste anchors
movie-goat-pp-cli recommend-for-me "Parasite" "Oldboy" "The Handmaiden"

# Plan a franchise marathon
movie-goat-pp-cli marathon "Lord of the Rings"

# Build a local archive for offline search
movie-goat-pp-cli sync
movie-goat-pp-cli search "time travel" --data-source local
```

## Unique Features

These are the workflows that make Movie GOAT more useful than a thin TMDb wrapper.

### `movies get`

Get a movie's core record plus appended credits, videos, recommendations, watch providers, and external IDs. With OMDb configured, the command can also attach IMDb, Rotten Tomatoes, and Metacritic ratings in the same result.

_Use this when you need one canonical view of a movie instead of checking TMDb, IMDb, and streaming apps separately._

```bash
movie-goat-pp-cli movies get 550
movie-goat-pp-cli movies get 550 --json
```

### `watch`

Look up where a movie or TV show is available to stream, rent, buy, or watch free with ads in a specific region.

_Useful when the real question is not "is this good?" but "can I actually watch it tonight in my country?"_

```bash
movie-goat-pp-cli watch "Inception" --region US
movie-goat-pp-cli watch "Breaking Bad" --type tv --region GB
```

### `recommend-for-me`

Takes two or more movies you already love, fetches recommendations and similars for each, then scores candidates by how well they bridge the input tastes.

_This is stronger than "more like this" because it tries to find overlap across multiple taste anchors instead of extending a single seed._

```bash
movie-goat-pp-cli recommend-for-me "The Matrix" "Mad Max: Fury Road" "Arrival"
```

### `versus`

Compares two movies side by side across ratings, runtime, genres, streaming, and, when OMDb is configured, additional review-score and box-office context.

_Use this when you are deciding between two options and want a direct answer, not two disconnected pages._

```bash
movie-goat-pp-cli versus "Barbie" "Oppenheimer"
```

### `blind`

The "surprise me" command. It discovers a random well-rated movie using filters for genre, decade, minimum rating, and streaming provider.

_Good when the bottleneck is decision fatigue, not lack of options._

```bash
movie-goat-pp-cli blind --genre 35 --decade 1990s --min-rating 7.0
```

### `tonight`

Builds a short ranked list from trending and popular titles, then sorts by a weighted popularity and rating score.

_This is the fast answer to "just give me a few strong picks right now."_

```bash
movie-goat-pp-cli tonight
movie-goat-pp-cli tonight --type tv
movie-goat-pp-cli tonight --genre 28
```

### `career`

Looks up a person's combined credits and turns them into a sortable filmography with summary stats.

_Useful for questions like "what are this director's best-rated works?" or "what was their busiest period?"_

```bash
movie-goat-pp-cli career "Denis Villeneuve" --sort rating
```

### `marathon`

Finds a TMDb collection, orders the films by release date, totals runtime, and suggests break points.

_Great for franchise planning because it turns "should we do this?" into a concrete time commitment._

```bash
movie-goat-pp-cli marathon "Harry Potter"
```

### `sync` + `search` + `analytics`

Movie GOAT can archive TMDb data locally into SQLite, then search it with FTS and run simple analytics on top.

_This is what turns the CLI from a one-shot API client into a reusable local movie dataset._

```bash
movie-goat-pp-cli sync
movie-goat-pp-cli search "time travel" --data-source local
movie-goat-pp-cli analytics --json
```

## Usage

Run `movie-goat-pp-cli --help` for the full command tree and all flags.

## Commands

### Discovery and recommendations

| Command | Description |
| --- | --- |
| `tonight` | Shortlist strong picks for tonight |
| `blind` | Random high-quality movie pick with filters |
| `recommend-for-me` | Blend multiple favorites into one recommendation set |
| `discover` | Discover movies with genre, year, rating, certification, and provider filters |
| `discover tv` | Discover TV shows with TV-specific filters |
| `trending` | Trending movies, TV, and people |

### Title and person lookup

| Command | Description |
| --- | --- |
| `movies get <id>` | Detailed movie record with appended context |
| `movies search <query>` | Search movies by title |
| `movies now-playing` | Movies in theaters now |
| `movies upcoming` | Upcoming movies |
| `movies top-rated` | Top-rated movies |
| `tv get <id>` | Detailed TV show record |
| `tv search <query>` | Search TV shows by title |
| `tv airing-today` | Shows airing today |
| `tv on-the-air` | Shows currently on the air |
| `tv top-rated` | Top-rated TV shows |
| `people` | Popular people |
| `people get <id>` | Person detail |
| `people search <query>` | Search people by name |
| `career <name-or-id>` | Full career summary and filmography |

### Planning and comparison

| Command | Description |
| --- | --- |
| `watch <title-or-id>` | Watch providers by region |
| `versus <title1> <title2>` | Side-by-side movie comparison |
| `marathon <collection>` | Franchise marathon planner |

### Local archive and power-user tools

| Command | Description |
| --- | --- |
| `sync` | Archive API data to local SQLite |
| `search <query>` | Search live API or local archive |
| `analytics` | Summaries over synced local data |
| `workflow archive` | Sync all resources for offline access |
| `workflow status` | Show local archive status |
| `export` | Export API data to JSONL or JSON |
| `import` | Import JSONL via API create/upsert calls |
| `tail <resource>` | Poll a resource and emit NDJSON events |
| `api` | Browse raw interface coverage |
| `doctor` | Check config, auth, and API reachability |
| `auth` | Show, set, or clear credentials |

## Output Formats

```bash
# Human-readable output in a terminal
movie-goat-pp-cli trending movies

# JSON for scripting and agents
movie-goat-pp-cli trending movies --json

# Select only the fields you need
movie-goat-pp-cli trending movies --json --select id,title,vote_average

# CSV output
movie-goat-pp-cli trending movies --csv

# Plain tab-separated text
movie-goat-pp-cli trending movies --plain

# Compact output for lower token usage
movie-goat-pp-cli trending movies --compact

# Show the request without sending it
movie-goat-pp-cli movies get 550 --dry-run

# Force local or live reads
movie-goat-pp-cli search "noir" --data-source local
movie-goat-pp-cli search "noir" --data-source live

# Agent mode = --json --compact --no-input --no-color --yes
movie-goat-pp-cli tonight --agent
```

## Agent Usage

This CLI is designed to behave well in scripts, MCP clients, and non-interactive agent runs:

- Non-interactive: every input can be provided via args, flags, or stdin
- Pipeable: data goes to stdout, operational messages go to stderr
- Filterable: `--select` reduces payload size
- Compactable: `--compact` keeps only key fields
- Previewable: `--dry-run` shows the request shape without sending it
- Cacheable: GET requests are cached unless `--no-cache` is set
- Source-aware: `--data-source auto|live|local` makes live/local behavior explicit
- Agent mode: `--agent` bundles the sane defaults for machine consumption

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use as MCP Server

This project also ships a companion MCP server.

### Claude Code

```bash
claude mcp add movie-goat movie-goat-pp-mcp -e TMDB_API_KEY=<your-key>
```

If you want OMDb enrichment there too, add `-e OMDB_API_KEY=<your-omdb-key>` as well.

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "movie-goat": {
      "command": "movie-goat-pp-mcp",
      "env": {
        "TMDB_API_KEY": "<your-key>",
        "OMDB_API_KEY": "<optional-omdb-key>"
      }
    }
  }
}
```

## Health Check

```bash
movie-goat-pp-cli doctor
```

This verifies:

- config loading
- auth presence
- API reachability
- version

## Configuration

Config file:

```text
~/.config/movie-goat-pp-cli/config.toml
```

Local archive database:

```text
~/.local/share/movie-goat-pp-cli/data.db
```

Environment variables:

- `TMDB_API_KEY`
- `OMDB_API_KEY`
- `MOVIE_GOAT_CONFIG`
- `MOVIE_GOAT_BASE_URL`

## Troubleshooting

**`doctor` says auth is not configured**

- Export `TMDB_API_KEY`
- Or run `movie-goat-pp-cli auth set-token <your-key>`
- Re-run `movie-goat-pp-cli doctor`

**I only see TMDb ratings**

- Set `OMDB_API_KEY`
- Re-run `movies get`, `blind`, or `versus`
- If the title has no IMDb mapping, OMDb enrichment may still be unavailable

**`watch` returns no providers for my region**

- Try another region with `--region`, for example `US`, `GB`, or `DE`
- Some titles have provider data only for a subset of regions

**Local search fails**

- Run `movie-goat-pp-cli sync` first
- Then search with `--data-source local`
- Confirm the DB exists at `~/.local/share/movie-goat-pp-cli/data.db`

**Search results are inconsistent**

- `--data-source auto` prefers live API search and falls back to local on network failure
- Use `--data-source live` or `--data-source local` when you need deterministic behavior

**You are hitting API limits or repeated network failures**

- Reduce request volume
- Use the built-in cache instead of `--no-cache`
- Use `--rate-limit` for gentler scripted runs

<!-- pr-218-features -->
## Agent workflow features

This CLI was patched to add these agent-workflow capabilities (see [`printing-press patch`](https://github.com/mvanhorn/cli-printing-press/pull/221)):

- **Named profiles** — save a set of flags under a name and reuse them: `movie-goat-pp-cli profile save <name> --<flag> <value>`, then `movie-goat-pp-cli --profile <name> <command>`. Flag precedence: explicit flag > env var > profile > default.
- **`--deliver`** — route command output to a sink other than stdout. Values: `file:<path>` writes atomically via tmp+rename; `webhook:<url>` POSTs as JSON (or NDJSON with `--compact`).
- **`feedback`** — record in-band feedback about the CLI. Entries append as JSON lines to `~/.movie-goat-pp-cli/feedback.jsonl`. When `MOVIE_GOAT_FEEDBACK_ENDPOINT` is set and either `--send` is passed or `MOVIE_GOAT_FEEDBACK_AUTO_SEND=true`, the entry is also POSTed upstream.
