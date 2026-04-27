---
name: pp-pagliacci
description: "Order Seattle's favorite pizza from the terminal — every endpoint, plus discount stacking, slice rotation across stores, and a local order history nobody else has. Trigger phrases: `order from pagliacci`, `what pagliacci slices are available`, `pagliacci rewards balance`, `pagliacci order history`, `use pagliacci`, `run pagliacci`."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["pagliacci-pizza-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/food-and-dining/pagliacci-pizza/cmd/pagliacci-pizza-pp-cli@latest","bins":["pagliacci-pizza-pp-cli"],"label":"Install via go install"}]}}'
---

# Pagliacci Pizza — Printing Press CLI

First and only CLI for the Pagliacci API. Browse menus and slice availability across all Seattle stores, build and price orders, manage your reward card and stored coupons, and replay past orders — all with offline search, agent-native output, and Chrome-cookie login (no manual token paste).

## When to Use This CLI

Use this CLI when an agent or user wants to interact with Pagliacci Pizza programmatically: browsing the menu, finding a store and time slot, building or pricing an order, checking rewards balance, or replaying a past order. The CLI is also the right choice for office-lunch automation (multi-address fan-out) and rewards-aware ordering (discount stacking).

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Local state that compounds

- **`slices today`** — See which Pagliacci slices are available right now at every Seattle store, sorted by proximity to your saved address.

  _When the agent is asked 'what slices can I get tonight?', this returns a single comparable list — no per-store iteration needed._

  ```bash
  pagliacci-pizza-pp-cli slices today --agent
  ```
- **`rewards stack`** — Compute the best application of stored coupons, reward redemption, and account credit for a given order total. Defaults to single-best-coupon + credit; multi-coupon stacking is flagged --experimental.

  _Agents helping with order placement can pick the optimal discount, not just the first valid coupon._

  ```bash
  pagliacci-pizza-pp-cli rewards stack --order-total 45.00 --agent
  ```
- **`orders summary`** — Aggregate order spend over a time range, with top items and store breakdown.

  _Agents helping with budgeting or reimbursement can produce a single roll-up report._

  ```bash
  pagliacci-pizza-pp-cli orders summary --since 90d --agent
  ```

### Time-aware composed lookups

- **`store tonight`** — List stores that are still open and can deliver to your saved address right now, sorted by ETA.

  _Late-night ordering: the agent only surfaces stores that will actually take the order._

  ```bash
  pagliacci-pizza-pp-cli store tonight --address-label home --agent
  ```
- **`address best-time`** — Resolve a saved address label to the next available delivery slot in one call.

  _Agents scheduling orders for a specific address don't need to discover the delivery zone separately._

  ```bash
  pagliacci-pizza-pp-cli address best-time --label home --agent
  ```

### Order workflows

- **`orders reorder`** — Re-create a past order as a fresh cart, with price revalidation since prices change. Add --send to also submit.

  _Agents can replay a routine order without rebuilding the cart line by line._

  ```bash
  pagliacci-pizza-pp-cli orders reorder --last --dry-run
  ```

## HTTP Transport

This CLI uses Chrome-compatible HTTP transport for browser-facing endpoints. It does not require a resident browser process for normal API calls.

## Discovery Signals

This CLI was generated with browser-observed traffic context.
- Capture coverage: 26 API entries from 26 total network entries
- Protocols: rest_json (98% confidence)
- Auth signals: composed — headers: Authorization — cookies: customerId, authToken
- Generation hints: requires_browser_auth, composed_auth
- Candidate command ideas: store list — GET /Store returned full inventory of locations; menu top — GET /MenuTop/{storeId} drives the home menu UI; menu cache — GET /MenuCache/{storeId} returns the full menu; menu slices — GET /MenuSlices returns today's slices across all stores; address lookup — POST /AddressInfo validates an address and resolves a delivery store; address list — GET /AddressName returns saved addresses; orders list — GET /OrderList/{page}/{size} returns paginated history; orders get — GET /OrderListItem/{id} returns full detail

## Command Reference

