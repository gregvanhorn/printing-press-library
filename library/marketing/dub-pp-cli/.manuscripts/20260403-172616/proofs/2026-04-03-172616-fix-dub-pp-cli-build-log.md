# Dub CLI Build Log

## What Was Built

### Priority 0 (Foundation) - Generator-built
- SQLite store with per-entity tables: links, domains, tags, folders, customers, partners, commissions, payouts, bounties, events, analytics
- FTS5 search indexes on links (description, title), folders (description, name), partners (description, name)
- Generic resources table with FTS5 for cross-entity search
- Sync state management, batch upsert, name resolution
- Archive/sync workflow command
- Export/import commands
- Tail (live polling) command

### Priority 1 (Absorbed - 52 features matched)
All 52 absorbed features from the manifest are implemented as generated commands:
- Links: create, list, get, get-info, get-count, update, delete, upsert, bulk-create, bulk-update, bulk-delete
- Domains: create, list, update, delete, register, check-status
- Tags: create, list, update, delete
- Folders: create, list, update, delete
- Customers: list, get, update, delete
- Partners: list, create, ban, deactivate, links (list, create, upsert), analytics
- Commissions: list, update, bulk-update
- Payouts: list
- Analytics: retrieve
- Events: list
- QR: get
- Track: lead, sale, open
- Bounties: submissions list, approve, reject
- Tokens: create referral embed
- Auth: login
- Config: show (via auth command)

### Priority 2 (Transcendence - 9 novel features)
- `links top` - Rank links by clicks/leads/sales from local data
- `links dead` - Health-check destination URLs for 404s/errors
- `links stale` - Find links with zero clicks in N days
- `workflow campaign` - Bulk-create tagged links from a URL/CSV file
- `workflow tags-report` - Aggregate tag performance (clicks/leads/sales per tag)
- `workflow domains-health` - Domain DNS + link count + click performance dashboard
- `workflow partners-leaderboard` - Rank partners by revenue/clicks/conversions
- `search` - Cross-entity FTS5 search (generator-built)
- `export`/`import` - Full offline backup/restore (generator-built)

### Priority 3 (Polish)
- Rewrote root command description to be user-facing
- Fixed duplicate analytics command registration
- Formatted all Go code with gofmt

## Skipped Complex Body Fields
- geo (complex geotargeting object)
- testVariants (A/B test config)
- webhookIds (array)
- linkProps (complex partner link properties)
- metadata (arbitrary JSON)
- data (bulk operations array)
- externalIds, linkIds, commissionIds (ID arrays)

These are available via --stdin JSON input.

## Deferred
- analytics trends (requires multiple sync snapshots over time — works after first sync + re-sync)
