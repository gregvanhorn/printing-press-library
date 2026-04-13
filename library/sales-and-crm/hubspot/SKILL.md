---
name: pp-hubspot
description: "Use this skill whenever the user asks about HubSpot CRM contacts, companies, deals, tickets, tasks, calls, emails, meetings, engagements, pipelines, deal velocity / stale deals / coverage, or wants to search across their CRM data. Also for creating / updating / deleting any HubSpot record or managing associations between objects. HubSpot CLI covering 15 HubSpot APIs with offline SQLite search and pipeline analytics. Requires a HubSpot access token. Triggers on phrasings like 'find contacts at Acme', 'show deals closing this month', 'which deals are stale', 'pipeline velocity this quarter', 'log a call for contact X', 'create a task for tomorrow'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["hubspot-pp-cli"],"env":["HUBSPOT_ACCESS_TOKEN"]},"primaryEnv":"HUBSPOT_ACCESS_TOKEN","install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/sales-and-crm/hubspot/cmd/hubspot-pp-cli@latest","bins":["hubspot-pp-cli"],"label":"Install via go install"}]}}'
---

# HubSpot — Printing Press CLI

Manage HubSpot CRM contacts, companies, deals, tickets, engagements, pipelines, and associations with offline search and pipeline analytics. Derived from 15 official HubSpot OpenAPI specs, covering the core CRM + engagements + marketing objects.

## When to Use This CLI

Reach for this when a user wants CRM lookups or mutations from outside the HubSpot web UI — especially when batched (find-then-update N contacts), analytical (pipeline velocity, stale deal detection), or chained with other tools via pipes. Agent-native output makes it ideal for "write this call note and tag the contact" style workflows.

Don't reach for this if the user has a dedicated workflow in HubSpot's own Workflows / Sequences tools that's cleaner, or if they need marketing-specific automation beyond the CRM object layer.

## Unique Capabilities

### Pipeline analytics

- **`deals velocity`** — Average time-in-stage per pipeline stage. Identifies stages where deals stall.

  _The "where is my funnel leaking" command. Dashboard visualizations hide the stage-level dwell time._

- **`deals stale [--days N]`** — Deals that haven't been updated in N days. Sales ops hygiene.

- **`deals coverage`** — Revenue coverage ratio — pipeline size vs quota. Forecasting tool.

### Search across CRM

- **`search "<query>"`** — Full-text search across synced contacts, companies, deals, tickets, notes, calls, emails. Single query spans all objects.

  _Much faster than HubSpot's per-object search UI; works offline after sync._

### Associations — the hard-to-use part

- **`associations <fromObjectType> <objectId> <toObjectType>`** — List associations from one record to another type (e.g., contacts on a deal, deals on a company).

  _Wraps HubSpot's quirky associations API (which requires specifying both ends of the type mapping) into a single call._

### Bulk operations

Every top-level object (`contacts`, `companies`, `deals`, `tickets`, etc.) supports create/update/delete with batch variants that the HubSpot API exposes. The CLI surfaces them consistently.

## Command Reference

CRM objects (each supports list, get, create, update, delete, search):

- `hubspot-pp-cli contacts`
- `hubspot-pp-cli companies`
- `hubspot-pp-cli deals` — plus `velocity`, `stale`, `coverage`
- `hubspot-pp-cli tickets`

Engagements:

- `hubspot-pp-cli tasks`
- `hubspot-pp-cli notes`
- `hubspot-pp-cli calls`
- `hubspot-pp-cli emails`
- `hubspot-pp-cli meetings`

Pipeline and schema:

- `hubspot-pp-cli pipelines` — List + get pipelines and stages
- `hubspot-pp-cli properties` — Property schema per object type
- `hubspot-pp-cli lists` — Static and active lists

Associations:

- `hubspot-pp-cli associations <fromType> <id> <toType>` — List related records

Marketing / misc:

- `hubspot-pp-cli owners` — CRM users / record owners
- `hubspot-pp-cli analytics` — Aggregate metrics

Local store + utility:

