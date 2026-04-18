# Printing Press Library

Printing Press Library is the published catalog for the [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press): agent-friendly CLIs, MCP servers, and plugin skills generated from real APIs and shipped as installable tools.

This repository currently includes 21 CLIs, 19 MCP servers, one mega-skill (`/ppl`), and 21 focused `pp-*` skills.

## Start Here

This repo is its own Claude Code plugin marketplace. Add the marketplace, then install the plugin:

```text
/plugin marketplace add mvanhorn/printing-press-library
/plugin install printing-press-library@printing-press-library
```

After install, the main router skill is:

```text
/ppl
```

That naming split is intentional:

- The repository and plugin are named `printing-press-library`
- The mega-skill you actually use is `/ppl`

If you also want to generate new tools from API specs, install the source project too:

```text
/plugin marketplace add mvanhorn/cli-printing-press
/plugin install cli-printing-press@cli-printing-press
```

## Quick Examples

```text
/ppl
/ppl sports scores
/ppl install espn cli
/ppl install espn mcp
/ppl espn lakers score
/ppl linear my open issues
/pp-espn lakers score
/pp-weather-goat phoenix forecast
```

Use `/ppl` when you want discovery, routing, or installation. Use a focused `pp-*` skill when you already know the tool you want.

## What This Repo Ships

- [`plugin/skills/ppl/SKILL.md`](plugin/skills/ppl/SKILL.md): the mega-skill that browses the catalog, installs CLIs and MCP servers, and routes user requests to the right tool
- [`plugin/skills/`](plugin/skills/): one focused skill per published tool, such as `pp-espn`, `pp-linear`, and `pp-weather-goat`
- [`library/`](library/): the actual Go modules for each CLI and MCP server
- [`registry.json`](registry.json): the machine-readable source of truth for the published catalog

The top-level README is the human guide. `registry.json` is the authoritative index.

## Catalog

Current published tools, grouped by how people usually reach for them. Each row links to the tool source and its focused plugin skill.

| Name | Skill | Auth | MCP | What it does |
|------|-------|------|-----|--------------|
| [`agent-capture`](library/developer-tools/agent-capture/) | [`/pp-agent-capture`](plugin/skills/pp-agent-capture/SKILL.md) | local only | no | Record, screenshot, and convert macOS windows and screens for agent evidence. |
| [`archive-is`](library/media-and-entertainment/archive-is/) | [`/pp-archive-is`](plugin/skills/pp-archive-is/SKILL.md) | none | full | Find and create Archive.today snapshots for URLs. |
| [`cal-com`](library/productivity/cal-com/) | [`/pp-cal-com`](plugin/skills/pp-cal-com/SKILL.md) | API key | full | Manage bookings, schedules, event types, and availability. |
| [`dominos-pp-cli`](library/commerce/dominos-pp-cli/) | [`/pp-dominos`](plugin/skills/pp-dominos/SKILL.md) | browser login | full | Order Domino's, browse menus, and track deliveries. |
| [`dub`](library/marketing/dub/) | [`/pp-dub`](plugin/skills/pp-dub/SKILL.md) | API key | full | Create short links, track analytics, and manage domains. |
| [`espn`](library/media-and-entertainment/espn/) | [`/pp-espn`](plugin/skills/pp-espn/SKILL.md) | none | full | Live scores, standings, schedules, and sports news. |
| [`flightgoat`](library/travel/flightgoat/) | [`/pp-flightgoat`](plugin/skills/pp-flightgoat/SKILL.md) | API key optional | full | Search flights, explore routes, and track flights. |
| [`hackernews`](library/media-and-entertainment/hackernews/) | [`/pp-hackernews`](plugin/skills/pp-hackernews/SKILL.md) | none | full | Browse stories, comments, jobs, and topic slices from Hacker News. |
| [`hubspot-pp-cli`](library/sales-and-crm/hubspot/) | [`/pp-hubspot`](plugin/skills/pp-hubspot/SKILL.md) | API key | full | Work with contacts, companies, deals, tickets, and pipelines. |
| [`instacart`](library/commerce/instacart/) | [`/pp-instacart`](plugin/skills/pp-instacart/SKILL.md) | browser session | no | Search products, manage carts, and shop Instacart from the terminal. |
| [`kalshi`](library/payments/kalshi/) | [`/pp-kalshi`](plugin/skills/pp-kalshi/SKILL.md) | API key | full | Trade markets, inspect portfolios, and analyze odds. |
| [`linear`](library/project-management/linear/) | [`/pp-linear`](plugin/skills/pp-linear/SKILL.md) | API key | full | Manage issues, cycles, teams, and projects with local sync. |
| [`movie-goat`](library/media-and-entertainment/movie-goat/) | [`/pp-movie-goat`](plugin/skills/pp-movie-goat/SKILL.md) | bearer token | full | Compare movie ratings, streaming availability, and recommendations. |
| [`pagliacci-pizza`](library/food-and-dining/pagliacci-pizza/) | [`/pp-pagliacci-pizza`](plugin/skills/pp-pagliacci-pizza/SKILL.md) | browser login | partial | Order Pagliacci and browse public menu and store data without login. |
| [`postman-explore`](library/developer-tools/postman-explore/) | [`/pp-postman-explore`](plugin/skills/pp-postman-explore/SKILL.md) | none | full | Search and browse the Postman API Network. |
| [`recipe-goat`](library/food-and-dining/recipe-goat/) | [`/pp-recipe-goat`](plugin/skills/pp-recipe-goat/SKILL.md) | API key | full | Find recipes across trusted sites and pull nutrition context. |
| [`slack`](library/productivity/slack/) | [`/pp-slack`](plugin/skills/pp-slack/SKILL.md) | API key | full | Send messages, search conversations, and monitor channels. |
| [`steam-web`](library/media-and-entertainment/steam-web/) | [`/pp-steam-web`](plugin/skills/pp-steam-web/SKILL.md) | API key | full | Look up Steam players, games, achievements, and stats. |
| [`trigger-dev`](library/developer-tools/trigger-dev/) | [`/pp-trigger-dev`](plugin/skills/pp-trigger-dev/SKILL.md) | API key | full | Monitor runs, trigger tasks, and inspect schedules and failures. |
| [`weather-goat`](library/other/weather-goat/) | [`/pp-weather-goat`](plugin/skills/pp-weather-goat/SKILL.md) | none | full | Forecasts, alerts, air quality, and activity verdicts. |
| [`yahoo-finance`](library/commerce/yahoo-finance/) | [`/pp-yahoo-finance`](plugin/skills/pp-yahoo-finance/SKILL.md) | none | full | Quotes, charts, fundamentals, options, and watchlists. |

