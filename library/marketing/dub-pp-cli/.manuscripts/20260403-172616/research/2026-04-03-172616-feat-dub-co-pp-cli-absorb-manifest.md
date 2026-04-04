# Dub.co CLI Absorb Manifest

## Sources Analyzed
1. **sujjeee/dubco** (TypeScript CLI) — 3 commands: login, config, link
2. **Gitmaxd/dubco-mcp-server** (MCP) — 3 tools: create_link, update_link, delete_link
3. **dubinc/dub-go** (Official Go SDK) — Full API coverage, 50+ methods
4. **dubinc/dub** (npm SDK) — Full API coverage
5. **dubinc/dub-python** (Python SDK) — Full API coverage

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Create short link | dubco CLI, MCP | `links create --url <url> [--key <slug>]` | --json, --dry-run, --stdin batch, auto-copy to clipboard |
| 2 | Update link | MCP | `links update <id> --url <url>` | --dry-run, --json, bulk via --stdin |
| 3 | Delete link | MCP | `links delete <id>` | --force, --dry-run, bulk delete |
| 4 | List links | Go SDK | `links list [--tag <tag>] [--domain <domain>]` | Offline search via SQLite, --json, --select, --limit |
| 5 | Get link info | Go SDK | `links get <id-or-key>` | Cached locally, --json, --select |
| 6 | Link count | Go SDK | `links count [--tag <tag>]` | Offline count from store |
| 7 | Bulk create links | Go SDK | `links bulk-create --file <csv/json>` | CSV/JSON import, --dry-run, progress bar |
| 8 | Bulk update links | Go SDK | `links bulk-update --file <csv/json>` | Same as above |
| 9 | Bulk delete links | Go SDK | `links bulk-delete --ids <id1,id2,...>` | --force confirmation |
| 10 | Upsert link | Go SDK | `links upsert --url <url> --key <slug>` | Idempotent, agent-safe |
| 11 | List domains | Go SDK | `domains list` | Offline, --json |
| 12 | Create domain | Go SDK | `domains create <domain>` | --dry-run |
| 13 | Update domain | Go SDK | `domains update <slug>` | --dry-run |
| 14 | Delete domain | Go SDK | `domains delete <slug>` | --force |
| 15 | Register domain | Go SDK | `domains register <domain>` | --dry-run |
| 16 | Check domain status | Go SDK | `domains status <slug>` | --json, cached |
| 17 | Retrieve analytics | Go SDK | `analytics [--group-by clicks\|leads\|sales]` | Offline cache, trend charts, --json |
| 18 | List events | Go SDK | `events list [--type click\|lead\|sale]` | Offline, --json, --limit |
| 19 | Create tag | Go SDK | `tags create <name> [--color <color>]` | --dry-run |
| 20 | List tags | Go SDK | `tags list` | --json |
| 21 | Update tag | Go SDK | `tags update <id> --name <name>` | --dry-run |
| 22 | Delete tag | Go SDK | `tags delete <id>` | --force |
| 23 | Create folder | Go SDK | `folders create <name>` | --dry-run |
| 24 | List folders | Go SDK | `folders list` | --json |
| 25 | Update folder | Go SDK | `folders update <id>` | --dry-run |
| 26 | Delete folder | Go SDK | `folders delete <id>` | --force |
| 27 | List customers | Go SDK | `customers list` | Offline, --json |
| 28 | Get customer | Go SDK | `customers get <id>` | --json |
| 29 | Update customer | Go SDK | `customers update <id>` | --dry-run |
| 30 | Delete customer | Go SDK | `customers delete <id>` | --force |
| 31 | List partners | Go SDK | `partners list` | --json |
| 32 | Create partner | Go SDK | `partners create` | --dry-run |
| 33 | Partner analytics | Go SDK | `partners analytics` | Offline trend |
| 34 | Ban partner | Go SDK | `partners ban <id>` | --force |
| 35 | Deactivate partner | Go SDK | `partners deactivate <id>` | --force |
| 36 | Partner links | Go SDK | `partners links list` | --json |
| 37 | Create partner link | Go SDK | `partners links create` | --dry-run |
| 38 | Upsert partner link | Go SDK | `partners links upsert` | Idempotent |
| 39 | List commissions | Go SDK | `commissions list` | --json |
| 40 | Update commission | Go SDK | `commissions update <id>` | --dry-run |
| 41 | Bulk update commissions | Go SDK | `commissions bulk-update` | --dry-run |
| 42 | List payouts | Go SDK | `payouts list` | --json |
| 43 | Get QR code | Go SDK | `qr <url> [--size <px>]` | Save to file, --format png/svg |
| 44 | Track lead | Go SDK | `track lead` | --dry-run |
| 45 | Track sale | Go SDK | `track sale` | --dry-run |
| 46 | Track open | Go SDK | `track open` | --dry-run |
| 47 | List bounty submissions | Go SDK | `bounties submissions list <bountyId>` | --json |
| 48 | Approve bounty submission | Go SDK | `bounties submissions approve <bountyId> <submissionId>` | --dry-run |
| 49 | Reject bounty submission | Go SDK | `bounties submissions reject <bountyId> <submissionId>` | --dry-run |
| 50 | Create referral embed token | Go SDK | `embed referral-token` | --json |
| 51 | Auth login | dubco CLI | `auth login [--token <key>]` | Env var + interactive |
| 52 | Config display | dubco CLI | `config show` | --json |

