# Trigger.dev CLI Brief

## API Identity
- Domain: Background jobs, durable workflows, AI agent orchestration
- Users: TypeScript/Python developers running background tasks, AI workflows, scheduled jobs
- Data profile: Runs (with status, timing, cost, errors, tags, metadata), tasks, schedules, queues, deployments, env vars, batches, waitpoint tokens, TRQL query results (ClickHouse-backed metrics)
- Auth: Bearer token (Secret API key, prefixes: tr_dev_, tr_prod_, tr_stg_). Env var: TRIGGER_SECRET_KEY
- Base URL: https://api.trigger.dev

## Reachability Risk
- None. Public REST API with documented endpoints. No 403 issues. Rate limit issues are about third-party integrations (Resend, Docker Hub), not the management API itself.

## Top Workflows
1. Monitor run failures in real time and get alerted instantly (USER PRIORITY)
2. Trigger tasks with payloads, check status, tail logs
3. Manage schedules (CRON jobs): create, update, pause, list upcoming
4. Query run history with TRQL for cost analysis, failure patterns, performance trends
5. Manage queues: pause during deploys, adjust concurrency, monitor backlog

## Table Stakes
- Trigger any task by identifier with JSON payload
- List/filter runs by status, task, tags, date range
- Retrieve run details including output, errors, attempts, related runs
- Cancel, replay, reschedule runs
- CRUD schedules with cron expressions and timezones
- CRUD environment variables across dev/staging/prod
- List queues, pause/resume, override concurrency
- Batch trigger tasks
- View deployments, promote between environments
- Execute TRQL queries and get results as JSON/CSV

## Competitors & Tools
- **Trigger.dev CLI** (npx trigger.dev): dev, deploy, init, login, trigger, env, promote, preview, analyze, mcp. No run management, no monitoring, no TRQL.
- **Trigger.dev MCP Server**: search_docs, trigger_task, list_runs, get_current_worker. Limited to dev-time IDE use. No offline, no alerting, no TRQL.
- **@trigger.dev/sdk** (npm): Full management API coverage via TypeScript. Requires a Node.js project. No CLI interface.
- **Inngest CLI**: Event-driven competitor. inngest dev, inngest deploy. No run querying, no TRQL equivalent.
- **Temporal CLI (tctl/temporal)**: Namespace/workflow management. Heavy enterprise tooling. No local data layer.

## Data Layer
- Primary entities: runs, tasks, schedules, queues, deployments, env_vars, batches, waitpoint_tokens
- Sync cursor: runs.list cursor-based pagination (after/before run IDs), createdAt filtering
- FTS/search: Full-text search across run errors, task identifiers, tags, metadata
- TRQL mirror: Cache TRQL query results locally for offline trend analysis

## User Vision
- Trigger.dev is bad at alerting in real time when there are run failures. The CLI should solve this with a watch/monitor command that polls runs and alerts immediately on failures.
- Project: proj_hrqzqujsznbqtlgxitck

## Product Thesis
- Name: trigger-dev-pp-cli
- Why it should exist: The official CLI handles dev/deploy workflow but has ZERO run management. The SDK requires a Node.js project. There is no way to monitor runs, query failure patterns, manage schedules, or watch for failures from the terminal. This CLI fills the gap between "deploy my tasks" and "understand what my tasks are doing" with offline search, real-time failure watching, TRQL from the command line, and agent-native output.

## Build Priorities
1. Foundation: SQLite store for runs, tasks, schedules, queues with sync and FTS5
2. Run management: list, get, cancel, replay, reschedule, tag, watch (real-time polling)
3. Task triggering: trigger with JSON payload, batch trigger, dry-run
4. Schedule management: CRUD with cron validation and timezone support
5. Queue management: list, pause/resume, concurrency override
6. TRQL: Execute queries from CLI, format as table/json/csv
7. Deployment management: list, get latest, promote
8. Env var management: list, create, update, delete, import across environments
9. Failure alerting: watch command with configurable polling, desktop notifications, sound alerts
10. Transcendence: failure patterns, cost analysis, stale detection, health dashboard