## Installation Paths

### Recommended: let `/ppl` handle it

```text
/ppl install <name> cli
/ppl install <name> mcp
```

That keeps users off of repo-path trivia and lets the skill read from [`registry.json`](registry.json).

### Direct CLI install

You need [Go 1.23+](https://go.dev/dl/).

```bash
go install github.com/mvanhorn/printing-press-library/<path>/cmd/<binary>@latest
```

Examples:

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn/cmd/espn-pp-cli@latest
go install github.com/mvanhorn/printing-press-library/library/project-management/linear/cmd/linear-pp-cli@latest
go install github.com/mvanhorn/printing-press-library/library/other/weather-goat/cmd/weather-goat-pp-cli@latest
```

### Direct MCP install

```bash
go install github.com/mvanhorn/printing-press-library/<path>/cmd/<mcp-binary>@latest
claude mcp add <mcp-binary> -- <mcp-binary>
```

Examples:

```bash
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn/cmd/espn-pp-mcp@latest
claude mcp add espn-pp-mcp -- espn-pp-mcp

go install github.com/mvanhorn/printing-press-library/library/other/weather-goat/cmd/weather-goat-pp-mcp@latest
claude mcp add weather-goat-pp-mcp -- weather-goat-pp-mcp
```

If a tool needs credentials, the focused skill and the tool README document the required environment variables.

## Repo Structure

```text
library/
  <category>/
    <tool>/
      cmd/
        <cli-binary>/
        <mcp-binary>/        # when available
      internal/
      README.md
      go.mod
      .printing-press.json
      .manuscripts/

plugin/
  skills/
    ppl/
      SKILL.md
    pp-*/
      SKILL.md

registry.json
```

Each published tool is self-contained: source code, a local README, provenance metadata, and manuscripts from the printing run.

## What "Endorsed" Means

Every published tool in this repo has passed:

1. Generation from an API spec or captured interface through the CLI Printing Press
2. Validation checks such as build, vet, help, and version
3. Provenance capture through `.printing-press.json` and `.manuscripts/`

Some tools are refined after generation. The generated artifacts remain in the tool directory so the provenance stays inspectable.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
