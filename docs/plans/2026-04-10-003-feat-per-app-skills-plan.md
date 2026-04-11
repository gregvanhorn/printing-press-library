---
title: "feat: Generate per-app skills for individual CLI tools"
type: feat
status: active
date: 2026-04-10
origin: docs/brainstorms/per-app-skills-requirements.md
---

# feat: Generate per-app skills for individual CLI tools

## Overview

Add a generator that produces individual per-app skills (`/pp-espn`, `/pp-kalshi`, etc.) from `registry.json` and per-app manifests. Each skill is a self-contained scoped router handling CLI installation, MCP setup, auth guidance, and direct usage for one app. The existing `/ppl` skill remains as the discovery/catalog layer.

## Problem Frame

Agents can only access Printing Press CLIs through the single `/ppl` router skill. This prevents discoverability from skill listings and blocks cross-skill composition (e.g., combining ESPN game data with Kalshi prediction markets). Individual skills with enriched descriptions solve both problems. (see origin: `docs/brainstorms/per-app-skills-requirements.md`)

## Requirements Trace

- R1. Each app gets its own skill as a scoped router (install CLI, install MCP, auth, direct use)
- R2. Skills follow PPL conventions (`go install`, `claude mcp add`, `--agent` flag, exit codes)
- R3. Each skill is fully self-contained — no dependency on `/ppl`
- R4. Descriptions combine registry description + domain commands from `--help`
- R5. Descriptions enable cross-skill composition from descriptions alone
- R6. Generator script in this repo reads registry + manifests and writes SKILL.md files
- R7. Generator uses `--help` for enriched descriptions; degrades gracefully when binary unavailable
- R8. Generator is the recommended way to produce skills; generated files can be hand-edited
- R9. Naming: `pp-<name>` with `-pp-cli` suffix stripped
- R10. `/ppl` remains for discovery; per-app skills handle direct usage

## Scope Boundaries

- Generator lives in this repo only — no changes to cli-printing-press
- No LLM-based description generation
- No CI automation — generator is a manual local script, re-run when `registry.json` or manifests change before committing
- `/ppl` is not modified — discovery and catalog browsing unchanged

### Deferred to Separate Tasks

- CI automation to auto-generate skills on merge: future iteration
- Adding `cli_binary` field to registry schema to eliminate manifest lookups: separate PR

## Context & Research

### Relevant Code and Patterns

- `skills/ppl/SKILL.md` — existing unified router skill (167 lines, 5 modes). Per-app skills reuse Modes 2-4 (Install CLI, Install MCP, Explicit Use) scoped to one app
- `registry.json` — 13 entries with `name`, `category`, `api`, `description`, `path`, and optional `mcp` block
- `library/<category>/<name>/.printing-press.json` — per-app manifests with `cli_name` (present in all 13) and `mcp_binary` (present in only ~3 of 13 — most manifests lack this field)
- `library/<category>/<name>/internal/cli/root.go` — Cobra root command. All standard CLIs share identical framework structure with `DO NOT EDIT` header
- `.claude-plugin/plugin.json` — `"skills": "./skills/"` auto-discovers all subdirectories containing SKILL.md

### Institutional Learnings

- Binary name derivation heuristic (try `<name>-pp-cli`, fall back to bare name) is documented as "temporary until `cli_binary` field is added to registry." Per-app skills can bypass this entirely by reading `.printing-press.json` manifests which have authoritative `cli_name`
- `hubspot-pp-cli` registry entry has incorrect path (`library/sales-and-crm/hubspot-pp-cli` vs. actual `library/sales-and-crm/hubspot/`). Must be fixed first (see origin: Dependencies/Assumptions)
- All standard Printing Press CLIs share framework commands that must be filtered from domain-specific commands during `--help` enrichment
- `allowed-tools: "Read Bash"` is the established convention for PPL skills

## Key Technical Decisions

