---
title: "feat: End-to-end checkout flow"
type: feat
status: completed
date: 2026-04-02
---

# feat: End-to-end checkout flow

## Overview

The dominos-pp-cli can browse menus, find stores, build carts, and price orders, but it cannot complete a checkout end-to-end. The auth login command exists but isn't wired up, the cart is local-only with no bridge to the order API, and there's no payment or confirmation flow. This plan fixes the gaps so an agent or human can go from `cart new` to `orders place_order` without leaving the CLI.

## Problem Frame

We tried to order a pizza tonight and hit every gap in sequence: no login flow (had to extract a browser token manually), no way to convert a local cart into an order payload, no coupon application during pricing, no payment method management, and no confirmation step before placing. Every competitor library (apizza, node-dominos, pizzapi, pizzamcp) has solved some subset of these problems.

## Competitor Research Summary

Research covered 8 libraries across Go, JavaScript, and Python. Many are old and may hit reCAPTCHA blocks, but the patterns are still instructive.

**Key findings:**

| Capability | apizza (Go) | node-dominos (JS) | pizzapi (Python) | pizzamcp (TS MCP) |
|---|---|---|---|---|
| Auth/login | OAuth2 to authproxy.dominos.com | None (guest only) | None (guest only) | None (guest only) |
| Card-on-file | Yes (via OAuth scopes) | No | No | No |
| Cart-to-order bridge | `store.NewOrder()` + `order.AddProduct()` | `new Order(customer)` + `addItem()` | `Order(store, customer, addr)` + `add_item()` | Builds order from local items array at place-time |
| Coupons | Not supported | `order.addCoupon(code)` | `order.add_coupon(code)` | Not supported |
| Validate/price/place | Three separate calls | Three separate async calls | Validate separate; price+place combined in `pay_with()` | All three inside `placeOrder()` |
| Payment amount | Derived from price response `Amounts.Customer` | Derived from `amountsBreakdown.customer` | Derived from `Amounts.Customer` | Derived from `amountsBreakdown.customer` |
| Safety mechanisms | None | Typed errors at each step | `pay_with()` as informal dry-run | reCAPTCHA detection + delivery-to-pickup fallback |

**Universal pattern across all libraries:** validate -> price (captures `Amounts.Customer`) -> attach payment with that amount -> place. The pricing step must run before place because the API requires the exact total on the payment object.

**Only apizza supports account login and card-on-file.** It uses OAuth2 against `authproxy.dominos.com/auth-proxy-service/login` with scopes like `customer:card:read` and `order:place:cardOnFile`. The token we extracted from the browser cookie had exactly these scopes.

**pizzamcp is most production-hardened:** handles reCAPTCHA blocks gracefully (returns store phone + URL for manual completion), auto-falls back from delivery to pickup, validates all fields before attempting placement.

## Requirements Trace

- R1. `auth login` works end-to-end: email + password in, token saved to config, subsequent commands authenticated
- R2. `cart checkout` (or similar) converts a local cart into a validated, priced order ready for placement
- R3. Coupons can be attached during the cart-to-order flow
- R4. `orders place_order` can use card-on-file from the authenticated account (no raw card numbers needed)
- R5. A confirmation step shows the total and requires explicit approval before placing (respects `--yes` flag for agents)
- R6. CREDITS.md documents the competitor libraries that informed the design

## Scope Boundaries

- No guest checkout (raw credit card entry). Card-on-file via authenticated account only.
- No reCAPTCHA solving. If the API returns a reCAPTCHA challenge, surface it clearly and suggest completing on dominos.com.
- No scheduled/future orders. Domino's API doesn't support them.
- No international support (Canada, UK variants). US only for now.

## Context & Research

### Relevant Code and Patterns

- `internal/cli/auth.go` - Auth parent command; `login` subcommand exists in `auth_login.go` but is NOT added to the parent
- `internal/cli/auth_login.go` - Hits `/power/login` with `--u` and `--p`, but doesn't save the returned token
- `internal/cli/cart.go` - Local cart stored in SQLite via `internal/store`; `cartRecord` has store ID, service method, address, items
- `internal/cli/orders_place_order.go` - Takes raw JSON via `--order` or `--stdin`, no cart integration
- `internal/cli/orders_price_order.go` - Same raw JSON interface
- `internal/config/` - Config with `SaveTokens(accessToken, refreshToken, token, ...)` and `AuthHeader()` methods

### Competitor Patterns Worth Adopting

- **apizza**: OAuth2 login flow with token persistence and card-on-file ordering
- **node-dominos**: Explicit three-step validate/price/place with typed errors at each stage
- **pizzapi**: `pay_with()` as a dry-run/preview step before committing
- **pizzamcp**: Pre-place validation checks, reCAPTCHA detection with graceful fallback message

## Key Technical Decisions

- **Wire up existing login command rather than building new OAuth flow**: `auth_login.go` already hits `/power/login`. The fix is: (1) add it to the parent command, (2) parse the response for access/refresh tokens, (3) save them via `config.SaveTokens()`. This matches how apizza's `gettoken()` works but uses the simpler `/power/login` endpoint.

