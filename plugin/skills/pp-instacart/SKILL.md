---
name: pp-instacart
description: "Printing Press CLI for Instacart. Natural-language Instacart CLI that talks directly to the web GraphQL API. Add items to your cart, search products, and manage carts across retailers without browser automation. Also caches your purchase history locally so 'add' resolves items you have bought before instead of guessing from live search. Trigger phrases: 'install instacart', 'use instacart', 'run instacart', 'add X to my Safeway cart', 'what did I buy last time', 'order the usual', 'add my regulars to Costco'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
---

# Instacart - Printing Press CLI

Natural-language Instacart CLI that talks directly to the web GraphQL API. Add items to your cart, search products, and manage carts across retailers without browser automation. Caches your purchase history locally so `add` resolves items you have bought before instead of picking whatever live search ranks first.

## When to Use This CLI

Reach for this when a user wants:

- Add a product to an Instacart cart by natural language ("add lemon sorbet to QFC")
- Add something they have bought before ("add my usual milk to Safeway")
- Show, search, or compare their active carts across retailers
- List or search their own Instacart order history
- Run an Instacart flow from a script, cron job, or agent loop

Do not reach for this if the user wants to actually check out. This CLI adds items to your cart; you still complete checkout in the Instacart app or web UI.

## Unique Capabilities

### History-first `add`

`add` checks your local purchase history FIRST and, when a confident match exists at the target retailer, skips the three-call live GraphQL chain entirely. Drops the cost of "add the lemon sorbet pops I usually get" from ~1.2s to ~200ms AND makes it resolve to the right SKU (the one you actually buy) instead of whatever live search ranks highest today.

Confidence rules:
- FTS5 match in your local purchased_items at that retailer
- Purchased within the last 365 days
- Was in stock on the last purchase

Falls through to today's live-search behavior when any condition fails. Pass `--no-history` to force live search.

Every successful `add` (history-resolved or live-resolved) writes back to `purchased_items` so the signal gets warmer without a full re-sync.

### Purchase history import + inspection

**Use `history import`, not `history sync`.** Instacart does not expose a clean order-history GraphQL op, so `sync` cannot work — it surfaces a clear error pointing at the working path. The working path is:

1. Run the browser-side dumper (see `docs/dumper.js` + `docs/extract-one.js`) against a logged-in Chrome tab
2. Download the resulting `instacart-orders.jsonl`
3. `instacart history import <path>` — upserts orders + items into the local DB

See the full playbook at [`docs/patterns/authenticated-session-scraping.md`](https://github.com/mvanhorn/printing-press-library/blob/main/docs/patterns/authenticated-session-scraping.md) (repo root).

`history list` / `history search` / `history stats` inspect whatever has been loaded.

### Natural-language `add`

Resolves a product from free-text via Instacart's own three-call GraphQL chain (ShopCollectionScoped -> Autosuggestions -> Items) and fires `UpdateCartItemsMutation`. No browser automation.

### Multi-retailer `carts`

`carts list` shows every active cart across retailers at once. Useful for agents that need to know where items live before adding to the right one.

## Command Reference

Authentication:

- `instacart auth login` - extract session cookies from Chrome
- `instacart auth status` - show current session state
- `instacart auth logout` - clear saved cookies
- `instacart auth paste` - paste cookie JSON manually (fallback for newer macOS Chrome)
- `instacart auth import-file <path>` - load cookies from a browser-use export JSON

Cart operations:

- `instacart add <retailer> <query...>` - add a product by natural language
- `instacart add <retailer> <query...> --no-history` - skip the history-first resolver
- `instacart add --item-id <id> <retailer>` - add by exact Instacart item id
- `instacart cart show <retailer>` - show current cart contents at a retailer
- `instacart cart remove <item-id> <retailer>` - remove an item from a cart
- `instacart carts list` - list every active cart across retailers

Discovery:

- `instacart search <query> --store <retailer>` - search products at a retailer
- `instacart retailers list` - list retailers available at your address
- `instacart retailers show <slug>` - cache one retailer locally

Purchase history:

- `instacart history import <path>` - load a JSONL order dump into the local DB (the working path)
- `instacart history import - --json` - read from stdin, JSON output for agent pipelines
- `instacart history import <path> --dry-run` - preview counts without writing
- `instacart history list` - top purchased items by count + recency
- `instacart history list --store <retailer> --limit 20` - filter + paginate
- `instacart history search <query>` - FTS search your purchase history
- `instacart history search <query> --store <retailer>` - scoped FTS search
- `instacart history stats` - counts + per-retailer state

Maintenance:

- `instacart doctor` - health check: config, store, ops, history, session, live ping
- `instacart capture` - refresh the GraphQL operation hash cache
- `instacart capture --remote` - merge fresh hashes from the community registry
- `instacart ops list` - show the operation-hash cache state

## Recipes

### First-time setup

```bash
instacart auth login                # extract cookies from Chrome
instacart doctor                    # verify auth + live ping
instacart capture                   # seed built-in op hashes
# History backfill: browser-side dump, then import. See
# docs/patterns/authenticated-session-scraping.md for the full walkthrough.
# After you have instacart-orders.jsonl:
instacart history import ~/Downloads/instacart-orders.jsonl
instacart history stats             # confirm what landed
```

### Add something you buy all the time

```bash
instacart add safeway "oat milk"    # resolves via local history if you have bought it before
```

Look for `via history` in the output. If you see `via live`, the FTS match did not pass the confidence check; check `instacart history search "oat milk" --store safeway` to see what is actually in your history.

### Force a fresh live search

```bash
instacart add safeway "oat milk" --no-history --dry-run --json
```

`--dry-run --json` is useful when debugging - the output includes `resolved_via` so you can see which path would have fired.

### Daily top-up from recent history

```bash
instacart history list --store safeway --limit 20 --json | jq -r '.[].name' \
  | while read item; do instacart add safeway "$item" --yes --json; done
```

## Auth Setup

Requires a logged-in Instacart session in Chrome. The CLI extracts cookies via kooky (no credential handling on our side). If Chrome is locked or you are on a system kooky cannot read:

```bash
instacart auth paste         # paste the full cookie JSON manually
instacart auth import-file <path>
```

Session lives at `~/.config/instacart/session.json` (0600).

## Agent Mode

The CLI is agent-native by default. Pass `--json` on any command for machine-readable output. `--dry-run` previews `add` without firing the mutation and surfaces which resolver (`history`, `live`, or `item-id`) would have fired.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error |
| 3 | Auth missing or rejected |
| 4 | Resource not found |
| 5 | API error / conflict |
| 7 | Rate limited or transient network |

## Argument Parsing

Given a free-form natural-language request:

1. Empty, `help`, or `--help` -> run `instacart --help`
2. Starts with `install` -> CLI install; ends with `mcp` -> MCP install
3. Anything else -> map to the best subcommand and run with `--json` when invoked from an agent

## CLI Installation

```bash
go install github.com/mvanhorn/printing-press-library/library/commerce/instacart/cmd/instacart-pp-cli@latest
instacart-pp-cli --version
```

Ensure `$HOME/go/bin` is on `$PATH`.

## Direct Use

1. Check installed: `which instacart-pp-cli`
2. Check auth: `instacart doctor`
3. Capture GraphQL hashes: `instacart capture`
4. (Optional but recommended) Sync history: follow `docs/history-ops-capture.md` then `instacart history sync`
5. Run your command with `--json` if invoked from an agent