**account** — Authentication and registration (no auth required for these endpoints)

- `pagliacci-pizza-pp-cli account confirm_email` — Confirm a new account by clicking the email-confirmation link's token
- `pagliacci-pizza-pp-cli account create_token` — Issue a session token (used internally by the SPA for token refresh)
- `pagliacci-pizza-pp-cli account login` — Authenticate with email/phone + password. Response sets customerId and authToken cookies.
- `pagliacci-pizza-pp-cli account logout` — Invalidate the current session
- `pagliacci-pizza-pp-cli account password_forgot` — Request a password reset email
- `pagliacci-pizza-pp-cli account password_reset` — Reset a password using a token from PasswordForgot email
- `pagliacci-pizza-pp-cli account register` — Create a new customer account

**address** — Address validation and saved address book

- `pagliacci-pizza-pp-cli address create` — Create a new saved address
- `pagliacci-pizza-pp-cli address delete` — Delete a saved address
- `pagliacci-pizza-pp-cli address get` — Get a saved address by ID
- `pagliacci-pizza-pp-cli address get_info` — Get address info by saved ID
- `pagliacci-pizza-pp-cli address list` — List the authenticated user's saved addresses
- `pagliacci-pizza-pp-cli address lookup` — Validate an address and check delivery zone (returns store ID if deliverable)

**cart** — Build and price an order before sending it

- `pagliacci-pizza-pp-cli cart get_quote_building` — Get the current cart/quote-building state by building ID
- `pagliacci-pizza-pp-cli cart price_order` — Compute the total price for an order (cart contents, taxes, fees, delivery) before sending
- `pagliacci-pizza-pp-cli cart send_order` — Submit an order. Requires payment information for guests; uses stored payment for authenticated users.
- `pagliacci-pizza-pp-cli cart update_quote_building` — Update cart contents (add/remove/modify items)

**credit** — Account credit balance and entries

- `pagliacci-pizza-pp-cli credit delete` — Remove an account credit entry
- `pagliacci-pizza-pp-cli credit get` — Get a single credit entry
- `pagliacci-pizza-pp-cli credit list` — List the authenticated user's account credit entries

**customer** — Customer profile and devices

- `pagliacci-pizza-pp-cli customer access_devices_delete` — Revoke a device's access to the account
- `pagliacci-pizza-pp-cli customer access_devices_list` — List devices that have access to this account
- `pagliacci-pizza-pp-cli customer get` — Get customer profile by ID
- `pagliacci-pizza-pp-cli customer migrate_answer` — Submit the answer to a migration question
- `pagliacci-pizza-pp-cli customer migrate_question` — Submit a security/migration question (legacy account migration flow)

**customer_feedback** — Customer feedback submissions to Pagliacci

- `pagliacci-pizza-pp-cli customer_feedback get` — Get a feedback submission by ID
- `pagliacci-pizza-pp-cli customer_feedback submit` — Submit customer feedback (guest or authenticated)

**gifts** — Stored gift cards, balance lookup, and transfer

- `pagliacci-pizza-pp-cli gifts check` — Check the balance of a gift card by ID and PIN (no auth required to check)
- `pagliacci-pizza-pp-cli gifts delete` — Remove a stored gift card from the account
- `pagliacci-pizza-pp-cli gifts get` — Get a single stored gift card by ID
- `pagliacci-pizza-pp-cli gifts list` — List the authenticated user's stored gift cards
- `pagliacci-pizza-pp-cli gifts transfer` — Transfer gift card balance to another account
- `pagliacci-pizza-pp-cli gifts value` — Get current value/balance of a saved gift card

**menu** — Menus, slices, and product pricing

- `pagliacci-pizza-pp-cli menu cache` — Get the full menu (categories, products, prices, descriptions, images) for a store
- `pagliacci-pizza-pp-cli menu product_price` — Calculate the price for a customized product (size, toppings, modifiers)
- `pagliacci-pizza-pp-cli menu slices` — Get available slices across all stores for the current day (perishable, rotates daily)
- `pagliacci-pizza-pp-cli menu top` — Get featured top-of-menu items for a store

