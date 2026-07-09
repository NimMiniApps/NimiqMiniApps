# Agent instructions — Nimiq Mini Apps

This repo is the **community catalog** for Nimiq Pay mini apps (Vue frontend, Go API, MCP server).
Read this file when helping developers **submit listings**, **change the API**, or **use the MCP tools**.

---

## Submit a developer's app to the catalog

Use the **public REST API** (no account, no admin token). Do not use admin endpoints for normal submissions.

### 1. Fetch the live contract

```bash
curl -s https://api.nimiqminiapps.com/openapi.json
```

Local dev: `http://localhost:8080/openapi.json`

Source of truth in git: `docs/openapi.yaml`

### 2. Pre-check (avoid wasted submits)

- **Slug free?** `GET /api/apps/{slug}` → 404 means available (or use MCP `get_app`)
- **Domain** — hostname only, no `https://` (e.g. `myapp.example.com`)
- **Category** — one of: `Games`, `Utilities`, `Finance`, `Maps`, `Social`, `Experiments`
- **Assets** — subset of: `NIM`, `USDT`, `USDC`, `BTC`, `ETH`
- Assemble the **full payload once**; do not spam retries (rate limit below)

### 3. Submit

```bash
curl -X POST https://api.nimiqminiapps.com/api/apps/submit \
  -H "Content-Type: application/json" \
  -d '{
    "slug": "my-mini-app",
    "name": "My Mini App",
    "domain": "myapp.example.com",
    "category": "Games",
    "developer_slug": "my-team",
    "developer_name": "My Team",
    "tagline": "One sentence pitch",
    "submitter_contact": "@telegram or you@example.com",
    "description": "Plain-text short summary for listings",
    "long_description": "## Features\n\nMarkdown for the detail page",
    "tags": ["games"],
    "assets": ["NIM"],
    "release_stage": "beta"
  }'
```

**Required:** `slug`, `name`, `domain`, `category`, `developer_slug`, `developer_name`, `tagline`, `submitter_contact`

**Rate limit:** 5 requests/hour per IP → HTTP 429. On 429, stop and tell the user to wait.

**After success:** track review at `GET /api/apps/{slug}/status` or `https://nimiqminiapps.com/status/{slug}`

### 4. Building the mini app itself

Catalog submission ≠ building the app. For wallet integration, scaffolding, and pre-ship checks, use the **mini-apps** skill (`references/scaffold.md`, `references/checklist.md`) and [nimiq.dev mini apps docs](https://nimiq.dev/mini-apps/).

---

## MCP server (Cursor)

Configured in `.cursor/mcp.json` → `nimiq-miniapps`. See `mcp/README.md`.

| Tool | Use for |
|------|---------|
| `list_apps`, `get_app` | Browse catalog, check slug collision |
| `list_categories`, `get_developer` | Discovery |
| `admin_*` | **Moderators only** — requires `MINIAPPS_ADMIN_TOKEN` in `.env` |

There is **no** `submit_app` MCP tool yet. Submit via `POST /api/apps/submit` (curl or `fetch`).

`MINIAPPS_API_URL` defaults to `https://api.nimiqminiapps.com`; use `http://localhost:8080` for local compose.

---

## Changing the API in this repo

When you add or change endpoints, request/response shapes, or validation rules:

1. **Edit** `docs/openapi.yaml` (single source of truth)
2. **Regenerate** embedded copies:
   ```bash
   ./scripts/gen-openapi.sh
   ```
3. **Commit** all three:
   - `docs/openapi.yaml`
   - `backend/openapi.yaml`
   - `backend/openapi.json`

CI fails if generated files are out of sync. The Go backend serves them at `GET /openapi.json` and `GET /openapi.yaml`.

Implement handlers in `backend/`, wire routes in `backend/main.go`, add tests in `backend/*_test.go`.

4. **Update** [`README.md`](README.md) when the change is user- or developer-facing (see below).

---

## Keep README.md up to date

When your work changes how people **use, run, or integrate with** this project, update `README.md` in the same change. Do not leave the README stale.

**Update README when you:**

- Add, remove, or rename API endpoints or public URLs
- Change quick start, ports, env vars, or default tokens
- Add features users see (submit flow, admin, MCP tools, OpenAPI, new pages)
- Change stack, deploy target, or CI/image names
- Add new top-level docs agents or developers should know about

**What to touch (match existing sections):**

| Section | Update if… |
|---------|------------|
| **Stack** | Dependencies or architecture change |
| **Quick start** | `docker compose`, ports, or first-run steps change |
| **How it works** | Browse / submit / moderate behavior changes |
| **API** | Public or admin endpoints, OpenAPI, MCP, or agent docs change |
| **CI** | Workflow, registry, or publish process changes |

**Also update when relevant:**

- `docs/DEV.md` — local dev details, curl examples, env tables
- `mcp/README.md` — MCP tools, env vars, field notes
- `AGENTS.md` — agent workflows (this file)

Keep edits short: one or two bullets or lines per change. Link to deeper docs (`docs/DEV.md`, `AGENTS.md`) instead of duplicating them.

**Skip README** for internal-only refactors, test-only changes, or fixes that don't change behavior developers see.

---

## Quick reference

| Task | Where |
|------|--------|
| OpenAPI spec | `docs/openapi.yaml`, live `/openapi.json` |
| Regenerate OpenAPI | `./scripts/gen-openapi.sh` |
| Dev setup | `docs/DEV.md` |
| Public submit endpoint | `POST /api/apps/submit` |
| Web submit form | `/submit` on the frontend |
| Admin moderation | `/admin` + `Authorization: Bearer $ADMIN_TOKEN` |
| README | Update when user/dev-facing behavior changes — see **Keep README.md up to date** |
