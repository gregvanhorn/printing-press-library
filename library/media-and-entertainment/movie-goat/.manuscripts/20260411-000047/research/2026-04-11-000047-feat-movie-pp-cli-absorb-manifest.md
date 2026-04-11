# Movie CLI Absorb Manifest

## Sources Analyzed
1. **tmdb-mcp** (XDwanj) — MCP server: search, details, discover, trending, recommendations
2. **imdb-mcp-server** (uzaysozen) — MCP server: search, details, top 250, popular, upcoming, cast/directors, regional
3. **mcp-server-imdb** (mrbourne777) — MCP server: search, poster, trailer (FM-DB API, no auth)
4. **tmdb-cli** (degerahmet, Go) — CLI: popular, top-rated, upcoming, now-playing
5. **tmdb-cli** (letsmakecakes, Go) — CLI: popular, top-rated, upcoming, now-playing
6. **tmdb-cli** (che1nov, Go) — CLI: popular, top-rated, upcoming, now-playing + Redis cache
7. **TMDB_CLI** (illegalbyte, Python) — CLI: movie/TV by ID, IMDB↔TMDB ID convert, JSON output
8. **tmdb-cli** (bhantsi) — CLI: search, trending, upcoming, movie details
9. **mediascore** (dkorunic, Go, archived) — CLI: multi-source ratings from OMDb/IMDB/RT/Metacritic
10. **BurntSushi/imdb-rename** (Rust) — Fuzzy search via BM25 across IMDb datasets
11. **tmdbv3api** (Python wrapper) — popular, details, search, similar, recommendations, discover, season
12. **TMDB-Trakt-Syncer** (Python) — Watchlist/ratings sync between TMDb and Trakt

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Search movies by title | tmdb-mcp search | `movie search <query>` | Multi-type (movie/TV/person), --type filter, --year filter, --json |
| 2 | Get movie details by ID | TMDB_CLI --movie | `movie get <id>` | Rich card display: rating, cast, where to watch, trailer link. --json, --select |
| 3 | Get TV show details | TMDB_CLI --television | `movie tv <id>` | Season/episode count, status, networks, next air date. --json |
| 4 | Get person details | tmdbv3api Person.details | `movie person <id-or-name>` | Full bio + filmography inline. --json |
| 5 | Popular movies | tmdb-cli popular | `movie popular` | Movies AND TV, --type filter, --page, --json |
| 6 | Top-rated movies | tmdb-cli top-rated | `movie top-rated` | Movies AND TV, --type filter, --genre, --json |
| 7 | Upcoming movies | tmdb-cli upcoming | `movie upcoming` | With streaming date info where available. --json |
| 8 | Now playing in theaters | tmdb-cli now-playing | `movie now-playing` | With runtime and certification. --json |
| 9 | Trending content | tmdb-mcp get_trending | `movie trending` | Daily or weekly, movies/TV/people/all. --window day\|week, --type |
| 10 | Discover movies with filters | tmdb-mcp discover_movies | `movie discover` | Genre, year, rating, votes, certification, cast, crew, keywords, providers. --json |
| 11 | Discover TV with filters | tmdb-mcp discover_tv | `movie discover --type tv` | Same rich filtering for TV shows |
| 12 | Recommendations | tmdb-mcp get_recommendations | `movie recommend <id>` | For any movie or TV show. --json |
| 13 | Similar titles | tmdbv3api similar | `movie similar <id>` | For any movie or TV show. --json |
| 14 | Watch providers | TMDb API | `movie watch <id>` | Stream/rent/buy grouped by provider with region. --region flag |
| 15 | Cast & crew | imdb-mcp-server get_cast | `movie credits <id>` | Full cast + crew, sorted by billing order. --role filter |
| 16 | Episode guide | tmdbv3api Season.details | `movie episodes <tv-id> [--season N]` | All seasons or specific, with ratings per episode |
| 17 | IMDB↔TMDB ID convert | TMDB_CLI --imdbidconvert | `movie id-convert <id>` | Auto-detect IMDB (tt*) vs TMDB format |
| 18 | Poster/image URLs | mcp-server-imdb poster | `movie images <id>` | Posters, backdrops, profiles. --type filter |
| 19 | Trailer/video URLs | mcp-server-imdb trailer | `movie videos <id>` | Trailers, teasers, clips, featurettes. YouTube links |
| 20 | JSON output | TMDB_CLI --json | `--json` on every command | Global flag, valid JSON, pipeable to jq |
| 21 | Genre listing | TMDb API | `movie genres` | Movie + TV genres with IDs. --json |
| 22 | On the air (TV) | TMDb API | `movie on-the-air` | Currently airing TV shows |
| 23 | Airing today (TV) | TMDb API | `movie airing-today` | TV shows with episodes airing today |
| 24 | Popular people | TMDb API | `movie popular --type person` | Popular actors/directors |

## Transcendence (only possible with our compound approach)

| # | Feature | Command | Score | Why Only We Can Do This |
|---|---------|---------|-------|------------------------|
| 1 | Where to watch dashboard | `movie watch <title>` | 9/10 | Shows ALL streaming platforms grouped by stream/rent/buy with prices. No existing CLI surfaces watch provider data at all — they only show ratings. |
| 2 | Career timeline | `movie career <person>` | 8/10 | Full filmography with ratings, revenue, genre distribution. Shows career arc — breakout role, peak years, genre shifts. Requires combining person credits with movie details. |
| 3 | Tonight picker | `movie tonight` | 8/10 | Smart "what should I watch?" based on trending, your watchlist, genre preferences, and what's available on your streaming services. Combines discover + trending + local prefs. |
| 4 | Head-to-head compare | `movie versus <title1> <title2>` | 7/10 | Side-by-side comparison: ratings, box office, cast overlap, runtime, streaming availability. No existing tool compares titles. |
| 5 | Marathon planner | `movie marathon <collection>` | 7/10 | Plan a franchise marathon: chronological or release order, total runtime, suggested break points. Works for any TMDb collection (MCU, Star Wars, etc). |
| 6 | Cast overlap finder | `movie cast-overlap <person1> <person2>` | 7/10 | Find every movie/show where two actors appeared together. Requires cross-referencing two filmographies — no single API call can do this. |
| 7 | Decade leaderboard | `movie decade <1990s>` | 6/10 | Best movies/shows of any decade, filterable by genre. Combines discover with date range filtering and presentation. |
| 8 | Watchlist with streaming alerts | `movie watchlist` | 6/10 | Local watchlist that shows current streaming availability for each item. "Inception is on Netflix." Updates on every check. |
