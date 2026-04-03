# Postman Explore CLI Brief

## API Identity
- **Domain:** API discovery and collection marketplace
- **Users:** Developers, API integrators, DevOps engineers, AI agent builders
- **Data profile:** 705K+ public collections, 309K+ workspaces, 25K+ APIs, 6K+ flows, 163K+ teams
- **Auth:** None required for public browsing/search. No API key needed.
- **Protocol:** All requests POST to `_api/ws/proxy` with `{service, method, path, body?}` — a proxy gateway pattern

## Top Workflows
1. **Search for a collection** — "Find me a Stripe payment collection I can fork"
2. **Browse by category** — "Show me all AI collections sorted by popularity this week"
3. **Discover publishers** — "Which verified teams publish the most-forked collections?"
4. **Track trending** — "What collections got the most forks this week?"
5. **Compare collections** — "Show me all Discord collections side-by-side with fork/view counts"

## Table Stakes
- Full-text search across collections, workspaces, APIs, flows, teams
- Category-based browsing (AI, DevOps, Payments, Communication, etc.)
- Sort by popularity, recency, weekly/all-time metrics
- Pagination through large result sets
- Publisher verification status
- Fork/view/watch metrics for popularity assessment
- Direct URLs to open collections in Postman

## Data Layer
- **Primary entities:** NetworkEntity (collection, workspace, api, flow), Category, Team, SearchResult
- **Sync cursor:** offset-based pagination for browse, from-based for search
- **FTS/search:** Full-text search via `search-all` POST endpoint with queryIndices filtering
- **Key metrics:** forkCount, viewCount, watchCount (total + monthly + weekly variants)

## Product Thesis
- **Name:** postman-explore-pp-cli
- **Why it should exist:** Postman Explore is the world's largest public API collection directory (700K+ collections) but has NO CLI or API. Developers who want to search/browse collections from the terminal — especially AI agents building API integrations — have zero tooling. The Postman CLI (newman) only *runs* collections, it can't *discover* them. The Postman MCP server manages your own workspace but can't search the public network. This CLI fills a gap that literally nothing else covers.

## Build Priorities
1. **Search** — `search "stripe payment" --type collection --limit 20 --json`
2. **Browse** — `browse collections --sort popular --category ai --limit 50`
3. **Categories** — `categories` to list, `category devops` for detail
4. **Teams** — `teams --sort popular` to discover publishers
5. **Stats** — `stats` for network-wide counts
6. **Store** — SQLite persistence for offline search, trending analysis, history
7. **Trending** — `trending --period week` (compare weekly vs all-time fork/view counts)
8. **Compare** — `compare "stripe" "braintree"` — side-by-side collection metrics

## Discovered API Endpoints

All via `POST https://www.postman.com/_api/ws/proxy` with body `{service, method, path, body?}`:

| Service | Internal Method | Path | Purpose |
|---------|----------------|------|---------|
| publishing | GET | `/v1/api/networkentity?entityType=<type>&limit=<n>&offset=<n>&sort=<sort>&categoryId=<id>` | Browse entities |
| publishing | GET | `/v1/api/networkentity/count?flattenAPIVersions=true` | Entity counts |
| publishing | GET | `/v2/api/category?sort=spotlighted` | List categories |
| publishing | GET | `/v2/api/category/<slug>` | Category detail |
| publishing | GET | `/v1/api/team?limit=<n>&sort=popular` | List publisher teams |
| search | POST | `/search-all` | Full-text search (body: queryText, size, from, domain, queryIndices) |
| notebook | GET | `/notebooks/count` | Notebook counts |

### Entity Types for Browse
- `collection` — Postman collections (most common, 705K+)
- `workspace` — Public workspaces (309K+)
- `api` — Published APIs (25K+)
- `flow` — Postman Flows (6K+)

### Query Indices for Search
- `runtime.collection` — Collections
- `collaboration.workspace` — Workspaces
- `runtime.request` — Individual requests
- `flow.flow` — Flows
- `apinetwork.team` — Publisher teams

### Sort Options for Browse
popular, recent, featured, new, week, alltime

### Known Categories (by ID)
1=Artificial Intelligence, 2=Communication, 3=Data Analytics, 4=Developer Productivity, 5=DevOps, 6=Financial Services, 7=Payments, 8=App Security, 9=Database, 10=Travel, 11=eSignature

## Proxy Pattern Note
The CLI must implement a custom HTTP client that wraps all requests as POST to the proxy URL
with the `{service, method, path, body?}` envelope. This is non-standard and cannot use a
typical REST client generator. The generator should produce the proxy wrapper and all commands
should use it.
