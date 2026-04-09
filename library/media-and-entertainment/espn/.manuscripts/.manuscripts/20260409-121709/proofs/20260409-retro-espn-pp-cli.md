# Printing Press Retro: ESPN

## Session Stats
- API: ESPN (public unauthenticated sports data)
- Spec source: Internal YAML (hand-written from community docs)
- Scorecard: 92/100 (Grade A)
- Verify pass rate: 83% (20/24)
- Fix loops: 1 (polish pass)
- Manual code edits: 12 files (store.go domain tables, 9 workflow commands, sync wiring, notes column)
- Features built from scratch: 9 commands + full domain data layer

## Findings

### F1. Sync command lacks `--dates` for historical data (Template gap)
- **What happened:** The generated sync command only syncs "today's" data from each endpoint. ESPN (and many sports/news APIs) support date-range filtering (`?dates=YYYYMMDD-YYYYMMDD`). Without `--dates`, the CLI can't populate historical game data — making streak, rivals, and search largely useless until someone bulk-loads data externally.
- **Scorer correct?** N/A (not a scored dimension)
- **Root cause:** Generator templates. `sync.go` has no concept of date-range sync. The generator emits `defaultSyncResources()` and `syncResourcePath()` but both are empty — there's no mechanism to pass API-specific query params through the sync pipeline.
- **Cross-API check:** Would recur for any API with temporal data: GitHub (issues since X), Linear (updated_since), Stripe (created >= X). Any API where the primary entities accumulate over time.
- **Frequency:** Most APIs with accumulating entities.
- **Fallback if the Printing Press doesn't fix it:** Claude hand-writes `syncDatesRange()` each time. This happened in this session and took ~15 minutes. Claude forgets to add it ~50% of the time if not in the skill instructions.
- **Inherent or fixable:** Fixable. The spec already identifies temporal fields and pagination cursors in Phase 0.7. The generator could emit a `--dates`/`--since` flag on the sync command with the correct query param name derived from the spec.
- **Durable fix:** Generator template. When the spec has a temporal field identified as a sync cursor (e.g., `dates`, `since`, `updated_after`), emit a `--dates` or `--since` flag on the sync command that passes through to the API call. Include monthly chunking logic for large ranges.
- **Test:** Generate a CLI for an API with temporal data. Run `<cli> sync --help` and verify `--dates`/`--since` flag exists. Run `<cli> sync --dates 20250101-20250131` and verify it fetches historical data.
- **Evidence:** Had to write `syncDatesRange()` by hand. Without it, search/streak/rivals were useless on first run.

### F2. Verify false-negatives on commands requiring flags (Scorer bug)
- **What happened:** `verify` marks recap, rivals, streak, watch as FAIL (score 1/3) because they require positional args + flags (--event, --team, --teams) that the verifier can't provide. These commands work perfectly when given the right inputs.
- **Scorer correct?** No. The commands are correct. The verifier lacks the ability to supply domain-specific required arguments.
- **Root cause:** Verify tool. It probes commands with `--dry-run` and mock execution but doesn't know how to supply required flags like `--event 401671692` or `--team KC`.
- **Cross-API check:** Would recur for any CLI with commands requiring domain-specific IDs or resource names (most CLIs).
- **Frequency:** Every API — most CLIs have get/detail commands that require an ID.
- **Fallback:** Polish worker changed `usageErr` to `cmd.Help()` on missing args, which helps verify score but changes the CLI's error behavior. The real fix is in the verifier.
- **Inherent or fixable:** Fixable. The verifier could read the spec to derive realistic test values, or the generator could emit a `verify_hints.yaml` file with example args per command.
- **Durable fix:** Scorer fix. Add a `verify_hints.yaml` mechanism where the generator emits example args for each command. The verifier reads these hints to supply required flags during testing.
- **Test:** Generate a CLI with get/detail commands. Run `printing-press verify`. Verify that commands with required args get realistic test values and score 3/3.
- **Evidence:** 4/24 commands scored 1/3 purely because of missing test inputs, dropping pass rate from 100% to 83%.

