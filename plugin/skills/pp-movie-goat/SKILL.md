---
name: pp-movie-goat
description: "Use this skill whenever the user asks about movies, TV shows, actors, directors, ratings, where to stream something, what to watch tonight, franchise marathons, hidden gems, or cross-taste recommendations. Multi-source movie CLI combining TMDb + OMDb for ratings from four critics at once plus streaming availability. Requires a free TMDb API key. Triggers on phrasings like 'what should I watch tonight', 'where can I stream Oppenheimer', 'compare Dune and Blade Runner', 'best movies from Denis Villeneuve', 'plan a Marvel marathon', 'what's trending on TMDb this week'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["movie-goat-pp-cli"],"env":["TMDB_API_KEY"]},"primaryEnv":"TMDB_API_KEY","install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-cli@latest","bins":["movie-goat-pp-cli"],"label":"Install via go install"}]}}'
---

# Movie Goat — Printing Press CLI

Multi-source movie ratings (TMDb + OMDb), streaming availability, and cross-taste recommendations. Look up any movie and see ratings from four sources in one view. Find where to stream it. Get recommendations by telling the CLI movies you love. Plan a franchise marathon with total runtime. Discover hidden gems by genre and decade. Compare two movies head-to-head.

## When to Use This CLI

Reach for this when a user wants movie or TV information — ratings, where to watch, recommendations, franchise planning, or cross-comparisons — and prefers a terminal-native flow over opening IMDB/Rotten Tomatoes/Letterboxd tabs one by one. Particularly valuable when a user can express taste as "I liked X and Y" and wants suggestions that bridge those.

Don't reach for this if the user wants real-time reviews from a specific critic (Roger Ebert etc.) or currently-showing-in-theaters data beyond what TMDb's release-date feeds provide.

## Unique Capabilities

Commands that combine TMDb + OMDb data or derive from streaming/watch-taste data.

### Multi-source ratings in one shot

- **`movies get <movieId>`** — Returns TMDb user score, IMDB rating, Metacritic, Rotten Tomatoes — all in one response.

  _Skips four tab-switches. Gives the agent a clean multi-source comparison to summarize._

- **`versus <movie1> <movie2>`** — Side-by-side rating comparison, box office, cast overlap, streaming availability. For "better movie" arguments and curation.

### Taste-aware recommendations

- **`recommend-for-me <movie> [<movie>...]`** — Give it 2-4 movies you love; it returns picks that bridge their shared tastes via TMDb's recommendation graph + genre/director overlap.

  _Solves "I liked Inception and Dune, what next?" better than Netflix homepage._

- **`tonight`** — Smart "what should I watch" — uses time of day, recent browsing patterns in the local cache, and currently-streaming availability on your providers to pick one film.

- **`blind`** — Random high-quality pick (rating >= 7.5, vote count >= 1000). Anti-decision-fatigue tool.

### Structured discovery

- **`discover [--with-genres <ids>] [--year <y>] [--vote-average-gte N] [--with-watch-providers <ids>]`** — Browse with structured filters. Combine genre IDs, year, rating floor, cast, and streaming provider IDs (e.g. `8`=Netflix, `337`=Disney+) in one query. Use `movies genres` to look up genre IDs.

- **`marathon <franchise>`** — Plan a franchise marathon with chronological or release order, total runtime, and where-to-stream for each title.

  _"Schedule a Star Wars weekend" → ordered list with total hours and streaming mix._

- **`career <person-name-or-id>`** — Complete filmography with genre breakdown, decades active, average rating, most-acclaimed titles.

- **`trending`** — TMDb's real-time trending list — movies, TV, and people.

### Streaming availability

- **`watch <movieId>`** — Where to stream, rent, or buy. Uses TMDb's JustWatch-powered provider data; scoped by region (defaults to US).

## Command Reference

Movies:

- `movie-goat-pp-cli movies` — Browse (popular, top-rated, upcoming, now-playing)
- `movie-goat-pp-cli movies search "<title>"` — Search by title
- `movie-goat-pp-cli movies get <movieId>` — Detail with multi-source ratings

TV:

- `movie-goat-pp-cli tv` — Browse TV shows
- `movie-goat-pp-cli tv get <seriesId>` — Series detail
- `movie-goat-pp-cli tv get <seriesId> <seasonNumber>` — Season detail
- `movie-goat-pp-cli tv airing-today` — Today's airings

People:

- `movie-goat-pp-cli people` — Browse/search
- `movie-goat-pp-cli people get <personId>` — Person detail
- `movie-goat-pp-cli career <person>` — Complete filmography with stats

