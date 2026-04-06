# Shipcheck Report: cal-com-pp-cli

## Dogfood
- Path Validity: 6/6 PASS
- Dead Flags: 0 PASS
- Dead Functions: 0 PASS (was 1 — formatCompact removed)
- Unregistered Commands: 0 PASS (was 2 — auth endpoints removed)
- Examples: 10/10 PASS
- Novel Features: 5/5 PASS
- Auth Protocol: MISMATCH (false positive — code uses Bearer correctly, scorer doesn't recognize the pattern)
- Data Pipeline: GOOD
- Verdict: PASS (after fixes)

## Verify
- Pass Rate: 86% (25/29)
- Critical Failures: 0
- Non-critical: 4 (conflicts, gaps, stats, workload — need synced data)
- Verdict: PASS

## Workflow-Verify
- Verdict: workflow-pass (no manifest)

## Scorecard
- Total: 90/100 — Grade A
- Auth Protocol: 3/10 (false positive from dogfood — actual auth works correctly)
- Type Fidelity: 3/5
- Terminal UX: 9/10
- Vision: 9/10
- All other dimensions: 10/10

## Fixes Applied
1. Removed dead `formatCompact` function
2. Removed unregistered auth endpoint files (oauth2-get-client.go, oauth2-token.go)
3. Fixed auth_source tracking for CalComToken

## Ship Recommendation: ship-with-gaps
The auth_protocol score is a false positive — `Bearer` auth works correctly at runtime (doctor confirms "credentials: valid"). Transcendence commands need synced data which will be validated in Phase 5 dogfood. Overall: 90/100, 86% verify pass rate, all novel features survived.
