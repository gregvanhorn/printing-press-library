---
name: pp-pokeapi
description: "PokeAPI as an agent-ready Pokemon knowledge graph, not just endpoint wrappers. Trigger phrases: `look up a pokemon`, `pokemon evolution`, `pokemon type matchup`, `pokemon team coverage`, `what moves can this pokemon learn`."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["pokeapi-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/pokeapi/cmd/pokeapi-pp-cli@latest","bins":["pokeapi-pp-cli"],"label":"Install via go install"}]}}'
---

# PokeAPI — Printing Press CLI

This CLI keeps the full official PokeAPI REST surface while adding graph commands for the workflows people actually ask about: profiles, evolutions, moves, matchups, and team coverage. It is public-API friendly and requires no authentication for normal reads.

## When to Use This CLI

Use PokeAPI when a user asks about Pokemon, moves, evolutions, types, or team composition. Prefer the graph commands for common questions; drop to v2 endpoint commands when you need raw API resources.

## When Not to Use This CLI

Do not activate this CLI for requests that require creating, updating, deleting, publishing, commenting, upvoting, inviting, ordering, sending messages, booking, purchasing, or changing remote state. This printed CLI exposes read-only commands for inspection, export, sync, and analysis.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Pokemon graph workflows

- **`pokemon profile`** — Build an agent-ready Pokemon profile by combining core pokemon data, species metadata, type names, abilities, stats, and move counts.

  _Use this when a user asks what a Pokemon is, what it does, or needs a compact structured summary._

  ```bash
  pokeapi-pp-cli pokemon profile pikachu --json
  ```
- **`pokemon evolution`** — Resolve a Pokemon's species and evolution chain into a readable evolution path.

  _Use this when a user asks what a Pokemon evolves into or from._

  ```bash
  pokeapi-pp-cli pokemon evolution eevee --json
  ```

### Battle planning

- **`pokemon matchups`** — Summarize type weaknesses, resistances, immunities, and offensive coverage for a Pokemon.

  _Use this for battle planning, weakness analysis, and type coverage questions._

  ```bash
  pokeapi-pp-cli pokemon matchups charizard --json
  ```
- **`pokemon moves`** — List and filter a Pokemon's moves by learn method, version group, and level learned.

  _Use this when a user asks what moves a Pokemon learns and how._

  ```bash
  pokeapi-pp-cli pokemon moves bulbasaur --method level-up --version-group red-blue --json
  ```
- **`team coverage`** — Analyze a comma-separated Pokemon team for shared weaknesses, resistances, immunities, and offensive type coverage.

  _Use this when a user asks whether a team is balanced or has dangerous shared weaknesses._

  ```bash
  pokeapi-pp-cli team coverage pikachu,charizard,blastoise --json
  ```

## Command Reference

**ability** — Manage ability

- `pokeapi-pp-cli ability list` — Abilities provide passive effects for Pokémon in battle or in the overworld. Pokémon have multiple possible...
- `pokeapi-pp-cli ability retrieve` — Abilities provide passive effects for Pokémon in battle or in the overworld. Pokémon have multiple possible...

**berry** — Manage berry

- `pokeapi-pp-cli berry list` — List berries
- `pokeapi-pp-cli berry retrieve` — Get a berry

**berry-firmness** — Manage berry firmness

- `pokeapi-pp-cli berry-firmness list` — List berry firmness
- `pokeapi-pp-cli berry-firmness retrieve` — Get berry by firmness

**berry-flavor** — Manage berry flavor

- `pokeapi-pp-cli berry-flavor list` — List berry flavors
- `pokeapi-pp-cli berry-flavor retrieve` — Get berries by flavor

**characteristic** — Manage characteristic

- `pokeapi-pp-cli characteristic list` — List charecterictics
- `pokeapi-pp-cli characteristic retrieve` — Get characteristic

**contest-effect** — Manage contest effect

- `pokeapi-pp-cli contest-effect list` — List contest effects
- `pokeapi-pp-cli contest-effect retrieve` — Get contest effect

**contest-type** — Manage contest type

- `pokeapi-pp-cli contest-type list` — List contest types
- `pokeapi-pp-cli contest-type retrieve` — Get contest type

**egg-group** — Manage egg group

- `pokeapi-pp-cli egg-group list` — List egg groups
- `pokeapi-pp-cli egg-group retrieve` — Get egg group

**encounter-condition** — Manage encounter condition

- `pokeapi-pp-cli encounter-condition list` — List encounter conditions
- `pokeapi-pp-cli encounter-condition retrieve` — Get encounter condition

