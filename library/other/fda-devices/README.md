# FDA Devices CLI

**Regulatory intelligence for FDA medical devices — predicate-chain auditing, competitive scanning, and safety pattern detection on a local mirror.**

fda-devices-pp-cli mirrors openFDA's 510(k), PMA, recall, MAUDE, classification, and establishment data into a local SQLite database, then surfaces it through agent-native commands like predicate-chain, competitors, and safety-pattern. Free public data, made queryable in compound ways the upstream API can't.

Learn more at [FDA Devices](https://open.fda.gov).

Printed by [@gregvanhorn](https://github.com/gregvanhorn) (Greg Van Horn).

## Install

The recommended path installs both the `fda-devices-pp-cli` binary and the `pp-fda-devices` agent skill in one shot:

```bash
npx -y @mvanhorn/printing-press install fda-devices
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press install fda-devices --cli-only
```


### Without Node

The generated install path is category-agnostic until this CLI is published. If `npx` is not available before publish, install Node or use the category-specific Go fallback from the public-library entry after publish.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/fda-devices-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-fda-devices --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-fda-devices --force
```

## Install for OpenClaw

Tell your OpenClaw agent (copy this):

```
Install the pp-fda-devices skill from https://github.com/mvanhorn/printing-press-library/tree/main/cli-skills/pp-fda-devices. The skill defines how its required CLI can be installed.
```

## Authentication

openFDA is unauthenticated by default (1,000 requests/day). Setting OPENFDA_API_KEY raises the cap to 120,000/day. No login flow.

## Quick Start

```bash
# One-time bulk mirror of every device endpoint
fda-devices-pp-cli sync --source all


# Recent clearances in a product code
fda-devices-pp-cli competitors --product-code DQA --since 2y --json


# Walk the ancestor tree
fda-devices-pp-cli predicate-chain K203456 --json


# Adverse-event rate ranking with field selection
fda-devices-pp-cli safety-pattern --product-code DQA --json --select results.product_code,results.events_per_year

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Competitive intelligence
- **`competitors`** — Every new 510(k) clearance in a product code, ranked by recency, with applicant info.

  _When an agent is asked 'who else cleared this kind of device recently', this gives a ranked answer in one call._

  ```bash
  fda-devices-pp-cli competitors --product-code DQA --since 2y --json
  ```
- **`applicant-history`** — Every clearance ever issued to a company, grouped by clinical area with go-to predicates.

  _Use this when asked about a vendor's regulatory footprint._

  ```bash
  fda-devices-pp-cli applicant-history 'Medtronic' --json
  ```
- **`market-entry`** — New entrants in a product code, ranked by time-to-clearance.

  _Use when asked about competitive timing or market dynamics in a device category._

  ```bash
  fda-devices-pp-cli market-entry --product-code DQA --since 1y --json
  ```

### Predicate audit
- **`predicate-chain`** — Recursive ancestor tree from a K-number, flagging any predicate that's been recalled.

  _Answers 'what is this device descended from' and 'has any ancestor been recalled' in one call._

  ```bash
  fda-devices-pp-cli predicate-chain K203456 --depth 10 --json
  ```

### Safety surveillance
- **`recall-link`** — Every recall on a device plus devices that used it as predicate.

  _Surfaces downstream exposure when a device or its ancestors get recalled._

  ```bash
  fda-devices-pp-cli recall-link K203456 --json
  ```
- **`safety-pattern`** — Devices in a category ranked by adverse-event count per year on market.

  _Quantitative answer to 'which devices in this category have the worst safety record per year on market'. Neutral statistic._

  ```bash
  fda-devices-pp-cli safety-pattern --product-code DQA --json
  ```

### Operational
- **`watch`** — Subscribe to product codes, applicants, predicate devices, or recall classes; pipes diff to file/Slack/webhook.

  _Use to set up ongoing surveillance for a category or vendor._

  ```bash
  fda-devices-pp-cli watch new --product-code DQA --notify slack
  ```

### Narrative
- **`story`** — One-paragraph briefing about a K-number: who cleared it, predicate chain, related recalls.

  _Quick context dump when an agent needs prose about a clearance._

  ```bash
  fda-devices-pp-cli story K203456
  ```

### Local data
- **`sql`** — Read-only SQL against the local mirror; join 510(k), recalls, MAUDE, classifications freely.

  _Escape hatch for arbitrary analytical questions an agent can compose on the fly._

  ```bash
  fda-devices-pp-cli sql 'SELECT applicant, COUNT(*) FROM clearances_510k WHERE product_code="DQA" GROUP BY applicant ORDER BY 2 DESC LIMIT 10'
  ```

## Usage

Run `fda-devices-pp-cli --help` for the full command reference and flag list.

## Commands

### device

Manage device

- **`fda-devices-pp-cli device list-classification`** - Search device product-code classifications
- **`fda-devices-pp-cli device list-establishment`** - Search establishment registrations
- **`fda-devices-pp-cli device list-maude`** - Search MAUDE adverse events
- **`fda-devices-pp-cli device list-pma`** - Search PMA approvals
- **`fda-devices-pp-cli device list-recall`** - Search device recalls
- **`fda-devices-pp-cli device list510k`** - Search 510(k) clearances


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
fda-devices-pp-cli device list-classification

# JSON for scripting and agents
fda-devices-pp-cli device list-classification --json

# Filter to specific fields
fda-devices-pp-cli device list-classification --json --select id,name,status

# Dry run — show the request without sending
fda-devices-pp-cli device list-classification --dry-run

# Agent mode — JSON + compact + no prompts in one flag
fda-devices-pp-cli device list-classification --agent
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

## Use with Claude Code

Install the focused skill — it auto-installs the CLI on first invocation:

```bash
npx skills add mvanhorn/printing-press-library/cli-skills/pp-fda-devices -g
```

Then invoke `/pp-fda-devices <query>` in Claude Code. The skill is the most efficient path — Claude Code drives the CLI directly without an MCP server in the middle.

<details>
<summary>Use as an MCP server in Claude Code (advanced)</summary>

If you'd rather register this CLI as an MCP server in Claude Code, install the MCP binary first:


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Then register it:

```bash
claude mcp add fda-devices fda-devices-pp-mcp
```

</details>

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/fda-devices-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "fda-devices": {
      "command": "fda-devices-pp-mcp"
    }
  }
}
```

</details>

## Health Check

```bash
fda-devices-pp-cli doctor
```

Verifies configuration and connectivity to the API.

## Configuration

Config file: `~/.config/fda-devices-pp-cli/config.toml`

Static request headers can be configured under `headers`; per-command header overrides take precedence.

## Troubleshooting
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific

- **429 Too Many Requests** — Set OPENFDA_API_KEY (free at open.fda.gov/apis/authentication) to raise the daily cap from 1k to 120k
- **predicate-chain returns 'no upstream references found'** — openFDA does not expose structured predicate fields; the chain depends on PDF letter scraping. Some clearance letters have no public PDF — this is honest empty data, not a bug
- **sync --source maude takes hours** — MAUDE is 11M+ records partitioned quarterly. Use --since 1y to limit to recent quarters; full sync stores ~5GB

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**Augmented-Nature/OpenFDA-MCP-Server**](https://github.com/Augmented-Nature/OpenFDA-MCP-Server) — TypeScript
- [**tsbischof/fda**](https://github.com/tsbischof/fda) — Python
- [**rOpenHealth/openfda**](https://github.com/rOpenHealth/openfda) — R
- [**coderxio/openfda**](https://github.com/coderxio/openfda) — Python
- [**PyMAUDE**](https://github.com/jhschwartz/PyMAUDE) — Python
- [**MAUDEMetrics**](https://github.com/MohamedMaroufMD/MAUDEMetrics) — Python
- [**autonlab/fda_maude**](https://github.com/autonlab/fda_maude) — Python
- [**FDA/openfda**](https://github.com/FDA/openfda) — Python

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
