---
name: pp-yahoo-finance
description: "Use this skill whenever the user asks about stock quotes, share prices, market data, options chains, analyst recommendations, earnings, fundamentals, dividends, portfolio tracking, or market movers — or whenever they mention a specific ticker (AAPL, MSFT, NVDA, SPY, TSLA, etc.) in a context that suggests looking it up. Yahoo Finance CLI with portfolio tracking, watchlist-driven digests, options moneyness filters, and a Chrome-session fallback for when Yahoo rate-limits the current IP. No API key. Works offline once synced. Triggers on natural phrasings like 'how's AAPL doing', 'track my portfolio', 'what are the most active stocks today', 'show me SPY options expiring soon', 'what's my unrealized P&L'."
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata: '{"openclaw":{"requires":{"bins":["yahoo-finance-pp-cli"]},"install":[{"id":"go","kind":"shell","command":"go install github.com/mvanhorn/printing-press-library/library/commerce/yahoo-finance/cmd/yahoo-finance-pp-cli@latest","bins":["yahoo-finance-pp-cli"],"label":"Install via go install"}]}}'
---

# Yahoo Finance — Printing Press CLI

Every Yahoo Finance feature, plus the things every existing tool is missing: portfolio tracking, watchlist-driven digests, options moneyness filters, and a Chrome-session fallback for when Yahoo blocks your IP. Powered by the reverse-engineered `query1.finance.yahoo.com` and `query2.finance.yahoo.com` endpoints that `yfinance`, `yahoo-finance2`, and `yahooquery` have proven over a decade. No API key required.

## When to Use This CLI

Reach for this when a user wants market data, price history, fundamentals, options chains, analyst recommendations, or wants to track a portfolio or watchlist locally. It's the only Yahoo Finance tool that can recover when Yahoo rate-limits the current IP, via the `auth login-chrome` flow that imports a live browser session.

Don't reach for this when the user needs a real-time streaming feed (Yahoo's endpoints are snapshot-only, ~15-minute delayed) or has a paid feed like Polygon or Alpaca that supplies cleaner data.

## Unique Capabilities

These aren't available in any other Yahoo Finance CLI, library, or MCP server.

### Local state that compounds

- **`portfolio add <symbol> <shares> <cost-per-share> [--purchased YYYY-MM-DD]`** — Record a purchase lot. Multiple lots per symbol supported.

  _Transforms a quote-lookup tool into a position-tracking tool. No other Yahoo Finance wrapper has lot-level state._

- **`portfolio perf`** — Current market value, cost basis, unrealized P&L per position, and portfolio total. Joins live quotes with local lots.

- **`portfolio gains`** — Per-lot unrealized gain/loss sorted by magnitude. Used for tax-lot selection when selling.

- **`watchlist create|add|show|list|remove|delete`** — Named collections of tickers backed by SQLite. Feed them into multi-symbol commands like `digest` and `compare`.

- **`sql "<query>"`** — Raw SQLite against the local database. Cross-entity queries work: `SELECT symbol, SUM(shares*cost_basis) FROM portfolio_lots GROUP BY symbol`.

### Commands that only make sense with local state

- **`digest --watchlist <name>`** — Biggest gainers, losers, and headline quotes across a watchlist. Morning briefing in a single line.

  _Compresses the "check my holdings" ritual into one agent call._

- **`compare <symbol> <symbol> [symbol...]`** — Side-by-side quote + 52w range + market cap across 2+ symbols, parallel-fetched.

- **`sparkline <symbol>`** — Unicode sparkline (`▁▂▃▄▅▆▇█`) of recent price action. Zero-config terminal chart.

- **`options-chain <symbol> [--moneyness otm|itm|atm] [--max-dte N] [--type calls|puts]`** — Options chain filtered to out-of-the-money calls expiring within 45 days, for example. Yahoo's raw endpoint returns everything; this filters to what traders actually want.

- **`fx <from> <to> [--amount N]`** — Currency conversion using Yahoo Finance's FX pairs. `fx USD EUR --amount 100` in one line.

### Reachability mitigation

- **`auth login-chrome`** — When Yahoo returns HTTP 429 for your IP (common on cloud providers and some residential ISPs), import a live Chrome session and the CLI uses its crumb + cookies instead.

  _This is the differentiator. IPs get blocked; without Chrome import, the CLI is dead. With it, one manual cookie export restores full functionality._

- **Adaptive rate limiter** — Starts conservative, ramps up on success, halves on 429, persists the ceiling per-session.

- **Automatic crumb bootstrap** — The `fc.yahoo.com` → `getcrumb` handshake that every working wrapper does, built-in and cached to disk for 24 hours.

## Command Reference

Base commands (spec-derived):