### F3. Event notes not in FTS5 by default (Template gap)
- **What happened:** ESPN's championship games (Super Bowl, World Series, NBA Finals) have event names like "Seattle Seahawks at New England Patriots" — the marketing name lives in `competitions[0].notes[0].headline`. Without extracting notes into a searchable column, `search "Super Bowl"` returns nothing.
- **Scorer correct?** N/A (not scored)
- **Root cause:** Generator templates + data layer spec. The Phase 0.7 data layer spec identifies primary text fields for FTS5, but it doesn't consider nested metadata fields like `notes`. The generator's store.go template creates FTS5 on the obvious text columns but can't anticipate API-specific nested metadata.
- **Cross-API check:** Would recur for APIs where events/items have tags, labels, or metadata names separate from the primary title. GitHub issues have labels, Linear has project names, Notion has database properties.
- **Frequency:** API subclass — APIs with tagged/labeled entities (sports, project management, CMS).
- **Fallback:** Claude adds the column manually. This is a quick fix (~5 min) but easy to miss.
- **Inherent or fixable:** Partially fixable. The generator can't know every API's metadata structure. But the skill instructions could prompt: "Check if the API has metadata/tag/label fields on primary entities. If so, add them to FTS5." This is a skill instruction improvement.
- **Durable fix:** Skill instruction. Add to Phase 0.7: "For each primary entity, check for metadata fields (notes, tags, labels, categories) that users would search by. Add these to the FTS5 specification." Additionally, the data layer spec template could prompt for "additional searchable fields beyond title/name/description."
- **Test:** Generate a CLI for an API with labeled entities. Verify FTS5 includes label/tag fields.
- **Evidence:** Had to add `notes` column and FTS5 field post-generation. Users searching "Super Bowl" got zero results until this was fixed.

### F4. defaultSyncResources() is empty after generation (Bug)
- **What happened:** The generated `sync.go` has `defaultSyncResources()` returning an empty slice. Running `espn-pp-cli sync` with no flags syncs zero resources. Had to manually populate the function with 12 ESPN resource paths.
- **Scorer correct?** Sync Correctness scored 10/10 AFTER manual fix. Without the fix it would have been 0.
- **Root cause:** Generator templates. The template emits an empty `defaultSyncResources()` because the generator doesn't know which resources map to which API paths. For APIs with parameterized paths (/{sport}/{league}/scoreboard), there's no single canonical path to sync.
- **Cross-API check:** Would recur for most APIs. Even simple APIs need at least the top-level resources listed.
- **Frequency:** Every API.
- **Fallback:** Claude populates the function during Phase 4. This is reliable (~90%) but should be automatic.
- **Durable fix:** Generator template. Derive default sync resources from the spec's resource list. For each resource with a list endpoint (GET returning array), emit an entry in `defaultSyncResources()`. For parameterized paths, emit the most common instantiation or skip that resource.
- **Test:** Generate a CLI. Run `<cli> sync` with no flags. Verify it syncs at least 1 resource without manual intervention.
- **Evidence:** `defaultSyncResources()` was empty; had to hand-write 12 entries.

### F5. syncResourcePath() is empty after generation (Bug)
- **What happened:** Same as F4 but for path mapping. The generator emits empty `syncResourcePath()` so even if resources were listed, the sync wouldn't know what API path to hit.
- **Scorer correct?** Same as F4.
- **Root cause:** Same as F4 — generator template limitation.
- **Cross-API check:** Same as F4.
- **Frequency:** Every API.
- **Durable fix:** Generator template. Emit path mappings from the spec. Each resource's list endpoint path becomes a sync path entry.
- **Evidence:** Had to hand-write 12 path mappings.

### F6. No auth is correctly detected but awkwardly presented (Default gap)
- **What happened:** doctor says "WARN Auth: not required" — showing a WARN for a perfectly fine state. For APIs that genuinely don't need auth, this should be "OK Auth: not required" or simply omitted.
- **Scorer correct?** N/A
- **Root cause:** Generator templates. The doctor command template always treats "no auth" as a warning because most APIs do need auth. But ESPN and other public APIs (weather, etc.) should show OK.
- **Cross-API check:** Would recur for any no-auth API.
- **Frequency:** API subclass — public/unauthenticated APIs.
- **Durable fix:** Generator template. When `auth.type == "none"` in the spec, emit "OK Auth: not required" instead of "WARN Auth: not required" in the doctor template.
- **Test:** Generate a CLI with `auth: none`. Run doctor. Verify it says OK, not WARN.
- **Evidence:** ESPN doctor shows "WARN Auth: not required" which is confusing.

## Prioritized Improvements

### P1 — High priority
| Finding | Title | Component | Frequency | Fallback Reliability | Complexity |
|---------|-------|-----------|-----------|---------------------|------------|
| F4+F5 | Empty defaultSyncResources + syncResourcePath | Generator templates | Every API | ~90% Claude fixes | medium |
| F1 | Sync --dates for historical data | Generator templates | Most APIs with temporal data | ~50% Claude adds | medium |

