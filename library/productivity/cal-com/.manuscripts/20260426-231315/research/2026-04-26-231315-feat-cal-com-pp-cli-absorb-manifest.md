# Cal.com Absorb Manifest

## Sources Cataloged
1. **bcharleson/calcom-cli** (TS, 61 tools) — closest competing CLI
2. **calcom/cal-mcp** (official, 167 endpoint mirrors, dormant) — vendor reference
3. **Composio Cal toolkit** (~140 hosted actions) — broadest checklist
4. **dsddet/booking_chest** (Python MCP, ~70 v2 tools) — second-best v2 coverage
5. **aditzel/caldotcom-api-v2-sdk** (TS SDK, 12 resource groups, 3 auth modes) — auth pattern reference
6. **terrylica/cc-skills calcom-commander** — adjacent agent CLI

## Absorbed (match or beat everything that exists)

### Auth & Profile (5)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Auth login/store key | bcharleson `auth login` | `auth set-token` writes to `~/.config/cal-com-pp-cli/config.toml` | Multi-source: env `CAL_COM_API_KEY`, flag `--api-key`, file. Doctor diagnostics for each. |
| 2 | Auth status | bcharleson `auth-status` | `auth status` + `doctor` | Doctor checks reachability, scopes, token shape, account match |
| 3 | Logout | bcharleson `auth logout` | `auth logout` | Clears stored creds |
| 4 | Get profile (me) | cal-mcp `getMe` | `me get` | Cached in store after first call; offline `me show` uses cache |
| 5 | Update profile | bcharleson `profile update` | `me update` | `--dry-run`, atomic field flags |

### Bookings (16)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 6 | List bookings | bcharleson `bookings list` | `bookings list` with all filters: `--status`, `--attendee-email`, `--event-type-id`, `--after-start`, `--before-end`, `--sort-start`, `--take`, `--skip` | Pagination wrapper, `--all` to auto-iterate, default reads from local store with `--live` to bypass |
| 7 | Get booking | cal-mcp `getBooking` | `bookings get <uid>` | Cached after sync |
| 8 | Create booking | bcharleson `bookings create` | `bookings create` with `--dry-run`, `--stdin` for batch | Show full request body in dry-run, machine-readable response |
| 9 | Reschedule | bcharleson `bookings reschedule` | `bookings reschedule <uid> --start ... [--reason]` | `--dry-run` |
| 10 | Cancel | bcharleson `bookings cancel` | `bookings cancel <uid> [--reason]` | `--dry-run` |
| 11 | Confirm pending | cal-mcp `confirmBooking` | `bookings confirm <uid>` | Lists pending if no uid given (transcendence overlap) |
| 12 | Decline pending | cal-mcp `declineBooking` | `bookings decline <uid> [--reason]` | `--dry-run` |
| 13 | Mark no-show | cal-mcp `markNoShow` | `bookings mark-absent <uid> [--host] [--attendees email,...]` | |
| 14 | Reassign (round-robin) | cal-mcp `reassignBooking` | `bookings reassign <uid> [--user-id N]` | |
| 15 | Update location | bcharleson `bookings update-location` | `bookings update-location <uid> --type ... --link ...` | |
| 16 | List attendees | bcharleson `bookings attendees` | `bookings attendees <uid>` | |
| 17 | Add attendee | bcharleson `bookings add-attendee` | `bookings attendees add <uid> --name --email --tz` | |
| 18 | Add guests | bcharleson `bookings add-guests` | `bookings guests add <uid> --emails a@b,c@d` | |
| 19 | Get calendar links (.ics + provider links) | cal-mcp `getCalendarLinks` | `bookings calendar-links <uid>` | Direct .ics download with `--save`  |
| 20 | List recordings | cal-mcp `getRecordings` | `bookings recordings <uid>` | |
| 21 | Get transcripts | cal-mcp `getTranscripts` | `bookings transcripts <uid>` | |

### Event Types (7)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 22 | List event types | bcharleson `event-types list` | `event-types list [--username u]` | Cached in store |
| 23 | Get event type | bcharleson `event-types get` | `event-types get <id>` | |
| 24 | Create event type | bcharleson `event-types create` | `event-types create --title --slug --length [...]` | `--dry-run` |
| 25 | Update event type | bcharleson `event-types update` | `event-types update <id> [...]` | `--dry-run` |
| 26 | Delete event type | bcharleson `event-types delete` | `event-types delete <id>` | `--dry-run` |
| 27 | Search event types (FTS) | absent in field | `event-types search <query>` | Local FTS over title/slug/description |
| 28 | Webhooks per event type | cal-mcp `EventTypeWebhooksController` | `event-types webhooks <id>` | List/create/delete scoped to one event type |

