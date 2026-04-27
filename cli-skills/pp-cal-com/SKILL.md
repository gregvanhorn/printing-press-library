---
name: pp-cal-com
description: "Every Cal.com feature, plus offline agendas, composed booking flows, and analytics no other Cal.com tool ships. Trigger phrases: `book a meeting on cal.com`, `what's on my cal.com agenda today`, `cancel cal.com booking`, `find available slots cal.com`, `cal.com no-show analytics`, `use cal-com`, `run cal-com`."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["cal-com-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-cli@latest","bins":["cal-com-pp-cli"],"label":"Install via go install"}]}}'
---

# Cal.com — Printing Press CLI

A Go single-binary CLI for Cal.com v2 with a local SQLite mirror that turns scheduling data into something queryable. Composed intents like `book` collapse slot-find / reserve / create / confirm into one safe call with `--dry-run`. The store powers `today`, `week`, `analytics`, and `conflicts` — views Cal.com's own API has no equivalent for.

## When to Use This CLI

Reach for cal-com-pp-cli when an agent needs Cal.com data for booking, agenda views, or analytics. The local store keeps repeated calls cheap, and composed intents like `book` and `slots find` collapse multi-step API chains into one operation. For raw API parity (any v2 endpoint at all) the absorbed surface matches the official MCP server. Skip this CLI for org-wide enterprise admin flows that require OAuth managed-user delegation — those are still better served via Cal.com's web UI today.

### Do NOT use this CLI for

- **Calendly bookings** — Calendly is a different scheduling service with a different API. Use a Calendly-specific tool.
- **Google Calendar / Outlook event creation** — these are external calendar APIs. Use those vendors' SDKs/CLIs directly. (Cal.com can `connect` external calendars for availability, but the CLI does not write events to them.)
- **Cal.com v1 API workflows** — v1 was deprecated April 8, 2026. This CLI targets `/v2` only.
- **Cal.com web-app navigation** — there is no headless browser; everything goes through the public REST API. UI-only operations (dashboard customization, settings UI flows) cannot be automated here.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Composed intents

- **`book`** — Find a slot and book it in a single command — no slot/reserve/create/confirm chain.

  _Reach for this when an agent or operator wants to book a meeting end-to-end without managing slot reservation state by hand._

  ```bash
  cal-com-pp-cli book --event-type-id 96531 --start "2026-05-01T14:00:00Z" --attendee-name "Jane Doe" --attendee-email "jane@example.com" --dry-run --json
  ```
- **`slots find`** — Find first available slot across multiple event types in one call, ranked by start time.

  _Pick this when the agent doesn't care which meeting type — only when the next available slot is._

  ```bash
  cal-com-pp-cli slots find --event-type-ids 96531,96532,96533 --start "2026-05-01" --end "2026-05-08" --first-only --json
  ```
- **`bookings pending`** — Pending-confirmation bookings sorted by age, with default 24h max-age cutoff.

  _Use when sweeping the pending queue to confirm/decline before the SLA window closes._

  ```bash
  cal-com-pp-cli bookings pending --max-age 24h --json
  ```

### Local state that compounds

- **`today`** — Today's bookings with status, attendees, and meeting links — read from the local store, no API call needed.

  _First-thing-in-the-morning view that works offline and stays cheap to call repeatedly._

  ```bash
  cal-com-pp-cli today --json
  ```
- **`week`** — 7-day calendar view of upcoming bookings, with conflict highlighting and per-day rollup counts.

  _Use when you need a one-look view of the upcoming week without paging through API responses._

  ```bash
  cal-com-pp-cli week --start monday --json
  ```
- **`analytics bookings|cancellations|no-show|density`** — Booking analytics over a time window. Each subcommand is a different metric; group with `--by event-type|attendee|weekday|hour|status`.

  _Reach for these when scoring a workflow's health, planning team capacity, or hunting cancellation patterns._

  ```bash
  cal-com-pp-cli analytics bookings --window 30d --by event-type --json
  cal-com-pp-cli analytics no-show --window 90d --by attendee --json
  cal-com-pp-cli analytics density --unit hour --window 90d --json
  ```
- **`conflicts`** — Detects overlaps between active Cal.com bookings and external calendar busy-times.

  _Spot last-minute calendar additions that didn't propagate into Cal.com availability._

  ```bash
  cal-com-pp-cli conflicts --window 7d --json
  ```
- **`gaps`** — Finds open windows in your schedule that are available but unbooked, filtered by minimum block size.

  _Use when looking for capacity to take new meetings or to spot underused windows worth promoting._

  ```bash
  cal-com-pp-cli gaps --window 7d --min-minutes 30 --json
  ```
- **`workload`** — Booking distribution across team members over a window — surfaces overloaded vs underutilized hosts.

  _Reach for this when tuning round-robin weights or planning team capacity._

  ```bash
  cal-com-pp-cli workload --team-id 42 --window 30d --json
  ```
- **`event-types stale`** — Event types with zero bookings in the last N days — candidates for removal.

  _Run during quarterly cleanups to retire dead booking links._

  ```bash
  cal-com-pp-cli event-types stale --days 30 --json
  ```