- **`cart checkout` as the bridge command**: Rather than making the user manually construct order JSON, a new `cart checkout` command reads the active cart, builds the Domino's order payload, runs validate -> price, shows the total, and on confirmation runs place. This mirrors pizzamcp's single `placeOrder()` that assembles everything internally.

- **Card-on-file only, no raw card entry**: The token we extracted had `order:place:cardOnFile` scope. The API supports placing orders with a saved card when authenticated. This avoids the security risk of handling raw card numbers in a CLI.

- **Coupons as a cart-level concept**: Add `--coupon` flag to `cart checkout` (and optionally `cart add-coupon`). Coupons go into the order payload's `Coupons` array, matching node-dominos and pizzapi's approach.

## Open Questions

### Resolved During Planning

- **Where does the token come from?** The `/power/login` endpoint returns it. The existing `auth_login.go` already calls this endpoint but doesn't save the response token. Confirmed by the JWT we extracted from the browser: it came from dominos.com's login flow with scopes including `order:place:cardOnFile`.

- **How does card-on-file work?** When the token has `order:place:cardOnFile` scope, the place-order payload can reference saved cards by type (e.g., `"Type": "CreditCard"`) without sending full card numbers. apizza confirms this pattern.

### Deferred to Implementation

- **Exact response shape from `/power/login`**: Need to capture the actual response to know where the access token and refresh token live in the JSON. The current `auth_login.go` just prints the raw response.
- **Token refresh flow**: The JWT we saw had a ~2 hour expiry. Need to determine if the refresh token endpoint exists and how it works. For v1, re-login when expired is acceptable.
- **reCAPTCHA frequency**: The unofficial API may require reCAPTCHA for login or order placement. Need to discover this during implementation and handle gracefully.

## Implementation Units

- [ ] **Unit 1: Wire up `auth login` and persist tokens**

  **Goal:** Make `dominos-pp-cli auth login --u email --p password` actually log in and save the token so all subsequent commands are authenticated.

  **Requirements:** R1

  **Dependencies:** None

  **Files:**
  - Modify: `internal/cli/auth.go` (add `newAuthLoginCmd` to parent)
  - Modify: `internal/cli/auth_login.go` (parse response, save tokens via config)
  - Modify: `internal/config/` (ensure `SaveTokens` populates the `token` field that `AuthHeader()` reads)
  - Test: `internal/cli/auth_login_test.go`

  **Approach:**
  - Add `cmd.AddCommand(newAuthLoginCmd(flags))` to `newAuthCmd`
  - After the POST to `/power/login`, parse the response JSON for access_token/refresh_token fields
  - Call `cfg.SaveTokens(...)` with the extracted values, ensuring the `token` field (not just `access_token`) is set since `AuthHeader()` reads `token`
  - Print confirmation with token expiry time

  **Patterns to follow:**
  - `internal/cli/auth.go` existing `set-token` command for config save pattern
  - apizza's `gettoken()` for the login endpoint contract

  **Test scenarios:**
  - Happy path: login with valid credentials returns token, token is saved to config, `auth status` shows authenticated
  - Error path: login with invalid credentials returns error, no token saved
  - Error path: login when API returns reCAPTCHA challenge, user gets clear message

  **Verification:**
  - `dominos-pp-cli auth login --u email --p pass` followed by `auth status` shows authenticated

- [ ] **Unit 2: Add `cart checkout` command (validate + price + confirm)**

  **Goal:** Bridge the local cart to the order API. Read the active cart, build the order payload, validate it, price it, show the breakdown, and wait for confirmation.

  **Requirements:** R2, R3, R5

  **Dependencies:** Unit 1 (needs auth for card-on-file)

  **Files:**
  - Create: `internal/cli/cart_checkout.go`
  - Modify: `internal/cli/cart.go` (add subcommand)
  - Test: `internal/cli/cart_checkout_test.go`

  **Approach:**
  - New `cart checkout [--coupon CODE]... [--tip AMOUNT]` command
  - Reads active cart via `loadActiveCart()` and `loadCartItems()`
  - Builds Domino's order JSON: maps `cartItem` structs to `Products` array with `Code`, `Qty`, `Options` (converting toppings to the `{code: {side: amount}}` format)
  - Attaches coupons from `--coupon` flags to `Coupons` array
  - Calls `/power/validate-order`, checks for errors, reports them
  - Calls `/power/price-order`, extracts `Amounts.Customer`, `Amounts.DeliveryFee`, `Amounts.Tax`
  - Prints a clear order summary: items, coupons applied, subtotal, delivery fee, tax, total
  - If not `--yes`, prompts "Place this order? [y/N]"
  - On confirmation, attaches card-on-file payment with the priced amount, calls `/power/place-order`
  - On success, prints order ID and estimated wait time

  **Patterns to follow:**
  - `internal/cli/orders_price_order.go` for API call pattern
  - pizzamcp's `placeOrder()` for the validate->price->payment->place sequence
  - node-dominos for the topping options format `{code: {"1/1": "1.0"}}`

  **Test scenarios:**
  - Happy path: cart with items -> validates -> prices -> user confirms -> order placed, order ID returned
  - Happy path: cart with coupon applied -> coupon reflected in price breakdown with discount
  - Edge case: empty cart -> error message before any API calls
  - Edge case: store closed -> validate returns StoreClosed status -> clear message with store hours
  - Error path: validation fails (invalid product code) -> error shown, no pricing attempted
  - Error path: reCAPTCHA required -> message with dominos.com URL and store phone
  - Agent path: `--yes` flag skips confirmation prompt

  **Verification:**
  - Full flow from `cart new` -> `cart add` -> `cart checkout` completes or fails with clear messages at each step

