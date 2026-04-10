# Domino's Pizza CLI Absorb Manifest

## Sources Cataloged
1. **apizza** (Go CLI) - harrybrwn/apizza - menu, cart, order, config
2. **node-dominos-pizza-api** (npm) - RIAEvangelist - stores, menu, order, tracking, payment, international
3. **pizzapi** (npm) - RIAEvangelist fork - stores, menu, order, payment
4. **pizzapi** (PyPI) - ggrammar - customer, address, store, menu, order, payment
5. **dominos** (PyPI) - tomasbasham - store, menu, basket, checkout (UK)
6. **mcpizza** (MCP) - GrahamMcBain - find_store, menu, add_to_order, customer, calculate_total
7. **pizzamcp** (MCP) - GrahamMcBain - order_pizza unified tool with 8 actions
8. **dominos-canada** (npm) - Canadian endpoint support
9. **ez-pizza-api** (npm) - simplified ordering wrapper
10. **Dominos GraphQL BFF** (sniff) - 24 operations including loyalty, deals, campaigns, upsells

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Find nearest store by address | apizza `config`, node-dominos NearbyStores | `stores find --address "421 N 63rd St, Seattle"` | Offline cache of visited stores, --json, distance sorting |
| 2 | Get store profile/details | node-dominos Store, dawg store.go | `stores get 7094` | SQLite-cached store data, hours check, capability flags |
| 3 | Store hours and availability | node-dominos store.info, sniff Store query | `stores hours 7094` | Formatted hours by day, open-now indicator, wait time estimate |
| 4 | Browse full menu by category | apizza `menu`, mcpizza get_menu | `menu browse --store 7094` | Offline FTS search, category filtering, --json |
| 5 | Search menu items by name | apizza `menu <category>`, pizzapi search | `menu search "pepperoni"` | FTS5 offline search across all items, toppings, descriptions |
| 6 | View item details with toppings | apizza `menu <code>`, node-dominos Item | `menu item S_PIZPH` | Full topping options with left/right/full, nutrition info |
| 7 | List toppings and codes | apizza `menu --toppings` | `menu toppings` | Searchable topping list with availability per item |
| 8 | Create new order/cart | apizza `cart new`, mcpizza add_to_order, sniff CreateCart | `cart new --store 7094 --service delivery --address "..."` | Named carts, --dry-run preview, SQLite persistence |
| 9 | Add item to cart | apizza `cart --add`, mcpizza add_to_order, sniff QuickAddProductMenu | `cart add S_PIZPH --size large --qty 2` | Topping syntax, quick-add by name, undo support |
| 10 | Remove item from cart | apizza `cart --remove`, pizzamcp remove_item | `cart remove <item-id>` | Undo buffer, item index reference |
| 11 | View cart contents | mcpizza view_order, sniff CartById | `cart show` | Formatted table with pricing, --json, item breakdown |
| 12 | Customize toppings | apizza topping syntax `P:full:2`, node-dominos Item options | `cart customize <item> --topping "P:left:1.5"` | Intuitive syntax: `--topping pepperoni:left:extra` |
| 13 | Validate order | node-dominos validate, dawg ValidateOrder, sniff CheckDraftDeal | `order validate` | Detailed validation errors, --json, fix suggestions |
| 14 | Price order | node-dominos price, dawg Price, sniff SummaryCharges | `order price` | Price breakdown: subtotal, tax, delivery fee, total |
| 15 | Place order | apizza `order`, node-dominos place, pizzamcp place_order | `order place --cvv 123` | --dry-run mandatory preview, confirmation prompt, receipt |
| 16 | Track order by phone | node-dominos Tracking, tracker endpoint | `track --phone 2065551234` | Real-time polling with status updates, progress bar |
| 17 | Customer setup | mcpizza set_customer_info, pizzapi Customer | `config set --name "Matt" --phone "..." --email "..."` | Stored in config, reused across orders |
| 18 | Address management | node-dominos Address, sniff Customer query | `address list` / `address add "421 N 63rd St"` | Multiple saved addresses, SQLite-backed |
| 19 | Payment management | node-dominos Payment, pizzapi PaymentObject | `payment add --card "..." --expiry "..." --zip "..."` | Encrypted local storage, multiple cards |
| 20 | Apply coupons | node-dominos Order coupons, sniff CheckDraftDeal | `cart coupon add <code>` | Auto-apply best deal, coupon stacking check |
| 21 | International support | node-dominos useInternational, dominos-canada | `config set --country canada` | US, Canada, custom endpoints |
| 22 | Config management | apizza `config set/get/--edit` | `config set/get/edit` | TOML config, env var override, per-profile support |
| 23 | List deals and coupons | sniff DealsList query | `deals list --store 7094` | All deals with pricing, eligibility, expiry dates |
| 24 | Loyalty points balance | sniff LoyaltyPoints query | `rewards points` | Current balance, pending points, account status |
| 25 | Available loyalty rewards | sniff LoyaltyRewards query | `rewards list` | Rewards by tier (20/40/60 pts), unlock status |
| 26 | Member-exclusive deals | sniff LoyaltyDeals query | `rewards deals` | Personal deals with expiry dates, --add to cart |
| 27 | Previous order suggestions | sniff PreviousOrderPizzaModal query | `orders recent` | Past orders for quick reorder |
| 28 | Upsell suggestions | sniff UpsellForOrder query | `cart suggest` | Contextual add-on suggestions based on cart contents |
| 29 | Account login/auth | dawg auth.go, power/login endpoint | `auth login` | OAuth token management, session persistence |
| 30 | Cart ETA | sniff CartEtaMinutes query | `cart eta` | Estimated wait time including prep and delivery |
| 31 | Menu categories list | sniff Category query | `menu categories --store 7094` | Category names, images, new-item indicators |