### Reachability mitigation

- **`webhooks coverage`** — Audits registered webhook triggers against the canonical set and reports lifecycle events with no subscriber.

  _Run before relying on webhooks in production to confirm every lifecycle stage has a handler._

  ```bash
  cal-com-pp-cli webhooks coverage --json
  ```

### Agent-native plumbing

- **`webhooks triggers`** — Static reference of every valid Cal.com webhook trigger constant, grouped by lifecycle stage.

  _Reach for this before writing webhook scaffolding, so trigger strings are exact._

  ```bash
  cal-com-pp-cli webhooks triggers --json
  ```

## Command Reference

**api-keys** — Refresh the active API key

- `cal-com-pp-cli api-keys` — Refresh API Key (promoted command; takes no args)

**auth** — Manage CAL_COM_TOKEN credentials

- `cal-com-pp-cli auth set-token <token>` — Save a Bearer token to the config file
- `cal-com-pp-cli auth status` — Show the active credential source and config path
- `cal-com-pp-cli auth logout` — Clear the stored token

**bookings** — Manage bookings

- `cal-com-pp-cli bookings create` — Create a booking
- `cal-com-pp-cli bookings get` — Get all bookings
- `cal-com-pp-cli bookings get-bookinguid` — Get a booking
- `cal-com-pp-cli bookings get-by-seat-uid` — Get a booking by seat UID

**calendars** — Manage calendars

- `cal-com-pp-cli calendars cal-unified-create-connection-event` — Create event on a connection
- `cal-com-pp-cli calendars cal-unified-delete-connection-event` — Delete event for a connection
- `cal-com-pp-cli calendars cal-unified-get-connection-event` — Get event for a connection
- `cal-com-pp-cli calendars cal-unified-get-connection-free-busy` — Get free/busy for a connection
- `cal-com-pp-cli calendars cal-unified-list-connection-events` — List events for a connection
- `cal-com-pp-cli calendars cal-unified-list-connections` — List calendar connections
- `cal-com-pp-cli calendars cal-unified-update-connection-event` — Update event for a connection
- `cal-com-pp-cli calendars check-ics-feed` — Check an ICS feed
- `cal-com-pp-cli calendars create-ics-feed` — Save an ICS feed
- `cal-com-pp-cli calendars get` — Get all calendars
- `cal-com-pp-cli calendars get-busy-times` — Get busy times

**conferencing** — Manage conferencing

- `cal-com-pp-cli conferencing get-default` — Get your default conferencing application
- `cal-com-pp-cli conferencing list-installed-apps` — List your conferencing applications

**destination-calendars** — Manage destination calendars

- `cal-com-pp-cli destination-calendars update` — Update destination calendars

**event-types** — Manage event types

- `cal-com-pp-cli event-types create` — Create an event type
- `cal-com-pp-cli event-types delete` — Delete an event type
- `cal-com-pp-cli event-types get` — Get all event types
- `cal-com-pp-cli event-types get-by-id` — Get an event type
- `cal-com-pp-cli event-types update` — Update an event type

**me** — Manage me

- `cal-com-pp-cli me get` — Get my profile
- `cal-com-pp-cli me update` — Update my profile

**oauth** — Manage oauth


**oauth-clients** — Manage oauth clients

- `cal-com-pp-cli oauth-clients create` — Create an OAuth client
- `cal-com-pp-cli oauth-clients delete` — Delete an OAuth client
- `cal-com-pp-cli oauth-clients get` — Get all OAuth clients
- `cal-com-pp-cli oauth-clients get-by-id` — Get an OAuth client
- `cal-com-pp-cli oauth-clients update` — Update an OAuth client

**organizations** — Manage organizations


**routing-forms** — Manage routing forms


**schedules** — Manage schedules

- `cal-com-pp-cli schedules create` — Create a schedule
- `cal-com-pp-cli schedules delete` — Delete a schedule
- `cal-com-pp-cli schedules get` — Get all schedules
- `cal-com-pp-cli schedules get-default` — Get default schedule
- `cal-com-pp-cli schedules get-scheduleid` — Get a schedule
- `cal-com-pp-cli schedules update` — Update a schedule

**selected-calendars** — Manage selected calendars

- `cal-com-pp-cli selected-calendars add` — Add a selected calendar
- `cal-com-pp-cli selected-calendars delete` — Delete a selected calendar

**slots** — Manage slots

- `cal-com-pp-cli slots delete-reserved` — Delete a reserved slot
- `cal-com-pp-cli slots get-available` — Get available time slots for an event type
- `cal-com-pp-cli slots get-reserved` — Get reserved slot
- `cal-com-pp-cli slots reserve` — Reserve a slot
- `cal-com-pp-cli slots update-reserved` — Update a reserved slot

**stripe** — Manage stripe

- `cal-com-pp-cli stripe check` — Check Stripe connection
- `cal-com-pp-cli stripe redirect` — Get Stripe connect URL
- `cal-com-pp-cli stripe save` — Save Stripe credentials

**teams** — Manage teams

