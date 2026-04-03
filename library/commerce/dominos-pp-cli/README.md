# dominos-pp-cli

Order pizza, browse menus, track deliveries, and manage rewards from the terminal. Offline menu search, saved order templates, and deal optimization included.

## Install

### Go

```
go install github.com/mvanhorn/dominos-pp-cli/cmd/dominos-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/dominos-pp-cli/releases).

## Quick Start

```bash
# Find your nearest store
dominos-pp-cli stores find_stores --s "421 N 63rd St" --c "Seattle, WA 98103"

# Browse the menu
dominos-pp-cli menu get_menu --storeID 7094 --json

# Search for items (works offline after first sync)
dominos-pp-cli menu search "pepperoni" --store 7094

# Build a cart and order
dominos-pp-cli cart new --store 7094 --service delivery --address "421 N 63rd St, Seattle, WA 98103"
dominos-pp-cli cart add S_PIZPH --size large --topping P:full:1.5
dominos-pp-cli cart show
dominos-pp-cli order validate
dominos-pp-cli order price

# Track your order
dominos-pp-cli track --phone 2065551234 --watch --interval 30s

# Check your rewards
dominos-pp-cli rewards points
dominos-pp-cli rewards list
```

No authentication is required for read-only operations (store lookup, menu browsing, deal listing). Account features (rewards, order history, saved addresses) require logging in:

```bash
dominos-pp-cli auth login
```

## Commands

### Stores

```bash
# Find nearby stores by address
dominos-pp-cli stores find_stores --s "1600 Pennsylvania Ave" --c "Washington, DC 20500" --type Carryout

# Get store details (hours, capabilities, wait times)
dominos-pp-cli stores get_store 7094
```

### Menu

```bash
# Full menu for a store
dominos-pp-cli menu get_menu --storeID 7094

# Search items offline (FTS5 full-text search)
dominos-pp-cli menu search "chicken wings" --store 7094 --limit 10

# See what changed since last sync
dominos-pp-cli menu diff --store 7094
```

### Cart

```bash
# Create a cart
dominos-pp-cli cart new --store 7094 --service delivery --address "421 N 63rd St, Seattle, WA 98103"

# Add items with toppings (code:side:amount)
dominos-pp-cli cart add S_PIZPH --size large --topping P:full:1.5 --topping J:left:1

# Remove item by index
dominos-pp-cli cart remove 2

# View cart with pricing
dominos-pp-cli cart show
dominos-pp-cli cart show --json
```

### Orders

```bash
# Validate and price before placing
dominos-pp-cli orders validate_order --stdin < order.json
dominos-pp-cli orders price_order --stdin < order.json

# Place an order (requires --dry-run to preview first)
dominos-pp-cli orders place_order --stdin < order.json --dry-run
dominos-pp-cli orders place_order --stdin < order.json
```

### Order Templates

```bash
# Save a reusable template
dominos-pp-cli template save "friday-night" --store 7094 --items S_PIZPH,S_MX --address "421 N 63rd St" --service delivery

# List and reorder
dominos-pp-cli template list
dominos-pp-cli template order "friday-night" --dry-run
```

### Tracking

```bash
# Track by phone number
dominos-pp-cli track --phone 2065551234

# Watch with live updates (polls every 30s, exits on delivery)
dominos-pp-cli track --phone 2065551234 --watch --interval 30s
```

### Deals

```bash
# List all deals for a store
dominos-pp-cli deals list --store 7094

# Find the best deal for your cart
dominos-pp-cli deals best --cart
```

### Rewards

```bash
# Check your points balance
dominos-pp-cli rewards points

# View available rewards by tier (20/40/60 points)
dominos-pp-cli rewards list

# See member-exclusive deals
dominos-pp-cli rewards deals
```

### Addresses

```bash
# Save an address
dominos-pp-cli address add "421 N 63rd St" --city Seattle --state WA --zip 98103 --label home --default

# List saved addresses
dominos-pp-cli address list
```

### Compare & Analyze

```bash
# Compare pricing across nearby stores
dominos-pp-cli compare-prices --address "421 N 63rd St, Seattle, WA 98103" --items S_PIZPH,S_MX

