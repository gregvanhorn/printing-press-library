# Dub.co CLI Brief

## API Identity
- Domain: Link management, attribution, conversion tracking, affiliate programs
- Users: Marketers, growth engineers, developer tools teams, affiliate managers
- Data profile: Links (core), analytics (clicks/leads/sales), domains, tags, folders, customers, partners, commissions, payouts, bounties, events, QR codes
- Spec: OpenAPI 3.0.3 at https://api.dub.co — 36 paths, well-documented
- Auth: Bearer token (dub_xxxxxx), 60 req/min rate limit
- Official SDKs: TypeScript, Python, Go, Ruby, PHP (all Speakeasy-generated)

## Reachability Risk
- None. Official API with published spec, official SDKs, active development. No 403 issues.

## Top Workflows
1. Create and manage short links (bulk create, upsert, update, delete)
2. Track analytics across links/domains/campaigns (clicks, leads, sales)
3. Manage custom domains (register, verify, configure)
4. Run partner/affiliate programs (create partners, track commissions, manage payouts)
5. Organize links with tags and folders

## Table Stakes
- Link CRUD (create, list, get, update, delete) with bulk operations
- Custom slugs and domains
- Analytics retrieval (clicks, browsers, devices, countries, cities, referers)
- Tag and folder management
- QR code generation
- Event tracking (leads, sales, opens)
- Partner/affiliate management

## Data Layer
- Primary entities: Links, Domains, Tags, Folders, Customers, Partners, Commissions, Payouts, Events
- Sync cursor: Links by updatedAt, Events by timestamp
- FTS/search: Link search by URL, key, tag, domain; Customer search

## Product Thesis
- Name: dub-co-pp-cli
- Why it should exist: No serious CLI exists for dub.co. The community CLI (sujjeee/dubco) only creates links. The MCP server only does CRUD on links. There is zero CLI coverage for analytics, domains, partners, commissions, tags, folders, customers, QR codes, or bulk operations. A CLI with offline analytics cache, bulk link management, and agent-native output would be the first tool that makes dub.co usable from the terminal.

## Build Priorities
1. Full link lifecycle (create, list, get, update, delete, bulk ops, upsert)
2. Analytics retrieval with offline caching
3. Domain management (register, verify, list, configure)
4. Partner/affiliate program management
5. Tag and folder organization
6. Customer and event tracking
7. QR code generation
8. Transcendence: offline analytics trends, link health monitoring, bulk import/export