**encounter-condition-value** — Manage encounter condition value

- `pokeapi-pp-cli encounter-condition-value list` — List encounter condition values
- `pokeapi-pp-cli encounter-condition-value retrieve` — Get encounter condition value

**encounter-method** — Manage encounter method

- `pokeapi-pp-cli encounter-method list` — List encounter methods
- `pokeapi-pp-cli encounter-method retrieve` — Get encounter method

**evolution-chain** — Manage evolution chain

- `pokeapi-pp-cli evolution-chain list` — List evolution chains
- `pokeapi-pp-cli evolution-chain retrieve` — Get evolution chain

**evolution-trigger** — Manage evolution trigger

- `pokeapi-pp-cli evolution-trigger list` — List evolution triggers
- `pokeapi-pp-cli evolution-trigger retrieve` — Get evolution trigger

**gender** — Manage gender

- `pokeapi-pp-cli gender list` — List genders
- `pokeapi-pp-cli gender retrieve` — Get gender

**generation** — Manage generation

- `pokeapi-pp-cli generation list` — List genrations
- `pokeapi-pp-cli generation retrieve` — Get genration

**growth-rate** — Manage growth rate

- `pokeapi-pp-cli growth-rate list` — List growth rates
- `pokeapi-pp-cli growth-rate retrieve` — Get growth rate

**item** — An item is an object in the games which the player can pick up, keep in their bag, and use in some manner. They have various uses, including healing, powering up, helping catch Pokémon, or to access a new area.

- `pokeapi-pp-cli item list` — List items
- `pokeapi-pp-cli item retrieve` — Get item

**item-attribute** — Manage item attribute

- `pokeapi-pp-cli item-attribute list` — List item attributes
- `pokeapi-pp-cli item-attribute retrieve` — Get item attribute

**item-category** — Manage item category

- `pokeapi-pp-cli item-category list` — List item categories
- `pokeapi-pp-cli item-category retrieve` — Get item category

**item-fling-effect** — Manage item fling effect

- `pokeapi-pp-cli item-fling-effect list` — List item fling effects
- `pokeapi-pp-cli item-fling-effect retrieve` — Get item fling effect

**item-pocket** — Manage item pocket

- `pokeapi-pp-cli item-pocket list` — List item pockets
- `pokeapi-pp-cli item-pocket retrieve` — Get item pocket

**language** — Manage language

- `pokeapi-pp-cli language list` — List languages
- `pokeapi-pp-cli language retrieve` — Get language

**location** — Locations that can be visited within the games. Locations make up sizable portions of regions, like cities or routes.

- `pokeapi-pp-cli location list` — List locations
- `pokeapi-pp-cli location retrieve` — Get location

**location-area** — Manage location area

- `pokeapi-pp-cli location-area list` — List location areas
- `pokeapi-pp-cli location-area retrieve` — Get location area

**machine** — Machines are the representation of items that teach moves to Pokémon. They vary from version to version, so it is not certain that one specific TM or HM corresponds to a single Machine.

- `pokeapi-pp-cli machine list` — List machines
- `pokeapi-pp-cli machine retrieve` — Get machine

**move** — Moves are the skills of Pokémon in battle. In battle, a Pokémon uses one move each turn. Some moves (including those learned by Hidden Machine) can be used outside of battle as well, usually for the purpose of removing obstacles or exploring new areas.

- `pokeapi-pp-cli move list` — List moves
- `pokeapi-pp-cli move retrieve` — Get move

**move-ailment** — Manage move ailment

- `pokeapi-pp-cli move-ailment list` — List move meta ailments
- `pokeapi-pp-cli move-ailment retrieve` — Get move meta ailment

**move-battle-style** — Manage move battle style

- `pokeapi-pp-cli move-battle-style list` — List move battle styles
- `pokeapi-pp-cli move-battle-style retrieve` — Get move battle style

**move-category** — Manage move category

- `pokeapi-pp-cli move-category list` — List move meta categories
- `pokeapi-pp-cli move-category retrieve` — Get move meta category

**move-damage-class** — Manage move damage class

- `pokeapi-pp-cli move-damage-class list` — List move damage classes
- `pokeapi-pp-cli move-damage-class retrieve` — Get move damage class

**move-learn-method** — Manage move learn method

- `pokeapi-pp-cli move-learn-method list` — List move learn methods
- `pokeapi-pp-cli move-learn-method retrieve` — Get move learn method

**move-target** — Manage move target

- `pokeapi-pp-cli move-target list` — List move targets
- `pokeapi-pp-cli move-target retrieve` — Get move target

