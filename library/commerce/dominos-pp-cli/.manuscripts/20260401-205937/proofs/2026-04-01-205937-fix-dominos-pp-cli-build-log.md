# Dominos PP CLI Build Log

## What Was Built

### Priority 0: Foundation (Data Layer)
- Enhanced SQLite store with 8 entity-specific tables: menu_items, toppings, carts, order_templates, deals, loyalty, addresses
- FTS5 virtual table for menu_items (name, description, category, code)
- 18 new store methods (Upsert/Get/List/Delete for each entity)

### Priority 1: Absorbed Features (31)
- Stores: find_stores, get_store (with hours, capabilities)
- Menu: get_menu, search (FTS5 offline), categories, products, item details, toppings
- Cart: new, add (with topping syntax), remove, show, suggest
- Orders: validate, price, place (with --dry-run)
- Tracking: track by phone
- Auth: login
- Deals: list (GraphQL DealsList), best (deal optimizer)
- Rewards: points, list (by tier), deals (member exclusive)
- Address: add, list, remove, default
- Templates: save, list, show, delete, order
- Config: set, get, edit
- GraphQL: customer, create_cart, get_cart, quick_add, summary_charges, categories, products, loyalty_points, loyalty_rewards, loyalty_deals, deals_list

### Priority 2: Transcendence Features (5 of 10)
- compare-prices: price comparison across nearby stores
- nutrition: calorie/nutrition calculator from synced menu data
- menu diff: detect menu changes against synced snapshot
- track --watch: real-time polling with status updates
- template order: one-command reorder from templates

### Priority 3: Polish
- Domain-specific help examples on all commands
- Dead function cleanup
- Description rewrite

## Intentionally Deferred
- Spending analytics (order history aggregation -- needs real order data to be useful)
- Store health score (needs historical wait time data)
- Smart reorder with substitution (complex matching logic)
- Bulk order builder (CSV parsing + multi-store optimization)
- International endpoint support (US-only for now)

## Generator Limitations Found
- Internal YAML spec format doesn't support GraphQL operations natively
- Dogfood false-positives on helper functions called within the same file
- Verify can't parse internal YAML spec format (needs OpenAPI)