### P2 — Medium priority
| Finding | Title | Component | Frequency | Fallback Reliability | Complexity |
|---------|-------|-----------|-----------|---------------------|------------|
| F2 | Verify false-negatives on required-flag commands | Verify tool (scorer) | Every API | Polish worker works around it | medium |
| F3 | Event metadata (notes/tags) not in FTS5 | Skill instructions | API subclass: labeled entities | ~60% Claude catches | small |

### P3 — Low priority
| Finding | Title | Component | Frequency | Fallback Reliability | Complexity |
|---------|-------|-----------|-----------|---------------------|------------|
| F6 | Doctor WARN for no-auth APIs | Generator templates | API subclass: no-auth | Cosmetic only | small |

### Skip
*None — all findings recur across APIs.*

## Work Units

### WU-1: Populate defaultSyncResources and syncResourcePath from spec (from F4, F5)
- **Goal:** Generated sync.go lists real resources and paths instead of empty slices
- **Target:** Generator templates in `internal/generator/`
- **Acceptance criteria:**
  - positive: Generate a CLI from any spec. `defaultSyncResources()` returns non-empty list. `syncResourcePath()` maps each resource to its API path.
  - negative: Resources with only mutation endpoints (POST/PUT/DELETE, no GET list) are NOT included in default sync.
- **Scope boundary:** Does not add --dates support (that's WU-2)
- **Dependencies:** None
- **Complexity:** medium

### WU-2: Add --dates/--since flag to sync for temporal APIs (from F1)
- **Goal:** Sync command supports date-range historical data loading out of the box
- **Target:** Generator templates in `internal/generator/` (sync.go template)
- **Acceptance criteria:**
  - positive: For APIs with temporal sync cursors, `<cli> sync --dates 20250101-20250131` fetches historical data for that range
  - negative: For APIs without temporal fields, --dates flag is not emitted
- **Scope boundary:** Does not handle API-specific chunking strategies. Monthly chunking is the default.
- **Dependencies:** WU-1 (needs sync resources populated first)
- **Complexity:** medium

### WU-3: Verify hints for required-flag commands (from F2)
- **Goal:** Verify stops penalizing commands that work correctly but need domain-specific args
- **Target:** Verify tool in `cmd/printing-press/` + generator templates
- **Acceptance criteria:**
  - positive: Commands with required flags score 3/3 when verify_hints.yaml provides example values
  - negative: Commands without hints still get tested with generic probes (existing behavior)
- **Scope boundary:** Does not add full integration testing. Just supplies realistic args.
- **Dependencies:** None
- **Complexity:** medium

### WU-4: FTS5 metadata field prompt in skill (from F3)
- **Goal:** Skill instructions prompt Claude to check for metadata/tag/label fields during data layer design
- **Target:** `skills/printing-press/SKILL.md` (Phase 0.7 / data layer section)
- **Acceptance criteria:**
  - positive: Phase 0.7 output includes metadata fields (notes, tags, labels) in FTS5 spec when the API has them
  - negative: APIs without metadata fields skip this cleanly
- **Scope boundary:** Skill instruction only — does not change generator templates
- **Dependencies:** None
- **Complexity:** small

### WU-5: Doctor shows OK for no-auth APIs (from F6)
- **Goal:** Doctor command shows "OK Auth: not required" instead of "WARN" for no-auth APIs
- **Target:** Generator templates in `internal/generator/` (doctor.go template)
- **Acceptance criteria:**
  - positive: CLI generated with `auth: none` shows "OK Auth: not required"
  - negative: CLI generated with auth still shows "WARN" when no token is configured
- **Dependencies:** None
- **Complexity:** small

## Anti-patterns
- Syncing "today only" by default makes the data layer useless on first run. Sync should either prompt for a date range or default to a reasonable lookback (e.g., last 30 days).
- The verify tool treating missing required args as test failures inflates the failure count and hides real issues in the noise.

## What the Printing Press Got Right
- **Clean generation from internal YAML spec.** The spec format handled ESPN's parameterized paths (/{sport}/{league}/resource) without issues. All 7 quality gates passed on first generation.
- **Domain-specific store infrastructure.** The generated store.go had the right structure (WAL mode, FTS5 virtual tables, sync_state) even though the domain tables needed to be filled in. The scaffolding was solid.
- **Agent-native defaults.** JSON-when-piped, --select, --dry-run, --compact, --agent flag all worked out of the box. No manual wiring needed.
- **Polish worker.** Autonomously improved examples, fixed verify compat, rewrote README — all without breaking any existing functionality.
- **92/100 scorecard.** The generator's baseline infrastructure is strong enough that manual improvements push into Grade A territory quickly.
