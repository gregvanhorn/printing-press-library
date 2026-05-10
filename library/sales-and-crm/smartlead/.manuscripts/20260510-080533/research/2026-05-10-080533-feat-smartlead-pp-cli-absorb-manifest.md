# Smartlead Absorb Manifest

## Sources Absorbed
| Source | Type | Stars | Coverage Contributed |
|--------|------|-------|---------------------|
| LeadMagic/smartlead-mcp-server | TS MCP | 18 | 113 tools across 9 modules |
| bcharleson/smartlead-cli | TS CLI | 2 | 142 commands; added master-inbox, CRM, smart-delivery DNS checks, lead-list import |
| jean-technologies MCP | TS MCP | 17 | n8n-style workflows (overlap with LeadMagic) |
| api.smartlead.ai/llms.txt | Mintlify | n/a | 28 canonical endpoints |
| smartlead-ai/API-Python-Library | Py | 0 | older surface; cross-checked |

## Absorbed Features (155 endpoints across 13 resources — match or beat everything)

Endpoint counts per resource (every endpoint becomes a typed CLI command + an MCP tool via cobratree mirror):

| Resource | Endpoints | Notes |
|----------|-----------|-------|
| campaigns | 20 | create / list / get / status / settings / schedule / sequences / analytics / export / by-lead |
| email_accounts | 28 | save / list / suspend / warmup / health / sending-limits / tags / per-campaign linking + bulk |
| leads | 30 | per-campaign + global; categories, blocklist, push-between-lists, message history, reply, forward |
| lead_lists | 7 | CRUD + import + tag-assignment |
| sequences | 2 | get / save (canonical paths, alongside campaign-nested variants) |
| analytics | 22 | global, per-campaign, per-client, per-mailbox, time-to-first-reply, follow-up rate |
| statistics | 21 | per-campaign deep stats: emails, responses, engagement, conversions, deliverability, A/B, benchmarks |
| webhooks | 18 | create / list / test / events / retry / analytics / delivery logs / per-campaign attachment |
| clients | 9 | list / create / API key mgmt (whitelabel) |
| crm | 10 | tags + notes + tasks per lead |
| master_inbox | 20 | replies / unread / archived / sent / scheduled / snoozed / important / reminders / category / read-status / push-to-subsequence |
| smart_senders | 16 | search-domain / vendors / auto-mailbox / orders / reputation / pause-resume / rotation / load-balancing |
| smart_delivery (spam-test) | 30 | manual + automated tests, results breakdowns (DKIM/SPF/rDNS/blacklists/headers/folders) + DNS checks |

**Universal flags on every command** (the agent-native moat over bcharleson):
- `--json` (structured) / `--select <fields>` / `--csv` / `--compact` / `--quiet` / `--dry-run` / `--limit N`
- Typed exit codes: 0 ok, 2 usage, 3 not-found, 4 auth, 5 rate-limited, 7 transport, 10 server-error
- Auto-chunking on `leads add` (>400 → silent batch)
- Adaptive rate limiter — surfaces `*cliutil.RateLimitError` instead of empty results

## Stubs (none planned)

Every absorbed endpoint will ship as a real working command. Smart-senders order placement (`POST /smart-senders/place-order`) involves real payment and will require `--confirm` and default to `--dry-run` showing the order — not a stub, but gated.

## Transcendence (only possible with our approach)

These are commands that REQUIRE the local SQLite + cross-campaign aggregation — Smartlead's API has no equivalent.

| # | Feature | Command | Why Only We Can Do This | Score |
|---|---------|---------|------------------------|-------|
| 1 | **Stale leads** — leads not touched in N days across ALL campaigns | `smartlead-pp-cli stale --days 14` | API has no global lead view; requires local join across leads × messages × last_event_at | 9/10 |
| 2 | **Overlap leads** — same email in multiple campaigns (duplicate outreach risk) | `smartlead-pp-cli overlap` | API only has `/leads/by-email` one at a time; cross-campaign dedup needs local index | 9/10 |
| 3 | **Mailbox burnout** — email accounts approaching warmup decline or rate-limit ceiling | `smartlead-pp-cli mailbox-burnout` | Requires time-series of warmup-stats + sending-limits snapshots; API gives point-in-time only | 8/10 |
| 4 | **Reply velocity** — time-to-first-reply per campaign, with historical baseline comparison | `smartlead-pp-cli reply-velocity --weeks 4` | Requires snapshotted historical analytics rolled up across campaigns | 7/10 |
| 5 | **Deliverability drift** — today vs 7d vs 30d spam-test placement per sender domain | `smartlead-pp-cli deliv-drift --domain acme.com` | Requires snapshot history of spam-test results; API returns only single-test snapshots | 7/10 |
| 6 | **Reach budget** — given target lead count, calculate days-to-completion per campaign from current sending-limits | `smartlead-pp-cli reach-budget --campaign 3217809` | Requires joining email-account daily limits × campaign rotation pool × pending leads count | 8/10 |
| 7 | **Agency roll-up** — aggregate stats across ALL client API keys, per-client and portfolio-level | `smartlead-pp-cli roll-up --keys-from ~/.smartlead/keys.txt` | Multi-key fan-out + local aggregation; agencies do this manually today | 8/10 |
| 8 | **Reply classify** — local FTS + keyword classifier on master inbox (positive / objection / OOO / unsub) | `smartlead-pp-cli reply-classify --since 7d` | Smartlead's UI categorization is coarse; local classifier + FTS gives fast, scriptable triage | 7/10 |

All transcendence features score ≥ 7/10. None score below threshold.

## Group themes

- **Local state that compounds**: stale, overlap, mailbox-burnout, reply-velocity, deliv-drift, reach-budget
- **Agent-native plumbing**: --json/--select/--csv on every command, typed exit codes, auto-chunking
- **Agency-scale**: roll-up, multi-key fanout, reply-classify
