# Allrecipes Absorb Manifest

## Tools Surveyed

| Tool | Type | Surface |
|------|------|---------|
| `hhursev/recipe-scrapers` (Python) | community parser, JSON-LD focused | 25+ AbstractScraper method slots |
| `remaudcorentin-dev/python-allrecipes` | wrapper (search + get) | 2 public methods |
| `ryojp/recipe-scraper` (Go, Colly) | full HTML+JSON-LD scraper | full Recipe struct + CLI |
| `jadkins89/Recipe-Scraper` (JS) | wrapper | name, ingredients, instructions, tags, servings, image, time |
| `marcon29/CLI-dinner-finder-grocery-list` (Ruby CLI) | interactive CLI | browse → select → grocery list |
| `cookbrite/Recipe-to-Markdown` (Python) | exporter | recipe → markdown |
| `Apify Allrecipes Advanced Scraper` | commercial | recipes + articles + galleries + reviews + author + collections |
| `Apify Allrecipes Search Scraper`, `Review Scraper` | commercial | search + reviews |

## Absorbed (match or beat everything)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-------------|--------------------|-------------|
| 1 | Search recipes by query | `python-allrecipes`, `Apify` | `search <query> --limit --page --json --select` | offline-cached results, ranking layer, agent-native |
| 2 | Get full recipe detail | `recipe-scrapers` allrecipes.py + AbstractScraper JSON-LD | `recipe <url-or-id> --json` | every Schema.org Recipe field; cache-on-fetch |
| 3 | Recipe ingredients (qty/unit/name) | `ryojp/recipe-scraper`, `recipe-scrapers` | `recipe <url> --json --select recipeIngredient` | structured (parsed qty+unit+name) |
| 4 | Recipe instructions (steps) | All | `recipe <url> --json --select recipeInstructions` or `instructions <url>` | `--markdown` step-numbered output |
| 5 | Recipe nutrition (cal/fat/carb/protein/sodium) | `recipe-scrapers` (nutrients()), `ryojp`, `Apify` | inside `recipe --json` and dedicated `nutrition <url>` | per-serving + scaled |
| 6 | Recipe rating + review count | All | inside `recipe --json` (`aggregateRating`) | feeds Bayesian-smoothed `top-rated` |
| 7 | Top reviews text | `Apify`, `Allrecipes Review Scraper` | `reviews <url> --limit N` | unauth-safe, JSON-clean |
| 8 | Made-It count | Allrecipes-specific (no other tool surfaces this) | extracted in `recipe --json` (per-recipe field) | unique competitive metric, surfaced as data |
| 9 | Browse by category (dish type) | `Apify`, `dinner-finder` | `category <slug> --top-rated --quick` | offline-rankable |
| 10 | Browse by cuisine | `Apify` | `cuisine <slug>` | same |
| 11 | Browse by main ingredient | `dinner-finder` | `ingredient <name>` | non-interactive (dinner-finder is interactive only) |
| 12 | Filter by dietary restriction | `recipe-scrapers` (dietary_restrictions()) | `dietary --type <gluten-free\|vegan\|low-carb>` | combines JSON-LD keywords + ingredient match |
| 13 | Browse by occasion (weeknight, holiday) | `Apify` | `occasion <slug>` | leverages JSON-LD keywords |
| 14 | Article extraction | `Apify` only (paid) | `article <url> --json` | shipped free |
| 15 | Gallery extraction (round-up posts) | `Apify` only (paid) | `gallery <url> --json` → recipes inside | shipped free |
| 16 | Cook profile | `Apify` only | `cook <slug>` (best-effort from public profile page) | none of the open-source tools have this |
| 17 | Markdown export | `cookbrite/Recipe-to-Markdown` (basic) | `recipe <url> --markdown` and `export <url>` | clean agent-readable markdown with attribution |
| 18 | Recipe scaling | NONE in any tool | `scale <url> --servings N` | math-aware fraction handling |
| 19 | Grocery list (multi-recipe) | `dinner-finder` (interactive Ruby) | `grocery-list <urls...> --json` | non-interactive, agent-callable, dedup'd |
| 20 | Sort by rating/popularity | All scrapers expose rating; sorting client-side | `top-rated <query-or-category>` | Bayesian smoothing kills the 5-star/1-review outlier |
| 21 | Equipment list | `recipe-scrapers` (equipment()) | inside `recipe --json` | rare metadata |
| 22 | Image URL | All | `recipe --json` includes `image` | enables agent multimodal use |
| 23 | Author / source attribution | `recipe-scrapers`, `ryojp` | inside `recipe --json` and on markdown export | proper attribution |
| 24 | Cooking method | `recipe-scrapers` (cooking_method()) | inside `recipe --json` | filterable for `--method bake` |
| 25 | Keywords (free tags) | `recipe-scrapers` (keywords()) | inside `recipe --json`, FTS-indexed | enables tag-based search |
| 26 | Time-capped search | NONE (Allrecipes UI cannot strict-cap) | `--max-minutes N` on `search`/`category`/`quick` | offline filter on cached total_time |
| 27 | Pagination | `python-allrecipes` page param | `search --page N --limit M` | consistent across every list command |
| 28 | Description / summary | `recipe-scrapers` (description()) | inside `recipe --json` | summary text for previews |

