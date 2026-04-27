# Allrecipes CLI

**Every Allrecipes recipe in your terminal — cached as data, with pantry-aware search, Bayesian-smoothed ranking, and one-line grocery lists.**

Search Allrecipes' 250k-recipe corpus from the command line, fetch a full recipe as parsed JSON-LD (ingredients with quantity+unit+name, instructions, nutrition, ratings, Made-It count), aggregate grocery lists from a meal plan, scale recipes, and export to clean markdown. Every recipe you fetch lands in a local SQLite store, which unlocks `pantry` (which recipes can I cook with what I have), `with-ingredient` (reverse index), `top-rated` with Bayesian smoothing (no more 1-review 5-star noise), and `cookbook` (export a category as a personal cookbook). Ships with a Chrome-impersonated TLS transport that walks past Cloudflare.

Learn more at [Allrecipes](https://www.allrecipes.com).

## Install

### Go

```
go install github.com/mvanhorn/printing-press-library/library/food-and-dining/allrecipes/cmd/allrecipes-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Authentication

No authentication required — every command works against public Allrecipes pages. The CLI stores no tokens and asks for no credentials.

## Quick Start

```bash
# Search the live site; results land in your local cache for offline reuse.
allrecipes-pp-cli search "brownies" --limit 5 --agent


# Fetch a full recipe as parsed JSON-LD; --select narrows the payload.
allrecipes-pp-cli recipe https://www.allrecipes.com/recipe/9599/quick-and-easy-brownies/ --agent --select recipeIngredient,totalTime,aggregateRating


# Rescale ingredients by servings.
allrecipes-pp-cli scale https://www.allrecipes.com/recipe/9599/quick-and-easy-brownies/ --servings 16


# Aggregate a multi-recipe shopping list.
allrecipes-pp-cli grocery-list https://www.allrecipes.com/recipe/9599/quick-and-easy-brownies/ https://www.allrecipes.com/recipe/16354/easy-meatloaf/ --agent


# Match cached recipes against what you already have.
allrecipes-pp-cli pantry --pantry-file ~/pantry.txt --query chicken --agent

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Local state that compounds

- **`pantry`** — Score Allrecipes recipes against your pantry — see which ones you can actually cook tonight without a grocery run.

  _When the user says 'what can I make with what I've got', this is the only command that knows the answer._

  ```bash
  allrecipes-pp-cli pantry --pantry-file ~/pantry.txt --query brownies --agent
  ```
- **`with-ingredient`** — Find every cached recipe that uses a given ingredient — a SQL view across your local corpus.

  _Use this when the user starts from an ingredient they want to use up, not a dish name._

  ```bash
  allrecipes-pp-cli with-ingredient buttermilk --top 10 --agent
  ```
- **`dietary`** — Filter cached recipes by gluten-free / vegan / low-carb using JSON-LD keywords plus ingredient-name patterns.

  _Use this when dietary restrictions are non-negotiable and the user wants more than what the site's diet category page surfaces._

  ```bash
  allrecipes-pp-cli dietary --type gluten-free --top 20 --agent
  ```

### Ranking that beats the website

- **`top-rated`** — Rank recipes by Bayesian-smoothed rating — proven popular wins over 1-review 5-star noise.

  _Pick this over raw search when the agent wants proven recipes, not freshly-uploaded 5-star outliers._

  ```bash
  allrecipes-pp-cli top-rated brownies --enrich --smooth-c 200 --limit 10 --agent
  ```
- **`quick`** — Top-rated recipes that fit a strict time cap — Allrecipes' UI cannot enforce one, but the local cache can.

  _Use this when the user's constraint is time, not dish — 'what can I make in 25 minutes that's actually good'._

  ```bash
  allrecipes-pp-cli quick --max-minutes 30 --query chicken --agent
  ```

### Agent-native plumbing

- **`cookbook`** — Compile a top-rated category into a single markdown cookbook with TOC, ingredients, and instructions.

  _When the user asks for a curated bundle (gifts, meal-plan packs), this builds it in one command._

  ```bash
  allrecipes-pp-cli cookbook --category italian --top 20 --output italian-cookbook.md
  ```
- **`grocery-list`** — Aggregate ingredients from many recipes into a deduped, agent-readable shopping list.

  _Use this at the end of a meal plan — one call replaces five scrolls through ingredient lists._

  ```bash
  allrecipes-pp-cli grocery-list https://www.allrecipes.com/recipe/9599/quick-and-easy-brownies/ https://www.allrecipes.com/recipe/16354/easy-meatloaf/ --agent
  ```

### Reachability mitigation

- **`doctor`** — Health check that names the Cloudflare 'Just a moment...' interstitial by inspecting the response body, then advises the browser-chrome transport.

  _When the CLI breaks because of bot detection, the agent gets a specific, actionable error rather than a generic timeout._

  ```bash
  allrecipes-pp-cli doctor
  ```

## Usage

Run `allrecipes-pp-cli --help` for the full command reference and flag list.

## Commands

### recipes

Public Allrecipes recipe pages with Schema.org Recipe JSON-LD markup

- **`allrecipes-pp-cli recipes get`** - Fetch a recipe by ID + slug; returns parsed JSON-LD Recipe
- **`allrecipes-pp-cli recipes search`** - Search Allrecipes for recipes matching a query


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
allrecipes-pp-cli recipes get

# JSON for scripting and agents
allrecipes-pp-cli recipes get --json

# Filter to specific fields
allrecipes-pp-cli recipes get --json --select id,name,status

# Dry run — show the request without sending
allrecipes-pp-cli recipes get --dry-run

# Agent mode — JSON + compact + no prompts in one flag
allrecipes-pp-cli recipes get --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `5` API error, `7` rate limited, `10` config error.

## Freshness

This CLI owns bounded freshness for registered store-backed read command paths. In `--data-source auto` mode, covered commands check the local SQLite store before serving results; stale or missing resources trigger a bounded refresh, and refresh failures fall back to the existing local data with a warning. `--data-source local` never refreshes, and `--data-source live` reads the API without mutating the local store.

Set `ALLRECIPES_NO_AUTO_REFRESH=1` to disable the pre-read freshness hook while preserving the selected data source.

JSON outputs that use the generated provenance envelope include freshness metadata at `meta.freshness`. This metadata describes the freshness decision for the covered command path; it does not claim full historical backfill or API-specific enrichment.

## Use as MCP Server

This CLI ships a companion MCP server for use with Claude Desktop, Cursor, and other MCP-compatible tools.

### Claude Code

```bash
claude mcp add allrecipes allrecipes-pp-mcp
```

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "allrecipes": {
      "command": "allrecipes-pp-mcp"
    }
  }
}
```

## Cookbook

Worked examples that combine the CLI's primitives. Each recipe uses verified flag names — copy and run.

### Plan a weeknight dinner from a half-empty fridge

```bash
# 1. Score cached recipes against your pantry, keep ones that need ≤2 extras.
allrecipes-pp-cli pantry --pantry chicken,rice,onion,garlic,lemon \
    --max-missing 2 --min-overlap 0.6 --agent

