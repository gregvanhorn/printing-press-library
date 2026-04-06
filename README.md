# Printing Press Library

The curated collection of CLIs and MCP servers built by the [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press).

Every entry in this library was generated from an API spec, verified through the press's quality gates, and submitted via the `/printing-press publish` skill. They're not wrappers — they have local SQLite sync, offline search, workflow commands, and agent-optimized output.

The printing press generates both CLIs and MCP servers from the same spec. CLIs are the efficiency layer — fewer tokens, composable with pipes, works with any shell-based agent. MCP servers are the discovery layer — show up in Claude Desktop, Cursor, and marketplace listings. Use the CLI to set up auth and explore interactively. Use the MCP to let your AI editor call the API.

## Published CLIs

### Works immediately (no auth required)

| CLI | MCP | API | Tools | What it does |
|-----|-----|-----|-------|-------------|
| **[espn-pp-cli](library/media-and-entertainment/espn-pp-cli/)** | espn-pp-mcp | ESPN | 3 | Sports data — scores, stats, standings, schedules, news, odds across 17 sports and 139 leagues. |
| **[postman-explore-pp-cli](library/developer-tools/postman-explore-pp-cli/)** | postman-explore-pp-mcp | Postman Explore | 9 | Search and browse the Postman API Network. |

### API key required

| CLI | MCP | API | Tools | What it does |
|-----|-----|-----|-------|-------------|
| **[dub-pp-cli](library/marketing/dub-pp-cli/)** | dub-pp-mcp | Dub | 53 | Create short links, track analytics, manage domains, and run affiliate programs. |
| **[linear-pp-cli](library/project-management/linear-pp-cli/)** | linear-pp-mcp | Linear | 63 | Issues, cycles, teams, projects via GraphQL. Local sync, stale detection, team health scoring. |
| **[steam-web-pp-cli](library/media-and-entertainment/steam-web-pp-cli/)** | steam-web-pp-mcp | Steam Web | 164 (29 public) | Look up Steam players, games, achievements, friends, and stats. 29 tools work without an API key. |
| **[cal-com-pp-cli](library/productivity/cal-com-pp-cli/)** | cal-com-pp-mcp | Cal.com | 288 | Manage bookings, event types, schedules, and availability. |

### Partial MCP (some tools work without auth)

| CLI | MCP | API | Tools | What it does |
|-----|-----|-----|-------|-------------|
| **[pagliacci-pizza-pp-cli](library/other/pagliacci-pizza-pp-cli/)** | pagliacci-pizza-pp-mcp | Pagliacci Pizza | 41 (10 public) | Order pizza, browse menus, manage rewards. 10 tools (stores, menus, pricing, scheduling) work without login. |

### CLI only (no MCP server)

| CLI | API | What it does |
|-----|-----|-------------|
| **[agent-capture-pp-cli](library/developer-tools/agent-capture-pp-cli/)** | agent-capture | Record, screenshot, and convert macOS windows and screens for AI agent evidence. |

## Install

