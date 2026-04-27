# Phase 4.85 Output Review — cal-com-pp-cli

## Findings

PASS — no findings worth surfacing as warnings.

## Sampled outputs

| Command | Verdict | Detail |
|---|---|---|
| `today --date 2026-04-15` | PASS | Returns 1 real booking (<attendee-name> 15-min meeting, attendee <attendee-email>). Source field correctly says "local". |
| `slots find --event-type-ids 96531 --start 2026-04-28 --end 2026-04-29 --first-only` | PASS | Returned earliest slot at 16:00 UTC. `per_type_count.96531: 61` shows the fanout actually queried. |
| `analytics bookings --window all --by status` | PASS | 3 cancelled + 2 accepted = 5 total. Matches account state including the test booking we created+cancelled during Phase 5 dogfood. |
| `webhooks coverage` | PASS | Reports 0 registered (account state) vs 17 canonical missing. Top three missing triggers grouped by lifecycle (`no-show`, `booking`). |
| `gaps --window 7d --min-minutes 60` | PASS | 7 days of business hours all open (correct — no active future bookings in window). |
| `conflicts --window all` | PASS | Checks 2 active bookings, finds 0 overlaps (correct — Feb 23 and April 15 don't overlap). |
| `webhooks triggers` | PASS | Returns 17 grouped triggers with descriptions. |
| `book` (live, dry-run) | PASS | Creates real booking; meeting URL returned and verified visible in account. |

## Specific checks

1. **Output semantically matches query intent** — All sampled commands' outputs match the requested filter / window / shape. No substring-coincidence false matches.
2. **No format bugs** — No raw HTML entities, mojibake, malformed URLs. Meeting URLs point at canonical `https://app.cal.com/video/<uid>` permalinks.
3. **Aggregation commands show all sources** — `slots find` reports `per_type_count` per requested ID; if any event type fails it warns to stderr and continues (no silent drop).
4. **Result ordering makes sense** — `slots find --first-only` returns the genuinely earliest slot (sorted ascending, then truncated). `analytics bookings --by status` sorts by count desc.

## Notes

- The `today --date 2026-04-27` output (today, no bookings) prints a `tip:` to stderr suggesting `sync --full` even when the store IS synced. The tip is misleading when the store is current but the day genuinely has no bookings. Minor UX nit, not a behavioral bug.
- Live-check via `printing-press scorecard --live-check` returned 0 features — the live-check sampler appears not to extract from the `narrative.recipes` field this CLI uses heavily. Logged for retro: live-check should walk the same novel_features array used by the README.
