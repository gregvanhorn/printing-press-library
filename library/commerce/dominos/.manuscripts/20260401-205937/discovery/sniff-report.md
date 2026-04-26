# Domino's Pizza Sniff Discovery Report

## User Goal Flow
- **Goal:** Order a pizza for delivery
- Steps completed:
  1. Homepage - clicked "Delivery" → triggered Store lookup APIs
  2. Address modal - selected saved address "421 N 63RD ST" → triggered Stores proximity query
  3. Confirm location - confirmed store 7094 → triggered Store profile, CartEta, LoyaltyPoints
  4. Menu - browsed Specialty Pizzas → triggered Category, Products, CampaignTilesByLocation
  5. Add to cart - added Spicy Chicken Bacon Ranch → triggered QuickAddProductMenu, CheckDraftDeal, SummaryCharges
  6. View cart - opened cart modal → triggered CartById, ProductQuantities, UpsellForOrder
  7. Checkout attempt - clicked "CHECK OUT" → redirected to homepage (SPA routing issue)
- Steps skipped: Payment entry (would place a real order), Checkout confirmation
- Secondary flows: Deals page (DealsList), My Rewards (LoyaltyRewards, LoyaltyDeals, LoyaltyPoints), Tracker page
- Coverage: 6 of 7 planned steps completed (checkout redirect was a routing issue, not a block)

## Pages & Interactions
1. Login page (dominos.com/pages/customer/#/customer/login) - user logged in manually
2. Homepage - clicked "Delivery" button
3. Address modal - selected saved address "421 N 63RD", clicked "CONTINUE FOR DELIVERY"
4. Confirm location modal - clicked "CONFIRM LOCATION"
5. Menu page - clicked "SPECIALTY PIZZAS" category
6. Specialty Pizzas - clicked "Add Spicy Chicken Bacon Ranch to Cart"
7. Cart modal - viewed cart contents, scrolled to checkout
8. Cart modal - clicked "CHECK OUT"
9. Deals page (via nav) - viewed all deals
10. My Rewards page (via nav) - viewed rewards tiers and member deals
11. Tracker page (via nav) - viewed tracker interface

## Sniff Configuration
- Backend: agent-browser v0.23.4
- Session: Headed browser, user logged in manually, state saved
- HAR recording: 654 total requests captured
- Proxy pattern: NOT detected (standard REST + GraphQL BFF)
- Effective rate: ~1 req/s (human-paced interaction)

## Endpoints Discovered

### GraphQL BFF (www.dominos.com/api/web-bff/graphql)
| Operation | Type | Description | Called |
|-----------|------|-------------|--------|
| CampaignAdvertisementTile | query | Store campaign ads | 1x |
| CampaignTilesByLocation | query | Location-based campaign tiles | 1x |
| CartById | query | Get cart with items and pricing | 9x |
| CartEtaMinutes | query | Estimated wait time for cart | 7x |
| CartSourceEvent | mutation | Log cart interaction events | 3x |
| Category | query | Menu categories for a store | 1x |
| CheckDraftDeal | mutation | Check and auto-apply deals | 5x |
| CreateCart | mutation | Create a new cart | 1x |
| Customer | query | Auth'd customer profile with saved addresses | 1x |
| DealsList | query | Available deals for store/service method | 2x |
| LoyaltyAvailabilityCounters | query | Loyalty counters for cart | 7x |
| LoyaltyDeals | query | Member-exclusive deals | 1x |
| LoyaltyPoints | query | Points balance and status | 6x |
| LoyaltyRewards | query | Available rewards by tier | 1x |
| PreviousOrderPizzaModal | query | Previous order suggestions | 2x |
| ProductQuantities | query | Cart product quantities | 6x |
| Products | query | Products in a category | 2x |
| QuickAddProductMenu | mutation | Quick-add product to cart | 1x |
| StJudeThanksAndGivingHomePage | query | St. Jude donation tracker | 1x |
| Store | query | Store details and availability | 10x |
| StoreSaltWarningEnabled | query | Salt warning for store | 6x |
| Stores | query | Find stores by proximity | 1x |
| SummaryCharges | query | Cart totals with tax/delivery | 7x |
| UpsellForOrder | query | Upsell suggestions for cart | 7x |

### Legacy REST endpoints (from community wrappers, not directly observed in sniff)
| Method | Path | Description |
|--------|------|-------------|
| GET | /power/store-locator | Find nearby stores |
| GET | /power/store/{id}/profile | Store profile |
| GET | /power/store/{id}/menu | Full menu |
| POST | /power/validate-order | Validate order |
| POST | /power/price-order | Price order |
| POST | /power/place-order | Place order |
| POST | /power/login | Account login |
| GET | /power/tracker | Track order |

## Coverage Analysis
- **Exercised:** Stores, menu categories, products, cart CRUD, deals, loyalty (points/rewards/deals), customer profile, upsells, campaigns
- **Missed:** Tracker API calls (no active order to track), payment/checkout flow (SPA redirect), coupon code validation, order history, gift cards, store hours detail

## Response Samples
- GraphQL responses follow standard `{data: {operationName: {...}}}` envelope
- Cart responses include full item breakdown, pricing, service method, timing
- Store responses include storeAvailability, address, hours, capabilities

## Rate Limiting Events
- No 429 responses encountered during sniff
- All requests returned 200 status

## Authentication Context
- **Authenticated session used:** Yes
- **Transfer method:** Headed browser login (user logged in manually)
- **Auth-only endpoints discovered:** Customer (profile/saved addresses), LoyaltyPoints, LoyaltyRewards, LoyaltyDeals, PreviousOrderPizzaModal
- **Session state excluded from archiving:** Will be cleaned before archive