Discovery / recommendations:

- `movie-goat-pp-cli discover [filters]` — Structured filtered browse
- `movie-goat-pp-cli trending` — Real-time trending
- `movie-goat-pp-cli recommend-for-me <titles>` — Taste-based recs
- `movie-goat-pp-cli tonight` — Smart single pick
- `movie-goat-pp-cli blind` — Random high-quality pick

Planning / comparison:

- `movie-goat-pp-cli marathon <franchise>` — Franchise watch-order
- `movie-goat-pp-cli versus <movie1> <movie2>` — Head-to-head
- `movie-goat-pp-cli watch <movieId>` — Streaming providers
- `movie-goat-pp-cli genres` — List genre IDs

Local / auth:

- `movie-goat-pp-cli sync` / `export` / `import` / `archive` — Local store
- `movie-goat-pp-cli auth set-token <TMDB_KEY>` — Save API key
- `movie-goat-pp-cli doctor` — Verify

## Recipes

### "What should I watch tonight?"

```bash
movie-goat-pp-cli tonight --agent
# or: bias toward recent favorites
movie-goat-pp-cli recommend-for-me "Dune: Part Two" "Oppenheimer" --limit 5 --agent
```

`tonight` picks one film considering time of day and streaming. `recommend-for-me` with 2-4 titles you liked returns a ranked shortlist that blends those tastes.

### Compare two movies before committing

```bash
movie-goat-pp-cli versus "Inception" "Tenet" --agent
```

Returns: TMDb score, IMDB, Metacritic, RT scores for each; box office; shared cast/crew; where each streams. Settles "which should I watch first" without five Google searches.

### Plan a weekend marathon

```bash
movie-goat-pp-cli marathon "Lord of the Rings" --agent
movie-goat-pp-cli marathon "Alien" --agent
```

Pass a TMDb collection name; returns titles in watch order, runtime per film, total weekend duration, and streaming mix so you know if one movie's only on a provider you don't have. Watch order follows TMDb's collection sequence.

### Research a director's full career

```bash
movie-goat-pp-cli career "Denis Villeneuve" --agent
```

Complete filmography, decades active, genres, average rating, highlights. Used as prep for deeper research or "what to watch next from them" picks.

### Discover-by-filter

```bash
movie-goat-pp-cli movies genres --agent                           # look up genre IDs first
# 878 = sci-fi, 8 = Netflix provider ID
movie-goat-pp-cli discover \
  --with-genres 878 --primary-release-year 2005 \
  --vote-average-gte 7.5 \
  --with-watch-providers 8 --watch-region US \
  --agent
```

TMDb's discover endpoint uses numeric IDs for genres and providers — look them up first via `movies genres`. Structured search beats search-by-title when the question is "what 2000s sci-fi with 7.5+ rating is on Netflix."

## Auth Setup

Requires a free TMDb API key. OMDb enrichment is optional.

```bash
# 1. Get a key: https://www.themoviedb.org/settings/api
export TMDB_API_KEY="your-tmdb-key"
movie-goat-pp-cli auth set-token "$TMDB_API_KEY"  # also persists to config
movie-goat-pp-cli doctor
```

Optional:
- `OMDB_API_KEY` — enables extra rating sources (IMDB, Metacritic, RT via OMDb)
- `MOVIE_GOAT_REGION` — streaming region (default `US`)

## Agent Mode

Add `--agent` to any command. Expands to `--json --compact --no-input --no-color --yes`. Supports `--select <fields>`, `--dry-run`, `--rate-limit N` (throttle requests), `--no-cache`.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error |
| 3 | Not found (movie, person, series) |
| 4 | Auth required (missing TMDB_API_KEY) |
| 5 | API error (upstream TMDb or OMDb) |
| 7 | Rate limited |
| 10 | Config error |

## Installation

### CLI

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-cli@latest
movie-goat-pp-cli auth set-token YOUR_TMDB_KEY
movie-goat-pp-cli doctor
```

### MCP Server

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/movie-goat/cmd/movie-goat-pp-mcp@latest
claude mcp add -e TMDB_API_KEY=<key> movie-goat-pp-mcp -- movie-goat-pp-mcp
```

## Argument Parsing

Given `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → run `movie-goat-pp-cli --help`
2. **`install`** → CLI; **`install mcp`** → MCP
3. **Anything else** → check `which movie-goat-pp-cli` (offer install if missing), verify `TMDB_API_KEY` is set (prompt for setup if not), match intent to a command (taste queries → `recommend-for-me`, comparison → `versus`, franchise → `marathon`, single pick → `tonight`), run with `--agent`.