- **Go generator over shell/Python**: The repo is pure Go with no scripting infrastructure. A Go program at `tools/generate-skills/` uses stdlib only (`encoding/json`, `os/exec`, `text/template`), matches the toolchain, and has zero external dependencies. It gets its own `go.mod` since the repo root has no module.
- **Layered metadata resolution**: For each app, resolve fields using this precedence: (1) `.printing-press.json` manifest for `cli_name`, (2) registry `mcp.binary` for MCP binary name (most manifests lack `mcp_binary`), (3) registry `mcp.auth_type` and `mcp.env_vars` for auth info (most manifests lack auth fields), (4) registry heuristic (append `-pp-cli`) as final fallback for `cli_name` if manifest is missing. Note: `agent-capture` manifest declares `cli_name: "agent-capture-pp-cli"` but the actual binary is `agent-capture` — the generator should validate that `cmd/<cli_name>/` exists in the source tree and fall back to the bare name if not.
- **Embed all metadata in each SKILL.md**: Each generated skill hardcodes its app's install path, binary name, MCP config, auth env vars, and enriched description. No runtime registry reads needed. At ~50-60 lines per skill, the duplication cost is low and keeps each skill fully self-contained (R3).
- **Hardcoded framework command filter**: Rather than dynamically detecting shared commands, maintain a static list of known framework commands to exclude during `--help` parsing. The list is stable (same Cobra template across all CLIs) and a false positive just means a slightly less rich description.
- **Separate template file**: The SKILL.md template lives at `tools/generate-skills/skill-template.md` rather than being embedded in Go code. Easier to iterate on the template without touching generator logic.

## Open Questions

### Resolved During Planning

- **Script language**: Go (`tools/generate-skills/main.go`) — consistent with repo, stdlib only, no external deps
- **`--help` parsing strategy**: Parse "Available Commands:" section line-by-line. Each line has format: `  <command>  <description>`. Filter against hardcoded framework commands list. Collect remaining as domain commands with descriptions
- **Shared vs. duplicated content**: Embed everything per skill. ~50-60 lines each is lean enough that duplication is acceptable and R3 self-containment is preserved
- **Naming collision** (`/pp-slack` vs `/slack`): Descriptions include "Printing Press CLI" phrasing to disambiguate from other plugins' skills
- **Binary name source**: Layered precedence — manifest `cli_name` → validate `cmd/` dir exists → bare name fallback → registry heuristic. Registry is primary for `mcp_binary`, `auth_type`, `env_vars` (most manifests lack these)

### Deferred to Implementation

- Exact enriched description wording per app — depends on `--help` output at generation time
- `agent-capture` binary name resolution — manifest declares `agent-capture-pp-cli` but actual binary is `agent-capture`. The `cmd/` directory validation in the generator's layered precedence handles this, but the exact fallback behavior should be confirmed at implementation time

## Output Structure

```
tools/
  generate-skills/
    main.go                  # Generator program
    go.mod                   # Standalone module (stdlib only)
    skill-template.md        # SKILL.md template with Go template syntax
skills/
  ppl/                       # Existing, unchanged
    SKILL.md
    registry.json -> ../../registry.json
  pp-espn/                   # Generated
    SKILL.md
  pp-kalshi/                 # Generated
    SKILL.md
  pp-linear/                 # Generated
    SKILL.md
  pp-dominos/                # Generated (stripped from dominos-pp-cli)
    SKILL.md
  ... (13 total per-app skill directories)
```

## High-Level Technical Design

> *This illustrates the intended approach and is directional guidance for review, not implementation specification. The implementing agent should treat it as context, not code to reproduce.*

