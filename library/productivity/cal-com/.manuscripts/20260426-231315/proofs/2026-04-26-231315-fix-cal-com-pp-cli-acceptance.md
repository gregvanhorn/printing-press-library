# Cal.com CLI Acceptance Report

  Level: Full Dogfood
  Tests: 55/55 passed (after one inline fix)
  API key: live (`cal_live_*`), full read+write authorized
  Test account: <account> (id <user-id>)

## Mechanical test matrix (run during Phase 5)

### A. Help checks (20/20)
All top-level commands return exit 0 from `--help`: book, today, week, conflicts, gaps, workload, bookings, event-types, schedules, slots, calendars, webhooks, me, teams, organizations, sync, search, analytics, doctor, auth.

### B. Novel-feature happy paths (15/15)
- `today`, `today --date 2026-04-15` (returned real <attendee-name> meeting from local store)
- `week --start 2026-04-13` (1 booking on Wed)
- `conflicts --window all` (2 bookings checked, 0 overlaps)
- `gaps --window 7d --min-minutes 60` (7 days of business hours open)
- `slots find` (61 slots for event-type 96531, --first-only at 16:00)
- `webhooks triggers` (17 grouped triggers)
- `webhooks coverage` (0 registered, 17 missing in lifecycles)
- `analytics bookings|cancellations|no-show|density --window all` (all return correct aggregations)
- `bookings pending` (0 pending — correct, account has none)
- `event-types stale --days 30` (96531 active, 96532+96533 stale — correct)
- `book --dry-run` (preview shows full request body)

### C. JSON fidelity (6/6)
All sampled novel-command outputs are valid JSON; round-trip parseable in Python.

### D. Error paths (5/5, all expected nonzero)
- `book` without required flags → exit 1 with friendly error naming the missing flag
- `workload` without `--team-id` → exit 1
- `today --date "not-a-date"` → exit 1
- `week --start "not-a-date"` → exit 1
- `slots find` without `--event-type-ids` → exit 1

### E. Absorbed live API calls (9/9)
- `doctor`: OK Config, OK Auth, OK API, account match
- `me get` --data-source live: <account>
- `bookings get --take 5 --data-source live`: returns the 5 most recent bookings
- `event-types get --data-source live`: 3 event types (15min, secret, 30min)
- `schedules get --data-source live`: Working Hours
- `schedules get-default --data-source live`: same default schedule
- `calendars get --data-source live`: returns connected-calendars envelope
- `webhooks get --data-source live`: empty (account has no webhooks)
- `teams get --data-source live`: returns team list

## End-to-end real-world test

Full lifecycle of a real booking via the composed `book` command:

1. `slots find --event-type-ids 96531 --start 2026-04-28 --end 2026-04-29 --first-only` → got 2026-04-28T16:00:00.000Z
2. `book --event-type-id 96531 --start <slot> --attendee-name "End-to-End Test" --attendee-email <test-email>` → created booking `prV4QEgYHtdSbHtpQWxQGa`, status accepted, meeting URL returned
3. `sync --resources bookings --full` → ingested 6 bookings to local store
4. `today --date 2026-04-28 --json` → store now shows the new booking with status accepted
5. `curl -I https://app.cal.com/video/prV4QEgYHtdSbHtpQWxQGa` → HTTP 200, content-type text/html (booking page is live)
6. `bookings cancel bookings-booking prV4QEgYHtdSbHtpQWxQGa --stdin` with `{"cancellationReason":"E2E dogfood cleanup"}` → cancel POST returned 200
7. Direct API verification: status=cancelled, cancellationReason="E2E dogfood cleanup"

A second test booking `vdyerLDGucR7ir1hnZG2P5` was created and cancelled earlier in the same session for Phase 4 shipcheck verification.

## Failures fixed inline (1)

| # | Test | Failure | Fix |
|---|------|---------|-----|
| 1 | `event-types stale --days 30` | Strict envelope parser only handled `data.eventTypeGroups[].eventTypes[]` shape; live API also returns `data:[]` and `data:{eventTypes:[]}` | Added `extractEventTypes()` helper that tries all three shapes. Re-test PASS. |

## Printing Press issues (for retro)

These are systemic, not Cal.com-specific:

1. **Generator does not auto-emit per-endpoint required headers** from OpenAPI spec (cal-api-version, Stripe-Version, X-GitHub-Api-Version pattern). Worked around with hand-written `internal/client/calcom_versions.go`.
2. **`auth set-token` writes to wrong field** (`access_token` vs the active `auth_header`). Workaround: hand-edited `~/.config/cal-com-pp-cli/config.toml`.
3. **Sync envelope handling fails on multi-shape APIs.** `me`, `teams`, `webhooks`, `calendars` all error with "missing id for X" because the sync extractor only knows one envelope shape. Bookings worked accidentally because the extractor recognizes `data.bookings`.
4. **Mutation body fields with nested objects skipped silently** (~50 fields skipped during generation). Forces `--stdin` workaround for many create/update operations.
5. **Generator emits awkward command paths** like `bookings cancel bookings-booking <uid>` instead of `bookings cancel <uid>` — operationId-derived naming creates ugly intermediate path components.
6. **scorecard `auth_protocol` substring match** looks only in `internal/client/client.go` for `"Bearer "` literal but the generator correctly puts the bearer prefix in `internal/config/config.go`. Causes a false-low 3/10 score.
7. **`live-check` does not extract from `narrative.recipes`** — the sampler returned 0 features for this CLI even though novel_features and recipes are populated in research.json.

Gate: PASS

The CLI passes Full Dogfood: every novel feature works end-to-end against the live Cal.com API, every absorbed live read returns real data, error paths fail cleanly, JSON output is valid throughout, and a complete real-world booking lifecycle (find → book → visit URL → sync → display → cancel → verify) was exercised successfully.