### Schedules (6)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 29 | List schedules | bcharleson `schedules list` | `schedules list` | Cached |
| 30 | Get default schedule | bcharleson `schedules default` | `schedules default` | |
| 31 | Get schedule | bcharleson `schedules get` | `schedules get <id>` | |
| 32 | Create schedule | bcharleson `schedules create` | `schedules create --name --availability ...` | `--dry-run`, accepts JSON or weekday-time syntax |
| 33 | Update schedule | bcharleson `schedules update` | `schedules update <id> ...` | `--dry-run` |
| 34 | Delete schedule | bcharleson `schedules delete` | `schedules delete <id>` | `--dry-run` |

### Slots (5)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 35 | Available slots | bcharleson `slots available` | `slots available --event-type-id N --start ... --end ...` | Default window = next 7 days, accepts natural-language dates |
| 36 | Reserve slot | bcharleson `slots reserve` | `slots reserve --event-type-id N --slot ...` | Returns reservation uid |
| 37 | Get reserved slot | bcharleson `slots get-reserved` | `slots reserved get <uid>` | |
| 38 | Update reserved | bcharleson `slots update-reserved` | `slots reserved update <uid> ...` | |
| 39 | Delete reserved | bcharleson `slots delete-reserved` | `slots reserved delete <uid>` | |

### Calendars (5)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 40 | List connected calendars | bcharleson `calendars list` | `calendars list` | Cached |
| 41 | Busy times | bcharleson `calendars busy` | `calendars busy --from ... --to ...` | Default = next 7 days |
| 42 | Check connection | bcharleson `calendars check` | `calendars check <credId>` | |
| 43 | Save Apple/CalDAV creds | bcharleson `calendars save-credentials` | `calendars save-credentials --type apple --username --password` | |
| 44 | Disconnect | bcharleson `calendars disconnect` | `calendars disconnect <credId>` | |

### Selected & Destination Calendars (3)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 45 | Add selected calendar | bcharleson `selected-calendars add` | `selected-calendars add --integration --external-id --credential-id` | |
| 46 | Delete selected calendar | bcharleson `selected-calendars delete` | `selected-calendars delete ...` | |
| 47 | Set destination calendar | bcharleson `destination-calendars update` | `destination-calendars set --integration --external-id` | |

### Webhooks (5)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 48 | List webhooks | bcharleson `webhooks list` | `webhooks list` | Cached, FTS over subscriber URL |
| 49 | Get webhook | bcharleson `webhooks get` | `webhooks get <id>` | |
| 50 | Create webhook | bcharleson `webhooks create` | `webhooks create --url --triggers BOOKING_CREATED,...  --secret` | `--dry-run`, `triggers list` to see all options |
| 51 | Update webhook | bcharleson `webhooks update` | `webhooks update <id> ...` | |
| 52 | Delete webhook | bcharleson `webhooks delete` | `webhooks delete <id>` | |

### Out-of-Office (4)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 53 | List OOO entries | bcharleson `out-of-office list` | `ooo list` | Cached |
| 54 | Create OOO | bcharleson `out-of-office create` | `ooo create --start --end [--notes]` | `--dry-run` |
| 55 | Update OOO | bcharleson `out-of-office update` | `ooo update <id> ...` | |
| 56 | Delete OOO | bcharleson `out-of-office delete` | `ooo delete <id>` | |

### Teams (5)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 57 | List teams | bcharleson `teams list` | `teams list` | Cached |
| 58 | Get team | bcharleson `teams get` | `teams get <id>` | |
| 59 | Create team | bcharleson `teams create` | `teams create --name --slug` | `--dry-run` |
| 60 | Update team | bcharleson `teams update` | `teams update <id> ...` | `--dry-run` |
| 61 | Delete team | bcharleson `teams delete` | `teams delete <id>` | `--dry-run` |

### Conferencing (4)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 62 | List apps | bcharleson `conferencing list` | `conferencing list` | |
| 63 | Get default app | bcharleson `conferencing default` | `conferencing default` | |
| 64 | Set default | bcharleson `conferencing set-default` | `conferencing set-default --app zoom` | |
| 65 | Connect/disconnect | bcharleson `conferencing connect/disconnect` | `conferencing connect|disconnect --app zoom` | |

### Stripe (3)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 66 | Check Stripe connection | bcharleson `stripe check` | `stripe check` | |
| 67 | Connect URL | bcharleson `stripe connect` | `stripe connect` | |
| 68 | Save creds | bcharleson `stripe save-credentials` | `stripe save-credentials --code` | |