```
Generator flow:
  1. Read registry.json → list of app entries
  2. For each entry:
     a. Derive skill name: strip "-pp-cli" suffix, prepend "pp-"
     b. Resolve metadata (layered precedence):
        - cli_name:  manifest .printing-press.json → validate cmd/<cli_name>/ exists
                     → if dir missing, try bare name → registry heuristic as last resort
        - mcp_binary: registry mcp.binary (most manifests lack this field)
        - auth_type:  registry mcp.auth_type → default "none" if no mcp block
        - env_vars:   registry mcp.env_vars → default empty if no mcp block
     c. Try: exec "<cli_name> --help" → parse domain commands
        On failure: set DomainCommands to nil; template omits domain keywords
        from description (uses registry description only)
     d. Build enriched description: registry description + domain command keywords
     e. Execute skill-template.md with app context → write skills/pp-<name>/SKILL.md
  3. Report: N skills generated, M with enriched descriptions, K degraded

Template variables:
  SkillName       — pp-<name> (e.g., pp-espn)
  APIName         — human name (e.g., ESPN)
  Description     — registry description
  EnrichedDesc    — description + domain keywords for composition
  CLIBinary       — resolved via layered precedence (see above)
  InstallPath     — registry path (e.g., library/media-and-entertainment/espn)
  HasMCP          — boolean (true when registry entry has mcp block)
  MCPBinary       — from registry mcp.binary
  AuthType        — from registry mcp.auth_type | default "none"
  EnvVars         — from registry mcp.env_vars | default empty
  DomainCommands  — filtered command list from --help (nil when degraded)

Framework commands to filter:
  api, auth, completion, doctor, export, help, import, load,
  orphans, sql, stale, sync, version, workflow
  (Note: "search" excluded from filter — it is domain-relevant across CLIs)
```

## Implementation Units

- [x] **Unit 1: Fix HubSpot registry path**

**Goal:** Correct the `hubspot-pp-cli` entry in `registry.json` so the path matches the actual directory.

**Requirements:** Prerequisite for R6 (generator reads registry paths)

**Dependencies:** None

**Files:**
- Modify: `registry.json`

**Approach:**
- Change `"path": "library/sales-and-crm/hubspot-pp-cli"` to `"path": "library/sales-and-crm/hubspot"`
- Verify the directory exists at the corrected path

**Patterns to follow:**
- Other registry entries use the actual directory name (e.g., `library/media-and-entertainment/espn`)

**Test scenarios:**
- Happy path: after fix, `library/sales-and-crm/hubspot/` resolves and contains `cmd/`, `internal/`, `.printing-press.json`

**Verification:**
- `ls library/sales-and-crm/hubspot/.printing-press.json` succeeds
- No other registry entry has a path mismatch (spot-check 2-3 others)

---

- [x] **Unit 2: Create SKILL.md template**

**Goal:** Design the per-app skill template that the generator stamps out for each app.

**Requirements:** R1, R2, R3, R4, R5

**Dependencies:** None (can be developed in parallel with Unit 1)

**Files:**
- Create: `tools/generate-skills/skill-template.md`

**Approach:**
- Template has 3 operating modes (no Discovery or Semantic Use — those stay in `/ppl`):
  1. CLI Installation — `go install` with hardcoded path and binary name
  2. MCP Installation — `claude mcp add` with env vars (conditional: skip if no MCP)
  3. Direct Use — `--help` discovery + `--agent` execution + exit code handling
- Frontmatter includes: `name` (pp-<name>), `description` (enriched), `argument-hint`, `allowed-tools: "Read Bash"`
- Description field combines registry description with domain command keywords and trigger phrases. Include "Printing Press CLI" to disambiguate from other plugins
- Auth guidance varies by `auth_type`: `none` (no auth section), `api_key` (export env var + `auth set-token`), `composed` (browser login via `auth login`)
- Target length: 50-60 lines per generated skill
- All app-specific values are Go template variables (e.g., `{{.CLIBinary}}`, `{{.InstallPath}}`)

**Patterns to follow:**
- `skills/ppl/SKILL.md` — frontmatter format, Mode 2/3/4 structure, exit code table, `--agent` flag convention
- Existing skill descriptions use trigger phrases for discoverability

**Test scenarios:**
- Happy path: template renders correctly for an `api_key` auth app with MCP (e.g., Kalshi — has api_key auth, 89 MCP tools, full mcp_ready)
- Edge case: template renders correctly for a `none` auth app (e.g., ESPN — no auth section, no env vars)
- Edge case: template renders correctly for an app with no MCP block (e.g., agent-capture — MCP section omitted entirely)
- Edge case: template renders correctly for a `composed` auth app (e.g., Dominos — browser login flow)
- Edge case: template renders correctly for a `partial` MCP readiness app (e.g., Pagliacci — shows public_tool_count vs total)
- Happy path: enriched description with domain commands is agent-parseable and contains composition-friendly keywords

