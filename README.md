# Nimiq Mini Apps

A community directory for [Nimiq Pay](https://nimiq.com) mini apps. Browse apps, open
them straight in your Nimiq Pay wallet, and submit your own for review.

## Stack

| Part | Tech |
|------|------|
| `frontend/` | Vue 3 + Vite + TypeScript + Tailwind |
| `backend/` | Go REST API (stdlib router, pgx) |
| Database | PostgreSQL (migrations + seed run automatically at startup) |
| Deploy | Docker Compose (Swarm-ready: stateless backend, env config, healthchecks) |

## Quick start

```bash
docker compose up --build
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- Default admin token: `dev-admin-token-change-me` (override with `ADMIN_TOKEN=...`)

See [docs/DEV.md](docs/DEV.md) for local development, testing from a phone on your
LAN, and API examples.

## How it works

- **Browse** — search, filter by category, and view app details. Every app gets an
  `Open in Nimiq Pay` link of the form `https://nimpay.app/miniapps/open/<domain>`.
- **Submit** — developers submit apps at `/submit` (rate-limited, no account needed).
  Submissions are hidden until reviewed.
- **Moderate** — admins approve, verify, or reject submissions at `/admin` using a
  bearer token.

## API

Public: `GET /api/apps` (with `q`, `category`, `status`, `featured`, `sort`),
`GET /api/apps/{slug}`, `GET /api/categories`, `GET /api/developers/{slug}`,
`POST /api/apps/submit`, `GET /health`.

Admin (`Authorization: Bearer <ADMIN_TOKEN>`): CRUD under `/api/admin/apps` plus
`/verify`, `/approve`, `/reject` actions.

**MCP** — see [`mcp/README.md`](mcp/README.md) for a Cursor MCP server that wraps this API.

## CI

GitHub Actions tests and builds both services on every push/PR, and publishes Docker
images to GHCR (`ghcr.io/nimminiapps/nimiq-mini-apps-{backend,frontend}`) on pushes
to `main`.