- `cal-com-pp-cli teams create` — Create a team
- `cal-com-pp-cli teams delete` — Delete a team
- `cal-com-pp-cli teams get` — Get teams
- `cal-com-pp-cli teams get-teamid` — Get a team
- `cal-com-pp-cli teams update` — Update a team

**verified-resources** — Manage verified resources

- `cal-com-pp-cli verified-resources user-get-verified-email-by-id` — Get verified email by id
- `cal-com-pp-cli verified-resources user-get-verified-emails` — Get list of verified emails
- `cal-com-pp-cli verified-resources user-get-verified-phone-by-id` — Get verified phone number by id
- `cal-com-pp-cli verified-resources user-get-verified-phone-numbers` — Get list of verified phone numbers
- `cal-com-pp-cli verified-resources user-request-email-verification-code` — Request email verification code
- `cal-com-pp-cli verified-resources user-request-phone-verification-code` — Request phone number verification code
- `cal-com-pp-cli verified-resources user-verify-email` — Verify an email
- `cal-com-pp-cli verified-resources user-verify-phone-number` — Verify a phone number

**webhooks** — Manage webhooks

- `cal-com-pp-cli webhooks create` — Create a webhook
- `cal-com-pp-cli webhooks delete` — Delete a webhook
- `cal-com-pp-cli webhooks get` — Get all webhooks
- `cal-com-pp-cli webhooks get-webhookid` — Get a webhook
- `cal-com-pp-cli webhooks update` — Update a webhook


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
cal-com-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes


### Book the next available 30-min meeting

```bash
cal-com-pp-cli slots find --event-type-ids 96532 --start "$(date +%Y-%m-%d)" --end "$(date -v+7d +%Y-%m-%d)" --first-only --json | jq -r '.slots[0].start' | xargs -I{} cal-com-pp-cli book --event-type-id 96532 --start {} --attendee-name "X" --attendee-email "x@y.com"
```

Combine `slots find` and `book` to grab the first opening across the next week.

### Today's agenda with only the fields agents care about

```bash
cal-com-pp-cli today --agent --select bookings.uid,bookings.title,bookings.start,bookings.attendees.email
```

`--agent` produces a compact JSON shape; `--select` narrows it to four fields, keeping context cheap on deeply-nested booking responses.

### Find no-show patterns over the last quarter

```bash
cal-com-pp-cli analytics no-show --window 90d --by attendee --json | jq '[.rows[] | select(.no_show_rate > 0.2)]'
```

Identify attendees with chronic no-show problems before they damage team capacity.

### Audit webhook coverage before going to production

```bash
cal-com-pp-cli webhooks coverage --json
```

Lists registered triggers vs canonical set; flags lifecycle stages without a subscriber.

### Cancel a booking with a reason

The cancel endpoint takes a body via stdin (Cal.com requires `cancellationReason`):

```bash
echo '{"cancellationReason":"Customer requested"}' | \
  cal-com-pp-cli bookings cancel bookings-booking <uid> --stdin --dry-run
echo '{"cancellationReason":"Customer requested"}' | \
  cal-com-pp-cli bookings cancel bookings-booking <uid> --stdin --json
```

Dry-run shows the exact request body, then the same command executed for real.

## Auth Setup

Cal.com uses Bearer API keys (the `cal_live_*` prefix from Settings → Developer → API Keys). Set `CAL_COM_TOKEN` in your environment, or run `cal-com-pp-cli auth set-token <token>` to save it. `doctor` will tell you which source is active and whether the key reaches `/v2/me`. Per-endpoint `cal-api-version` headers (Bookings 2024-08-13, Slots 2024-09-04, etc.) are handled automatically.

Run `cal-com-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  cal-com-pp-cli bookings get --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal — piped/agent consumers get pure JSON on stdout.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
cal-com-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
cal-com-pp-cli feedback --stdin < notes.txt
cal-com-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.cal-com-pp-cli/feedback.jsonl`. They are never POSTed unless `CAL_COM_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `CAL_COM_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
cal-com-pp-cli profile save briefing --json
cal-com-pp-cli --profile briefing bookings get
cal-com-pp-cli profile list --json
cal-com-pp-cli profile show briefing
cal-com-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `cal-com-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → CLI installation
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## CLI Installation

1. Check Go is installed: `go version` (requires Go 1.23+)
2. Install:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-cli@latest
   
   # If `@latest` installs a stale build (Go module proxy cache lag), install from main:
   GOPRIVATE='github.com/mvanhorn/*' GOFLAGS=-mod=mod \
     go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-cli@main
   ```
3. Verify: `cal-com-pp-cli --version`
4. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-mcp@latest
   
   # If `@latest` installs a stale build (Go module proxy cache lag), install from main:
   GOPRIVATE='github.com/mvanhorn/*' GOFLAGS=-mod=mod \
     go install github.com/mvanhorn/printing-press-library/library/productivity/cal-com/cmd/cal-com-pp-mcp@main
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add cal-com-pp-mcp -- cal-com-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which cal-com-pp-cli`
   If not found, offer to install (see CLI Installation above).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   cal-com-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `cal-com-pp-cli <command> --help`.
