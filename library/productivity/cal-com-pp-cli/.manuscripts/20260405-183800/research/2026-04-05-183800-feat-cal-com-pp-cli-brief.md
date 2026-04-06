# Cal.com CLI Brief

## API Identity
- Domain: Scheduling infrastructure (open-source Calendly alternative)
- Users: Developers building scheduling into apps, teams managing availability, solo professionals managing bookings
- Data profile: 285 operations across 181 paths. Resources: bookings (19 paths), event-types (6), schedules (3), slots (3), calendars (17), teams (20), organizations (78), conferencing (7), webhooks (2), stripe (3)
- Base URL: https://api.cal.com/v2
- Auth: Bearer token (`cal_live_*` prefix), no `securitySchemes` in spec — requires parser inference from Authorization header params
- Per-endpoint versioning: `cal-api-version` header required on 37/285 ops with different values per resource group: Bookings `2024-08-13`, Event Types `2024-06-14`, Schedules `2024-06-11`, Slots `2024-09-04`
- Rate limits: 120 req/min (API key), 500 req/min (OAuth)

## Reachability Risk
- **Low** — API is actively maintained, rate limits documented, 403 issues are operation-specific (event type deletion, embed cancellation), not API-wide. No systemic blocking.

## Top Workflows
1. **Booking lifecycle** — Create, list, reschedule, cancel, confirm, decline bookings. Mark no-shows. Manage attendees and guests.
2. **Event type configuration** — Create/update event types with duration, scheduling rules, hosts. Team event types with round-robin.
3. **Availability management** — Create/update schedules with weekly slots, check available slots for specific date ranges.
4. **Calendar sync** — Connect external calendars (Google, Outlook), manage busy times, destination calendar routing.
5. **Team administration** — Create teams, manage memberships, roles, verified resources, conferencing apps.

## Table Stakes
- Booking CRUD (create, list, get, reschedule, cancel, confirm, decline, mark absent)
- Event type CRUD (create, list, get, update, delete)
- Schedule management (create, list, get default, update, delete)
- Slot availability queries
- Calendar connections and busy times
- Webhook management (create, list, update, delete per event-type or org)
- Team management (create, list, manage memberships)
- Profile/me endpoint
- OAuth client management
- Conferencing app connections
- Stripe integration status

## Data Layer
- Primary entities: Bookings, EventTypes, Schedules, Slots (ephemeral), Calendars, Teams, Users, Webhooks
- Sync cursor: Bookings have afterStart/status filters; event-types and schedules are relatively static
- FTS/search: Bookings (by title, attendee name, notes), Event Types (by title/slug), Teams (by name)
- Analytics: Booking volume over time, no-show rates, event type popularity, team utilization

## Product Thesis
- Name: cal-com-pp-cli
- Why it should exist: Cal.com has an official MCP (141 tools) but it's stateless — every query hits the API. No existing tool offers offline booking search, cross-entity analytics (which event types drive the most no-shows?), schedule conflict detection, or team workload visibility. The CLI with SQLite backing makes scheduling data queryable, composable, and agent-native.

## Build Priorities
1. Foundation: data layer for bookings, event-types, schedules, teams + sync + FTS search
2. Absorb: every feature from @calcom/cal-mcp (9 default + 141 extended), Composio toolkit, calendly-cli
3. Transcend: compound features only possible with local data (conflict detection, booking analytics, availability gaps, team workload, no-show patterns)
