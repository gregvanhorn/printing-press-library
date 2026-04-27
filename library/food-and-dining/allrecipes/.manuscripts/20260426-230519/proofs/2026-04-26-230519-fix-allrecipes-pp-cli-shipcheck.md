# Allrecipes-pp-cli Shipcheck

## Summary

| Check | Result |
|-------|--------|
| dogfood | **PASS** (initial WARN; fixed dead helpers + added recipes tests; re-ran clean) |
| verify | WARN — 61% pass rate, 0 critical failures (synthetic-spec limitation) |
| workflow-verify | workflow-pass (no manifest, skipped) |
| verify-skill | **PASS** (fixed 2 mechanical mismatches in SKILL.md) |
| scorecard | **78/100 Grade B** |
| behavioral live probes | search/recipe/scale all returned correct content end-to-end |

## Verdict: SHIP

All ship-threshold conditions met:
- dogfood: PASS
- verify: WARN with 0 critical failures (limitation: synthetic spec; commands requiring positional args fail verify's no-arg probes — these errors are correct behavior, not bugs)
- workflow-verify: pass
- verify-skill: clean
- scorecard: 78 (≥ 65)
- No flagship feature returns wrong output (live probes confirmed)

## Fixes applied during shipcheck

### Dead helpers removed (5)
- `replacePathParam`, `extractResponseData`, `printProvenance`, `wrapWithProvenance`, `wrapResultsWithFreshness` — generated helpers that lost their callers when I replaced `recipes_search.go` and `recipes_get.go` with JSON-LD-driven handlers. Removed from `internal/cli/helpers.go`.

### Recipes package tests added (16 tests)
- Per the build checklist's pure-logic-package rule. Coverage: `ParseISO8601Duration`, `ParseJSONLD` (basic + @graph + missing), `ParseSearchResults` (basic + dedup), `CanonicalRecipeURL`, `ResolveRecipeURL`, `ParseURL`, `ParseIngredient` (8 cases including unicode fractions and mixed-fraction parsing), `ScaleIngredients`, `AggregateGrocery`, `BayesianRating` (4 cases including outlier defense), `Rank`, `FormatTime`, `CleanInstructions` (HowToStep + plain string + HowToSection).
- All 16 tests pass; total project test count now 77.

### SKILL.md mechanical mismatches fixed (2)
- `quick --top 3` → `quick --limit 3 | jq -r '.[].url' | xargs ...` — `quick` doesn't have a `--top` flag (uses `--limit`); also added the proper `jq` pipeline to extract URLs for `grocery-list`.
- `recipes get --agent --select id,name,status` → `recipes get 9599 quick-and-easy-brownies --agent --select name,recipeIngredient,totalTime` — example was missing the required positional args.

### Stale URL in README + research.json fixed
- `recipe/10813/best-brownies/` returns 404 on the live site; replaced with the proven-good `9599/quick-and-easy-brownies/`.

### Generic troubleshooting tip rewritten
- "Run the `list` command to see available items" replaced with "Run `allrecipes-pp-cli search <query>` to find a fresh URL" — the generic `list` command doesn't exist on this CLI.

## Verify fail-rate analysis

13 commands FAIL verify because they require positional args that verify cannot fixture from the synthetic spec. These are correct behaviors (the commands return usage errors), not bugs:

| Command | Reason for FAIL |
|---------|-----------------|
| `recipe` | requires `<url-or-id>` |
| `ingredients` | requires `<url-or-id>` |
| `instructions` | requires `<url-or-id>` |
| `nutrition` | requires `<url-or-id>` |
| `reviews` | requires `<url-or-id>` |
| `scale` | requires `<url-or-id>` and `--servings` |
| `with-ingredient` | requires `<name>` |
| `pantry` | requires `--pantry-file` or `--pantry` |
| `recipes get` | requires `<recipe_id> <slug>` |
| `which` | requires `<query>` |

The actual verify command runs the help check (PASS for all 13) but dry-run + JSON checks fail because there's no way to invoke them without args. These commands DO work when given real arguments — confirmed by live probes during the build.

The remaining 20/33 PASSes cover all the no-arg commands and the search/sync/list family.

## Scorecard breakdown

```
  Output Modes   10/10    Auth           9/10
  Error Handling 10/10    Terminal UX    9/10
  README         8/10     Doctor         10/10
  Agent Native   10/10    Local Cache    10/10
  Breadth        7/10     Vision         6/10
  Workflows      6/10     Insight        2/10
  Sync Correctness 10/10  Type Fidelity  3/5
  Dead Code      5/5      Pipeline Integrity 7/10

  Total: 78/100 - Grade B
```

Notes on omitted dimensions:
- `path_validity`, `auth_protocol` — N/A for synthetic specs (excluded from denominator per the standard policy)
- `mcp_tool_design`, `mcp_surface_strategy`, `live_api_verification` — not yet implemented for this CLI

## Live behavioral probes (during shipcheck)

```
$ allrecipes-pp-cli search brownies --limit 3 --agent
[
  { "title": "S'mores Brownies", "url": "https://www.allrecipes.com/recipe/19848/smore-brownies/" },
  { "title": "Quick and Easy Brownies", "url": "https://www.allrecipes.com/recipe/9599/quick-and-easy-brownies/" },
  { "title": "Vegan Brownies", "url": "https://www.allrecipes.com/recipe/68436/vegan-brownies/" }
]

$ allrecipes-pp-cli recipe 9599/quick-and-easy-brownies --agent --select name,recipeIngredient,totalTime,aggregateRating
{
  "name": "Quick and Easy Brownies",
  "recipeIngredient": ["baking spray", "2 cups white sugar", "1.5 cups all-purpose flour", ...],
  "totalTime": 1800,
  "aggregateRating": { "value": 4.7, "count": 2040 }
}

$ allrecipes-pp-cli scale 9599/quick-and-easy-brownies --servings 32 --agent
# Returns factor=2 with all 10 ingredients correctly doubled

$ allrecipes-pp-cli doctor --json
{ "api": "reachable (via browser-chrome transport)", "auth": "not required", ... }
```

## Generator gaps for retro

- Generated `doctor` uses stdlib HTTP, not the configured transport — fails cleanly against Cloudflare-fronted sites.
- Generated HTML-extract `mode: links` produces raw HTML in result fields; usable as a starting point but always replaced for real-world site scraping.
- `auth` subcommand is unconditionally generated even when spec has `auth.type: none` — currently requires manual delete + unregister.
- Generated helpers (`replacePathParam`, `extractResponseData`, `printProvenance`, `wrap*`) become dead when generated `recipes_search.go`/`recipes_get.go` are replaced by hand-written equivalents — currently requires manual cleanup.
