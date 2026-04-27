# Cal.com CLI Build Log

## What was built

### Generator output (Priority 0/1)
- 285-endpoint CLI from cached OpenAPI spec at `cal-com-openapi-v2.json`
- Resource groups generated: bookings, calendars, conferencing, event-types, me, oauth, oauth-clients, organizations, routing-forms, schedules, selected-calendars, slots, stripe, teams, verified-resources, webhooks
- Standard scaffolding: doctor, auth (set-token/logout/status), sync, search, analytics (generic), workflow, api-discovery, agent-context, profile, feedback, which, export, import, tail
- Local SQLite store with typed tables for me, schedules, selected_calendars, teams, bookings, event_types, invite, memberships, verified_resources, routing_forms, calculate_slots, conferencing, connect, default, disconnect, oauth, slots, webhooks, stripe, check, calendars
- Auth: Bearer token via env `CAL_COM_TOKEN` and config `~/.config/cal-com-pp-cli/config.toml`
- README with display name "Cal.com", quickstart, troubleshoots, novel-features section

### Hand-built Cal.com novel features (Priority 2)

All 12 transcendence commands from the absorb manifest (no stubs):

| # | Command | Status | Wire location |
|---|---------|--------|---------------|
| T1 | `book` | DONE — composed slot-check → reserve → create → confirm; verified end-to-end with real booking | `internal/cli/calcom_novel.go` + `root.go` |
| T2 | `today` | DONE — store-backed agenda with live fallback | same |
| T3 | `week` | DONE — 7-day rollup, conflict-aware | same |
| T4 | `slots find` | DONE — fanout across multiple event-type IDs, dedup, sort | wired under `slots` |
| T5 | `analytics bookings/cancellations/no-show/density` | DONE — 4 sub-commands extending generic analytics | wired under `analytics` |
| T6 | `webhooks coverage` | DONE — registered triggers vs canonical | wired under `webhooks` |
| T7 | `conflicts` | DONE — overlap detection over local store | top-level |
| T8 | `gaps` | DONE — open windows in business hours | top-level |
| T9 | `workload` | DONE — bookings per host (live API call) | top-level |
| T10 | `event-types stale` | DONE — types with no recent bookings | wired under `event-types` |
| T11 | `bookings pending` | DONE — pending-confirmation sorted by age | wired under `bookings` |
| T12 | `webhooks triggers` | DONE — static catalog of 17 trigger constants (`// pp:novel-static-reference`) | wired under `webhooks` |

### Patches to generated code

1. **`internal/client/calcom_versions.go` (new)** — Path-prefix → `cal-api-version` header map, applied automatically by the client `do()` function. Cal.com requires per-endpoint API versioning (Bookings 2024-08-13, Slots 2024-09-04, etc.). Without this header, the API returns v1-fallback shapes (empty arrays) silently. **Logged for retro: the generator should auto-emit per-endpoint required headers from the OpenAPI spec.**
2. **`internal/client/client.go`** — Two small patches in `do()` and `dryRun()` to merge `requiredHeadersForPath()` into outgoing requests + dry-run output.

## End-to-end verification (real Cal.com API)

- **`doctor`**: OK Config, OK Auth, OK API reachable, account=<account>
- **`bookings get`** with various filters: returns real bookings (no longer empty-fallback shape since cal-api-version header)
- **`book`**: Created real booking `vdyerLDGucR7ir1hnZG2P5` for tomorrow 17:00 UTC against event-type 96531 (15min). Meeting link generated: `https://app.cal.com/video/vdyerLDGucR7ir1hnZG2P5`. Booking confirmed visible in account.
- **`bookings get-bookinguid <uid>`**: returns the same booking with attendee details
- **`bookings cancel`** (via `--stdin`): cancelled the test booking with reason "Cleanup after dogfood". Confirmed via direct API check.
- **`today`** / **`week`** / **`conflicts`**: render local-store data correctly when store is populated
- **`slots find` --first-only**: queried 3 event types, returned earliest slot
- **`webhooks triggers`**: returns 17 grouped triggers
- **`analytics bookings --by status`**: 2 cancelled + 2 accepted = 4 total

## Intentionally deferred

None for the novel feature set. Every Phase 1.5 transcendence feature is wired and runtime-tested.

## Skipped body fields (generator warnings)

The generator skipped ~50 complex body fields (`bookerLayouts`, `bookingFields`, `calVideoSettings`, `hosts`, `locations`, `attendees`, `metadata`, `serviceAccountKey`, `availability`, `overrides`, `enabled`, `response`, `activation`, `steps`, `destinationCalendar`, `disableCancelling`, `disableRescheduling`, `emailSettings`, `color`) from various event-type and webhook create/update endpoints because they're nested objects/arrays. Workaround: agents use `--stdin` to pass full JSON bodies when these fields are needed. **Logged for retro: nested-object body fields should at least produce a `--<field>-json` flag that accepts a raw JSON string.**

## Generator-level issues found (for retro)

These are systemic Printing Press issues, not Cal.com-specific:

1. **Per-endpoint required headers from OpenAPI not auto-emitted.** Cal.com declares `cal-api-version: required: true` on 27 paths. Stripe's `Stripe-Version`, GitHub's `X-GitHub-Api-Version`, etc. follow the same pattern. The generator should walk `parameters[in=header, required=true, schema.default]` and produce a path → header-map at codegen time.
2. **`auth set-token` writes to wrong field.** Saves to `access_token` but the auth resolver reads `auth_header` (full "Bearer ..." string). Fix: derive `auth_header` from `access_token` at save time.
3. **Sync envelope shape extraction.** Cal.com's `/v2/event-types`, `/v2/teams`, `/v2/webhooks`, `/v2/me`, `/v2/calendars` use various envelope shapes (`data.eventTypeGroups[*].eventTypes`, `data:[]`, single object). The default sync extractor handles the `data.bookings` case but fails on others with "missing id for X".
4. **Mutation body fields with nested objects skipped silently.** Forces `--stdin` workaround.
5. **`bookings cancel` aliasing.** The generator emits `bookings cancel bookings-booking <uid>` instead of `bookings cancel <uid>`. The intermediate path component `bookings-booking` is ugly noise from operationId-derived naming.

## File counts

- Total Go files in `internal/cli/`: ~270
- Generated source: ~268
- Hand-written: 2 (`calcom_novel.go`, `internal/client/calcom_versions.go`)
- Patched: 1 (`internal/client/client.go`)

## Build status

- `go build ./...`: PASS
- Binary built: `./cal-com-pp-cli` (~21 MB)
- All 12 novel commands registered and runtime-tested
- All Cal.com API auth/version headers correctly applied
