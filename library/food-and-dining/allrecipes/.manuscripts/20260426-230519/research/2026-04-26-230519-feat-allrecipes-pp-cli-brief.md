# Allrecipes CLI Brief

## API Identity
- **Domain**: allrecipes.com — Dotdash Meredith property, Cloudflare-fronted recipe website. No official public API.
- **Users**: home cooks ("what should I make with chicken thighs?"), meal planners (week of dinners + grocery list), dietary-restricted cooks (gluten-free, low-carb), agentic shoppers asking an LLM to plan a meal.
- **Data profile**: ~250k recipes with full Schema.org Recipe JSON-LD on every recipe page (title, ingredients with quantity+unit+name, instructions, prep/cook/total time, yield, nutrition, ratings, review counts, "Made It" counts, author, images, cuisine, category, keywords). Rich review text per recipe. Categories: dish type, cuisine, ingredient, occasion, diet, holiday.

## Reachability Risk
- **HIGH** — direct curl with a default UA returns HTTP 403 with a Cloudflare "Just a moment..." interstitial. TLS-fingerprint detection.
- **Mitigation**: `http_transport: browser-chrome` (Surf, Chrome-impersonated TLS). recipe-goat already uses this successfully against allrecipes.com, food52.com, and other Dotdash properties — the spec.yaml even names AllRecipes in the transport comment.
- **Evidence**: probe at `https://www.allrecipes.com/search?q=brownies` → `HTTP=403`, body `<title>Just a moment...</title>` with `cloudflare` and `challenges.cloudflare.com` references. recipe-goat ships and works.

## Top Workflows
1. **Find a recipe** — `search "brownies"`, sorted by rating × review count, optionally constrained by time/cuisine/diet.
2. **Get the full recipe as data** — `recipe https://www.allrecipes.com/recipe/10813/best-brownies/` returns ingredients/instructions/times/nutrition/rating as parsed JSON, not HTML.
3. **Build a grocery list from a meal plan** — pick N recipes → aggregate and de-dup ingredient quantities into a single shopping list.
4. **Scale a recipe** — change servings; quantities and (best-effort) units rescale.
5. **Browse what's good in a category/cuisine** — top-rated Italian, top-rated weeknight, top-rated under 30 min.
6. **Export to markdown / shareable file** — agent writes a recipe to a file in clean markdown for cooking-mode reading.

## Table Stakes (incumbent features we must match)
- Search (`q`, pagination), category browse, cuisine browse, ingredient browse.
- Full recipe extraction: title, ingredients (qty + unit + name), instructions, prep/cook/total time, servings/yield, author, image, nutrition, rating, review count, "Made It" count.
- Reviews list (top reviews, recency, photos count).
- JSON output everywhere; markdown export.
- Robust to Cloudflare TLS challenge.

## Data Layer
- **Primary entities**: `recipes`, `ingredients` (per-recipe rows: qty, unit, name, raw_text), `reviews`, `categories`, `cuisines`, `nutrition`.
- **Sync cursor**: not strictly needed (no auth, no per-user state). But a local cache keyed by recipe ID + URL is essential — every successful fetch should populate the cache so search/scale/export work offline thereafter.
- **FTS**: title, ingredient names, cuisine, category, keywords from JSON-LD. Enables transcendence features (pantry-match, dietary-filter, swap-aware search).

## Codebase Intelligence
- **recipe-scrapers (hhursev/recipe-scrapers)** — `allrecipes.py` is intentionally minimal; inherits `AbstractScraper` which uses JSON-LD with HTML selector fallbacks. Field set: title, ingredients, instructions, prep/cook/total time, yields, image, language, nutrients, author, ratings, review_count. **JSON-LD is the primary surface.**
- **remaudcorentin-dev/python-allrecipes** — surfaces `AllRecipes.search(query)` (returns name/url/image/rating per result) and `AllRecipes.get(url)` (returns name/rating/ingredients/steps/prep_time/cook_time/total_time/nb_servings). README notes "search only supports text — other options are not available on allrecipes website anymore." So: search is text-only; filtering happens client-side.
- **ryojp/recipe-scraper (Go, Colly)** — full-coverage Go scraper. Extracts title, summary, url, image, author, ingredients (qty+unit+name), directions, prep/cook/total times, servings, rating, review count, photo count, calories/fat/carbs/protein.
- **marcon29/CLI-dinner-finder-grocery-list (Ruby CLI)** — only known CLI tool. Browse by ingredient → select recipes → aggregate grocery list. This is the closest CLI competitor.
- **Apify Allrecipes Data Extractor** — paid commercial scraper covering recipes, articles, galleries, reviews. Scope hints: there's value beyond recipes (articles, photo galleries, holiday round-ups).
- **Auth patterns**: none — the JSON-LD on public pages doesn't require auth. Authenticated features (saved recipes, meal plans, profile) are explicitly out of scope per user instruction.

## User Vision
> "do it for unauthenticated scenarios only. do not require authentication"

The CLI ships with no `auth` subcommand, no token storage, no login flow. Every command works against public pages.

## Product Thesis
- **Name**: Allrecipes Pocket (binary `allrecipes-pp-cli`)
- **Why it should exist**: Allrecipes is the largest crowd-rated recipe corpus on the web — 250k+ recipes, thousands of "Made It!" datapoints per popular recipe — but the website is ad-heavy, story-driven, and slow on mobile. Power users want recipes-as-data: search, scale, shop, export, filter by ingredient, all from the terminal or their agent. JSON-LD makes the data clean; SQLite makes it fast and composable; a TLS-impersonated transport bypasses the bot wall. Nothing in the existing tool zoo combines these — Python wrappers don't cache, recipe-scrapers doesn't index for search, Ruby dinner-finder is interactive only. The differentiator is "every Allrecipes recipe you've ever fetched is now in your local store, queryable in SQL, scalable, and shoppable."

## Build Priorities
1. **Foundation**: shared transport (Surf/`browser-chrome`), JSON-LD extractor lifted from recipe-goat's `recipes/jsonld.go` shape, SQLite schema for recipes/ingredients/reviews/categories.
2. **Absorbed (P1)**: `search`, `recipe`, `category`, `cuisine`, `reviews`, `export` (markdown), `nutrition`, `scale`, `grocery-list` from N recipes.
3. **Transcendence (P2)**: `pantry` (match recipes against a pantry CSV), `quick` (recipes under N minutes from cache), `cookbook` (export categorized markdown bundle), `made-it` (leaderboard by Made-It count), `dietary` (gluten-free / vegetarian / low-carb filter on cached corpus), `top-rated` (Bayesian-smoothed ranking — hide 1-review 5-star outliers), `ingredients` (which recipes use this ingredient — reverse index), `swap` (suggest recipes similar to one you have but with one ingredient swapped), `since` (recipes added to local store in the last N days), `cache` (introspect/dedupe local cache).
4. **Polish (P3)**: SKILL recipes pairing `--agent` and `--select`, README cookbook section, doctor that names "Cloudflare interstitial" failure mode.
