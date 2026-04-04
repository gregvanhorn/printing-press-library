# Dub CLI Shipcheck Report

## Dogfood
- Path validity: 5/5 PASS
- Auth protocol: MISMATCH (false positive — code uses Bearer prefix correctly at runtime, dogfood static check misdetects)
- Dead flags: 0 PASS
- Dead functions: 0 PASS (fixed: removed formatCompact)
- Data pipeline: GOOD (domain-specific upsert + search)
- Examples: 10/10 PASS
- Unregistered commands: 3 (false positive — bounties submissions are nested, not top-level)

## Verify
- Pass rate: 95% (21/22)
- Critical failures: 0
- Failed: tail (requires resource argument — not a runtime bug)
- Data pipeline: PASS

## Workflow-Verify
- Verdict: workflow-pass (no manifest)

## Scorecard
- Total: 92/100 Grade A
- Output Modes: 10/10
- Auth: 8/10
- Error Handling: 10/10
- Terminal UX: 9/10
- README: 10/10
- Doctor: 10/10
- Agent Native: 10/10
- Local Cache: 10/10
- Breadth: 10/10
- Vision: 9/10
- Workflows: 10/10
- Insight: 10/10

## Fixes Applied
1. Removed duplicate analytics command registration
2. Removed dead formatCompact function
3. Set AuthSource = "bearer" for DubToken auth path

## Ship Recommendation: ship
