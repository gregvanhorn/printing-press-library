# IMDb CLI Brief

## API Identity
- Domain: Entertainment metadata — movies, TV shows, episodes, people, ratings
- Users: Movie enthusiasts, media organizers, data analysts, developers, film students, recommendation engine builders
- Data profile: ~11M titles, ~13M people, daily-refreshed datasets. Mix of real-time API (OMDb) + bulk datasets (IMDb Non-Commercial)

## Data Sources (Dual-Layer Architecture)

### Layer 1: OMDb REST API (real-time)
- Base URL: https://www.omdbapi.com/
- Auth: API key (free tier: 1,000 req/day, paid: unlimited)
- Endpoints: search by title (/?s=), get by title (/?t=), get by IMDb ID (/?i=)
- Returns: title, year, rated, released, runtime, genre, director, writer, actors, plot, language, country, awards, poster, ratings (IMDb, Rotten Tomatoes, Metacritic), boxOffice, DVD, production
- OpenAPI spec: https://github.com/zuplo-samples/OMDb-OpenAPI/blob/main/omdb-v3.json

### Layer 2: IMDb Non-Commercial Datasets (bulk/offline)
- Source: https://datasets.imdbws.com/
- Format: gzipped TSV, UTF-8, refreshed daily
- Files:
  - title.basics.tsv.gz — tconst, titleType, primaryTitle, originalTitle, isAdult, startYear, endYear, runtimeMinutes, genres
  - title.ratings.tsv.gz — tconst, averageRating, numVotes
  - title.crew.tsv.gz — tconst, directors, writers
  - title.episode.tsv.gz — tconst, parentTconst, seasonNumber, episodeNumber
  - title.principals.tsv.gz — tconst, ordering, nconst, category, job, characters
  - name.basics.tsv.gz — nconst, primaryName, birthYear, deathYear, primaryProfession, knownForTitles
  - title.akas.tsv.gz — titleId, ordering, title, region, language, types, attributes, isOriginalTitle
- License: Non-commercial personal use only
- Raw size: ~5.5 GB uncompressed, ~19 GB with SQLite indices

## Reachability Risk
- **Low** — OMDb API is stable, free tier works reliably. IMDb datasets are public downloads. No 403 issues when using these approved paths. Direct IMDB.com scraping shows 403 blocks (but we don't scrape — we use the official datasets and OMDb API).

## Top Workflows
1. **Search & discover**: Find movies/shows by title, genre, year, rating threshold — "show me all sci-fi movies rated above 8.0 from the 2010s"
2. **Deep dive on a title**: Full cast, crew, ratings from multiple sources, episode guide, related titles
3. **Filmography explorer**: Browse an actor/director's complete career with ratings, sort by year/rating/votes
4. **Rating analytics**: Genre trends over decades, director batting averages, franchise score trajectories
5. **Media library management**: Match local media files to IMDb entries, identify unrated/low-rated content

## Table Stakes (from competitors)
- Search by title (all existing CLIs)
- Get title details by ID (all existing CLIs)
- Show ratings from multiple sources (mediascore)
- Fuzzy search across millions of titles (BurntSushi/imdb-rename)
- Bulk dataset import to SQLite (imdb-sqlite)
- Top 250 movies/shows lists (uzaysozen/imdb-mcp-server)
- Cast and crew lookup (uzaysozen/imdb-mcp-server)
- Episode guide for TV series (imdb-api npm)
- Poster/image URLs (mcp-server-imdb)
- --json output (none have this)
- Pagination (uzaysozen/imdb-mcp-server has it for MCP)

## Data Layer
- Primary entities: titles, people, ratings, episodes, crew, akas
- Sync cursor: dataset file timestamps (daily refresh). OMDb lookups are on-demand cache.
- FTS/search: FTS5 on primaryTitle, originalTitle, primaryName for instant fuzzy search across all 11M+ titles and 13M+ people

## Product Thesis
- Name: imdb-pp-cli
- Why it should exist: No existing tool combines real-time OMDb API lookups with a local SQLite database of the entire IMDb catalog. Current tools are either thin API wrappers (3 endpoints, limited) or dataset importers (no query interface). This CLI gives you the entire IMDb in a searchable local database with FTS5, plus real-time enrichment via OMDb, plus compound analytics (genre trends, career timelines, franchise tracking) that are only possible when all the data lives together locally.

## Build Priorities
1. Dataset import pipeline (download, decompress, SQLite ingest) — the foundation for everything
2. FTS5 search across titles and people — the core UX
3. OMDb API integration for real-time enrichment (ratings, plots, posters)
4. Filmography, episode guides, and cross-entity queries
5. Analytics: genre trends, director stats, decade rankings, franchise tracking
