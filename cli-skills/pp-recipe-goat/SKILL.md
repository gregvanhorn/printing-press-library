---
name: pp-recipe-goat
description: "Find the best version of any recipe across 37 trusted sites — then plan, shop, and cook with a local kitchen companion. Trigger phrases: `find me the best recipe for`, `what can I make with`, `substitute for`, `set up a meal plan`, `shopping list from my meal plan`, `tonight's dinner`, `use recipe-goat`, `run recipe-goat`."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["recipe-goat-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/food-and-dining/recipe-goat/cmd/recipe-goat-pp-cli@latest","bins":["recipe-goat-pp-cli"],"label":"Install via go install"}]}}'
---

# Recipe Goat — Printing Press CLI

Recipe GOAT aggregates 37 of the web's most trusted recipe sites (King Arthur, Serious Eats, Budget Bytes, Smitten Kitchen, Food52, AllRecipes, Food Network, Simply Recipes, EatingWell, BBC Good Food, Bon Appétit, Epicurious, and 25 more), ranks results by merged trust + rating + review-count signals, and builds a local SQLite cookbook that powers pantry match, cook log, meal plans, and aisle-grouped shopping lists. Unique commands like `goat` (best-version ranker), `sub` (cross-site substitution lookup), `tonight` (decision-fatigue killer), and `cookbook match --have` (pantry match) solve problems no single recipe site can.

## When to Use This CLI

Reach for Recipe GOAT when the user needs the best version of a dish without site-hopping, wants to query their own cookbook by pantry contents, needs a shopping list for a meal plan, or wants ingredient substitution lookups. The ranking weights real reader signal (rating × log(reviews)) at 80% and editorial site-trust at 15%, so tie-break favors curated chef/baker sites over crowdsourced aggregators like AllRecipes.

## When Not to Use This CLI