# Calculate nutrition for your cart
dominos-pp-cli nutrition --cart

# Analyze your ordering patterns
dominos-pp-cli analytics summary
```

## Output Formats

```bash
# Human-readable table (default)
dominos-pp-cli stores find_stores --s "421 N 63rd St" --c "Seattle, WA 98103"

# JSON for scripting and agents
dominos-pp-cli stores find_stores --s "421 N 63rd St" --c "Seattle, WA 98103" --json

# Filter specific fields
dominos-pp-cli stores find_stores --s "421 N 63rd St" --c "Seattle, WA 98103" --json --select StoreID,MinDistance

# Dry run (show request without sending)
dominos-pp-cli orders place_order --stdin < order.json --dry-run
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** -- never prompts, every input is a flag
- **Pipeable** -- `--json` output to stdout, errors to stderr
- **Filterable** -- `--select id,name` returns only fields you need
- **Previewable** -- `--dry-run` shows the request without sending
- **Confirmable** -- `--yes` for explicit confirmation of destructive actions
- **Piped input** -- `echo '{"key":"value"}' | dominos-pp-cli orders place_order --stdin`
- **Cacheable** -- GET responses cached for 5 minutes, bypass with `--no-cache`
- **All-in-one** -- `--agent` sets `--json --compact --no-input --no-color --yes`

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
dominos-pp-cli doctor
```

## Configuration

Config file: `~/.config/dominos-pp-cli/config.toml`

Environment variables:
- `DOMINOS_TOKEN` -- OAuth token for authenticated operations (rewards, order history)

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `dominos-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $DOMINOS_TOKEN`
- Re-authenticate: `dominos-pp-cli auth login`

**Not found errors (exit code 3)**
- Check the store ID or product code is correct
- Run `dominos-pp-cli stores find_stores` to find valid store IDs
- Run `dominos-pp-cli menu search` to find valid product codes

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- Use `--rate-limit 1` to throttle requests
- If persistent, wait a few minutes and try again

## Cookbook

### Order a pizza in 4 commands

```bash
dominos-pp-cli stores find_stores --s "421 N 63rd St" --c "Seattle, WA 98103" --json --select StoreID,MinDistance
dominos-pp-cli cart new --store 7094 --service delivery --address "421 N 63rd St, Seattle, WA 98103"
dominos-pp-cli cart add S_PIZPH --size large
dominos-pp-cli order price
```

### Reorder your Friday night pizza

```bash
dominos-pp-cli template order "friday-night"
```

### Find the cheapest store for your order

```bash
dominos-pp-cli compare-prices --address "421 N 63rd St, Seattle, WA 98103" --items S_PIZPH,S_MX --json
```

### Watch your delivery in real-time

```bash
dominos-pp-cli track --phone 2065551234 --watch --interval 15s
```

## Sources & Inspiration

This CLI was built by studying these community projects and resources:

- [**node-dominos-pizza-api**](https://github.com/RIAEvangelist/node-dominos-pizza-api) -- JavaScript. The most comprehensive reverse-engineered documentation of Domino's endpoints, tracking, international support, and error codes.
- [**apizza**](https://github.com/harrybrwn/apizza) -- Go. CLI for ordering Domino's with cart management, topping syntax, and the `dawg` Go API wrapper library that mapped the `/power/` REST endpoints.
- [**pizzapi (Python)**](https://github.com/ggrammar/pizzapi) -- Python. Customer, address, store, menu, and order flow.
- [**mcpizza**](https://github.com/GrahamMcBain/mcpizza) -- Python MCP server. Store finder, menu browsing, cart management with safety-first ordering (real orders disabled by default).
- [**pizzamcp**](https://github.com/GrahamMcBain/pizzamcp) -- JavaScript MCP server. End-to-end ordering with payment processing via the `dominos` npm package.
- [**dominos (PyPI)**](https://github.com/tomasbasham/dominos) -- Python. UK API variant. Flagged the 403 reachability risk that informed our rate-limiting approach.

API endpoints were also discovered by sniffing authenticated traffic from dominos.com, which revealed a GraphQL BFF (`/api/web-bff/graphql`) with 24 operations (loyalty, deals, campaigns, upsells) not documented by any community project.

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