**orders** — Order history and details

- `pagliacci-pizza-pp-cli orders clone` — Get order data shaped for re-ordering (transforms a past order into a new cart)
- `pagliacci-pizza-pp-cli orders get` — Get the full detail of a single past order (items, prices, store, time)
- `pagliacci-pizza-pp-cli orders list` — List the authenticated user's order history (paginated)
- `pagliacci-pizza-pp-cli orders list_gift_cards` — List orders that purchased gift cards
- `pagliacci-pizza-pp-cli orders list_pending` — List orders that are currently in flight (placed but not yet delivered/picked up)
- `pagliacci-pizza-pp-cli orders suggestion` — Get personalized order suggestions for a customer

**rewards** — Loyalty card, rewards history, and stored coupons

- `pagliacci-pizza-pp-cli rewards card` — Get the authenticated user's reward card balance, points, and available rewards
- `pagliacci-pizza-pp-cli rewards coupon_lookup` — Look up a coupon by its serial number (validate before applying)
- `pagliacci-pizza-pp-cli rewards history` — Get reward earning/redemption history (most recent N entries)
- `pagliacci-pizza-pp-cli rewards stored_coupons` — List coupons saved to the authenticated user's account

**scheduling** — Delivery and pickup time windows

- `pagliacci-pizza-pp-cli scheduling slot_list` — List available time-window slots for a store and service type
- `pagliacci-pizza-pp-cli scheduling slot_list_for_date` — List allowed slot times for a specific delivery/pickup date (YYYYMMDD)
- `pagliacci-pizza-pp-cli scheduling window_days` — List available delivery or pickup days for a store. serviceType is DEL (delivery) or PICK (pickup)

**store** — Pagliacci store locations, hours, and quote info

- `pagliacci-pizza-pp-cli store compute_quote` — Compute a quote for a specific store with cart contents (returns Delivery, Drone, Pickup wait values)
- `pagliacci-pizza-pp-cli store get` — Get a single store by its numeric ID
- `pagliacci-pizza-pp-cli store get_quote` — Get quote-store metadata (delivery fee, drone status, pickup wait time) for a single store
- `pagliacci-pizza-pp-cli store list` — List all Pagliacci store locations with addresses, hours, GPS, amenities, and available slices
- `pagliacci-pizza-pp-cli store list_quotes` — List quote-store metadata (delivery fee, drone status, pickup wait) for all stores

**system** — System information and announcements

- `pagliacci-pizza-pp-cli system site_wide_message` — Get site-wide announcement banner text (closures, holiday hours, etc.)
- `pagliacci-pizza-pp-cli system version` — Get the current API version


## Freshness Contract

This printed CLI owns bounded freshness only for registered store-backed read command paths. In `--data-source auto` mode, those paths check `sync_state` and may run a bounded refresh before reading local data. `--data-source local` never refreshes. `--data-source live` reads the API and does not mutate the local store. Set `PAGLIACCI_NO_AUTO_REFRESH=1` to skip the freshness hook without changing source selection.

Covered paths:

- `pagliacci-pizza-pp-cli address`
- `pagliacci-pizza-pp-cli address get`
- `pagliacci-pizza-pp-cli address list`
- `pagliacci-pizza-pp-cli address search`
- `pagliacci-pizza-pp-cli credit`
- `pagliacci-pizza-pp-cli credit get`
- `pagliacci-pizza-pp-cli credit list`
- `pagliacci-pizza-pp-cli credit search`
- `pagliacci-pizza-pp-cli customer`
- `pagliacci-pizza-pp-cli customer get`
- `pagliacci-pizza-pp-cli customer list`
- `pagliacci-pizza-pp-cli customer search`
- `pagliacci-pizza-pp-cli gifts`
- `pagliacci-pizza-pp-cli gifts get`
- `pagliacci-pizza-pp-cli gifts list`
- `pagliacci-pizza-pp-cli gifts search`
- `pagliacci-pizza-pp-cli orders`
- `pagliacci-pizza-pp-cli orders get`
- `pagliacci-pizza-pp-cli orders list`
- `pagliacci-pizza-pp-cli orders search`
- `pagliacci-pizza-pp-cli store`
- `pagliacci-pizza-pp-cli store get`
- `pagliacci-pizza-pp-cli store list`
- `pagliacci-pizza-pp-cli store search`

