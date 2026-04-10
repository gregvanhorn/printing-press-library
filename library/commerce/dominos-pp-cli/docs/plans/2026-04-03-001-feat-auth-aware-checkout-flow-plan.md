---
title: "feat: Auth-aware checkout flow with new user experience"
type: feat
status: active
date: 2026-04-03
---

# Auth-Aware Checkout Flow with New User Experience

## Overview

The CLI has a fully functional ordering pipeline (cart -> validate -> price -> place) and auth token storage, but they are completely disconnected. An authenticated user still has to manually construct raw JSON with their name, phone, email, and credit card to place an order. The CLI never checks auth status during the ordering flow, never pulls customer profile or saved cards, and provides no guidance to new users on how to get started.

This plan adds an auth-aware `checkout` command that bridges cart and order placement, plus a first-run experience that guides new users through setup.

## Problem Frame

Three distinct failures in the current experience:

1. **Auth blindness in ordering**: The token config has scopes like `customer:profile:read:extended`, `customer:card:read`, and `order:place:cardOnFile`, but nothing in the ordering flow uses them. A logged-in user is treated identically to an anonymous one.

2. **No checkout bridge**: There is no command that takes a local cart and converts it into a validated, priced, payment-attached order. Users must manually construct the full Domino's order JSON object and pipe it through three separate commands.

3. **No new user experience**: Running `dominos` for the first time gives you a wall of subcommands with no guidance. There is no `dominos setup`, no first-run detection, no progressive disclosure of "here's how to order your first pizza."

## Requirements Trace

- R1. When a user has a valid auth token, the checkout flow must automatically fetch their profile (name, email, phone) and saved payment methods
- R2. The checkout flow must allow selecting a saved card on file without re-entering card details
- R3. A new `checkout` command must convert a local cart into a validated, priced order with payment attached, ready to place
- R4. The `checkout` command must show a clear summary and require explicit confirmation before placing
- R5. First-time users must get guided setup that walks through: auth, address, first store lookup
- R6. The GraphQL customer endpoint must work (fix the double `/api` path bug or use the REST profile endpoint)
- R7. The `--order` flag bug in validate/price/place commands must be fixed (passes string instead of parsed JSON)

## Scope Boundaries

- Not building a full OAuth login flow with username/password (token is already imported from browser)
- Not building a payment entry form for new cards (users without saved cards use `--payment-type cash` or add cards via the website)
- Not changing the existing low-level `orders validate_order` / `price_order` / `place_order` commands, just fixing the `--order` flag bug
- Not building order history or reorder functionality in this iteration

## Context & Research

### Relevant Code and Patterns

- `internal/config/config.go` - Token storage, `AuthHeader()` method. Token includes JWT with `Email` and `CustomerID` claims
- `internal/client/client.go:254-260` - Auth header already attached to all requests automatically
- `internal/cli/cart.go` - Local cart in SQLite with `cartRecord` struct (store, address, items)
- `internal/cli/orders_validate_order.go:52-56` - **Bug**: `--order` flag sets `body["Order"] = bodyOrder` as raw string instead of parsing JSON
- `internal/cli/graphql_customer.go:36` - GraphQL path `/api/web-bff/graphql`. The 404 error shows the server responding with `Cannot POST /api/api/web-bff/graphql`, suggesting the Domino's BFF server sits behind a proxy that adds `/api`
- `internal/cli/graphql_customer.go:119-120` - `operationname` flag is marked required but has a default value of "Customer", creating a confusing UX
- Token JWT payload contains: `Email`, `CustomerID`, scopes including `customer:profile:read:extended`, `customer:card:read`, `order:place:cardOnFile`

### Key API Observations

- The Power API (`/power/*`) works correctly with the base URL `https://order.dominos.com`
- The GraphQL BFF likely needs base URL `https://order.dominos.com` with path `/web-bff/graphql` (drop the leading `/api`)
- Alternative: use the Power API's customer profile endpoint if one exists, or decode the JWT directly for basic profile info (email, customer ID)