- [ ] **Unit 3: Add `cart add-coupon` and `cart remove-coupon` commands**

  **Goal:** Let users attach coupons to their cart before checkout, and see them in `cart show`.

  **Requirements:** R3

  **Dependencies:** None (can parallel with Unit 2, coupons stored on cart)

  **Files:**
  - Create: `internal/cli/cart_coupon.go`
  - Modify: `internal/cli/cart.go` (add subcommands, extend `cartRecord` or items JSON)
  - Modify: `internal/cli/cart_show.go` (display coupons)
  - Test: `internal/cli/cart_coupon_test.go`

  **Approach:**
  - Store coupons as part of the cart record (either extend `cartRecord` with a `CouponsJSON` field or add to the existing items structure)
  - `cart add-coupon CODE` appends to the list
  - `cart remove-coupon CODE` removes it
  - `cart show` displays coupons section
  - `cart checkout` reads coupons from cart (in addition to `--coupon` flag)

  **Patterns to follow:**
  - node-dominos `order.addCoupon(code)` / `order.removeCoupon(code)` for the interface
  - Existing `cart_add.go` and `cart_remove.go` for the command structure

  **Test scenarios:**
  - Happy path: add coupon, show cart displays it, checkout includes it
  - Happy path: remove coupon, show cart no longer displays it
  - Edge case: add duplicate coupon code -> no-op or warning
  - Edge case: remove non-existent coupon -> clear error

  **Verification:**
  - `cart add-coupon 1121` -> `cart show` includes coupon -> `cart checkout` sends it in the order

- [ ] **Unit 4: Add CREDITS.md**

  **Goal:** Document the open-source projects that informed the CLI's design with proper attribution.

  **Requirements:** R6

  **Dependencies:** None

  **Files:**
  - Create: `CREDITS.md`

  **Approach:**
  - List each project with: name, author, GitHub URL, language, what we learned
  - Include the note that many of these projects are old and may not work against the current API, but their patterns informed the design
  - Note the GraphQL BFF discovery from the sniff session

  **Test scenarios:**
  - N/A (documentation only)

  **Verification:**
  - File exists, all 8 projects from the credits table are listed with URLs

## System-Wide Impact

- **Auth flow**: Login command will modify config file on disk. Other commands already read config via `AuthHeader()`, so they'll automatically pick up the token.
- **Cart storage**: Adding coupons to the cart extends the SQLite schema or JSON structure. Existing carts without coupons should still load fine (empty/nil coupons list).
- **API surface parity**: The MCP server (`cmd/dominos-mcp/`) should eventually get equivalent checkout tools, but that's out of scope for this plan.
- **Error propagation**: Each step in the checkout flow (validate, price, place) can fail independently. Errors should surface clearly with the Domino's status codes and messages, not generic HTTP errors.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| `/power/login` may require reCAPTCHA on headless requests | Detect and surface clearly. Suggest browser login + token extraction as fallback (what we did tonight). |
| Token expiry (~2 hours) means checkout must happen within that window | Show expiry time on login. For v2, add token refresh. |
| Domino's API is unofficial and undocumented; endpoints may change | The Printing Press sniff session captured the current contract. Pin to known working shapes. |
| Card-on-file payment format is inferred from apizza, not confirmed | Test with a real order during implementation. Use `--dry-run` first. |

## Sources & References

- **apizza** (Go): github.com/harrybrwn/apizza - Cart management, topping syntax, Go API wrapper (dawg library), OAuth2 auth flow
- **node-dominos-pizza-api** (JS): github.com/RIAEvangelist/node-dominos-pizza-api - Most comprehensive endpoint docs, tracking, international support, error codes, coupon API
- **pizzapi** (Python): github.com/ggrammar/pizzapi - Customer/address/store/order flow, `pay_with()` dry-run pattern
- **dominos** (PyPI): github.com/tomasbasham/dominos - UK API variant, 403 reachability risk signal
- **mcpizza** (Python MCP): github.com/GrahamMcBain/mcpizza - MCP tool patterns, safety-first ordering
- **pizzamcp** (JS MCP): github.com/GrahamMcBain/pizzamcp - End-to-end ordering via MCP, payment flow, reCAPTCHA handling
- **dominos-canada** (JS): Canadian endpoint patterns
- **ez-pizza-api** (JS): Simplified ordering wrapper
- Related: GraphQL BFF with 24 operations discovered via sniff session (not documented by any community tool)
