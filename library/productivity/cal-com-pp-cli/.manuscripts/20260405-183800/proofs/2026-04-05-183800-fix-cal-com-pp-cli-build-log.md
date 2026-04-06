# Build Log: cal-com-pp-cli

## What Was Built

### Priority 0 (Foundation)
- SQLite store with bookings, event-types, schedules, calendars, teams, webhooks tables
- FTS5 full-text search indexes on calendars and event-types
- Sync command with parallel workers, resumable cursors, incremental sync
- Per-endpoint version headers in sync (bookings: 2024-08-13, event-types: 2024-06-14, schedules: 2024-06-11)
- Fixed sync resource paths (removed non-functional users/{clientId} path, added me endpoint)

### Priority 1 (Absorbed — 330 generated commands)
- All 285 API operations mapped to CLI commands
- Bookings CRUD (create, list, get, reschedule, cancel, confirm, decline, mark-absent, reassign)
- Event types CRUD (create, list, get, update, delete)
- Schedules CRUD (create, list, get default, update, delete)
- Slots availability queries
- Calendar connections, busy times, destination calendar
- Teams CRUD with membership management
- Webhooks CRUD
- Conferencing connections
- OAuth client management
- Profile (me) endpoint
- Search, export, import, tail, analytics, workflow commands

### Priority 2 (Transcendence — 7 hand-built commands + 1 generated)
1. `today` — Daily schedule dashboard with attendee details, conferencing links, time-until
2. `conflicts` — Double-booking and schedule overlap detection
3. `stats` — Booking analytics (volume, busiest hours, cancellation rates, avg duration)
4. `noshow` — No-show pattern analysis by event type, day, hour
5. `gaps` — Unbooked availability window finder
6. `workload` — Team member booking distribution
7. `stale` — Event types with no recent bookings
8. `search` — FTS5 offline search (generated)

### Priority 3 (Polish)
- Rewrote CLI Short description from spec description to user-facing text
- Added QueryJSON method to store for transcendence command data access

## Intentionally Deferred
- Complex body fields (availability overrides, booking fields, hosts arrays) — skipped by generator
- Per-endpoint cal-api-version headers on individual non-sync commands — will be caught by verify
- Organization-scoped admin endpoints — require org-level auth

## Generator Limitations Found
- No securitySchemes in spec → auth enrichment needed before generation (machine has the fix in PR #136)
- No servers block → manual enrichment needed (retro finding F5)
- Per-endpoint version headers not propagated to individual command files (sync fixed manually)