### Sync & Search (4)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 69 | Sync all data | absent in field | `sync` (incremental) / `sync --full` | None of the competitors have offline mirror |
| 70 | Cross-entity search | absent in field | `search <query>` | FTS5 across bookings, event_types, teams, webhooks |
| 71 | Resource list | (machine-driven) | `resources list` | Generator-emitted manifest of available resources |
| 72 | Doctor | (machine-driven) | `doctor` | Multi-mode auth diagnostics, reachability, store integrity |

**Total absorbed: 72 commands** (vs bcharleson's 61, dsddet's 70, cal-mcp's 167 raw mirrors).

## Transcendence (only possible with our approach)

Reconciled with the prior April 5 cal-com CLI (which shipped 8 novel features). Kept what served the real Cal.com user (solo professionals, small sales/support teams running round-robin, AI agents booking on someone's behalf), added two from prior I had missed (`gaps`, `workload`), folded `noshow` into `analytics cancellations` (a flag mode), kept `search` as part of the absorbed FTS surface.

| # | Feature | Command | Why Only We Can Do This | Score |
|---|---------|---------|------------------------|-------|
| T1 | One-shot booking flow | `book --event-type-id N --start "tomorrow 2pm" --attendee-name "X" --attendee-email "y@z"` | Composed: find slots, validate, optionally reserve, create, confirm — all in one. No competing tool composes endpoint calls. | 9/10 |
| T2 | Today's agenda | `today [--user-id N]` | Local-store query joining bookings + event types + attendees, sorted by start, with status badges. No API call needed once synced. | 8/10 |
| T3 | Week view | `week [--start mon]` | 7-day calendar view from store, with conflict highlighting. | 7/10 |
| T4 | Cross-event-type slot search | `slots find --event-type-ids 96531,96532,96533 --start ... --end ... [--first-only]` | The API only takes one event type at a time. Our command fans out, merges, dedups, and ranks. | 8/10 |
| T5 | Booking analytics | `analytics --window 30d [--by event-type|attendee|weekday|hour] [--metric volume|cancellations|no-show|duration]` | Aggregations across all bookings: no-show rate, cancellation rate, booking density. Replaces prior `stats` and `noshow` commands as one tool with metric flags. The API has no analytics endpoints; Cal.com Insights is paid-tier. | 8/10 |
| T6 | Webhook coverage | `webhooks coverage` | Compares registered triggers against the full canonical set; flags lifecycle events with no subscriber. Composio doesn't have this. | 7/10 |
| T7 | Conflict detection | `conflicts [--user-id N] [--window 7d]` | Joins active bookings against external calendar busy-times to find overlaps that the booking flow may have missed (e.g. last-minute calendar additions). | 8/10 |
| T8 | Availability gap finder | `gaps [--event-type-id N] [--window 7d] [--min-minutes 30]` | Finds open windows in a schedule that are available but unbooked. Solo professionals optimizing their week, sales teams identifying capacity. From prior CLI. | 7/10 |
| T9 | Team workload balance | `workload --team-id N [--window 30d]` | Booking distribution across team members for round-robin tuning. Identifies overloaded vs underutilized hosts. From prior CLI. | 6/10 |
| T10 | Stale event types | `event-types stale [--days 30]` | Event types with zero bookings in the last N days. Used to clean up dead links. From prior CLI. | 6/10 |
| T11 | Pending review | `bookings pending [--max-age 24h]` | Pending-confirmation bookings approaching expiration; default sort by oldest. | 6/10 |
| T12 | Webhook trigger catalog | `webhooks triggers` | Static reference of all valid `triggers` enum values, grouped by lifecycle stage. Saves agents from guessing exact strings. (`// pp:novel-static-reference`) | 5/10 |

**Total transcendence: 12 commands.** All score >= 5/10 — every one belongs in the manifest.

### Reconciliation notes vs prior CLI (2026-04-05)
- **Kept (overlap):** `today`, `conflicts`, `stale` (rebadged as `event-types stale`), analytics (prior `stats`+`noshow` folded into one `analytics` command with `--metric` flags), `search` (now part of absorbed FTS surface as `search`, `event-types search`, etc.).
- **Added from prior:** `gaps`, `workload` — both serve real personas the new list missed.
- **New ideas (not in prior):** `book` (composed booking), `week`, `slots find --event-type-ids` (cross-type), `webhooks coverage`, `bookings pending`, `webhooks triggers` (static reference).
- **Cut from prior:** `noshow` as a standalone command (folded into `analytics --metric no-show`); `search` as a transcendence feature (it's table-stakes, lives in absorbed surface).

## Stubs
None. Every absorbed and transcendence feature is implementable with the OpenAPI spec + local store. No paid-API gating, no headless browser required.

## Summary
- 72 absorbed + 12 transcendence = **84 commands**
- 38% more than bcharleson's 61 (closest competing CLI)
- Composed intents, offline store, and analytics make this the agent-native, intent-driven Cal.com CLI that no existing tool delivers.