**Verification:**
- Manually render the template for ESPN, Kalshi, and agent-capture and confirm each produces a valid, readable SKILL.md
- Each rendered skill covers all 3 modes and handles its auth type correctly

---

- [x] **Unit 3: Build the generator**

**Goal:** Create a Go program that reads registry + manifests + optional CLI help output and produces SKILL.md files for all apps.

**Requirements:** R6, R7, R8, R9

**Dependencies:** Unit 2 (template must exist)

**Files:**
- Create: `tools/generate-skills/main.go`
- Create: `tools/generate-skills/go.mod`

**Approach:**
- Standalone Go module (`module generate-skills`) with stdlib-only dependencies
- Program flow:
  1. Read `registry.json` from repo root (path relative to working directory)
  2. For each entry: derive skill name (strip `-pp-cli`, prepend `pp-`), resolve metadata via layered precedence (manifest `cli_name` → validate `cmd/` dir → bare name fallback; registry for `mcp.binary`, `mcp.auth_type`, `mcp.env_vars`)
  3. Try running `<cli_name> --help`, parse "Available Commands:" section, filter framework commands. Before generating all 13 skills, audit the filter list against at least one live `--help` output to catch any missing framework commands
  4. Build template context with all variables, execute template, write to `skills/pp-<name>/SKILL.md`
  5. Print summary: N generated, M enriched, K degraded
- Framework command filter: hardcoded string set of known framework commands (see High-Level Technical Design for the list)
- Graceful degradation (R7): when binary not found or `--help` fails, set `DomainCommands` to nil. Template conditionally omits domain command keywords from description. Log a warning but continue
- Create skill directory if it doesn't exist; overwrite existing SKILL.md

**Patterns to follow:**
- Go `text/template` for template execution
- Go `os/exec` for running CLI `--help`
- Go `encoding/json` for registry and manifest parsing

**Test scenarios:**
- Happy path: generator reads registry.json with 13 entries and produces 13 skill directories under `skills/`
- Happy path: generated skill for ESPN has enriched description including domain commands like "boxscore", "scoreboard", "standings"
- Happy path: generated skill for Kalshi includes MCP install section with `KALSHI_API_KEY` env var
- Edge case: CLI binary not installed — generator logs warning and produces skill with registry-only description
- Edge case: `--help` returns non-zero exit code — treated same as binary not found (graceful degradation)
- Edge case: `.printing-press.json` missing for an app — fall back to registry heuristic for binary name (append `-pp-cli`)
- Edge case: registry entry with `-pp-cli` suffix (dominos-pp-cli) produces skill named `pp-dominos`, not `pp-dominos-pp-cli`
- Edge case: app with no MCP block (agent-capture) produces skill with MCP section omitted
- Edge case: manifest `cli_name` doesn't match actual binary (agent-capture) — generator detects missing `cmd/` dir and falls back to bare name
- Error path: `registry.json` not found — exit with clear error message
- Error path: template file not found — exit with clear error message

