# Trigger.dev CLI Build Log

## Generated (Priority 0 + Priority 1)
- Data layer: SQLite store with per-entity tables (runs, tasks, schedules, queues, batches, envvars, deployments, waitpoints, query) + FTS5 search
- Sync: Full sync with cursor-based pagination
- Runs: list, get, cancel, replay, reschedule, add-tags, update-metadata
- Tasks: trigger (with payload, options), batch-trigger
- Schedules: list, get, create, update, delete, activate, deactivate, timezones
- Env vars: list, create, delete, import
- Queues: list, get, pause, resume, override-concurrency, reset-concurrency
- Deployments: get, latest, promote
- Batches: create, get
- Waitpoints: list, get, create, complete
- Query: execute (TRQL)
- Auth: login, whoami
- Doctor: health check
- Export/Import: JSONL backup
- Search: FTS5 across all entities
- API discovery: browse endpoints by interface

## Built (Priority 2 - Transcendence)
1. **watch** - Real-time failure monitoring with polling, desktop notifications (macOS/Linux), sound alerts
2. **failures** - Failure pattern analysis by task, status, hour with cost aggregation
3. **health** - Task health dashboard: success rate, avg/p95 duration, cost trends
4. **costs** - Cost analysis by task/status with percentage breakdown
5. **schedules stale** - Stale schedule detection via schedule + run cross-reference
6. **queues bottleneck** - Queue bottleneck detection: backlog ratio, concurrency limits, paused state
7. **runs timeline** - ASCII timeline visualization of run patterns
8. **envvars diff** - Cross-environment variable comparison with masked values

## Generator Issues Fixed
- `usageErr` function missing from helpers.go (added)
- Unused variable `data` in config.go (removed)

## Intentionally Deferred
- Live API mutation testing (Phase 5)
- TRQL query caching in SQLite (could be added as enhancement)
