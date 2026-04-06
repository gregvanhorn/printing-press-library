# Cal.com CLI Absorb Manifest

## Sources Analyzed
1. **@calcom/cal-mcp** (official MCP, 9 default / 141 extended tools) — github.com/calcom/cal-mcp
2. **Composio Cal toolkit** (141 actions) — composio.dev/toolkits/cal
3. **mcpmarket.com Cal.com Calendar** (MCP) — mcpmarket.com/es/server/cal-com-calendar
4. **LobeHub cal-com-automation** (skill) — lobehub.com/skills
5. **calendly-cli** (competitor CLI) — github.com/iloveitaly/calendly-cli
6. **@calcom/sdk** (official npm SDK) — npmjs.com/package/@calcom/sdk
7. **@modelcontext/cal-com-api-v2** (community MCP) — npmjs.com/package/@modelcontext/cal-com-api-v2

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | List bookings with filters | cal-mcp getBookings | `bookings` with --status, --after, --before, --attendee | Offline search, SQLite-backed, --json/--csv/--select |
| 2 | Get booking by UID | cal-mcp getBooking | `bookings get <uid>` | Cached in store, instant re-lookup |
| 3 | Create booking | cal-mcp createBooking | `bookings create --event-type --start --attendee` | --dry-run, --json, agent-native, stdin batch |
| 4 | Reschedule booking | cal-mcp rescheduleBooking | `bookings reschedule <uid> --start` | --dry-run preview of time change |
| 5 | Cancel booking | cal-mcp cancelBooking | `bookings cancel <uid> --reason` | --dry-run, typed exit codes |
| 6 | Confirm booking | Composio confirm_booking | `bookings confirm <uid>` | --dry-run, batch via stdin |
| 7 | Decline booking | Composio decline_booking | `bookings decline <uid> --reason` | Agent-native, --json |
| 8 | Mark no-show | Composio mark_absent | `bookings mark-absent <uid> --attendees` | Feeds noshow analytics |
| 9 | Reassign booking | Composio reassign_booking | `bookings reassign <uid> --to-user` | --dry-run |
| 10 | List event types | cal-mcp getEventTypes | `event-types` with --json/--select | Offline, synced to SQLite |
| 11 | Get event type by ID | cal-mcp getEventTypeById | `event-types get <id>` | Cached lookup |
| 12 | Create event type | Composio create_event_type | `event-types create --title --slug --duration` | --dry-run, full param support |
| 13 | Update event type | cal-mcp updateEventType | `event-types update <id> --title --duration` | --dry-run |
| 14 | Delete event type | cal-mcp deleteEventType | `event-types delete <id>` | --dry-run, confirmation prompt |
| 15 | List schedules | Composio get_schedules | `schedules` | Offline, synced |
| 16 | Get default schedule | Composio get_default_schedule | `schedules default` | Quick access |
| 17 | Create schedule | Composio create_schedule | `schedules create --name --timezone --availability` | --dry-run |
| 18 | Update schedule | Composio update_schedule | `schedules update <id>` | --dry-run |
| 19 | Delete schedule | Composio delete_schedule | `schedules delete <id>` | --dry-run |
| 20 | Get available slots | cal-mcp (extended) | `slots --event-type --start --end` | Offline gap analysis possible |
| 21 | Reserve slot | Composio reserve_slot | `slots reserve --event-type --start` | --dry-run |
| 22 | Calendar connections list | Composio retrieve_calendar_list | `calendars` | Synced to store |
| 23 | Connect calendar | Composio connect_calendar | `calendars connect` | Guided setup |
| 24 | Disconnect calendar | Composio disconnect_calendar | `calendars disconnect <id>` | --dry-run |
| 25 | Calendar busy times | Composio calendar_busy_times | `calendars busy --start --end` | Local caching |
| 26 | Destination calendar | Composio update_destination | `calendars destination --calendar-id` | |
| 27 | Selected calendars | Composio add_selected | `calendars select --add/--remove` | |
| 28 | ICS feed management | Composio ics_feed | `calendars ics-feed` | |
| 29 | Create team | Composio create_team | `teams create --name` | --dry-run |
| 30 | List teams | Composio get_teams | `teams` | Offline, synced |
| 31 | Get team details | Composio get_team_by_id | `teams get <id>` | Cached |
| 32 | Update team | Composio update_team | `teams update <id>` | --dry-run |
| 33 | Delete team | Composio delete_team | `teams delete <id>` | --dry-run |
| 34 | Team memberships | Composio team_memberships | `teams members <id>` | |
| 35 | Add team member | Composio add_member | `teams add-member <id> --user --role` | --dry-run |
| 36 | Remove team member | Composio delete_membership | `teams remove-member <id> --user` | |
| 37 | Team event types | Composio team_event_types | `teams event-types <id>` | |
| 38 | Create webhook | Composio create_webhook | `webhooks create --url --triggers` | --dry-run |
| 39 | List webhooks | Composio list_webhooks | `webhooks` | |
| 40 | Update webhook | Composio update_webhook | `webhooks update <id>` | |
| 41 | Delete webhook | Composio delete_webhook | `webhooks delete <id>` | |
| 42 | Connect conferencing | Composio connect_conferencing | `conferencing connect --app` | |
| 43 | Default conferencing | Composio default_conferencing | `conferencing default --app` | |
| 44 | Conferencing list | Composio conferencing_info | `conferencing` | |
| 45 | Stripe status | Composio check_stripe | `stripe status` | |
| 46 | User profile (me) | Composio retrieve_my_info | `me` | Cached |
| 47 | Update profile | Composio update_profile | `me update --name --timezone` | |
| 48 | API keys management | spec api-keys endpoints | `api-keys` / `api-keys refresh` | |
| 49 | Timezones list | Composio get_timezones | `timezones` | Offline lookup |
| 50 | Available times (copy/paste) | calendly-cli | `slots --format text` | Human-readable availability |
| 51 | OAuth client management | Composio oauth tools | `oauth-clients` (admin) | |
| 52 | Booking attendees | spec bookings/attendees | `bookings attendees <uid>` | |
| 53 | Booking guests | spec bookings/guests | `bookings guests add <uid>` | |
| 54 | Booking recordings | spec bookings/recordings | `bookings recordings <uid>` | |
| 55 | Booking transcripts | spec bookings/transcripts | `bookings transcripts <uid>` | |
| 56 | Booking references | spec bookings/references | `bookings references <uid>` | |
| 57 | Booking location update | spec bookings/location | `bookings location <uid> --type --link` | |
| 58 | Booking conferencing sessions | spec bookings/conferencing | `bookings video <uid>` | |
| 59 | Get booking by seat UID | spec bookings/get-by-seat | `bookings get-by-seat <uid>` | |
| 60 | Booking calendar links | spec bookings/calendar-links | `bookings calendar-link <uid>` | |
| 61 | Routing forms | spec routing-forms | `routing-forms` | |
| 62 | Verified resources | spec verified-resources | `verified-resources` | |
| 63 | Organization management | Composio org tools | `orgs` (admin) | |