## Key Technical Decisions

- **Decode JWT for basic profile instead of fixing GraphQL first**: The auth token already contains `Email` and `CustomerID`. For name and phone, we can try the GraphQL fix but fall back to prompting the user to save their info locally on first checkout. This avoids blocking on the GraphQL path bug.

- **`checkout` command wraps the three-step flow**: Rather than expecting users to pipe JSON between validate/price/place, a single `checkout` command reads the active cart, builds the order object, calls validate -> price, shows a summary, and on confirmation calls place.

- **Profile stored locally after first use**: Once we have the user's name, phone, email (from JWT + first prompt), save to config. Subsequent checkouts skip the prompt.

- **First-run `doctor` enhancement over separate `setup` command**: Rather than a new `setup` command, enhance `doctor` to detect first-run state (no token, no saved address) and print actionable next steps. Add a `quickstart` alias that runs the guided flow.

## Open Questions

### Resolved During Planning

- **Q: Should we fix the GraphQL BFF or work around it?** Resolution: Try fixing the path (likely `/web-bff/graphql` instead of `/api/web-bff/graphql`), but the core checkout flow should not depend on it. JWT decode + local profile storage is the primary path.

- **Q: How to handle saved cards?** Resolution: The Domino's API supports `cardOnFile` payment type when placing orders with a valid auth token. The checkout command should attempt to list saved cards via the customer profile endpoint. If that fails, offer cash payment as fallback.

### Deferred to Implementation

- **Q: Exact GraphQL query shape for fetching saved cards** - needs experimentation with the corrected endpoint
- **Q: Whether `order:place:cardOnFile` scope allows referencing cards by ID or just uses the default card** - needs API testing

## Implementation Units

- [ ] **Unit 1: Fix the --order flag JSON parsing bug**

  **Goal:** Make `--order` flag properly parse JSON instead of passing it as a raw string, across all three order commands.

  **Requirements:** R7

  **Dependencies:** None

  **Files:**
  - Modify: `internal/cli/orders_validate_order.go`
  - Modify: `internal/cli/orders_price_order.go`
  - Modify: `internal/cli/orders_place_order.go`
  - Test: `internal/cli/orders_validate_order_test.go`

  **Approach:**
  In all three files, the `--order` flag handler does `body["Order"] = bodyOrder` where `bodyOrder` is a string. It should instead `json.Unmarshal` the string into a `map[string]any` and set that as the value, matching the `--stdin` code path behavior.

  **Patterns to follow:**
  - The `--stdin` code path in the same files already does proper JSON unmarshaling

  **Test scenarios:**
  - Happy path: `--order '{"Order":{...}}'` produces the same API request body as `--stdin` with identical JSON
  - Error path: `--order 'not-json'` returns a clear parse error instead of sending malformed request

  **Verification:**
  - `dominos orders validate_order --order '<json>' --dry-run` shows properly nested JSON body, not a string-escaped Order field

