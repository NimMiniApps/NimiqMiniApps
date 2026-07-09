# Nimiq Mini Apps MCP Server

Thin MCP wrapper around the Nimiq Mini Apps REST API so Cursor (and other MCP clients) can browse and edit the catalog.

## Setup

```bash
cd mcp
npm install
npm run build
```

## Cursor configuration

This repo includes a `.cursor/mcp.json` that runs `mcp/run-with-env.sh`.
The launcher loads the repository `.env`, maps `ADMIN_TOKEN` to
`MINIAPPS_ADMIN_TOKEN`, and defaults `MINIAPPS_API_URL` to
`https://api.nimiqminiapps.com`.

Add to `.cursor/mcp.json` (project) or `~/.cursor/mcp.json` (global):

```json
{
  "mcpServers": {
    "nimiq-miniapps": {
      "command": "/absolute/path/to/NimiqMiniApps/mcp/run-with-env.sh",
      "args": []
    }
  }
}
```

For local dev with `docker compose up`, set `MINIAPPS_API_URL=http://localhost:8080`
in `.env`. For production, keep `ADMIN_TOKEN` or `MINIAPPS_ADMIN_TOKEN` in `.env`.
Never commit real tokens.

During development you can run without building:

```json
{
  "command": "npx",
  "args": ["tsx", "/absolute/path/to/NimiqMiniApps/mcp/src/index.ts"]
}
```

## Tools

| Tool | Auth | Description |
|------|------|-------------|
| `miniapps_info` | — | API URL and whether admin token is configured |
| `health_check` | — | API / DB health |
| `list_apps` | — | Public catalog with filters |
| `get_app` | — | Single app by slug |
| `list_categories` | — | Category counts |
| `get_developer` | — | Developer profile + apps |
| `list_developers` | — | All developers with public app counts |
| `get_related_apps` | — | Up to 4 related public apps |
| `admin_list_apps` | admin | All apps, any status |
| `admin_search_users` | admin | Find wallet owners by display name or address |
| `admin_add_app_owner` | admin | Link a wallet as co-owner |
| `admin_remove_app_owner` | admin | Unlink a wallet (can remove last owner) |
| `admin_create_app` | admin | Create app |
| `admin_update_app` | admin | Partial update (merges with existing) |
| `admin_delete_app` | admin | Delete app |
| `admin_approve_app` | admin | Set status to approved |
| `admin_verify_app` | admin | Set status to verified |
| `admin_reject_app` | admin | Set status to rejected |

There is **no** `submit_app` MCP tool. Public submission requires a **wallet session cookie** (`POST /api/apps/submit` after `POST /api/auth/verify`) — direct developers to `/submit` in the browser, or use `admin_create_app` / `admin_update_app` plus `admin_add_app_owner` for agent workflows.

## Catalog field notes

When creating or updating apps via MCP (or the REST API):

| Field | Format | Notes |
|-------|--------|--------|
| `tagline` | Plain text | Shown on app cards |
| `description` | Plain text | Short summary; listings / meta |
| `long_description` | **Markdown** | Rendered on the app detail page (bold, lists, links, headings, code). HTML is stripped. |
| `reward_assets` | array of `NIM`, `USDT`, `USDC`, `BTC`, `ETH` | Assets users can actually receive from the app: daily rewards, leaderboard prizes, payouts, tips, faucets, or similar receive-side flows. Leave empty when the app merely accepts, displays, swaps, or supports a token. |
| `domain` | Host/path | `https://` / `http://` stripped automatically if pasted |
| `owner_wallet_addresses` | array of Nimiq addresses (read-only) | Wallets that can self-service manage this app (My apps, request-update). Manage via `admin_add_app_owner` / `admin_remove_app_owner`, not through create/update. |
| `developer_name` / `developer_slug` | Plain text | Public catalog identity — set directly; unaffected by ownership. |
| `submitter_contact` | Plain text | Private; admin responses only. Required for browser wallet submit, not for admin/MCP create. |
| `socials` | `{ platform, url }[]` | App's public social links — not the same as submitter contact |

Public `get_app` / `list_apps` responses omit `submitter_contact`. Use `admin_list_apps` to see it. `owner_wallet_addresses` is included on public app objects (empty array when unclaimed).

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `MINIAPPS_API_URL` | `http://localhost:8080` | Backend base URL |
| `MINIAPPS_ADMIN_TOKEN` | _(empty)_ | Bearer token for admin tools |

**OpenAPI** — `GET /openapi.json` on the API host documents all endpoints, schemas, and rate limits.
Wallet-authenticated routes (`/api/apps/submit`, `/api/my/apps`, `/api/apps/{slug}/request-update`) are documented there but are not exposed as MCP tools.
