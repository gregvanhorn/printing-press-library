# Capturing Instacart history hashes — SUPERSEDED

**This document is superseded.**

The original PR #78 assumed Instacart exposed dedicated GraphQL ops named `BuyItAgainPage` and `CustomerOrderHistory`. It does not. The `history sync` command that relied on those ops cannot be made to work because the underlying operations are fictional — see [`docs/solutions/best-practices/instacart-orders-no-clean-graphql-op.md`](../../../docs/solutions/best-practices/instacart-orders-no-clean-graphql-op.md).

## Use this instead

For the working path to backfill your Instacart purchase history, see:

- **Playbook:** [`docs/patterns/authenticated-session-scraping.md`](../../../docs/patterns/authenticated-session-scraping.md) — the canonical reference for this class of problem across any printed CLI.
- **Dumper:** `docs/dumper.js` + `docs/extract-one.js` — the browser-side JS to walk orders and export JSONL.
- **Importer:** `instacart history import <path>` — the CLI-side command that reads the JSONL and populates the history tables.

## Quick-start (10 minutes end-to-end)

1. Open `https://www.instacart.com/store/account/orders` in Chrome, logged in.
2. Run `dumper.js` to collect order IDs (via DevTools console or agent-driven Chrome MCP).
3. For each order ID, navigate to `/store/orders/<id>` and run `extract-one.js`.
4. Run the exporter snippet at the bottom of `extract-one.js` to download `instacart-orders.jsonl`.
5. `instacart history import ~/Downloads/instacart-orders.jsonl`.
6. `instacart add <retailer> "<something you've bought>" --dry-run --json` → look for `"resolved_via": "history"`.

The full walkthrough with tooling rationale is in the playbook.
