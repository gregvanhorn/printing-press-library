# Dub CLI Live Smoke Test Report

## Environment
- API: https://api.dub.co
- Auth: Bearer token (DUB_TOKEN env var)
- Workspace: ws_1KMSGMTYQ2W6N521CEGANEMSP

## Tests

| Test | Result | Notes |
|------|--------|-------|
| --help | PASS | Shows correct description |
| doctor | PASS | Auth configured, API reachable (credentials warning is a doctor-specific check) |
| links create | PASS | Created link at dub.sh/qNbrz0K → returned full JSON response |
| links get (list) | PASS | Listed created link correctly |
| links delete | PASS | Deleted test link, confirmed |
| domains list | PASS | Empty (no custom domains in workspace) |
| tags get (list) | PASS | Empty (no tags in workspace) |

## Summary
All live API operations succeed. Create → List → Delete lifecycle fully verified.
Workspace is clean (new account with no pre-existing data), which is why list operations return empty arrays.
