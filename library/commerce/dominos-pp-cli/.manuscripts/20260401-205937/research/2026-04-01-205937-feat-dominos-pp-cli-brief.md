# Domino's Pizza CLI Brief

## API Identity
- Domain: Food ordering (pizza delivery/carryout)
- Users: Developers, power users, CLI enthusiasts, agents automating food orders
- Data profile: Stores, menus (products/variants/toppings), orders, coupons, customers, order tracking status
- Auth: OAuth 2.0 password grant via `authproxy.dominos.com`, or unauthenticated for read-only operations

## Reachability Risk
- **Medium** — The UK Python wrapper (tomasbasham/dominos) has a 403 issue, but the US-focused node wrapper (RIAEvangelist) and Go wrapper (harrybrwn/dawg) remain functional. Rate limiting exists but is manageable. Geographic blocking for non-US requests reported.

## API Endpoints (Reverse-Engineered, Stable)

### Base: `https://order.dominos.com`
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/power/store-locator?s={line1}&c={line2}&type={type}` | Find nearby stores |
| GET | `/power/store/{storeID}/profile` | Store details, hours, capabilities |
| GET | `/power/store/{storeID}/menu?lang={lang}&structured=true` | Full menu with categories |
| POST | `/power/validate-order` | Validate order contents |
| POST | `/power/price-order` | Price an order |
| POST | `/power/place-order` | Submit order |
| POST | `/power/login` | Account login |

### Auth: `https://authproxy.dominos.com`
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/auth-proxy-service/login` | OAuth token (password grant) |

### Tracking: `https://tracker.dominos.com` (or `trkweb.dominos.com`)
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/orderstorage/GetTrackerData?Phone={phone}` | Track order by phone |

## Top Workflows
1. **Order a pizza for delivery** — find store, browse menu, build order, validate, price, pay, track
2. **Find nearest store & check hours** — locate by address/zip, check if open, delivery vs carryout availability
3. **Browse menu & search items** — explore categories, search by name, view toppings/options
4. **Track an active order** — monitor prep/bake/box/delivery stages in real-time
5. **Reorder a favorite** — save order templates locally, replay with one command

## Table Stakes (from competitors)
- Store locator with distance (apizza, node-dominos, mcpizza)
- Full menu browsing by category (apizza `menu`, mcpizza `get_menu`)
- Item customization with toppings (apizza topping syntax `P:full:2`)
- Cart management with named orders (apizza `cart new/add/remove`)
- Order validation and pricing (all wrappers)
- Order placement with payment (apizza, pizzamcp)
- Order tracking by phone (node-dominos)
- Config management for saved addresses/payment (apizza `config`)

## Data Layer
- Primary entities: Stores, MenuItems (Products + Variants), Toppings, Orders, OrderItems, Coupons
- Sync cursor: Menu data per store (changes infrequently), order history per user
- FTS/search: Menu item search across name, description, category. Topping search.

## Product Thesis
- Name: **dominos-pp-cli**
- Why it should exist: Every existing tool is either abandoned (apizza last commit years ago), language-specific (node/python only), or limited to MCP (mcpizza). No maintained CLI offers offline menu search, order templates, real-time tracking with polling, or agent-native output. The compound features (price comparison across stores, order history analytics, smart reorder from templates) are impossible without a local data layer.

## Build Priorities
1. Store locator + menu browsing with offline FTS search
2. Full order workflow: build, validate, price, place
3. Order tracking with real-time polling
4. Saved order templates and reorder
5. Coupon discovery and application
6. Account auth with order history sync
