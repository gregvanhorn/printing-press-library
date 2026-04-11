---
date: 2026-04-10
topic: per-app-skills
---

# Per-App Skills for Printing Press Library

## Problem Frame

The printing-press-library ships a single `/ppl` skill that routes to all 13+ apps. Agents can't discover individual tools from skill listings — they must invoke `/ppl` first and navigate the catalog. Cross-skill composition (e.g., combining ESPN scores with Kalshi prediction markets) doesn't work because agents can't reason about tool capabilities from skill descriptions alone.

Individual per-app skills with rich descriptions unlock two things: direct discoverability (agents see `/pp-espn` in the skill listing and immediately know what it does) and cross-skill composition (agents combine `/pp-espn` + `/pp-kalshi` based on complementary descriptions). `/ppl` remains the catalog for browsing and searching the full library — per-app skills are for agents that already know which tool they need or discover one via the skill listing.

## Requirements

**Skill Content**
- R1. Each app in `registry.json` gets its own skill as a scoped router — handling CLI installation, MCP installation, authentication guidance, and direct CLI usage for that single app.
- R2. Skills follow the same conventions as the existing PPL skill: `go install` for CLIs, `claude mcp add` for MCP servers, `--agent` flag for execution, shared exit code handling.
- R3. Each skill is self-contained. An agent encountering `/pp-espn` for the first time can install, authenticate, and start querying without knowing `/ppl` exists.

**Descriptions for Composition**
- R4. Skill descriptions combine the registry description with data domain signals extracted from the CLI's command tree (via `--help`). For example, ESPN's description would include "game schedules, team matchups, player stats, live odds" alongside the base description.
- R5. Descriptions must be specific enough that an agent can identify composition opportunities across skills from descriptions alone (e.g., ESPN's game data + Kalshi's prediction markets).

**Generation**
- R6. A generation script/template in this repo reads `registry.json` and produces a `SKILL.md` per app. No LLM required — content is formulaic.
- R7. The generator runs each CLI's `--help` to extract command names and data domains for enriched descriptions (R4). When a CLI binary is unavailable or `--help` fails, the generator gracefully degrades to the registry description alone rather than failing the entire run.
- R8. Generation can be run locally. The template is the recommended way to produce and update skills, but generated files are normal SKILL.md files that can be hand-edited if needed.

**Naming and Structure**
- R9. Per-app skills use the naming convention `pp-<name>` where `<name>` is the registry entry name with any `-pp-cli` suffix stripped (e.g., `espn` → `/pp-espn`, `dominos-pp-cli` → `/pp-dominos`, `hubspot-pp-cli` → `/pp-hubspot`).
- R10. The existing `/ppl` skill continues as the discovery and catalog layer. Per-app skills handle direct usage. Both coexist in the same plugin.

## Success Criteria

- An agent with no prior context can invoke `/pp-espn`, install the CLI, and retrieve live sports scores in a single session.
- An agent can combine two per-app skills (e.g., `/pp-espn` + `/pp-kalshi`) to answer a cross-domain question like "what Kalshi prediction markets relate to tomorrow's NBA games?" without the user manually orchestrating between tools.
- Adding a new app to the library and re-running the generator produces a working skill with no manual intervention beyond having the CLI binary available.

## Scope Boundaries

- The generator lives in this repo only — no changes to cli-printing-press or CI infrastructure.
- No LLM-based description generation. Descriptions are assembled from registry metadata + CLI help output.
- Per-app skills do not replace `/ppl`. Discovery and catalog browsing stay in the mega skill.
- The generator does not auto-run on merge. It is a manual local script (CI automation can be added later).

## Key Decisions

- **Scoped router over minimal skill**: Each skill handles install + MCP + auth + usage, not just a pointer to `--help`. This makes each skill fully self-contained.
- **Template in this repo over CI or cli-printing-press**: Keeps generation simple and decoupled. No new infrastructure needed.
- **Keep PPL alongside per-app skills**: PPL becomes the index/catalog, per-app skills are the entries. Clean separation of discovery vs. usage.
- **`pp-` prefix over `ppl-`**: Shorter, maps to "Printing Press" branding, distinct from the `/ppl` catalog skill.
- **Enhanced descriptions from CLI help**: Richer composition surface without manual curation. Requires CLI binaries available during generation.

## Dependencies / Assumptions

- CLI binaries should be installed (or buildable) on the machine running the generator for enriched descriptions (R7). If unavailable, the generator falls back to registry-only descriptions.
- The plugin.json `skills` field (`"./skills/"`) already supports subdirectories — confirmed by examining other installed plugins (compound-engineering, iterative-engineering). No plugin config changes needed.
- The `hubspot-pp-cli` registry entry has an incorrect `path` (`library/sales-and-crm/hubspot-pp-cli` vs. actual `library/sales-and-crm/hubspot/`). This should be fixed before or during implementation, as the generator uses registry paths to locate source code.

## Outstanding Questions

### Deferred to Planning
- [Affects R6][Technical] What is the best format for the generation script — shell, Go, or Python? Should consider what's already in the repo's toolchain.
- [Affects R7][Needs research] How to best parse `--help` output to extract data domain keywords for descriptions. Cobra help output is structured but varies by app complexity. Key challenge: filtering out shared framework commands (api, auth, completion, doctor, export, help, import, sync, version, workflow, etc.) that appear in every CLI, keeping only domain-specific commands.
- [Affects R1][Technical] How much of the PPL skill's SKILL.md content should be shared vs. duplicated per app? A shared reference file could reduce template size but adds indirection. Note: at 13 skills, duplicated boilerplate means ~1K+ lines of near-identical install/MCP/error-handling content — context window cost if multiple skills are loaded in one session.
- [Affects R5][Technical] Some per-app skill names overlap with existing skills from other plugins (e.g., `/pp-slack` vs. the `/slack` plugin skill). Descriptions should disambiguate — e.g., "Printing Press CLI/MCP for the Slack API" vs. browser-based Slack automation.

## Next Steps

`-> /ce:plan` for structured implementation planning