- [ ] **Unit 2: Add JWT decode helper and local profile storage**

  **Goal:** Extract user info (email, customer ID) from the stored JWT token and add local profile fields to config for name/phone.

  **Requirements:** R1

  **Dependencies:** None

  **Files:**
  - Modify: `internal/config/config.go`
  - Create: `internal/config/jwt.go`
  - Test: `internal/config/jwt_test.go`

  **Approach:**
  - Add a `DecodeJWTClaims()` function that base64-decodes the JWT payload (no signature verification needed since we're reading our own stored token) and extracts `Email`, `CustomerID`, and `exp`
  - Add `FirstName`, `LastName`, `Phone`, `Email` fields to Config struct with TOML tags
  - Add `Profile()` method that returns merged data: config fields take precedence, JWT fields as fallback for email
  - Add `SaveProfile(first, last, phone, email)` method

  **Patterns to follow:**
  - Existing `Config.SaveTokens()` method for config persistence pattern
  - Standard Go `encoding/base64` and `encoding/json` for JWT decode

  **Test scenarios:**
  - Happy path: `DecodeJWTClaims` correctly extracts email and customer ID from a valid JWT
  - Happy path: `Profile()` merges config fields with JWT fallback
  - Edge case: expired JWT still returns claims (we want the profile info regardless)
  - Edge case: malformed token returns empty claims without error (graceful degradation)
  - Edge case: empty token returns empty claims

  **Verification:**
  - Unit tests pass for JWT decode with real token format from Domino's

- [ ] **Unit 3: Fix GraphQL BFF endpoint path**

  **Goal:** Fix the double `/api` path issue so the GraphQL customer endpoint works.

  **Requirements:** R6

  **Dependencies:** None

  **Files:**
  - Modify: `internal/cli/graphql_customer.go`
  - Modify: all `internal/cli/graphql_*.go` files that use the same path pattern
  - Test: `internal/cli/graphql_customer_test.go`

  **Approach:**
  - Change GraphQL path from `/api/web-bff/graphql` to `/web-bff/graphql` in all GraphQL commands
  - If that still 404s, the BFF may use a different base URL entirely (e.g., `https://www.dominos.com/api/web-bff/graphql`). Add a `graphql_base_url` config field as escape hatch
  - Also remove the `MarkFlagRequired("operationname")` call since the flag already has a default value

  **Patterns to follow:**
  - Power API commands that successfully use paths like `/power/validate-order`

  **Test scenarios:**
  - Happy path: `graphql customer` returns customer profile data
  - Happy path: `graphql customer` works without explicitly passing `--operationname` (uses default)
  - Error path: unauthenticated request returns clear "not logged in" message

  **Verification:**
  - `dominos graphql customer --json` returns profile data instead of 404

- [ ] **Unit 4: Add `checkout` command - cart to order bridge**

  **Goal:** Single command that takes the active cart, attaches customer info and payment, validates, prices, summarizes, and optionally places the order.

  **Requirements:** R1, R2, R3, R4

  **Dependencies:** Unit 1, Unit 2

  **Files:**
  - Create: `internal/cli/checkout.go`
  - Test: `internal/cli/checkout_test.go`
  - Modify: `internal/cli/root.go` (register command)

  **Approach:**

  The `checkout` command flow:
  1. Load active cart (fail with helpful message if no cart)
  2. Load profile from config + JWT (prompt for missing name/phone on first use, save to config)
  3. Build the Order JSON object from cart data (address, store, products, service method) plus customer info (name, email, phone)
  4. Call validate-order endpoint
  5. Call price-order endpoint
  6. Display order summary: items, address, estimated wait, subtotal, delivery fee, tax, total
  7. Show payment options: saved card (if authenticated with cardOnFile scope), cash
  8. On confirmation, call place-order with selected payment
  9. Display order confirmation with tracking info

  Flags:
  - `--payment-type <card|cash>` - skip payment selection prompt
  - `--yes` - skip confirmation (for agents/scripts)
  - `--dry-run` - show everything but don't place
  - `--json` - output order summary as JSON (for agent consumption)

  **Patterns to follow:**
  - Cart commands for reading active cart from SQLite
  - Order commands for API call patterns
  - `--yes` and `--dry-run` flag patterns from existing commands

  **Test scenarios:**
  - Happy path: authenticated user with complete profile checks out with saved card, order placed
  - Happy path: `--dry-run` shows full summary without placing
  - Happy path: `--json` outputs structured order summary
  - Happy path: `--payment-type cash` skips payment selection
  - Edge case: no active cart shows helpful "create a cart first" message with example command
  - Edge case: missing profile info (first checkout) prompts for name/phone and saves
  - Edge case: expired auth token shows "re-authenticate" message
  - Error path: validate-order fails shows API error clearly
  - Error path: user declines confirmation, order not placed
  - Integration: cart items correctly map to Domino's Product options format (toppings as `{"P": {"1/1": "1"}}`)

  **Verification:**
  - `dominos cart show` -> `dominos checkout --dry-run` shows correct order summary matching cart contents

- [ ] **Unit 5: New user experience - `quickstart` command and enhanced `doctor`**

  **Goal:** Guide first-time users through setup and first order with progressive disclosure.

  **Requirements:** R5

  **Dependencies:** Unit 2, Unit 4

  **Files:**
  - Create: `internal/cli/quickstart.go`
  - Modify: `internal/cli/doctor.go`
  - Modify: `internal/cli/root.go` (register command)
  - Test: `internal/cli/quickstart_test.go`

  **Approach:**

  `quickstart` command walks through:
  1. Check auth status - if no token, explain how to get one (browser cookie extraction instructions)
  2. Ask for delivery address, save as default
  3. Find nearest delivery store
  4. Browse popular menu categories, help pick an item
  5. Create cart, add item
  6. Run checkout flow

  Enhanced `doctor` adds first-run detection:
  - No token -> "Run `dominos quickstart` to get started, or `dominos auth set-token` to add your auth token"
  - Token but no profile -> "Run `dominos checkout` on your next order to set up your profile"
  - Token but no saved address -> "Run `dominos address add` or `dominos quickstart` to save your address"

  When run non-interactively (`--no-input`), `quickstart` should output a checklist of what's missing and the commands to fix each item.

  **Patterns to follow:**
  - Existing `doctor` command structure for health checks
  - Interactive prompt patterns from cart commands

  **Test scenarios:**
  - Happy path: fresh install with no config runs through full quickstart flow
  - Happy path: partially configured user (has token, no address) skips completed steps
  - Happy path: fully configured user told "you're all set" with example order command
  - Edge case: `--no-input` mode outputs actionable checklist instead of interactive prompts
  - Edge case: `doctor` shows appropriate next-step hints based on config state

  **Verification:**
  - Fresh config directory + `dominos quickstart` guides through to a cart ready for checkout
  - `dominos doctor` output includes actionable hints when setup is incomplete

## System-Wide Impact

- **Interaction graph:** `checkout` becomes the new primary ordering entry point, calling validate/price/place internally. Existing low-level commands remain for scripting/agent use.
- **Error propagation:** Checkout must surface API errors from validate/price/place clearly, not wrap them in generic messages. Auth failures (401/403) should specifically suggest re-authentication.
- **State lifecycle risks:** Profile data in config.toml is persisted alongside auth tokens. Token expiry should be checked before starting checkout (not mid-flow).
- **API surface parity:** The `--agent` flag should work with `checkout` (skip prompts, JSON output, use `--payment-type` and `--yes` flags).
- **Unchanged invariants:** The three-step `orders validate_order` / `price_order` / `place_order` commands remain unchanged in behavior (only the `--order` flag parsing is fixed).

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| GraphQL BFF path fix may not be just dropping `/api` - the BFF might be on a different domain entirely | JWT decode + local profile is the primary path; GraphQL customer endpoint is a nice-to-have enhancement |
| `cardOnFile` payment type may require specific card ID from profile | Test with the fixed endpoint first; fall back to cash payment if card-on-file API shape is unclear |
| Token expiry not tracked (config shows `0001-01-01`) | Decode `exp` claim from JWT in Unit 2; warn user before starting checkout if expired |
| Cart-to-order product mapping may have edge cases with complex toppings | Start with the common case (standard toppings), document known gaps |

## Sources & References

- Config with token: `~/.config/dominos-pp-cli/config.toml`
- JWT token contains: `Email`, `CustomerID`, `exp`, scopes including `customer:card:read`, `order:place:cardOnFile`
- Domino's Power API base: `https://order.dominos.com/power/*`
- GraphQL BFF (needs path fix): `https://order.dominos.com/web-bff/graphql` (tentative)
