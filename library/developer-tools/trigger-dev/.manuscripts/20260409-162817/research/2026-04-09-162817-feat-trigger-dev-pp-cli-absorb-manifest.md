# Trigger.dev CLI Absorb Manifest

## Tools Analyzed
1. **Trigger.dev CLI** (npx trigger.dev) - Official CLI: dev, deploy, init, login, trigger, env, promote, preview, analyze, mcp, whoami, switch, update
2. **Trigger.dev MCP Server** - search_docs, trigger_task, list_runs, get_current_worker, project management
3. **@trigger.dev/sdk** (npm) - Full management API: runs, tasks, schedules, queues, deployments, env vars, batches, waitpoints, TRQL query
4. **Inngest CLI** - Competing workflow tool: dev, deploy, serve
5. **Temporal CLI (tctl)** - Enterprise workflow: namespace, workflow, activity management

## Absorbed (match or beat everything that exists)

| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | List runs with filters | SDK runs.list() | runs list --status FAILED --task my-task --tag prod | Offline FTS, SQLite cache, --json/--csv/--select |
| 2 | Get run details | SDK runs.retrieve() | runs get <run-id> | Offline cached, shows attempts/errors/related runs |
| 3 | Get run result/output | SDK runs.retrieveResult() | runs result <run-id> | Pretty-printed JSON, --select field filtering |
| 4 | Cancel a run | SDK runs.cancel() | runs cancel <run-id> --dry-run | Dry-run preview, batch cancel with filters |
| 5 | Replay a run | SDK runs.replay() | runs replay <run-id> | Replay with modified payload, --dry-run |
| 6 | Reschedule a run | SDK runs.reschedule() | runs reschedule <run-id> --delay 1h | Human-friendly delay syntax |
| 7 | Add tags to run | SDK runs.addTags() | runs tag <run-id> --add incident-42 | Bulk tagging with filters |
| 8 | Update run metadata | SDK runs.updateMetadata() | runs metadata <run-id> --set key=value | JSON patch support |
| 9 | Trigger a task | SDK tasks.trigger(), MCP trigger_task | trigger <task-id> --payload '{}' | HTTPie-style params, --dry-run, stdin pipe |
| 10 | Batch trigger tasks | SDK tasks.batchTrigger() | trigger <task-id> --batch --payload-file items.json | File-based batch, progress bar |
| 11 | Create schedule | SDK schedules.create() | schedules create <task-id> --cron "0 * * * *" --tz US/Eastern | Human-readable cron validation, next-run preview |
| 12 | List schedules | SDK schedules.list() | schedules list | Table with next run time, active status |
| 13 | Update schedule | SDK schedules.update() | schedules update <sched-id> --cron "0 0 * * *" | Preview changes before applying |
| 14 | Delete schedule | SDK schedules.del() | schedules delete <sched-id> | Confirmation prompt, --force |
| 15 | Activate/deactivate schedule | SDK schedules.activate/deactivate() | schedules pause/resume <sched-id> | Bulk pause/resume |
| 16 | Get timezones | SDK schedules.timezones() | schedules timezones | Searchable timezone list |
| 17 | List env vars | SDK envvars.list() | env list --env prod | Cross-env diff, masked values |
| 18 | Create/update/delete env var | SDK envvars.create/update/del() | env set KEY=VALUE --env prod | Bulk set from .env file |
| 19 | Import env vars | SDK envvars.import() | env import .env.production --env prod | .env file parsing, conflict detection |
| 20 | List queues | SDK queues.list() | queues list | Running/queued counts, pause status |
| 21 | Pause/resume queue | SDK queues.pause/resume() | queues pause/resume <queue-name> | Bulk pause for maintenance |
| 22 | Override concurrency | SDK queues.overrideConcurrency() | queues concurrency <queue-name> --limit 5 | Preview before apply |
| 23 | Get deployment | SDK deployments.get() | deployments get <deploy-id> | Task list, version, timing |
| 24 | Get latest deployment | SDK deployments.getLatest() | deployments latest | Quick status check |
| 25 | Promote deployment | SDK deployments.promote() | deployments promote <deploy-id> --env prod | Dry-run, confirmation |
| 26 | Execute TRQL query | SDK query.execute() | query "SELECT task_identifier, count(*) FROM runs GROUP BY task_identifier" | Table/CSV/JSON output, query history |
| 27 | Create batch | SDK batches.create() | batches create <task-id> --payload-file items.json | Progress tracking |
| 28 | Retrieve batch | SDK batches.retrieve() | batches get <batch-id> | Status summary with counts |
| 29 | Create waitpoint token | SDK waitpoints.create() | waitpoints create --tags review-needed | Human approval workflows |
| 30 | Complete waitpoint | SDK waitpoints.complete() | waitpoints complete <token-id> --output '{"approved": true}' | Quick approve/reject |
| 31 | List waitpoint tokens | SDK waitpoints.list() | waitpoints list --pending | Filter by status |
| 32 | Login/auth | Trigger CLI login | auth login | Profile management |
| 33 | Whoami | Trigger CLI whoami | auth whoami | Current project/env info |
| 34 | Dev server | Trigger CLI dev | (defer - not in scope, this is dev tooling) | -- |
| 35 | Deploy | Trigger CLI deploy | (defer - not in scope, this is dev tooling) | -- |
| 36 | Search docs | MCP search_docs | docs search "retry configuration" | Offline cached docs search |
| 37 | Retrieve run trace | SDK runs.retrieveTrace() | runs trace <run-id> | Visual trace timeline in terminal |
| 38 | Retrieve run events | SDK runs.retrieveEvents() | runs events <run-id> | Filtered event stream |
| 39 | Doctor/health check | -- | doctor | Auth, connectivity, project validation |
| 40 | Sync all data | -- | sync --full | Full SQLite sync of runs, schedules, queues |

