# Credits

This document acknowledges the open-source projects whose work informed the design of dominos-pp-cli. We studied these tools to understand the Domino's API surface, common ordering flows, and the design tradeoffs involved in wrapping a fast-food API.

## Community Projects

| Project | Author | Language | What We Learned |
|---------|--------|----------|-----------------|
| [apizza](https://github.com/harrybrwn/apizza) | harrybrwn | Go | Cart management, topping syntax, Go API wrapper (dawg library), auth flow |
| [node-dominos-pizza-api](https://github.com/RIAEvangelist/node-dominos-pizza-api) | RIAEvangelist | JavaScript | Most comprehensive endpoint docs, tracking, international support, error codes |
| [pizzapi](https://github.com/ggrammar/pizzapi) | ggrammar | Python | Customer/address/store/order flow |
| [dominos](https://github.com/tomasbasham/dominos) | tomasbasham | Python | UK API variant, 403 reachability risk signal |
| [mcpizza](https://github.com/GrahamMcBain/mcpizza) | GrahamMcBain | Python | MCP tool patterns, safety-first ordering |
| [pizzamcp](https://github.com/GrahamMcBain/pizzamcp) | GrahamMcBain | JavaScript | End-to-end ordering via MCP, payment flow |
| dominos-canada | - | JavaScript | Canadian endpoint patterns |
| ez-pizza-api | - | JavaScript | Simplified ordering wrapper |

Note: dominos-canada and ez-pizza-api were referenced during research but do not have confirmed GitHub URLs.

## A Note on Freshness

Many of these projects are years old and may not work against the current Domino's API. Domino's regularly changes endpoints, validation rules, and response formats. That said, the design patterns and API endpoint documentation these projects provided were invaluable for building dominos-pp-cli. We are grateful to every author who took the time to reverse-engineer and document what they found.

## The GraphQL BFF Discovery

The sniff discovery (logged into dominos.com, walked the ordering flow) revealed a GraphQL BFF with 24 operations that none of the community tools had documented. This layer sits between the public-facing website and the underlying REST APIs, and it represents a more structured and potentially more stable interface than the raw endpoints most community tools target.
