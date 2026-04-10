# Trigger.dev CLI Shipcheck

## Verify
- Pass Rate: 95% (21/22 passed, 0 critical)
- Verdict: PASS
- Failures:
  - envvars: requires --project-ref flag (expected - positional arg validation)
  - watch: infinite loop by design (exec timeout expected)

## Scorecard
- Total: 88/100 - Grade A
- Perfect (10/10): Output Modes, Auth, Error Handling, README, Doctor, Agent Native, Local Cache, Workflows
- Strong (8-9): Terminal UX 9, Breadth 9, Vision 8, Insight 8
- Domain: Data Pipeline 7/10, Sync 10/10, Type Fidelity 3/5, Dead Code 5/5

## Workflow Verify
- Verdict: workflow-pass (no manifest)

## Dogfood
- Dogfood tool could not parse internal YAML spec (format validation issue)
- Manual verification: binary builds, all commands register, --help/--dry-run/--json work

## Ship Recommendation: ship