**Verification:**
- `cd tools/generate-skills && go build .` succeeds with no external dependencies (note: no root `go.mod` exists, so `./` relative paths from repo root won't work)
- From repo root: `go run ./tools/generate-skills/main.go` produces 13 `skills/pp-*/SKILL.md` files
- Each generated file has valid SKILL.md frontmatter (name, description, argument-hint, allowed-tools)

---

- [x] **Unit 4: Generate all skills, commit, and validate**

**Goal:** Run the generator to produce all 13 per-app skills, commit them as static artifacts, and verify they work as a Claude Code plugin. Generated files are committed to the repo so users get them on plugin install. Re-run the generator and re-commit when registry or manifests change.

**Requirements:** R1-R10 (end-to-end validation)

**Dependencies:** Units 1, 2, 3

**Files:**
- Create: `skills/pp-espn/SKILL.md` (generated)
- Create: `skills/pp-kalshi/SKILL.md` (generated)
- Create: `skills/pp-linear/SKILL.md` (generated)
- Create: `skills/pp-dominos/SKILL.md` (generated)
- Create: `skills/pp-hubspot/SKILL.md` (generated)
- Create: `skills/pp-cal-com/SKILL.md` (generated)
- Create: `skills/pp-dub/SKILL.md` (generated)
- Create: `skills/pp-slack/SKILL.md` (generated)
- Create: `skills/pp-steam-web/SKILL.md` (generated)
- Create: `skills/pp-agent-capture/SKILL.md` (generated)
- Create: `skills/pp-pagliacci-pizza/SKILL.md` (generated)
- Create: `skills/pp-postman-explore/SKILL.md` (generated)
- Create: `skills/pp-trigger-dev/SKILL.md` (generated)

**Approach:**
- Run `go run ./tools/generate-skills/main.go` from repo root
- Review generated skills for correctness: spot-check frontmatter, install paths, auth sections, MCP sections
- Verify the plugin loads all new skills by checking skill count
- Test at least one skill end-to-end: invoke `/pp-espn` and confirm it can guide CLI installation and usage

**Test scenarios:**
- Happy path: all 13 skills generated with no errors
- Happy path: `/pp-espn` skill description contains domain keywords like "scoreboard", "standings", "boxscore"
- Happy path: `/pp-kalshi` skill description contains "prediction markets", "events", "portfolio"
- Happy path: `/pp-agent-capture` skill correctly omits MCP installation section
- Happy path: `/pp-dominos` skill shows `composed` auth flow (browser login)
- Happy path: `/pp-pagliacci-pizza` skill notes partial MCP readiness
- Integration: invoking `/pp-espn install cli` in Claude Code triggers correct `go install` command
- Integration: invoking `/pp-espn` with a sports query runs the CLI with `--agent` flag

**Verification:**
- 13 `skills/pp-*/SKILL.md` files exist, each with valid frontmatter
- Plugin loads all skills (existing `/ppl` + 13 new per-app skills = 14 total)
- At least one skill works end-to-end for install and usage

## System-Wide Impact

- **Interaction graph:** No callbacks or middleware. Skills are static SKILL.md files discovered by Claude Code's plugin loader. The generator runs offline and writes files.
- **Error propagation:** Generator errors are local (warnings for missing binaries, fatal only for missing registry/template). Generated skills use the same exit code handling as `/ppl`.
- **State lifecycle risks:** None. Generated skills are stateless markdown files. No caching, no partial writes (each file is atomic write-or-skip).
- **API surface parity:** Per-app skills expose the same install/MCP/usage capabilities as `/ppl` Modes 2-4 but scoped to one app. No new capabilities are added.
- **Unchanged invariants:** The `/ppl` skill is not modified. Its discovery, catalog, and semantic routing modes continue to work as before. The `registry.json` schema is unchanged (only the HubSpot path fix in Unit 1).

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Not all 13 CLI binaries installed on generator machine | R7 graceful degradation — produces skill with registry-only description. Log which apps were degraded. |
| `.printing-press.json` schema varies across apps — most lack `mcp_binary` and auth fields | Layered precedence: manifest for `cli_name`, registry for `mcp.binary`/`auth_type`/`env_vars`, heuristic as last resort. |
| `agent-capture` manifest declares wrong `cli_name` (`agent-capture-pp-cli` vs actual `agent-capture`) | Generator validates `cmd/<cli_name>/` dir exists; falls back to bare name when dir is missing. |
| Framework command list drifts as new shared commands are added | Filter list is easy to update. A missed framework command just means slightly noisier descriptions — not a functional failure. |
| Context window cost if agent loads many per-app skills | Skills are ~50-60 lines each. At 13 skills, only loaded on demand (not all at once). Descriptions appear in skill listings (~100 tokens each). |
| `agent-capture` has non-standard CLI structure | Template conditional sections handle no-MCP case. The simpler CLI structure doesn't affect skill content since skills use `--help` discovery. |

## Sources & References

- **Origin document:** [docs/brainstorms/per-app-skills-requirements.md](docs/brainstorms/per-app-skills-requirements.md)
- Related plan: [docs/plans/2026-04-10-002-feat-claude-code-plugin-plan.md](docs/plans/2026-04-10-002-feat-claude-code-plugin-plan.md) — established plugin infrastructure
- Existing skill: `skills/ppl/SKILL.md` — template reference for conventions
- Registry: `registry.json` — primary data source for generator