## Transcendence (only possible with our local data layer)

| # | Feature | Command | Why Only We Can Do This | Score | Evidence |
|---|---------|---------|------------------------|-------|----------|
| 1 | Price comparison across stores | `compare-prices --address "..." --items S_PIZPH` | Requires syncing menu/pricing from multiple nearby stores into SQLite and joining | 9/10 | No existing tool compares prices across stores; community wrappers only query one store at a time |
| 2 | Order templates & quick reorder | `template save "friday-night" && template order "friday-night"` | SQLite-stored named order templates with item codes, toppings, store, address | 9/10 | apizza has cart persistence but no named templates; node-dominos has no persistence at all |
| 3 | Deal optimizer | `deals best --cart` | Cross-references cart contents against all available deals + loyalty deals to find the cheapest combination | 9/10 | No tool optimizes deal application; sniff showed CheckDraftDeal only checks one deal at a time |
| 4 | Menu diff (what changed) | `menu diff --store 7094` | Compares current menu against last-synced SQLite snapshot to show new items, removed items, price changes | 8/10 | No existing tool tracks menu changes; power users on HN thread mentioned wanting to know about new items |
| 5 | Spending analytics | `analytics --period 30d` | Aggregates order history in SQLite: spending trends, favorite items, order frequency, average order value | 8/10 | Order history is auth-gated; no existing tool persists or analyzes it |
| 6 | Store health score | `stores health 7094` | Combines wait times, hours, capabilities, and historical delivery times into a composite health score | 7/10 | Wait time data from CartEtaMinutes + store profile; no tool aggregates this |
| 7 | Smart reorder with substitution | `reorder --last --substitute-unavailable` | Replays last order but substitutes unavailable items with closest menu match using FTS similarity | 7/10 | Unique to having both order history and full menu in SQLite; mcpizza and apizza have no substitution logic |
| 8 | Bulk order builder | `order bulk --csv orders.csv` | Reads CSV of orders (for teams/parties) and batches them, finding the optimal store for the group | 7/10 | No existing tool supports multi-order batch; agent workflow gap for party planning |
| 9 | Live delivery tracker with polling | `track --watch --interval 30s` | Polls tracker endpoint and streams status updates: prep → bake → quality check → out for delivery | 8/10 | node-dominos has tracking but no polling/watch mode; 4iar/mockinos exists specifically because people want to programmatically watch status |
| 10 | Nutrition calculator | `nutrition --cart` | Sums calories, protein, fat, carbs across all cart items using synced menu nutrition data | 7/10 | Menu endpoint includes nutrition data (410 cal/slice observed in sniff); no existing tool aggregates it |

## Feature Counts
- Absorbed: 31 features from 10 tools
- Transcendence: 10 novel features (all scoring >= 7/10)
- Total: 41 features
