# Printing Press Retro: allrecipes

## Session Stats
- API: allrecipes (Cloudflare-fronted recipe website, no official API)
- Spec source: synthetic (hand-written internal YAML; declared `kind: synthetic`, `http_transport: browser-chrome`, 2 endpoints, 24 extra_commands)
- Scorecard: 85/100 Grade A (after polish; 78 before)
- Verify pass rate: 100% (after polish; 64% before)
- Dogfood: PASS (data pipeline GOOD after polish; PARTIAL before)
- Live acceptance gate: PASS (64/65 Phase 5 tests; 1 in-session fix)
- Fix loops: 2 (polish round 1: dry-run + dead-helper sweep; polish round 2: search.go rename + Cookbook section)
- Manual code edits: ~20 files touched after generation (foundation: 6 files in internal/recipes; commands: 8 cmd_*.go files; root + doctor + helpers patches)
- Features built from scratch: 27 hand-written top-level commands (the spec only defined `recipes search` and `recipes get`)

## Findings

### F1. Generated `doctor` uses stdlib HTTP, not the configured transport (Bug / Default gap)

- **What happened:** The generated `doctor` command does its reachability probe with `&http.Client{}` from `net/http`. Against any Cloudflare-fronted (or generally bot-detected) site, this reports "API unreachable" even when the actual CLI commands work fine through the configured Surf/`browser-chrome` transport. I had to patch `doctor.go` to use `flags.newClient()` and add explicit Cloudflare interstitial detection.
- **Scorer correct?** N/A — neither dogfood nor scorecard probes the doctor's transport. The defect would have shipped silently if I hadn't manually invoked `doctor` against the live site during shipcheck.
- **Root cause:** `internal/generator/templates/doctor.go.tmpl` lines 19/152 — hardcoded `import "net/http"` and `httpClient := &http.Client{Timeout: 5 * time.Second}` regardless of the spec's `http_transport`. The same template is reused for every CLI.
- **Cross-API check:** Recurs on every CLI where the spec sets `http_transport` to anything other than `standard` — including any browser-chrome/browser-clearance/Surf-impersonated CLI. Today: at least allrecipes, recipe-goat, and any future Dotdash/Cloudflare-fronted site.
- **Frequency:** subclass: every CLI with `http_transport != standard`. That's 5-15% of catalog now and likely 25%+ as more reverse-engineered/sniffed CLIs ship.
- **Fallback if the Printing Press doesn't fix it:** Claude has to remember to patch `doctor.go` after every browser-chrome generation. Easy to forget; doctor failures only surface when a user runs `doctor`, which is often only after they've already hit a real failure. High miss rate.
- **Worth a Printing Press fix?** Yes. The fix is small; the cost of NOT fixing is "doctor lies to users on every Cloudflare-fronted CLI."
- **Inherent or fixable:** Fixable. The doctor template should call `flags.newClient()` (the same client every other command uses) instead of constructing a stdlib http.Client.
- **Durable fix:** In `doctor.go.tmpl`, replace the stdlib `httpClient := &http.Client{...}` block with `c, err := flags.newClient()` and use `c.Get("/", nil)` for the reachability probe. Detect challenge responses by inspecting the body for known interstitial signatures (Cloudflare `Just a moment`, Akamai BMP, Vercel challenge, AWS WAF, DataDome). Emit a domain-aware error message that names the transport and the protection vendor.
- **Test:** Positive — for `http_transport: browser-chrome`, generated doctor uses `flags.newClient()`; reachability probe succeeds against allrecipes.com. Negative — for `http_transport: standard`, doctor still works (Surf path is a superset; a stdlib Client can't probe Cloudflare anyway).
- **Evidence:** During Phase 4 shipcheck I patched `internal/cli/doctor.go` lines 44-95 to use Surf + Cloudflare detection. The diff is in the manuscripts.

### F2. `auth` subcommand is generated even when `auth.type: none` (Default gap / Scorer bug)

- **What happened:** Spec declared `auth.type: none`. The generator still emitted a full `internal/cli/auth.go` with `auth set-token`, `auth status`, `auth logout` subcommands. I deleted the file and removed `rootCmd.AddCommand(newAuthCmd(&flags))` from `root.go`. Scorer also docked us 1 point on the Auth dimension because the "auth subcommand" check fired without checking whether the spec needed auth.
- **Scorer correct?** Partially. The scorer is right that "no auth subcommand" is a sign of incomplete CLIs *for APIs that need auth*. It's wrong to dock a CLI whose spec explicitly says `auth.type: none`. Both should be fixed.
- **Root cause:** `internal/generator/generator.go:1239-1255` — `renderAuthFiles()` is called unconditionally; it picks an auth template (`auth.go.tmpl`, `auth_simple.go.tmpl`, or `auth_browser.go.tmpl`) based on auth subtypes but never short-circuits on `auth.type: none` for non-cookie, non-composed flows.
- **Cross-API check:** Every no-auth CLI ships dead UI. Other examples in the catalog: any public-data API (weather, news feeds, ESPN unauth endpoints, Hacker News, public Discord webhook readers, cooking sites).
- **Frequency:** subclass: `auth.type: none` AND `auth.type != cookie` AND `auth.type != composed`. Likely 10-20% of the catalog.
- **Fallback if the Printing Press doesn't fix it:** Claude must remember to delete `auth.go` and unregister `newAuthCmd` after every no-auth generation. Two-step manual change; high miss rate. The dead `auth set-token` UI is also a smell — agents may try to call it expecting it to work.
- **Worth a Printing Press fix?** Yes. Both gaps:
  - Generator: when `auth.type == "none"` and no GraphQL persisted-query state and no traffic-analysis browser hint, skip emitting `auth.go` AND skip the `rootCmd.AddCommand(newAuthCmd(&flags))` line. Add a guard in `auth_simple.go.tmpl` so it can detect this state if the unconditional render path is preserved for backward compatibility.
  - Scorer: in the Auth dimension, check whether `auth.type: none` and exempt the "no auth subcommand" deduction.
- **Inherent or fixable:** Fixable in both spots.
- **Durable fix:** In `generator.go:renderAuthFiles()`, early-return when `g.Spec.Auth.Type == "none"` AND there are no traffic-analysis hints requiring browser-aware auth. In `root.go.tmpl`, gate `newAuthCmd` registration with the same condition. In `scorecard.go:scoreAuth`, when the parsed spec has `auth.type: none`, don't dock for missing auth subcommand — give full credit if reachability passes.
- **Test:** Positive — generate from a spec with `auth.type: none` → no `internal/cli/auth.go`, no `newAuthCmd` registration, scorecard auth=10. Negative — generate from a spec with `auth.type: api_key` → auth.go still emitted with full subcommand tree.
- **Evidence:** Removed `internal/cli/auth.go` and patched `internal/cli/root.go:159` (commented out registration) during Phase 3. Scorecard auth: 9/10.

### F3. Generated helpers go dead when `recipes_search.go`/`recipes_get.go` are replaced (Default gap / Recurring friction)

- **What happened:** The generator emitted `internal/cli/helpers.go` with five helpers (`replacePathParam`, `extractResponseData`, `printProvenance`, `wrapWithProvenance`, `wrapResultsWithFreshness`) that exist solely to support the generic `resolveRead → extractHTMLResponse → wrapWithProvenance` pipeline used by spec-derived `*_search.go`/`*_get.go` handlers. When I replaced those handlers with custom JSON-LD-driven equivalents, all five helpers became dead code. Dogfood flagged them under "Dead Functions: 5 dead (WARN)" on the first run.
- **Scorer correct?** Yes — they ARE dead. The scorer is right to flag them. The fix isn't in the scorer; it's in the generator's emission strategy.
- **Root cause:** `helpers.go.tmpl` emits these helpers unconditionally. The expected use case is "the generated handler calls them"; the actual reality for HTML-scrape and reverse-engineered CLIs is "the handler is replaced by hand and the helpers go orphaned."
- **Cross-API check:** Every CLI that hand-writes its `*_search.go` and `*_get.go` (HTML-scrape, browser-sniff, RPC-style proxies, GraphQL persisted-query). Today: allrecipes, recipe-goat, any future synthetic+kind CLI. Plus "wrapper-only" catalog entries (krisukox/google-flights-api etc.) where the generated handlers are wholly inappropriate.
- **Frequency:** subclass: synthetic specs + HTML-scrape + browser-sniffed CLIs. 20%+ of the catalog and growing.
- **Fallback if the Printing Press doesn't fix it:** Claude has to delete 5 functions from helpers.go after replacing handlers. Easy to forget; the dogfood warning eventually catches it but it's noise.
- **Worth a Printing Press fix?** Yes. Two paths:
  - Best: emit these helpers only when at least one generated handler actually calls them. Static analysis at generation time: parse the rendered handlers; if no one calls `wrapWithProvenance`, don't emit it.
  - Cheaper: move them to a sub-package (`internal/cli/provenance/`) that handlers import on demand. If unused, the package's symbols don't appear in the dead-code scan.
- **Inherent or fixable:** Fixable.
- **Durable fix:** Move the provenance helpers into a separate package (`internal/cli/provenance/`) so they don't pollute the dead-code scan when unused. Simpler than parsing every rendered handler.
- **Test:** Positive — generate a synthetic-spec CLI, replace the search handler, run dogfood → no dead-code warnings about provenance helpers. Negative — generate a normal REST CLI, keep the generated handlers, the helpers are still callable via the new package import.
- **Evidence:** Phase 3 build log notes 5 dead helpers removed manually. Dogfood on first run: `Dead Functions: 5 dead (WARN)`.

### F4. HTML-extract `mode: links` returns raw HTML in result fields (Bug)

- **What happened:** The generated `recipes search` used `extractHTMLResponse(data, htmlExtractionOptions{Mode: "links", ...})` against allrecipes' search page. The output had `name` and `text` fields containing literal `<img src="..." />` elements, not clean text:
  ```json
  { "name": "<img src=\"https://...\" alt=\"S'mores Brownies\" /> S'mores Brownies 342 Ratings", ... }
  ```
  I replaced the entire handler with a custom Allrecipes-aware parser that extracts clean title, image URL (separately), rating, and review count.
- **Scorer correct?** N/A — no scorer probes for "is the parsed HTML output clean?" Output review (Phase 4.85) catches it but is wave-B warnings only.
- **Root cause:** `internal/generator/templates/html_extract.go.tmpl` line 83's `case "links":` branch — extracts text from anchors but doesn't strip nested `<img>`/`<span>`/`<div>` tags from the link's inner HTML. The scrape pipeline yields the raw textContent including HTML tag literals.
- **Cross-API check:** Every CLI using `response_format: html` with `mode: links` against modern templated pages (search results, category browse, gallery). Old-school Wikipedia-style pages would parse cleanly; Vue/React-rendered cards with image cards inside anchors fail.
- **Frequency:** subclass: every HTML-scrape CLI with image+text cards. Most modern recipe/news/listing sites use this layout.
- **Fallback if the Printing Press doesn't fix it:** Claude has to write a custom search-card parser for every HTML-scrape CLI. That's a 100-200 line file per CLI, often the most complex hand-written component.
- **Worth a Printing Press fix?** Yes. The link extractor is the entry point for every search/category/list command in HTML-scrape CLIs. Cleaning its output saves significant per-CLI hand-work.
- **Inherent or fixable:** Fixable. The `htmlLink` struct should carry separate `Title` (clean text) and `Image` (first nested `<img src="..." />` URL) fields. The text extraction should call a tag-stripping helper that decodes entities and collapses whitespace.
- **Durable fix:** In `html_extract.go.tmpl`, when extracting a link's title:
  1. Walk the anchor's children with html parser.
  2. Concatenate only TextNode values; skip ElementNodes.
  3. Decode HTML entities, collapse whitespace, trim.
  4. If a child is `<img>`, capture its `src` attribute into a separate `Image` field on the result struct.
  5. Add a `Rating` and `ReviewCount` field if the inner content matches a recipe-card-rating pattern (`mntl-recipe-card-meta__rating`, "N Ratings"). Make this domain-agnostic by making the regex configurable via `html_extract.options` in the spec.
- **Test:** Positive — generate a spec with `response_format: html, mode: links`, fetch a real search page with image cards, get clean titles + image URLs in separate fields. Negative — fetch a plain link list (sitemap, no images), still works.
- **Evidence:** First search probe returned `name: "<img src='...' /> S'mores Brownies 342 Ratings"`. Replaced with custom parser at `internal/recipes/search.go`. The Allrecipes-specific patterns (recipe-card-rating, "N Ratings") could become spec-driven configuration.

### F5. Hand-written transcendence commands need `--dry-run` short-circuit boilerplate (Skill instruction gap / Default gap)

- **What happened:** The first polish-worker pass added `if flags.dryRun { return nil }` to 12 hand-written commands so verify's `--dry-run` probe could reach exit 0. Many of those commands also had `Args: cobra.MinimumNArgs(1)` or `MarkFlagRequired(...)` that had to be removed because cobra evaluates them BEFORE the RunE function runs, so the dry-run guard couldn't reach. That's a pattern: every hand-written transcendence command needs the same scaffolding for verify compatibility.
- **Scorer correct?** Yes — verify is right to want dry-run-able commands. The fix isn't to game the scorer; it's to make the build pattern emit-friendly.
- **Root cause:** Skill phase 3 build instructions don't enumerate this pattern. The cmd_helpers.go file I created (which is itself a useful pattern) doesn't include a `dryRunHelp(cmd, flags) error` helper, so each command rewrites the same 3 lines.
- **Cross-API check:** Every CLI with hand-written novel-feature commands (which is most synthetic-spec CLIs and many normal CLIs that ship transcendence features). 50%+ of generated CLIs ship Phase 3 hand-written commands.
- **Frequency:** every CLI with hand-written commands (most of the catalog).
- **Fallback if the Printing Press doesn't fix it:** First polish-worker pass adds the boilerplate, but the iteration cycle is wasted. Verify pass-rate sits artificially low until polish runs.
- **Worth a Printing Press fix?** Yes. Cheap to fix; recurring on every generation.
- **Inherent or fixable:** Fixable.
- **Durable fix:** Two layers:
  - Generator: in `cmd_helpers.go` (or wherever shared helpers land), emit a `dryRunOK(cmd *cobra.Command, flags *rootFlags) error` helper that returns nil if `flags.dryRun` is true and the command is in a dry-run-safe state. Document it in the helper's comments as the canonical first line of any RunE that requires positional args or required flags.
  - Skill: in `skills/printing-press/SKILL.md` Phase 3 build instructions, add a "Verify-friendly RunE template" section showing the `if flags.dryRun { return nil }` pattern and the rule "do NOT use `Args: cobra.MinimumNArgs(N)` or `MarkFlagRequired(...)`; check inside RunE and fall through to `cmd.Help()` for help-only invocations."
  - Even better: emit a Cobra middleware that wraps every hand-registered command with the dry-run check. Generator-side change in root.go.tmpl.
- **Test:** Positive — generate a CLI with a synthetic spec, hand-author a command following the helper pattern, run verify → dry-run path passes without polish. Negative — verify still catches genuine command bugs (RunE that should fail returns its actual error).
- **Evidence:** Polish-worker round 1 modified 12 commands across 3 files (cmd_recipe.go, cmd_pantry.go, cmd_cookbook.go) plus which.go. All edits were the same 3-line pattern.

### F6. Dogfood data-pipeline check uses filename heuristic `internal/cli/search.go` (Scorer bug)

- **What happened:** The polish-worker round 2 renamed `cmd_search.go` to `search.go` to satisfy `dogfood.go:1225` (`searchData, _ := os.ReadFile(filepath.Join(dir, "internal", "cli", "search.go"))`). The actual content at both filenames was identical and used the same `recipes.QueryIndex` + `recipes.FetchSearch` calls. Pure file-name ceremony.
- **Scorer correct?** No. The scorer's grep is brittle: it asks "does the file at the canonical path use a domain-specific store method?" but doesn't search for search.go-class files. A CLI with `cmd_search.go`, `commands/search.go`, or `searches.go` fails the same way.
- **Root cause:** `internal/pipeline/dogfood.go:1225` — single hardcoded path.
- **Cross-API check:** Every CLI that uses non-canonical filenames for its search command (most synthetic-spec CLIs that I've seen).
- **Frequency:** every CLI not perfectly aligned with the heuristic's filename assumption.
- **Fallback if the Printing Press doesn't fix it:** Renaming files is a 30-second fix per CLI. But it gates the data-pipeline-integrity score, which is a Tier-2 dimension.
- **Worth a Printing Press fix?** Yes. Scorer fix is cheap and durable.
- **Inherent or fixable:** Fixable.
- **Durable fix:** In `dogfood.go`, find search.go-class files by scanning for `func newSearchTopCmd|func newSearchCmd|cobra.Command{Use: "search"...}` patterns, OR aggregate evidence by reading every `internal/cli/*.go` and looking for calls into the store package + a `search` cobra command. Don't gate on filename.
- **Test:** Positive — a CLI named `cmd_search.go` calling Store.Search receives full data-pipeline credit. Negative — a CLI with a `search.go` that calls a stubbed function (no real store call) does NOT receive credit.
- **Evidence:** Polish-worker round 2 result includes the rename. Verified that the rename was the only change that lifted data_pipeline_integrity from PARTIAL to GOOD.

### F7. Scorer's `insight` dimension uses filename prefixes (Scorer bug)

- **What happened:** Scorecard's insight dimension scored 2/10. Inspection: `scorecard.go:1208 scoreInsight()` looks for filename prefixes like `health`, `bottleneck`, `trends`, `analytics`, `velocity`, etc. Our genuinely insight-producing commands — `top-rated` (Bayesian-smoothed ranking), `quick` (cache+rating filter), `pantry` (overlap scoring), `with-ingredient` (reverse index) — don't match any of those names, so they don't count.
- **Scorer correct?** No (mostly). The scorer's content fallback (Signal 2: "store + SQL aggregation") catches some of these, but it specifically looks for `COUNT()` / `SUM()` / `GROUP BY`. Our SQL is a different shape — Bayesian-smoothing is in Go code, not SQL aggregation. The detection misses real insight.
- **Root cause:** `scorecard.go:scoreInsight()` lines 1217-1255 — the prefix list and the SQL-aggregation regex are too narrow.
- **Cross-API check:** Every CLI that ships ranking/scoring features without using the canonical filename prefixes. Given that many APIs warrant Bayesian smoothing (recipe sites, product reviews, restaurant ratings, sports stats), this is widespread.
- **Frequency:** every CLI with non-aggregation insight features.
- **Fallback if the Printing Press doesn't fix it:** Claude could rename files (game the score) or add commands the scorer expects (feature creep). Neither is right.
- **Worth a Printing Press fix?** Yes. Scorer fix.
- **Inherent or fixable:** Fixable.
- **Durable fix:** In `scoreInsight()`, add a Signal 3 that detects ranking/scoring algorithms by content patterns:
  - `BayesianRating`, `bayesSmooth`, `priorMean` symbols → ranking insight
  - `Rank(`, `sortBy`, custom comparator functions over rating + count → ranking insight
  - `score := (...)` patterns combining 3+ fields → composite-score insight
  - Reverse-index queries (FTS5 + JOIN against typed table) → reverse-lookup insight
  Each pattern is domain-evidence; counting any one of them as insight credit is reasonable.
- **Test:** Positive — a CLI with a `top-rated` command that uses Bayesian smoothing scores ≥6/10 on insight. Negative — a CLI with no analytical commands still scores 0-2/10.
- **Evidence:** Polish-worker round 2 explicitly skipped this finding as "needs a scorer change." Scorecard.go:1208-1255 confirms the narrow detection.

### F8. Scorer's `type_fidelity` regex over-matches (Scorer bug)

- **What happened:** Scorecard's `type_fidelity` scored 3/5 because the per-flag-description average word count was 2.88. Manual inspection: actual flag descriptions in our CLI are reasonably detailed (e.g., `--smooth-c "Bayesian credibility weight (higher = stricter; needs more reviews to leave the prior)"`). The scorer's regex extracts the description string AND counts adjacent variable names (e.g., `flagMaxMissing`) as if they were part of the description.
- **Scorer correct?** No.
- **Root cause:** `scorecard.go:scoreTypeFidelity()` — needs source inspection but the symptom is regex over-match.
- **Cross-API check:** Every CLI with hand-written flags using the `cmd.Flags().StringVar(&flagFoo, "foo", "", "Description")` pattern. The variable name `flagFoo` gets pulled in.
- **Frequency:** Every CLI.
- **Fallback if the Printing Press doesn't fix it:** Type fidelity is artificially capped for every CLI.
- **Worth a Printing Press fix?** Yes. Scorer fix.
- **Durable fix:** Tighten the regex to capture only the third argument string literal of `.StringVar/.IntVar/.BoolVar/.Float64Var/...`. Use the Go AST instead of regex if regex fragility persists.
- **Test:** Positive — a CLI with detailed flag descriptions averages > 8 words per description. Negative — a CLI with terse `"foo bar"` descriptions still scores low.
- **Evidence:** Polish-worker round 2 noted this as "scorer bug; not fixable without scorer change."

### F9. Cookbook README section was missing (Skill instruction gap)

- **What happened:** Polish-worker round 2 added a "Cookbook" section to README with 9 worked examples. The 5 standard sections the scorecard expects are: Quick Start, Agent Usage, Health Check, Troubleshooting, Cookbook. Our initial README had the first 4 but no Cookbook.
- **Scorer correct?** Yes.
- **Root cause:** The skill's research.json `narrative.recipes` field IS used to drive the Cookbook section in some templates but not in this CLI's generation. The README template either didn't render the Cookbook section or rendered it under a different name (it had "Unique Features" and "Cookbook"-style content under that name).
- **Cross-API check:** Every CLI generated from a research.json with a `recipes` array. Currently inconsistent which template branch fires.
- **Frequency:** unclear — could be a one-off for this run, could be common.
- **Fallback if the Printing Press doesn't fix it:** Polish picks it up.
- **Worth a Printing Press fix?** Yes. The scorecard's standard-section check should be the definitive list, and the README template should always emit a Cookbook section when `narrative.recipes` is non-empty (even if it duplicates content from "Unique Features" — the scorecard reads section names, not content).
- **Durable fix:** README template should emit `## Cookbook` heading + each entry from `narrative.recipes`. Don't merge with "Unique Features" — they serve different scorer signals.
- **Test:** Positive — a CLI with `narrative.recipes` non-empty has a `## Cookbook` section. Negative — a CLI with empty recipes has neither (and the scorecard exempts that dimension or tolerates absence).
- **Evidence:** Polish round 2 manually wrote the Cookbook section.

### F10. Phase 1.5 absorb manifest user-cut workflow (What went right)

- **What happened:** When I presented the absorb manifest with 10 novel features, the user replied "cut #7, #9". The Phase Gate flow accepted ad-hoc cuts cleanly: I edited the manifest, edited research.json, re-validated, and proceeded.
- **Worth flagging:** Yes — protect this. The user-friendly cut syntax is a strength of the gate workflow.

### F11. Browser-Sniff Gate marker file enforcement (What went right)

- **What happened:** Phase 1.7 wrote a `browser-browser-sniff-gate.json` marker. Phase 1.5 verified its presence before proceeding. The marker survived archiving. Worked smoothly.
- **Worth flagging:** Yes — keep the contract.

### F12. `cli_description` → `root.Short` (What went right)

- **What happened:** I added `cli_description: "Search and fetch Allrecipes recipes as structured data, scale ingredients, build grocery lists, and rank by Bayesian-smoothed popularity."` to the spec. It landed cleanly in `root.go`'s `Short:` field, providing user-friendly framing.
- **Worth flagging:** Yes — keep it. Generator-side support for `cli_description` is a recent addition and worked perfectly.

### F13. `research.json` narrative.* drives README + SKILL with high fidelity (What went right)

- **What happened:** Every field in `narrative` (`display_name`, `headline`, `value_prop`, `auth_narrative`, `quickstart`, `troubleshoots`, `recipes`, `trigger_phrases`, `when_to_use`) landed in the right place in README + SKILL. The README's "Authentication" section was correctly auto-generated as "No authentication required..." instead of generic auth boilerplate.
- **Worth flagging:** Yes — keep it.

### F14. Surf/`browser-chrome` transport for Cloudflare-fronted sites (What went right)

- **What happened:** A single line `http_transport: browser-chrome` in the spec produced a CLI that walked past Cloudflare's TLS fingerprint detection on every request. No additional configuration needed.
- **Worth flagging:** Yes — this is the Printing Press at its best. Specific: keep this as the canonical pattern; the doctor fix (F1) should default-detect this transport's presence.

### F15. Lock acquire/update/promote pattern (What went right)

- **What happened:** Lock acquire at Phase 2 start; heartbeat updates at each phase; promote at Phase 5.6 atomically swapped working dir into library. Smooth.
- **Worth flagging:** Yes — protect.

## Prioritized Improvements

### P1 — High priority

| Finding | Title | Component | Frequency | Fallback Reliability | Complexity | Guards |
|---------|-------|-----------|-----------|---------------------|------------|--------|
| F1 | doctor uses stdlib HTTP | Generator template (`doctor.go.tmpl`) | every browser-chrome CLI | low (Claude must remember manual patch) | small | activate when `http_transport != standard` |
| F2 | `auth` subcommand emitted for `auth.type:none` | Generator (`generator.go:renderAuthFiles`) + Scorer (`scorecard.go:scoreAuth`) | every no-auth CLI | low (two-step manual change) | small | activate when `auth.type == "none"` AND no GraphQL persisted-query AND no traffic-analysis browser hint |
| F4 | HTML-extract `mode:links` returns raw HTML | Generator (`html_extract.go.tmpl`) | every HTML-scrape CLI with image+text cards | low (custom parser per CLI) | medium | activate for `response_format: html, mode: links` |
| F5 | Hand-written commands need `--dry-run` boilerplate | Generator (`cmd_helpers.go.tmpl`) + Skill (`SKILL.md` Phase 3 build) | every CLI with hand-written commands | medium (polish-worker catches it) | small | none — pure helper emission |

### P2 — Medium priority

| Finding | Title | Component | Frequency | Fallback Reliability | Complexity | Guards |
|---------|-------|-----------|-----------|---------------------|------------|--------|
| F7 | Insight scorer uses filename prefixes | Scorer (`scorecard.go:scoreInsight`) | every CLI with non-aggregation insight | medium (rename or accept score) | medium | none — pure scorer fix |
| F8 | Type fidelity scorer regex over-matches | Scorer (`scorecard.go:scoreTypeFidelity`) | every CLI | low (artificial cap) | small-medium | none — pure scorer fix |
| F6 | Data pipeline scorer uses hardcoded filename `search.go` | Scorer (`dogfood.go:1225`) | every synthetic-spec CLI | medium (renaming files works) | small | none — pure scorer fix |
| F3 | Generated provenance helpers go dead when handlers replaced | Generator (move helpers to subpackage) | every synthetic-spec / HTML-scrape CLI | medium (dogfood catches at warn level) | small-medium | none — refactor into a separate package |
| F9 | Cookbook README section missing | Generator (README template) | inconsistent | medium (polish catches) | small | none |

### P3 — Low priority

(none surfaced — every "Do" finding is P1 or P2)

### Skip

| Finding | Title | Why unlikely to recur |
|---------|-------|----------------------|
| (printed-CLI SQL bug) | QueryIndex ambiguous column | Hand-written code in this CLI's localstore.go; not generator-emitted; only relevant when the agent writes a JOIN-with-FTS5 helper. Not recurring across the catalog. |

## Work Units

### WU-1: Doctor template uses configured transport (from F1)
- **Goal:** Generated `doctor` reachability probe uses the same transport that other commands use, so it can verify connectivity through Cloudflare/Surf/etc. and report transport-specific failure modes.
- **Target:** `internal/generator/templates/doctor.go.tmpl` lines 19, 152-157, 167, 220.
- **Acceptance criteria:**
  - positive: generate from a spec with `http_transport: browser-chrome`, run `doctor` against an https://www.allrecipes.com base URL, report includes "reachable (via browser-chrome transport)".
  - negative: generate from a spec with `http_transport: standard`, doctor still works (Surf is a superset; can hit normal endpoints).
  - regression: existing CLIs (notion, github, stripe) still pass doctor with the new template.
- **Scope boundary:** Do not change the `doctor` flag interface or the JSON output schema. Only the HTTP probe's transport.
- **Dependencies:** none.
- **Complexity:** small.

### WU-2: Skip `auth` subcommand and registration when `auth.type == "none"` (from F2)
- **Goal:** No-auth CLIs ship without dead `auth` UI; scorecard auth dimension gives full credit when the spec says no auth is required.
- **Target:** `internal/generator/generator.go:renderAuthFiles`; `internal/generator/templates/root.go.tmpl` (registration line); `internal/pipeline/scorecard.go:scoreAuth`.
- **Acceptance criteria:**
  - positive: spec with `auth.type: none` → no `internal/cli/auth.go`, no `newAuthCmd` registration, scorecard auth = 10/10.
  - negative: spec with `auth.type: api_key` → auth.go emitted with full subcommand tree.
  - guard: spec with `auth.type: none` BUT a traffic-analysis hint of `graphql_persisted_query` → still emits the browser-aware auth template (existing exception).
- **Scope boundary:** Do not change the auth command for any auth type other than `none`.
- **Dependencies:** none.
- **Complexity:** small.

### WU-3: HTML extract `mode: links` produces clean text + separate image URL (from F4)
- **Goal:** Generated `recipes search` (and any HTML-scrape link-extraction handler) returns clean string fields without raw HTML; image URLs in a separate field.
- **Target:** `internal/generator/templates/html_extract.go.tmpl` lines 21-30 (struct), 83-150 (link mode logic).
- **Acceptance criteria:**
  - positive: extract from a real allrecipes search page, every result has clean `Title`, separate `Image` URL, no nested HTML in any string field.
  - negative: extract from a plain `<ul><li><a>...</a></li></ul>` list still works (no images, just clean titles).
  - regression: golden tests for existing link-mode CLIs still pass.
- **Scope boundary:** Don't add domain-specific extraction (rating, review count) to the default extractor — leave that to per-CLI parsers. But document a hook so spec authors can declare additional fields via `html_extract.field_extract` rules.
- **Dependencies:** none.
- **Complexity:** medium.

### WU-4: Verify-friendly RunE pattern for hand-written commands (from F5)
- **Goal:** Hand-written novel-feature commands pass verify on first run — no polish-worker touch-up needed for `--dry-run` compatibility.
- **Target:** `internal/generator/templates/cmd_helpers.go.tmpl` (or wherever cmd_helpers is emitted); `skills/printing-press/SKILL.md` Phase 3 build instructions.
- **Acceptance criteria:**
  - positive: emit a `dryRunOK(cmd, flags) bool` helper. Add a Phase 3 build template "Verify-friendly RunE" showing the pattern.
  - positive: skill instruction explicitly tells Claude to NOT use `Args: cobra.MinimumNArgs(N)` or `MarkFlagRequired(...)` for hand-written commands; instead check inside RunE and fall through to `cmd.Help()` for help-only invocations.
  - negative: existing generated commands (which already use Args correctly) are unchanged.
- **Scope boundary:** Don't auto-rewrite existing commands. Just emit the helper + document the pattern.
- **Dependencies:** none.
- **Complexity:** small.

### WU-5: Insight scorer detects content patterns, not just filename prefixes (from F7)
- **Goal:** CLIs with ranking/scoring/reverse-index features get insight credit even when their filenames don't match the canonical prefixes.
- **Target:** `internal/pipeline/scorecard.go:scoreInsight()` lines 1208-1255.
- **Acceptance criteria:**
  - positive: a CLI with `BayesianRating`, `Rank(`, `sortBy` patterns scores ≥ 6/10 on insight.
  - positive: a CLI with reverse-index queries (FTS5 + JOIN against typed table) scores ≥ 4/10.
  - negative: a CLI with no analytical commands stays at 0-2/10.
  - regression: existing CLIs (planet-stats, github-velocity, etc.) maintain their current scores.
- **Scope boundary:** Don't lower the bar for genuinely empty CLIs. The patterns must be specific enough that "insight" still means insight.
- **Dependencies:** none.
- **Complexity:** medium.

### WU-6: Type fidelity scorer captures only the description literal (from F8)
- **Goal:** Average word count per flag description reflects only the description string, not adjacent variable names.
- **Target:** `internal/pipeline/scorecard.go:scoreTypeFidelity()`.
- **Acceptance criteria:**
  - positive: a CLI with detailed flag descriptions (>8 words avg) scores 5/5.
  - positive: existing CLIs that scored well continue to score well (no false negatives).
  - negative: a CLI with terse `"foo"` descriptions still scores < 3/5.
- **Scope boundary:** Use Go AST if regex remains brittle.
- **Dependencies:** none.
- **Complexity:** small-medium.

### WU-7: Data pipeline scorer detects search by content, not filename (from F6)
- **Goal:** Dogfood's data-pipeline check finds search.go-class commands by content (cobra `Use: "search"` + store call), not by hardcoded path.
- **Target:** `internal/pipeline/dogfood.go:1225`.
- **Acceptance criteria:**
  - positive: a CLI with `cmd_search.go` (or `commands/search.go`, etc.) that uses Store.Search* still gets data-pipeline credit.
  - negative: a CLI whose search command stubs the store call (no real persistence) does NOT get credit.
  - regression: existing CLIs unchanged.
- **Scope boundary:** Don't widen the search to false-positive on every command containing the substring "search".
- **Dependencies:** none.
- **Complexity:** small.

### WU-8: Move provenance helpers to a subpackage so they don't pollute dead-code scans (from F3)
- **Goal:** When generated handlers are replaced by hand, the unused provenance helpers don't trigger dogfood dead-code warnings.
- **Target:** Generator templates emit `internal/cli/provenance/` package; rendered handlers import it on demand.
- **Acceptance criteria:**
  - positive: generate a synthetic-spec CLI, replace `*_search.go` and `*_get.go` with hand-written equivalents, run dogfood → no dead-code warnings about provenance helpers.
  - negative: generate a normal REST CLI, leave handlers intact → handlers correctly call `provenance.WrapWith*`, etc.
  - regression: existing CLIs that import the old helpers still build (provide a transitional shim or migrate them all).
- **Scope boundary:** Migrate exactly: `extractResponseData`, `printProvenance`, `wrapWithProvenance`, `wrapResultsWithFreshness`, `replacePathParam`. Do not move other helpers.
- **Dependencies:** none.
- **Complexity:** medium.

## Anti-patterns
- **Hardcoded `httpClient := &http.Client{Timeout: 5*time.Second}` in generated code where the rest of the CLI uses Surf.** Always reuse `flags.newClient()` (or whatever the canonical client constructor is) — that's the only way to honor `http_transport`.
- **Hardcoded filename heuristics in scorers.** `dogfood.go:1225` reading exactly `internal/cli/search.go` is the canonical example — anti-pattern is filename-as-contract.
- **Unconditional code generation when the spec says "don't."** Auth, MCP intent stubs, anything where the spec encodes "this isn't applicable" should suppress the matching template, not emit it.
- **Filename prefix matching for capability detection.** Insight scoring should look at content (Bayesian smoothing, rank-by-composite-score, reverse-index queries), not at file basenames.
- **Required cobra args on hand-written commands.** `Args: cobra.MinimumNArgs(N)` and `MarkFlagRequired` block the `--dry-run` short-circuit. Either avoid them in hand-written commands OR teach Cobra to honor `--dry-run` via a PreRun.

## What the Printing Press Got Right
- **`http_transport: browser-chrome` was a one-line spec change that produced a Cloudflare-bypassing CLI.** Surf integration in the generated client is solid; the rest of the catalog could lean on this more.
- **`research.json` narrative drives README + SKILL with high fidelity.** display_name, headline, value_prop, auth_narrative, troubleshoots, recipes, trigger_phrases — all landed in the right places.
- **`cli_description` flag in spec → `root.Short` in CLI.** Recent addition; works perfectly.
- **Phase 1.5 absorb manifest gate with user-cut syntax.** "cut #7, #9" was clean and the manifest accepted edits without rebuilding research.
- **Browser-Sniff Gate marker file contract.** Survives archiving; clear pre-flight check at Phase 1.5; legacy-resume tolerance.
- **Lock acquire / update / promote.** Heartbeat across phases; atomic library swap at promote.
- **Synthetic spec + extra_commands as scaffolding for hand-built CLIs.** The user-facing flow felt natural; declaring `kind: synthetic` correctly relaxed dogfood's path-validity check and scorer's tier-2 denominator.
- **Auto-suggest novel features framework.** I came up with 10 novel features; user accepted 8/10. The structured Score/Evidence/Group format made the gate decision cheap.
- **Quality gates (7 mechanical checks) passed on first generation.** No retry loop needed.