## Transcendence (only possible with our local data layer)

| # | Feature | Command | Why Only We Can Do This | Score | Evidence |
|---|---------|---------|------------------------|-------|----------|
| 1 | Real-time failure watch | watch --failures --notify | Polls runs, detects new failures, desktop notifications + sound. No existing tool does this. | 10/10 | USER PRIORITY: "bad at alerting you in real time when there are run failures". Dashboard alerts are webhook-only with no CLI/terminal option. |
| 2 | Failure pattern analysis | failures --period 7d | Cross-references failed runs by task, error message, time-of-day, and machine type. Requires local joins across run + error + task data. | 9/10 | Trigger.dev alerts only tell you "run failed" - no pattern analysis. TRQL can query but requires writing SQL. This summarizes automatically. |
| 3 | Task health dashboard | health | Shows per-task success rate, avg duration, p95 duration, failure rate, cost trend - all from SQLite. One command, full picture. | 8/10 | Dashboard requires clicking through individual tasks. SDK has no health summary endpoint. Only possible by aggregating synced run data locally. |
| 4 | Cost analysis | costs --period 30d --by task | Aggregates costInCents and baseCostInCents across runs by task, time period, machine type. Finds cost spikes and anomalies. | 8/10 | Run objects include costInCents. No existing tool aggregates this. TRQL can but requires manual SQL. |
| 5 | Stale schedule detection | schedules stale | Finds schedules that haven't produced runs recently, disabled schedules with no recent activity, schedules with high failure rates. | 7/10 | Requires joining schedule data with run history. No existing tool cross-references these. |
| 6 | Queue bottleneck detection | queues bottleneck | Identifies queues with growing backlogs, high wait times, or concurrency limits causing delays. Requires local time-series data. | 7/10 | Queue API shows current counts. Trend analysis requires historical snapshots only possible with local sync. |
| 7 | Run timeline | runs timeline --task my-task --period 24h | Visual ASCII timeline showing run starts/ends/failures over time. Spot patterns at a glance. | 6/10 | Runs have startedAt/finishedAt. Visual timeline requires local aggregation across many runs. |
| 8 | Environment diff | env diff dev prod | Side-by-side comparison of env vars between environments. Highlights missing, different, and matching values. | 6/10 | Requires fetching from both environments and joining locally. SDK has no diff endpoint. |