**nature** — Manage nature

- `pokeapi-pp-cli nature list` — List natures
- `pokeapi-pp-cli nature retrieve` — Get nature

**pal-park-area** — Manage pal park area

- `pokeapi-pp-cli pal-park-area list` — List pal park areas
- `pokeapi-pp-cli pal-park-area retrieve` — Get pal park area

**pokeathlon-stat** — Manage pokeathlon stat

- `pokeapi-pp-cli pokeathlon-stat list` — List pokeathlon stats
- `pokeapi-pp-cli pokeathlon-stat retrieve` — Get pokeathlon stat

**pokedex** — Manage pokedex

- `pokeapi-pp-cli pokedex list` — List pokedex
- `pokeapi-pp-cli pokedex retrieve` — Get pokedex

**pokemon** — Pokémon are the creatures that inhabit the world of the Pokémon games. They can be caught using Pokéballs and trained by battling with other Pokémon. Each Pokémon belongs to a specific species but may take on a variant which makes it differ from other Pokémon of the same species, such as base stats, available abilities and typings. See [Bulbapedia](http://bulbapedia.bulbagarden.net/wiki/Pok%C3%A9mon_(species)) for greater detail.

- `pokeapi-pp-cli pokemon list` — List pokemon
- `pokeapi-pp-cli pokemon retrieve` — Get pokemon

**pokemon-color** — Manage pokemon color

- `pokeapi-pp-cli pokemon-color list` — List pokemon colors
- `pokeapi-pp-cli pokemon-color retrieve` — Get pokemon color

**pokemon-form** — Manage pokemon form

- `pokeapi-pp-cli pokemon-form list` — List pokemon forms
- `pokeapi-pp-cli pokemon-form retrieve` — Get pokemon form

**pokemon-habitat** — Manage pokemon habitat

- `pokeapi-pp-cli pokemon-habitat list` — List pokemom habitas
- `pokeapi-pp-cli pokemon-habitat retrieve` — Get pokemom habita

**pokemon-shape** — Manage pokemon shape

- `pokeapi-pp-cli pokemon-shape list` — List pokemon shapes
- `pokeapi-pp-cli pokemon-shape retrieve` — Get pokemon shape

**pokemon-species** — Manage pokemon species

- `pokeapi-pp-cli pokemon-species list` — List pokemon species
- `pokeapi-pp-cli pokemon-species retrieve` — Get pokemon species

**region** — Manage region

- `pokeapi-pp-cli region list` — List regions
- `pokeapi-pp-cli region retrieve` — Get region

**stat** — Manage stat

- `pokeapi-pp-cli stat list` — List stats
- `pokeapi-pp-cli stat retrieve` — Get stat

**super-contest-effect** — Manage super contest effect

- `pokeapi-pp-cli super-contest-effect list` — List super contest effects
- `pokeapi-pp-cli super-contest-effect retrieve` — Get super contest effect

**type** — Manage type

- `pokeapi-pp-cli type list` — List types
- `pokeapi-pp-cli type retrieve` — Get types

**version** — Manage version

- `pokeapi-pp-cli version list` — List versions
- `pokeapi-pp-cli version retrieve` — Get version

**version-group** — Manage version group

- `pokeapi-pp-cli version-group list` — List version groups
- `pokeapi-pp-cli version-group retrieve` — Get version group


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
pokeapi-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes


### Build a Pokemon profile

```bash
pokeapi-pp-cli pokemon profile pikachu --json
```

Combines core Pokemon fields into a compact agent-readable summary.

### Check battle matchups

```bash
pokeapi-pp-cli pokemon matchups charizard --json
```

Combines type damage relations into weaknesses, resistances, and offensive coverage.

## Auth Setup

No authentication required.

Run `pokeapi-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  pokeapi-pp-cli ability list --agent --select id,name,status
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
pokeapi-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
pokeapi-pp-cli feedback --stdin < notes.txt
pokeapi-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.pokeapi-pp-cli/feedback.jsonl`. They are never POSTed unless `POKEAPI_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `POKEAPI_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

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
pokeapi-pp-cli profile save briefing --json
pokeapi-pp-cli --profile briefing ability list
pokeapi-pp-cli profile list --json
pokeapi-pp-cli profile show briefing
pokeapi-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `pokeapi-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/pokeapi/cmd/pokeapi-pp-cli@latest
   ```
3. Verify: `pokeapi-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/pokeapi/cmd/pokeapi-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add pokeapi-pp-mcp -- pokeapi-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which pokeapi-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   pokeapi-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `pokeapi-pp-cli <command> --help`.
