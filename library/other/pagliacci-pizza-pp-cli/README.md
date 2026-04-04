# Pagliacci Pizza CLI

REST API for Pagliacci Pizza - Seattle's favorite pizza. Browse stores, menus, slices, time windows, manage rewards, view order history, and place orders. Authenticated endpoints use the custom PagliacciAuth scheme with a required Version-Num header.

Learn more at [Pagliacci Pizza](https://pagliacci.com).

## Install

### Go

```
go install github.com/mvanhorn/printing-press-library/library/other/pagliacci-pizza-pp-cli/cmd/pagliacci-pizza-pp-cli@latest
```

### Binary

Download from [Releases](https://github.com/mvanhorn/printing-press-library/releases).

## Quick Start

### 1. Install

See [Install](#install) above.

### 2. Set Up Credentials

Get your API key from your API provider's developer portal. The key typically looks like a long alphanumeric string.

```bash
export PAGLIACCI_PIZZA_PAGLIACCI_AUTH="<paste-your-key>"
```

You can also persist this in your config file at `~/.config/pagliacci-pizza-pp-cli/config.toml`.

### 3. Verify Setup

```bash
pagliacci-pizza-pp-cli doctor
```

This checks your configuration and credentials.

### 4. Try Your First Command

```bash
pagliacci-pizza-pp-cli access-device list
```

## Usage

<!-- HELP_OUTPUT -->

## Commands

### access-device

Manage access device

- **`pagliacci-pizza-pp-cli access-device get`** - Get device access info

### address-info

Manage address info

- **`pagliacci-pizza-pp-cli address-info lookup-address`** - Validate and look up a delivery address

### address-name

Manage address name

- **`pagliacci-pizza-pp-cli address-name list-saved-addresses`** - List saved delivery addresses

### customer

Manage customer

- **`pagliacci-pizza-pp-cli customer get`** - Get customer profile

### feedback

Manage feedback

- **`pagliacci-pizza-pp-cli feedback submit`** - Submit customer feedback

### login

Manage login

- **`pagliacci-pizza-pp-cli login login`** - Log in to account

### logout

Manage logout

- **`pagliacci-pizza-pp-cli logout logout`** - Log out of account

### menu-cache

Manage menu cache

- **`pagliacci-pizza-pp-cli menu-cache get`** - Get full cached menu

### menu-slices

Manage menu slices

- **`pagliacci-pizza-pp-cli menu-slices list`** - List available pizza slices

### menu-top

Manage menu top

- **`pagliacci-pizza-pp-cli menu-top get`** - Get featured menu items

### migrate-answer

Manage migrate answer

- **`pagliacci-pizza-pp-cli migrate-answer migrate_answer`** - Answer account migration question

### migrate-question

Manage migrate question

- **`pagliacci-pizza-pp-cli migrate-question migrate_question`** - Get account migration question

### order-list

Manage order list

- **`pagliacci-pizza-pp-cli order-list list-orders`** - List order history (paginated)

### order-list-item

Manage order list item

- **`pagliacci-pizza-pp-cli order-list-item get-order-details`** - Get specific order details

### order-list-pending

Manage order list pending

- **`pagliacci-pizza-pp-cli order-list-pending list-pending-orders`** - List active/pending orders

### order-price

Manage order price

- **`pagliacci-pizza-pp-cli order-price price-order`** - Price an order
- **`pagliacci-pizza-pp-cli order-price price-order-guest`** - Price an order as guest

### order-send

Manage order send

- **`pagliacci-pizza-pp-cli order-send send-order`** - Submit an order
- **`pagliacci-pizza-pp-cli order-send verify-order`** - Verify an order before submission

### order-suggestion

Manage order suggestion

- **`pagliacci-pizza-pp-cli order-suggestion get`** - Get suggested orders

### password-forgot

Manage password forgot

- **`pagliacci-pizza-pp-cli password-forgot forgot-password`** - Request password reset email

### password-reset

Manage password reset

- **`pagliacci-pizza-pp-cli password-reset reset-password`** - Reset password

### product-price

Manage product price

- **`pagliacci-pizza-pp-cli product-price get`** - Get product pricing

### quote-building

Manage quote building

- **`pagliacci-pizza-pp-cli quote-building get`** - Get delivery building quote

### quote-store

Manage quote store

- **`pagliacci-pizza-pp-cli quote-store get`** - Get store-specific pricing
- **`pagliacci-pizza-pp-cli quote-store list`** - List store availability and pricing

### register

Manage register

- **`pagliacci-pizza-pp-cli register register`** - Create a new account

### reward-card

Manage reward card

- **`pagliacci-pizza-pp-cli reward-card get`** - Get loyalty rewards card

### site-wide-message

Manage site wide message

- **`pagliacci-pizza-pp-cli site-wide-message get`** - Get system-wide message

### store

Manage store

- **`pagliacci-pizza-pp-cli store list`** - List all Pagliacci Pizza stores

### stored-coupons

Manage stored coupons

- **`pagliacci-pizza-pp-cli stored-coupons list`** - List saved coupons

### stored-credit

Manage stored credit

- **`pagliacci-pizza-pp-cli stored-credit get`** - Get account credit balance

### stored-gift

Manage stored gift

- **`pagliacci-pizza-pp-cli stored-gift get`** - Get gift card info

### time-window-days

Manage time window days

- **`pagliacci-pizza-pp-cli time-window-days get`** - Get available days for service

### time-windows

Manage time windows

- **`pagliacci-pizza-pp-cli time-windows get-by-date`** - Get time slots for a specific day
- **`pagliacci-pizza-pp-cli time-windows get-today`** - Get today's time slots

### transfer-gift

Manage transfer gift

- **`pagliacci-pizza-pp-cli transfer-gift transfer_gift`** - Transfer a gift card

### version

Manage version

- **`pagliacci-pizza-pp-cli version get`** - Get API version


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
pagliacci-pizza-pp-cli access-device list

# JSON for scripting and agents
pagliacci-pizza-pp-cli access-device list --json

# Filter to specific fields
pagliacci-pizza-pp-cli access-device list --json --select id,name,status

# Dry run — show the request without sending
pagliacci-pizza-pp-cli access-device list --dry-run

# Agent mode — JSON + compact + no prompts in one flag
pagliacci-pizza-pp-cli access-device list --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Retryable** - creates return "already exists" on retry, deletes return "already deleted"
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - `echo '{"key":"value"}' | pagliacci-pizza-pp-cli <resource> create --stdin`
- **Cacheable** - GET responses cached for 5 minutes, bypass with `--no-cache`
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set
- **Progress events** - paginated commands emit NDJSON events to stderr in default mode

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Cookbook

Common workflows and recipes:

```bash
# List resources as JSON for scripting
pagliacci-pizza-pp-cli access-device list --json

# Filter to specific fields
pagliacci-pizza-pp-cli access-device list --json --select id,name,status

# Dry run to preview the request
pagliacci-pizza-pp-cli access-device list --dry-run

# Sync data locally for offline search
pagliacci-pizza-pp-cli sync

# Search synced data
pagliacci-pizza-pp-cli search "query"

# Export for backup
pagliacci-pizza-pp-cli export --format jsonl > backup.jsonl
```

## Health Check

```bash
pagliacci-pizza-pp-cli doctor
```

<!-- DOCTOR_OUTPUT -->

## Configuration

Config file: `~/.config/pagliacci-pizza-pp-cli/config.toml`

Environment variables:
- `PAGLIACCI_PIZZA_PAGLIACCI_AUTH`

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `pagliacci-pizza-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $PAGLIACCI_PIZZA_PAGLIACCI_AUTH`

**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- If persistent, wait a few minutes and try again

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
