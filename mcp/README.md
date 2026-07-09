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
| `admin_list_apps` | admin | All apps, any status |
| `admin_create_app` | admin | Create app |
| `admin_update_app` | admin | Partial update (merges with existing) |
| `admin_delete_app` | admin | Delete app |
| `admin_approve_app` | admin | Set status to approved |
| `admin_verify_app` | admin | Set status to verified |
| `admin_reject_app` | admin | Set status to rejected |

## Catalog field notes

When creating or updating apps via MCP (or the REST API):

| Field | Format | Notes |
|-------|--------|--------|
| `tagline` | Plain text | Shown on app cards |
| `description` | Plain text | Short summary; listings / meta |
| `long_description` | **Markdown** | Rendered on the app detail page (bold, lists, links, headings, code). HTML is stripped. |
| `submitter_contact` | Plain text | Required for public submit; admin-only (Telegram, email, etc.) |
| `socials` | `{ platform, url }[]` | App's public social links — not the same as submitter contact |

Public `get_app` / `list_apps` responses omit `submitter_contact`. Use `admin_list_apps` to see it.

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `MINIAPPS_API_URL` | `http://localhost:8080` | Backend base URL |
| `MINIAPPS_ADMIN_TOKEN` | _(empty)_ | Bearer token for admin tools |