# 2. Pick a recipe URL from the result and fetch the full ingredient list.
allrecipes-pp-cli recipe 9599/quick-and-easy-brownies --markdown
```

### Build a curated cookbook from a category

```bash
# Pre-warm the cache for a category so cookbook has data to draw from.
allrecipes-pp-cli category italian --limit 30 --agent | \
    jq -r '.[].url' | \
    xargs -I {} allrecipes-pp-cli recipe {} --agent > /dev/null

# Compile the top 20 cached italian recipes into a single markdown file.
allrecipes-pp-cli cookbook --cuisine italian --top 20 \
    --title "Italian Top 20" --output italian.md
```

### Pick the best recipe, not the loudest

```bash
# Bayesian-smoothed top 5 — needs --enrich to get accurate ratings.
allrecipes-pp-cli top-rated brownies --enrich --smooth-c 200 --limit 5 --agent
```

### Recipes that fit a strict time cap

```bash
# Allrecipes' UI cannot enforce numeric time caps; the local cache can.
allrecipes-pp-cli quick --max-minutes 25 --query chicken --agent
```

### Aggregate a multi-recipe shopping list

```bash
# Fetch each, parse ingredients, sum quantities, dedupe.
allrecipes-pp-cli grocery-list 9599 16354 23456 --output markdown
```

### Search offline against the local cache

```bash
# --cache-only skips the network and queries the FTS index directly.
allrecipes-pp-cli search "chicken thighs" --cache-only --limit 20 --agent
```

### Reverse-lookup recipes by an ingredient you want to use up

```bash
# Find every cached recipe that uses buttermilk.
allrecipes-pp-cli with-ingredient buttermilk --top 10 --agent
```

### Filter by dietary restriction

```bash
# Heuristic-based; not a substitute for reading the full ingredient list.
allrecipes-pp-cli dietary --type gluten-free --max-minutes 30 --top 20 --agent
```

### Scale a recipe for a dinner party

```bash
allrecipes-pp-cli scale 9599/quick-and-easy-brownies --servings 16 --agent
```

## Health Check

```bash
allrecipes-pp-cli doctor
```

Verifies configuration and connectivity to the API.

## Configuration

Config file: `~/.config/allrecipes-pp-cli/config.toml`

## Troubleshooting
**Not found errors (exit code 3)**
- Check that the recipe URL or ID is correct
- Run `allrecipes-pp-cli search <query>` to find a fresh URL

### API-specific

- **doctor reports 'Cloudflare interstitial detected'** — The default Surf transport should bypass this. If you're hitting it anyway, check that you're on the latest binary; older builds shipped with stdlib HTTP which Cloudflare blocks.
- **search returns 0 results** — Allrecipes search is text-only; filter syntax that worked on the website does not work in the URL anymore. Use plain words, then narrow with --max-minutes and --top-rated client-side.
- **recipe fetch returns 403** — Run `allrecipes-pp-cli doctor` to confirm Cloudflare bypass is working; if it isn't, your network may be on a flagged IP — try from a different connection.
- **grocery-list output has duplicate ingredients with different units** — Unit normalization is best-effort; pass --raw-quantities to see the source strings and aggregate manually for edge cases.

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**hhursev/recipe-scrapers**](https://github.com/hhursev/recipe-scrapers) — Python (1900 stars)
- [**jadkins89/Recipe-Scraper**](https://github.com/jadkins89/Recipe-Scraper) — JavaScript (100 stars)
- [**remaudcorentin-dev/python-allrecipes**](https://github.com/remaudcorentin-dev/python-allrecipes) — Python (60 stars)
- [**ryojp/recipe-scraper**](https://github.com/ryojp/recipe-scraper) — Go (30 stars)
- [**cookbrite/Recipe-to-Markdown**](https://github.com/cookbrite/Recipe-to-Markdown) — Python (20 stars)
- [**marcon29/CLI-dinner-finder-grocery-list**](https://github.com/marcon29/CLI-dinner-finder-grocery-list) — Ruby (5 stars)

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