## Transcendence (only possible with our local data layer)

| # | Feature | Command | Why Only We Can Do This |
|---|---------|---------|------------------------|
| 1 | Link performance ranking | `links top [--by clicks\|leads\|sales] [--period 7d\|30d]` | Requires local join of links + analytics over time |
| 2 | Dead link detection | `links dead` | Requires local link store + HTTP health checks against destination URLs |
| 3 | Analytics trends & deltas | `analytics trends [--period 7d\|30d]` | Requires historical analytics snapshots in SQLite to compute week-over-week changes |
| 4 | Domain health dashboard | `domains health` | Requires local join of domains + DNS checks + link counts + analytics per domain |
| 5 | Tag usage report | `tags report` | Requires local join of tags + links + analytics to show which tags drive the most clicks |
| 6 | Campaign builder | `campaign create --name <name> --urls <file> --tag <tag> --domain <domain>` | Bulk-creates tagged links under a domain from a URL list — compound of links + tags + domains |
| 7 | Link export/import | `links export [--format csv\|json]` / `links import --file <path>` | Full offline backup/restore of link inventory with metadata |
| 8 | Partner leaderboard | `partners leaderboard [--by revenue\|clicks\|conversions]` | Requires local join of partners + commissions + analytics |
| 9 | Stale link finder | `links stale [--days 30]` | Finds links with zero clicks in N days — requires local analytics history |
| 10 | Cross-entity search | `search <query>` | FTS5 across links, tags, folders, customers, domains simultaneously |

## Auto-Suggested Novel Features

### Scored Candidates (>= 5/10)

1. **Link performance ranking** (9/10) — Every link manager wants "what's working?" No existing tool answers this from the terminal.
2. **Dead link detection** (8/10) — Marketers create hundreds of links and forget about them. Health-checking destination URLs catches 404s before users do.
3. **Analytics trends** (8/10) — The dub.co dashboard shows point-in-time analytics. Week-over-week delta requires storing snapshots locally.
4. **Campaign builder** (8/10) — Creating 50 tagged links for a campaign is tedious via API calls. One command from a CSV file.
5. **Stale link finder** (7/10) — "Show me links nobody clicked in 30 days" requires historical data.
6. **Cross-entity search** (7/10) — Search across links, tags, customers, domains in one query.
7. **Link export/import** (7/10) — Backup/migrate link inventory. No existing tool does this.
8. **Domain health dashboard** (6/10) — DNS + analytics + link count per domain in one view.
9. **Tag usage report** (6/10) — Which tags drive the most traffic? Requires join across entities.
10. **Partner leaderboard** (6/10) — Rank partners by revenue/clicks. Requires local commission data.