When JSON output uses the generated provenance envelope, freshness metadata appears at `meta.freshness`. Treat it as current-cache freshness for the covered command path, not a guarantee of complete historical backfill or API-specific enrichment.

### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
pagliacci-pizza-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes


### What can I eat tonight?

```bash
pagliacci-pizza-pp-cli slices today --json --select store_name,slice_name,price
```

Returns today's available slices across all stores with just the high-gravity fields for fast agent decision making.

### Replay last order

```bash
pagliacci-pizza-pp-cli orders reorder --last --dry-run --json
```

Pulls your most recent OrderListItem, transforms it via OrderClone, re-prices it, and returns a sendable cart without submitting.

### Maximize discount before checkout

```bash
pagliacci-pizza-pp-cli rewards stack --order-total 42.50 --agent
```

Computes the best application of saved coupons, reward redemption, and account credit for a given order total.

### Spend summary for last 90 days

```bash
pagliacci-pizza-pp-cli orders summary --since 90d --json --select total_spent,top_items,by_store
```

Aggregates local order history into a single roll-up; --select narrows the response to high-gravity fields. (Use --json instead of --agent here: --agent enables --compact, which strips fields before --select can pick them.)

### Delivery time for home address

```bash
pagliacci-pizza-pp-cli address best-time --label home --agent
```

One-call resolution from saved address label to next available delivery slot for the resolved store.

## Auth Setup

Pagliacci has no public API and uses a custom composed `PagliacciAuth {customerId}|{authToken}` header constructed from cookies. Run `pagliacci-pizza-pp-cli auth login --chrome` while logged into pagliacci.com in Chrome — the CLI reads the auth cookies and constructs the header for you. No manual token paste required.

Run `pagliacci-pizza-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs. Pair `--select` with `--json` (not `--agent`) so the compact pass doesn't strip fields before they can be picked:

  ```bash
  pagliacci-pizza-pp-cli address list --json --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag

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
pagliacci-pizza-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
pagliacci-pizza-pp-cli feedback --stdin < notes.txt
pagliacci-pizza-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.pagliacci-pizza-pp-cli/feedback.jsonl`. They are never POSTed unless `PAGLIACCI_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `PAGLIACCI_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

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
pagliacci-pizza-pp-cli profile save briefing --json
pagliacci-pizza-pp-cli --profile briefing address list
pagliacci-pizza-pp-cli profile list --json
pagliacci-pizza-pp-cli profile show briefing
pagliacci-pizza-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `pagliacci-pizza-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.25+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/food-and-dining/pagliacci-pizza/cmd/pagliacci-pizza-pp-cli@latest
   
   # If `@latest` installs a stale build (Go module proxy cache lag), install from main:
   GOPRIVATE='github.com/mvanhorn/*' GOFLAGS=-mod=mod \
     go install github.com/mvanhorn/printing-press-library/library/food-and-dining/pagliacci-pizza/cmd/pagliacci-pizza-pp-cli@main
   ```
3. Verify: `pagliacci-pizza-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/food-and-dining/pagliacci-pizza/cmd/pagliacci-pizza-pp-mcp@latest
   
   # If `@latest` installs a stale build (Go module proxy cache lag), install from main:
   GOPRIVATE='github.com/mvanhorn/*' GOFLAGS=-mod=mod \
     go install github.com/mvanhorn/printing-press-library/library/food-and-dining/pagliacci-pizza/cmd/pagliacci-pizza-pp-mcp@main
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add pagliacci-pizza-pp-mcp -- pagliacci-pizza-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which pagliacci-pizza-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   pagliacci-pizza-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `pagliacci-pizza-pp-cli <command> --help`.
