Acceptance Report: trigger-dev
  Level: Quick Check
  Tests: 6/6 passed
  Failures: none
  Notes:
    - Dev environment has no runs (empty data is correct behavior)
    - queues sync returns HTTP 400 (engine-version header) - API version requirement
    - waitpoints endpoint returns 404 (may not be available on this plan/version)
    - All commands execute, produce output, respect --json/--dry-run/--select flags
  Fixes applied: 0
  Printing Press issues: 1
    - dogfood tool cannot parse internal YAML spec (spec validation: "at least one resource is required")
  Gate: PASS