This CLI does not change remote state. Do not use it to order groceries (use Instacart-pp-cli), order delivery (use Domino's-pp-cli or another delivery CLI), book restaurant reservations, send recipe links by email or message, or post to social platforms. Recipe GOAT does build local SQLite state — saved cookbook, tags, cook log, meal plan slots — and supports `import` (POSTs to USDA endpoints), so local-side mutation is fully expected.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Cross-site intelligence

- **`goat`** — Query any dish across 37 recipe sites and rank results by normalized rating × review count × author trust × site trust × recency.

  _Use this when you need the single best version of a dish — the agent gets structured results with provenance and trust signals instead of guessing from a web search._

  ```bash
  recipe-goat-pp-cli goat "chicken tikka masala" --limit 5 --json
  ```
- **`sub`** — Curated ingredient-substitution table sourced from King Arthur, Serious Eats, Budget Bytes, Minimalist Baker, and AllRecipes community reviews. Hand-curated and shipped with the binary (no live fetching at query time); ranked by source trust with ratios and context filters.

  _When a recipe needs a sub, agents can pick the best one given the cooking context (baking vs marinade) instead of suggesting the first hit on Google._

  ```bash
  recipe-goat-pp-cli sub buttermilk --context baking
  ```
- **`recipe reviews`** *(planned, work-in-progress — emits a stub today)* — Will surface the top modifications cooks made to a recipe once the source-review fetcher is wired. Today the command returns a clearly-labeled placeholder so agents don't depend on it.

  ```bash
  recipe-goat-pp-cli recipe reviews <id> --limit 10  # emits stub message in v1
  ```
- **`recipe get --nutrition`** — When a site omits nutrition, parse ingredients, match USDA FoodData Central IDs, compute per-serving macros locally.

  _Agents can answer 'is this recipe high-protein?' reliably even when the source doesn't publish macros._

  ```bash
  recipe-goat-pp-cli recipe get https://www.budgetbytes.com/creamy-mushroom-pasta/ --nutrition
  ```
- **`recipe get`** — Flag out-of-season ingredients inline ("⚠ asparagus is out of season in November — peak April–June") and suggest in-season swaps.

  _Agents surface cost + quality signals the user wouldn't otherwise see._

  ```bash
  recipe-goat-pp-cli recipe get <url>  # seasonal flag appears automatically
  ```

### Local state that compounds

- **`cookbook match`** — Find recipes in the local cookbook that you can make right now with listed ingredients, or with ≤N missing ingredients.

  _When the user says 'what can I make with what's in my fridge,' the agent gets ranked candidates with missing-ingredient counts instead of guessing._

  ```bash
  recipe-goat-pp-cli cookbook match --have "chicken,rice,broccoli" --missing-max 2
  ```
- **`tonight`** — Pick dinner in 2 seconds: filter cookbook by time budget, recency from cook log, and dietary/kid-friendly flags.

  _Ends the 20-minute 'what are we having' debate with three data-backed candidates._

  ```bash
  recipe-goat-pp-cli tonight --max-time 30m --no-repeat-within 7d --kid-friendly
  ```
- **`search --kid-friendly`** — Filter recipes against an editable ingredient-exclusion list (capers, anchovies, excess heat, raw fish, etc.). Personalizable per-child.

  _Parents get results actually edible by their kids, not marketing's idea of 'kid-friendly'._

  ```bash
  recipe-goat-pp-cli search "chicken dinner" --kid-friendly --limit 10
  ```
- **`meal-plan shopping-list`** — Aggregate ingredients across planned meals, reconcile units (2 cup + 1 cup milk → 3 cup), group by grocery aisle.

  _The agent hands the user a complete shopping list ready for grocery day, aisle-grouped._

  ```bash
  recipe-goat-pp-cli meal-plan shopping-list --week --export md
  ```
- **`recipe cost`** *(approximate, work-in-progress — placeholder heuristic only)* — Will eventually estimate cost per serving from Budget Bytes line-item data plus USDA retail averages. Today the command emits a clearly-labeled stub with a note that ingredient-price data integration is not yet wired.

  ```bash
  recipe-goat-pp-cli recipe cost <id>  # emits placeholder + wip note in v1
  ```

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Command Reference

**foods** — USDA FoodData Central — ingredient nutrition lookups

- `recipe-goat-pp-cli foods get` — Get a specific food by FDC ID
- `recipe-goat-pp-cli foods list` — List foods paginated
- `recipe-goat-pp-cli foods search` — Search USDA FoodData Central for foods matching a query


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
recipe-goat-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes


### Find the best version of a dish

```bash
recipe-goat-pp-cli goat "carbonara" --limit 5
```

Fetches each site's listing for the query, validates JSON-LD on each candidate, and ranks survivors by `0.55 × rating + 0.25 × log(reviews+1)/log(1000) + 0.15 × site_trust + 0.05 × recency`. Rating/reviews come from each page's Schema.org `aggregateRating`. `site_trust` is hand-curated 0.70–0.95 (editorial chef/baker sites 0.9–0.95; crowdsourced aggregators 0.70–0.75). Two trust-aware adjustments run before scoring: curated sites with no Schema.org rating get an imputed 4.5/100 baseline (editorial vetting ≈ 100 implicit favorable reviews); aggregator-site ratings are Bayesian-smoothed toward 4.0 with credibility C=200 (so a 5.0/100 from AllRecipes effectively becomes 4.33, while a 4.7/5000 stays 4.67). Net effect: niche curated recipes with no ratings can outrank mid-tier AllRecipes results; heavily-reviewed AllRecipes blockbusters still rank where their reader signal earns them.

### Save and tag for weeknight

```bash
recipe-goat-pp-cli save <url> --tags weeknight,pasta
```

Persist the recipe locally with tags you'll filter on later.

### Plan 5 dinners with kid-safe ingredients

```bash
recipe-goat-pp-cli tonight --max-time 30m --kid-friendly --limit 5
```

Three candidates from your cookbook that match time and ingredient constraints.

### Aggregate shopping list for the week

```bash
recipe-goat-pp-cli meal-plan shopping-list --week --aisle
```

Shopping list grouped by grocery aisle with unit reconciliation.

### Substitute an out-of-stock ingredient

```bash
recipe-goat-pp-cli sub eggs --context baking --vegan
```

Context-aware substitutions ranked by source trust.

## Auth Setup

USDA FoodData Central (free, 3,500 req/hr) enables nutrition backfill when a recipe site omits macros. Get a key at https://fdc.nal.usda.gov/api-key-signup and export USDA_FDC_API_KEY. All other features work without any authentication.

Run `recipe-goat-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  recipe-goat-pp-cli foods list --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal — piped/agent consumers get pure JSON on stdout.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
recipe-goat-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
recipe-goat-pp-cli feedback --stdin < notes.txt
recipe-goat-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.recipe-goat-pp-cli/feedback.jsonl`. They are never POSTed unless `RECIPE_GOAT_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `RECIPE_GOAT_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
recipe-goat-pp-cli profile save briefing --json
recipe-goat-pp-cli --profile briefing foods list
recipe-goat-pp-cli profile list --json
recipe-goat-pp-cli profile show briefing
recipe-goat-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `recipe-goat-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.25+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/food-and-dining/recipe-goat/cmd/recipe-goat-pp-cli@latest
   
   # If `@latest` installs a stale build (Go module proxy cache lag), install from main:
   GOPRIVATE='github.com/mvanhorn/*' GOFLAGS=-mod=mod \
     go install github.com/mvanhorn/printing-press-library/library/food-and-dining/recipe-goat/cmd/recipe-goat-pp-cli@main
   ```
3. Verify: `recipe-goat-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/food-and-dining/recipe-goat/cmd/recipe-goat-pp-mcp@latest
   
   # If `@latest` installs a stale build (Go module proxy cache lag), install from main:
   GOPRIVATE='github.com/mvanhorn/*' GOFLAGS=-mod=mod \
     go install github.com/mvanhorn/printing-press-library/library/food-and-dining/recipe-goat/cmd/recipe-goat-pp-mcp@main
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add recipe-goat-pp-mcp -- recipe-goat-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which recipe-goat-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   recipe-goat-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `recipe-goat-pp-cli <command> --help`.