- `hubspot-pp-cli sync` / `export` / `import` / `archive` — Offline data ops
- `hubspot-pp-cli search <query>` — Cross-object full-text
- `hubspot-pp-cli auth set-token <HUBSPOT_ACCESS_TOKEN>`
- `hubspot-pp-cli doctor` — Verify

## Recipes

### Morning pipeline review

```bash
hubspot-pp-cli deals stale --days 14 --agent           # neglected deals
hubspot-pp-cli deals velocity --period 30d --agent     # stage-level timing
hubspot-pp-cli deals coverage --quota 500000 --agent   # vs quota
```

Stale deals to nudge, velocity to spot funnel leaks, coverage to see if pipeline is healthy against the target.

### Find a contact and log a call

```bash
hubspot-pp-cli contacts search --filter "email:alice@acme.com" --agent
CONTACT_ID=$(hubspot-pp-cli contacts search --filter "email:alice@acme.com" --agent | jq -r '.results[0].id')
hubspot-pp-cli calls create --contact-id "$CONTACT_ID" --duration 20m --notes "Discussed pricing, sending proposal by Thursday" --agent
```

Search returns the contact; pipe through jq to extract ID; `calls create` logs the engagement with automatic association.

### What does Acme look like in our CRM?

```bash
COMPANY=$(hubspot-pp-cli companies search --filter "name:Acme" --agent | jq -r '.results[0].id')
hubspot-pp-cli companies get "$COMPANY" --agent
hubspot-pp-cli associations companies "$COMPANY" contacts --agent   # people at Acme
hubspot-pp-cli associations companies "$COMPANY" deals --agent      # deals with Acme
```

One company lookup → associated contacts → associated deals. A full 360° CRM view for a single org.

### Cross-object search

```bash
hubspot-pp-cli search "hormone therapy" --agent
```

Returns matching contacts, companies, deals, tickets, and engagements (notes, calls, emails) in one response. Useful for domain-spanning queries.

## Auth Setup

HubSpot uses private app access tokens. Create one at [app.hubspot.com/l/settings/account-settings/integrations](https://app.hubspot.com/l/settings/account-settings/integrations) → Private Apps.

```bash
export HUBSPOT_ACCESS_TOKEN="pat-..."
hubspot-pp-cli auth set-token "$HUBSPOT_ACCESS_TOKEN"
hubspot-pp-cli doctor
```

Grant the private app the scopes you need (typically: `crm.objects.contacts.*`, `crm.objects.companies.*`, `crm.objects.deals.*`, `crm.objects.tickets.*`, plus engagement scopes). Missing scopes surface as exit code 4 with a specific scope name.

## Agent Mode

Add `--agent` to any command. Expands to `--json --compact --no-input --no-color --yes`. Useful flags: `--select`, `--dry-run`, `--rate-limit N` (HubSpot is aggressive on rate limits; throttle for bulk ops), `--filter <expr>` for structured filtering.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error |
| 3 | Not found (contact, company, deal, etc.) |
| 4 | Auth required / missing scope |
| 5 | API error |
| 7 | Rate limited |
| 10 | Config error |

## Installation

### CLI

```bash
go install github.com/mvanhorn/printing-press-library/library/sales-and-crm/hubspot/cmd/hubspot-pp-cli@latest
hubspot-pp-cli auth set-token YOUR_HUBSPOT_ACCESS_TOKEN
hubspot-pp-cli doctor
```

### MCP Server

```bash
go install github.com/mvanhorn/printing-press-library/library/sales-and-crm/hubspot/cmd/hubspot-pp-mcp@latest
claude mcp add -e HUBSPOT_ACCESS_TOKEN=<token> hubspot-pp-mcp -- hubspot-pp-mcp
```

## Argument Parsing

Given `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → run `hubspot-pp-cli --help`
2. **`install`** → CLI; **`install mcp`** → MCP
3. **"find contact / company / deal"** → `<object> search --filter <expr> --agent`
4. **"pipeline health"** → `deals velocity | stale | coverage`
5. **"log call / task / note"** → `<engagement> create --contact-id ... --agent`
6. **Anything else** → check install + auth, route by CRM object verb, run with `--agent`.