Each CLI is a standalone Go module with both a CLI and MCP binary. You need [Go 1.23+](https://go.dev/dl/) installed.

### CLI (go install)

```bash
# ESPN — sports scores, stats, standings (no auth)
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn-pp-cli/cmd/espn-pp-cli@latest

# Dub — link management (set DUB_TOKEN env var)
go install github.com/mvanhorn/printing-press-library/library/marketing/dub-pp-cli/cmd/dub-pp-cli@latest

# Linear — project management (set LINEAR_API_KEY env var)
go install github.com/mvanhorn/printing-press-library/library/project-management/linear-pp-cli/cmd/linear-pp-cli@latest

# Cal.com — scheduling (set CAL_COM_TOKEN env var)
go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com-pp-cli/cmd/cal-com-pp-cli@latest

# Steam Web — gaming data (set STEAM_WEB_API_KEY env var)
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/steam-web-pp-cli/cmd/steam-web-pp-cli@latest

# Postman Explore — API network browser (no auth)
go install github.com/mvanhorn/printing-press-library/library/developer-tools/postman-explore-pp-cli/cmd/postman-explore-pp-cli@latest

# Pagliacci Pizza — pizza ordering (browser login for full access)
go install github.com/mvanhorn/printing-press-library/library/other/pagliacci-pizza-pp-cli/cmd/pagliacci-pizza-pp-cli@latest
```

### MCP Server (Claude Desktop / Cursor)

```bash
# Install the MCP binary
go install github.com/mvanhorn/printing-press-library/library/media-and-entertainment/espn-pp-cli/cmd/espn-pp-mcp@latest

# Add to Claude Code
claude mcp add espn-pp-mcp -- espn-pp-mcp

# With API key (example: Dub)
claude mcp add dub-pp-mcp -e DUB_TOKEN=your-key -- dub-pp-mcp

# With API key (example: Steam Web)
claude mcp add steam-web-pp-mcp -e STEAM_WEB_API_KEY=your-key -- steam-web-pp-mcp
```

The binary lands in your `$GOPATH/bin` (or `$HOME/go/bin` by default). Make sure that's on your `PATH`.

### From source

```bash
git clone https://github.com/mvanhorn/printing-press-library.git
cd printing-press-library/library/<category>/<cli-name>
go install ./cmd/<cli-name>       # CLI
go install ./cmd/<cli-name-pp-mcp>   # MCP server (if available)
```

Check each CLI's own README for usage, configuration, and required environment variables.

## Structure

```
library/
  <category>/
    <cli-name>/
      cmd/
        <cli-name>/           # CLI binary
        <cli-name-pp-mcp>/    # MCP server binary
      internal/
      .printing-press.json    # provenance manifest (includes MCP metadata)
      .manuscripts/           # research + verification artifacts
        <run-id>/
          research/
          proofs/
          discovery/
      README.md
      go.mod
      ...
```

CLIs are organized by category. Each CLI folder is self-contained — it includes the full source code, the provenance manifest, and the manuscripts (research briefs, shipcheck results, discovery provenance) from the printing run.

## Categories

| Category | What goes here |
|----------|---------------|
| `developer-tools` | SCM, CI/CD, feature flags, hosting |
| `monitoring` | Error tracking, APM, alerting, product analytics |
| `cloud` | Compute, DNS, CDN, storage, infrastructure |
| `project-management` | Tasks, sprints, issues, roadmaps |
| `productivity` | Docs, wikis, databases, scheduling |
| `social-and-messaging` | Chat, SMS, voice, social, streaming, media |
| `sales-and-crm` | Pipelines, contacts, deals |
| `marketing` | Email campaigns, automation |
| `payments` | Billing, transactions, banking, fintech |
| `auth` | Identity, SSO, user management |
| `commerce` | Storefronts, inventory, orders, shopping |
| `ai` | LLMs, inference, ML, computer vision |
| `devices` | Smart home, wearables, hardware |
| `media-and-entertainment` | Streaming, sports, video, music, content platforms |
| `other` | Anything that doesn't fit above |

## What "Endorsed" Means

Every CLI in this library has passed:

1. **Generation** — Built by the CLI Printing Press from an API spec
2. **Validation** — `go build`, `go vet`, `--help`, and `--version` all pass
3. **Provenance** — `.printing-press.json` manifest and `.manuscripts/` artifacts are present

CLIs may be improved after generation (emboss passes, manual refinements). The manuscripts show what was originally generated, and the diff shows what changed.

## Registry

`registry.json` at the repo root is a machine-readable index of all CLIs with MCP metadata:

```json
{
  "schema_version": 1,
  "entries": [
    {
      "name": "espn-pp-cli",
      "category": "media-and-entertainment",
      "api": "ESPN",
      "description": "ESPN sports CLI with live scores, standings, stats, and offline search",
      "path": "library/media-and-entertainment/espn-pp-cli",
      "mcp": {
        "binary": "espn-pp-mcp",
        "transport": "stdio",
        "tool_count": 3,
        "public_tool_count": 3,
        "auth_type": "none",
        "env_vars": [],
        "mcp_ready": "full"
      }
    }
  ]
}
```

## Want to generate your own?

The [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press) has 18 APIs in its catalog ready to go, and can generate CLIs from any OpenAPI spec, GraphQL schema, or even sniffed browser traffic.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to submit a CLI.

## License

MIT