**No stubs.** Every absorbed row is shipping scope.

## Transcendence (only possible with our approach)

| # | Feature | Command | Why Only We Can Do This | Score |
|---|---------|---------|-------------------------|-------|
| 1 | Pantry match | `pantry --pantry-file pantry.txt --query brownies` | Requires SQL JOIN over cached recipes × ingredient names with token overlap scoring — pure local-store feature, not on the website | 9/10 |
| 2 | Bayesian top-rated | `top-rated <query> --smooth-c 200` | Allrecipes "sort by rating" surfaces 5-star/1-review noise; Bayesian smoothing toward prior `mean=4.0, C=200` returns proven-popular recipes — algorithm transparency the website doesn't expose | 9/10 |
| 3 | Reverse ingredient index | `with-ingredient buttermilk --top 10` | Native search is title-only; reverse-index lookup is a SQL feature only the cache enables | 8/10 |
| 4 | Quick weeknight | `quick --max-minutes 30 --top` | Allrecipes UI has no strict time cap on search; offline filter on cached total_time + Bayesian rating | 8/10 |
| 5 | Cookbook export | `cookbook --category italian --top 20 --output italian.md` | Compounds category browse + top-rated + markdown export → a personal cookbook bundle, no tool ships this | 8/10 |
| 6 | Dietary filter on cache | `dietary --type gluten-free --top 20` | Allrecipes' diet pages are incomplete; we combine JSON-LD `keywords` + ingredient-name pattern matching across the local corpus | 7/10 |
| 7 | Cache introspection | `cache list \| cache stats \| cache clear` | Required local-data feature; powers offline reproducibility | 5/10 |
| 8 | Doctor with Cloudflare diagnosis | `doctor` | Names "Cloudflare interstitial" by inspecting body for "Just a moment" + advises browser-chrome transport — domain-specific reachability check | 6/10 |

8 transcendence features, all scoring ≥ 5/10. **No stubs.** (User cut Made-It leaderboard and Since-cached during the Phase Gate 1.5 review.)

## Total Scope

- 28 absorbed + 8 transcendence = **36 features**
- Best existing tool: marcon29/CLI-dinner-finder-grocery-list (~4 commands) → **9.5× feature count**
- Most-comprehensive parser: hhursev/recipe-scrapers (25+ JSON-LD methods) → matched 1:1 + agent-native + cached + searchable

## User-First Personas (informs the build emphasis)

1. **The agentic shopper** — asks an LLM to plan a week of dinners and produce one grocery list. Needs: composable JSON, `grocery-list`, `--select`, `dietary`.
2. **The weeknight cook** — has 30 minutes, wants proven recipes. Needs: `quick --max-minutes 30 --top-rated`, no scrolling.
3. **The pantry-driven cook** — has chicken thighs + lemons. Needs: `pantry`, `with-ingredient`.
4. **The cookbook builder** — wants a personal Italian top-20 cookbook in markdown. Needs: `cookbook --category italian --output italian.md`.

## Reachability Note

`http_transport: browser-chrome` is mandatory. Direct curl returns Cloudflare 403 ("Just a moment..."); Surf TLS-impersonated transport bypasses it. This is the same approach recipe-goat uses successfully against allrecipes.com today. The `doctor` command should detect and name the Cloudflare interstitial when the transport is misconfigured.
