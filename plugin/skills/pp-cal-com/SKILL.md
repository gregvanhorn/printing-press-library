---
name: pp-cal-com
description: "Use this skill whenever the user asks about their Cal.com schedule, bookings, upcoming meetings, event types, availability, calendar conflicts, booking analytics, or wants to create / reschedule / cancel a Cal.com booking. Cal.com CLI covering 285 API operations with offline search, booking analytics, conflict detection, and 7 insight commands for scheduling intelligence. Requires a Cal.com API key. Triggers on phrasings like 'what's on my Cal.com today', 'any conflicts in my calendar next week', 'how many bookings did I have this month', 'show my no-shows', 'which event types convert best'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["cal-com-pp-cli"],"env":["CAL_COM_API_KEY"]},"primaryEnv":"CAL_COM_API_KEY","install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-cli@latest","bins":["cal-com-pp-cli"],"label":"Install via go install"}]}}'
---

# Cal.com — Printing Press CLI

Manage bookings, event types, schedules, and availability via the Cal.com API. Covers 285 API operations across 181 paths with a SQLite data layer for offline search and 7 insight commands no other Cal.com tool offers.

## When to Use This CLI

Reach for this when a user wants to manage or analyze their Cal.com scheduling from outside the web UI — checking today's agenda, spotting conflicts, reviewing no-shows, analyzing conversion rates by event type, or batch-managing bookings. Also useful for agent-driven workflows that schedule or reschedule via natural language.

Don't reach for this when the user wants a Google Calendar or Outlook lookup (use their respective CLIs / MCPs) or when the scheduling platform in question isn't Cal.com.

## Unique Capabilities

The 7 insight commands that require the local SQLite data layer.

### Scheduling intelligence

- **`today`** — Today's schedule with attendee details and conferencing links. The daily kickoff command.

  _Compresses the "check my day" ritual into a single JSON blob: meeting time, attendee emails, Zoom/Meet links, agenda notes._

- **`conflicts`** — Detects overlapping bookings or bookings inside blocked-off windows. Runs against the synced local store so it sees cross-calendar conflicts too.

- **`gaps`** — Free-time slots between bookings. Useful for "when can I take a break this week."

### Post-meeting analysis

- **`stats`** — Booking analytics: total bookings, no-show rate, average duration, by event-type breakdown.

- **`noshow`** — Dedicated no-show reporting with per-attendee history.

- **`workload`** — Calendar density analysis. How many hours booked per day, trend over time.

### Hygiene

- **`stale`** — Event types nobody's booked in N days. Candidates for removal or promotion.

## Command Reference

Core resources (each supports list/get/create/update/delete):

- `cal-com-pp-cli bookings` — Meetings
- `cal-com-pp-cli event-types` — Bookable event types
- `cal-com-pp-cli schedules` — Availability schedules
- `cal-com-pp-cli slots` — Available time slots for an event type
- `cal-com-pp-cli calendars` — Connected calendars (Google, Outlook, etc.)
- `cal-com-pp-cli me` — Profile
- `cal-com-pp-cli teams` — Teams
- `cal-com-pp-cli webhooks` — Webhook subscriptions
- `cal-com-pp-cli attendees` / `attributes` — Attendee and custom attribute management
- `cal-com-pp-cli api-keys` — API key management

Booking operations:

- `cal-com-pp-cli booking-add <bookingUid>` — Add attendee to booking
- `cal-com-pp-cli booking-get-booking <bookingUid>` — Detail
- `cal-com-pp-cli booking-update-booking <bookingUid>` — Modify

Unique insight commands:

- `cal-com-pp-cli today` — Today's schedule
- `cal-com-pp-cli conflicts` — Overlap detection
- `cal-com-pp-cli gaps` — Free time
- `cal-com-pp-cli stats` — Analytics
- `cal-com-pp-cli noshow` — No-show report
- `cal-com-pp-cli workload` — Density analysis
- `cal-com-pp-cli stale` — Unused event types

Utility:

- `cal-com-pp-cli sync` / `export` / `import` / `archive` — Local store
- `cal-com-pp-cli search "<query>"` — Full-text across bookings/types/attendees
- `cal-com-pp-cli auth set-token <CAL_COM_API_KEY>`
- `cal-com-pp-cli doctor` — Verify

## Recipes

### Morning schedule check

```bash
cal-com-pp-cli today --agent
cal-com-pp-cli conflicts --agent  # catch overlapping slots
```

`today` returns today's meetings with conferencing links inline; `conflicts` checks for any overlaps you might have missed.

### Weekly workload review

```bash
cal-com-pp-cli workload --period 7d --agent
cal-com-pp-cli gaps --days 7 --min-gap 60 --agent
```

Workload shows total booked hours per day; `gaps` with a 60-minute floor surfaces open blocks that could fit deep work. (Workload uses `--period` as a free-form window; gaps uses `--days` for the lookahead and `--min-gap` in minutes.)

### Post-month analytics for event-type tuning

```bash
cal-com-pp-cli stats --period 30d --agent
cal-com-pp-cli noshow --agent
cal-com-pp-cli stale --days 60 --agent
```

Stats shows which event types get booked most; no-show flags problematic attendees or event types; `stale` identifies event types to delete or promote.

### Find a gap and book

```bash
# List available slots between two ISO-8601 UTC datetimes for an event type:
cal-com-pp-cli slots \
  --event-type-slug "intro-call" --username alice \
  --start 2026-05-15T00:00:00Z --end 2026-05-15T23:59:59Z --agent

# Create a booking by piping a JSON body on stdin (bookings create is
# stdin-only — every field goes into the JSON payload):
cat <<'EOF' | cal-com-pp-cli bookings create --stdin --agent
{
  "start": "2026-05-15T14:00:00Z",
  "eventTypeId": 123,
  "attendee": {"name": "Alice", "email": "alice@example.com", "timeZone": "America/Los_Angeles"}
}
EOF
```

Ask for open slots by event-type slug + user + date window, then POST a booking body via stdin.

## Auth Setup

Cal.com uses API keys. Get one at [cal.com/settings/developer/api-keys](https://app.cal.com/settings/developer/api-keys).

```bash
export CAL_COM_API_KEY="cal_..."
cal-com-pp-cli auth set-token "$CAL_COM_API_KEY"
cal-com-pp-cli doctor
```

Optional:
- `CAL_COM_BASE_URL` — override API base (for self-hosted Cal.com v2 instances)

## Agent Mode

Add `--agent` to any command. Expands to `--json --compact --no-input --no-color --yes`. Useful flags: `--select`, `--dry-run`, `--period <duration>` for analytics windows, `--data-source auto|live|local` to force live API vs local store.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error |
| 3 | Not found (booking, event-type, user) |
| 4 | Auth required |
| 5 | API error |
| 7 | Rate limited |
| 10 | Config error |

## Installation

### CLI

```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-cli@latest
cal-com-pp-cli auth set-token YOUR_CAL_COM_API_KEY
cal-com-pp-cli doctor
```

### MCP Server

```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-mcp@latest
claude mcp add -e CAL_COM_API_KEY=<key> cal-com-pp-mcp -- cal-com-pp-mcp
```

## Argument Parsing

Given `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → run `cal-com-pp-cli --help`
2. **`install`** → CLI; **`install mcp`** → MCP
3. **Agenda / today / schedule queries** → `today --agent`
4. **Conflict / overlap queries** → `conflicts --agent`
5. **Anything else** → check install + auth, match intent to a command, run with `--agent`.