- `yahoo-finance-pp-cli quote --symbols AAPL,MSFT,NVDA` — Current quotes (batched, comma-separated)
- `yahoo-finance-pp-cli chart <symbol>` — Historical OHLCV price data
- `yahoo-finance-pp-cli fundamentals <symbol>` — EPS, revenue, margins, cash flow time series
- `yahoo-finance-pp-cli insights <symbol>` — Technical events, valuation, research reports
- `yahoo-finance-pp-cli options <symbol>` — Raw options chain (calls and puts)
- `yahoo-finance-pp-cli recommendations <symbol>` — Symbols that share analyst recommendations
- `yahoo-finance-pp-cli screener <id>` — Predefined screener (`day_gainers`, `most_actives`, etc.)
- `yahoo-finance-pp-cli trending <region>` — Top trending symbols in a region (e.g., `US`)
- `yahoo-finance-pp-cli search <query>` — Full-text search across synced data or live API
- `yahoo-finance-pp-cli autocomplete <prefix>` — Legacy symbol autocomplete (faster than search)
- `yahoo-finance-pp-cli sync` / `export` / `import` — Local-store management
- `yahoo-finance-pp-cli doctor` — Verify setup and credentials

Unique commands (see Unique Capabilities above): `portfolio`, `watchlist`, `digest`, `compare`, `sparkline`, `sql`, `fx`, `options-chain`, `auth login-chrome`.

## Recipes

Multi-step flows that combine commands.

### Morning briefing over a watchlist

```bash
yahoo-finance-pp-cli watchlist create tech
yahoo-finance-pp-cli watchlist add tech AAPL MSFT NVDA GOOG META
yahoo-finance-pp-cli digest --watchlist tech --agent
```

Builds a named watchlist once, then `digest` surfaces the biggest overnight movers and quote snapshots across every ticker. Run daily from cron or a morning kickoff. Add/remove symbols any time without rebuilding.

### Track a real portfolio with cost basis

```bash
yahoo-finance-pp-cli portfolio add AAPL 50 185.50 --purchased 2024-06-15
yahoo-finance-pp-cli portfolio add MSFT 20 340.00 --purchased 2024-03-01
yahoo-finance-pp-cli portfolio perf --agent
yahoo-finance-pp-cli portfolio gains --agent
```

Record every purchase as a lot (symbol, shares, cost-per-share, purchase date). `perf` returns unrealized P&L per position and portfolio total, joining live quotes with local lots. `gains` breaks down per-lot gain/loss for tax-lot selection.

### Fallback when Yahoo blocks your IP

```bash
# 1. Open finance.yahoo.com in Chrome. Accept cookies. Stay signed out.
# 2. Export cookies for *.yahoo.com as JSON (use a cookie-export browser extension).
# 3. Get a crumb from DevTools console on finance.yahoo.com:
#    fetch('/v1/test/getcrumb').then(r => r.text()).then(console.log)
yahoo-finance-pp-cli auth login-chrome --cookies ~/yahoo-cookies.json --crumb abc123
yahoo-finance-pp-cli doctor  # verify the imported session works
```

One-time setup when running from a cloud IP or otherwise-blocked address. The imported session persists to `~/.config/yahoo-finance-pp-cli/session.json`. Automatic crumb bootstrap takes over once the session works.

## Auth Setup

Yahoo Finance has no official API and no API key. The CLI bootstraps a session automatically by visiting `fc.yahoo.com`, extracting A1/B1 cookies, and fetching a crumb from `/v1/test/getcrumb` — the same pattern every working wrapper uses.

If your IP is rate-limited (common on cloud providers; Yahoo returns HTTP 429 on every request), use the Chrome-import recipe above. The CLI then uses the imported crumb and cookies on every request, bypassing the automatic handshake.

Run `yahoo-finance-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. It expands to `--json --compact --no-input --no-color --yes` — structured output, no prompts, machine-parseable. Every command also accepts `--select <fields>` for cherry-picked fields, `--dry-run` to preview the request, and `--no-cache` to bypass the 5-minute GET cache.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream Yahoo issue) |
| 7 | Rate limited (wait and retry, or switch to `auth login-chrome`) |
| 10 | Config error |

## Installation

### CLI

1. Check Go: `go version` (requires Go 1.23+)
2. Install: `go install github.com/mvanhorn/printing-press-library/library/commerce/yahoo-finance/cmd/yahoo-finance-pp-cli@latest`
3. Verify: `yahoo-finance-pp-cli --version`
4. Ensure `$GOPATH/bin` is on `$PATH`.

### MCP Server

```bash
go install github.com/mvanhorn/printing-press-library/library/commerce/yahoo-finance/cmd/yahoo-finance-pp-mcp@latest
claude mcp add yahoo-finance-pp-mcp -- yahoo-finance-pp-mcp
claude mcp list  # verify
```

## Argument Parsing

Given `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → run `yahoo-finance-pp-cli --help`
2. **`install`** → CLI installation; **`install mcp`** → MCP installation
3. **Anything else** → check `which yahoo-finance-pp-cli` (offer to install if missing), match the user's intent to a command from Unique Capabilities or Command Reference above, then run it with `--agent` for structured output. Drill into subcommand help if the mapping is ambiguous.