## Transcendence (only possible with our local data layer)

| # | Feature | Command | Why Only We Can Do This | Score | Evidence |
|---|---------|---------|------------------------|-------|----------|
| 1 | Schedule conflict detection | `conflicts` | Requires local join across bookings + event-types + schedules to find double-bookings and overlap across all calendar sources | 10/10 | Cal.com GitHub Issue #23549 (403 on cancel due to conflicts), community complaints about double-booking |
| 2 | Today's schedule dashboard | `today` | Single command showing today's bookings with attendee details, conferencing links, prep time. Requires local booking + calendar + event-type data | 10/10 | calendly-cli only shows availability, no "today" view in any existing tool |
| 3 | Booking analytics | `stats` | Volume over time, busiest days/hours, avg duration, cancellation rate, booking-to-completion rate. Requires historical booking data in SQLite | 8/10 | Cal.com Insights is paid-tier only; no free analytics in any existing tool |
| 4 | No-show pattern radar | `noshow` | Track no-show rates by event type, day of week, time of day. Predict which upcoming bookings are high no-show risk | 8/10 | mark-absent API exists (spec endpoint), no tool analyzes patterns |
| 5 | Full-text booking search | `search` | Instant offline search across booking titles, attendee names, notes, descriptions. FTS5 index | 8/10 | Cal.com API has no search endpoint; MCP tools require per-query API calls |
| 6 | Availability gap finder | `gaps` | Find schedule windows that are available but chronically unbooked. Identify underutilized time slots for schedule optimization | 7/10 | Slots + booking history correlation; scheduling optimization is common pain point |
| 7 | Team workload balance | `workload` | Booking distribution across team members. Identify overloaded or underutilized members for round-robin tuning | 7/10 | Composio has team tools but no analytics; team management is top-5 feature area |
| 8 | Stale event type cleanup | `stale` | Find event types that haven't received a booking in N days. Identify scheduling page cruft for cleanup | 6/10 | No existing tool offers event-type usage analytics |

## Feature Totals
- Absorbed: 63 features
- Transcendence: 8 features (all scoring >= 6/10)
- Total: 71 features
- Best existing tool: Composio Cal (141 actions, but stateless — no offline, no analytics, no search)
