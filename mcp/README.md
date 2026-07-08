# Nimiq Mini Apps MCP Server

Thin MCP wrapper around the Nimiq Mini Apps REST API so Cursor (and other MCP clients) can browse and edit the catalog.

## Setup

```bash
cd mcp
npm install
npm run build
```

## Cursor configuration

Add to `.cursor/mcp.json` (project) or `~/.cursor/mcp.json` (global):

```json
{
  "mcpServers": {
    "nimiq-miniapps": {
      "command": "node",
      "args": ["/absolute/path/to/NimiqMiniApps/mcp/dist/index.js"],
      "env": {
        "MINIAPPS_API_URL": "http://localhost:8080",
        "MINIAPPS_ADMIN_TOKEN": "dev-admin-token-change-me"
      }
    }
  }
}
```

For local dev with `docker compose up`, use the default API URL and admin token from the root README.

For production, set `MINIAPPS_API_URL` to your deployed API origin and use a strong `MINIAPPS_ADMIN_TOKEN`. Never commit real tokens.

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

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `MINIAPPS_API_URL` | `http://localhost:8080` | Backend base URL |
| `MINIAPPS_ADMIN_TOKEN` | _(empty)_ | Bearer token for admin tools |
