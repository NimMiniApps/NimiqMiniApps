# Development Guide

## Stack

- `frontend/` — Vue 3 + Vite + TypeScript + Tailwind
- `backend/` — Go REST API (stdlib router + pgx)
- PostgreSQL — migrations + seed data run automatically at backend startup

Default local admin token: `dev-admin-token-change-me`

## Docker Compose (easiest)

```bash
docker compose up --build
```

- Frontend: http://localhost:5173 (nginx, production build; proxies `/api` to backend)
- Backend: http://localhost:8080
- Postgres: internal only (uncomment the ports block in `docker-compose.yml` to expose 5432)

Migrations and seed data apply automatically on backend startup. Override secrets via env:

```bash
ADMIN_TOKEN=my-secret docker compose up --build
```

If port 5173 is taken on your host (e.g. another Vite dev server), pick a different one:

```bash
FRONTEND_PORT=8090 docker compose up -d
```

## Local development (no Docker)

You still need Postgres — easiest is to run only that in Docker:

```bash
docker compose up -d postgres
# then expose it: uncomment the ports block in docker-compose.yml, or run:
docker run -d --name nimiq-pg -p 5432:5432 \
  -e POSTGRES_USER=nimiq -e POSTGRES_PASSWORD=nimiq -e POSTGRES_DB=nimiq_miniapps \
  postgres:17-alpine
```

### Backend

```bash
cd backend
export DATABASE_URL="postgres://nimiq:nimiq@localhost:5432/nimiq_miniapps?sslmode=disable"
export ADMIN_TOKEN="dev-admin-token-change-me"
export WALLET_AUTH_SECRET="dev-wallet-auth-secret-change-me"
export HTTP_ADDR=":8080"   # binds 0.0.0.0:8080
go run .
```

### Frontend

```bash
cd frontend
npm install
npm run dev            # Vite runs on 0.0.0.0:5173 and proxies /api to localhost:8080
```

Open http://localhost:5173. No `VITE_API_BASE_URL` needed thanks to the dev proxy;
set it only if the backend runs on another machine:

```bash
VITE_API_BASE_URL=http://192.168.1.50:8080 npm run dev
```

## Testing from your phone (LAN)

1. Find your LAN IP:
   ```bash
   hostname -I        # or: ip addr
   ```
2. Start the backend (binds `0.0.0.0:8080` by default) and the frontend:
   ```bash
   npm run dev -- --host 0.0.0.0
   ```
3. Open `http://<LAN-IP>:5173` on your phone.

The Vite dev server proxies API calls to the backend, so the phone only needs to
reach port 5173. The same applies to the Docker setup (nginx proxies `/api`).

If you point the phone directly at the backend (port 8080), add your LAN origin to CORS:

```bash
export CORS_ALLOWED_ORIGINS="http://localhost:5173,http://127.0.0.1:5173,http://<LAN-IP>:5173"
```

## OpenAPI

- **Source:** [`docs/openapi.yaml`](openapi.yaml) (edit this file)
- **Live:** `GET /openapi.json` and `GET /openapi.yaml` on the API host
- **Regenerate** embedded copies for the backend Docker image:

```bash
./scripts/gen-openapi.sh
```

CI validates the YAML and fails if `backend/openapi.{yaml,json}` are out of sync with `docs/openapi.yaml`.

## API examples

```bash
TOKEN=dev-admin-token-change-me
API=http://localhost:8080

# OpenAPI spec
curl $API/openapi.json

# public submission (wallet login required; forced to status=submitted, featured=false; 5/hour per IP)
# developer_slug/developer_name are derived from the caller's profile display_name, not sent here
curl -X POST $API/api/apps/submit -H "Content-Type: application/json" \
  -H "Cookie: wallet_session=<value from POST /api/auth/verify>" \
  -d '{"slug":"my-app","name":"My App","domain":"myapp.example.com","category":"Utilities","tagline":"Does a thing.","submitter_contact":"you@example.com"}'

# public reads
curl "$API/api/apps?q=game&category=Games&sort=newest"
curl "$API/api/apps?paginate=1&limit=20&offset=0"
curl $API/api/apps/nimbomber
curl $API/api/categories
curl $API/api/developers/maestro
curl $API/health
```

### Admin API examples

```bash
TOKEN=dev-admin-token-change-me
API=http://localhost:8080

# create
curl -X POST $API/api/admin/apps \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{
    "slug": "my-app",
    "name": "My App",
    "domain": "myapp.example.com",
    "category": "Utilities",
    "developer_slug": "me",
    "developer_name": "Me",
    "tagline": "Does a thing.",
    "tags": ["utility"],
    "assets": ["NIM"],
    "status": "submitted"
  }'

# update (PUT and PATCH both merge over the existing app)
curl -X PATCH $API/api/admin/apps/my-app \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"tagline": "Does a better thing.", "featured": true}'

# moderation
curl -X POST $API/api/admin/apps/my-app/approve -H "Authorization: Bearer $TOKEN"
curl -X POST $API/api/admin/apps/my-app/verify  -H "Authorization: Bearer $TOKEN"
curl -X POST $API/api/admin/apps/my-app/reject  -H "Authorization: Bearer $TOKEN"

# delete
curl -X DELETE $API/api/admin/apps/my-app -H "Authorization: Bearer $TOKEN"

# admin: list everything including submitted/rejected
curl $API/api/admin/apps -H "Authorization: Bearer $TOKEN"
```

## Notes

- Open-in-Nimiq-Pay URLs are generated as `https://nimpay.app/miniapps/open/<domain>`
  (no scheme after `/open/`; the API rejects domains containing `://`).
- Public listings only show `approved`/`verified`/`experimental` apps. `submitted` and
  `rejected` ones are visible via an explicit `?status=` query or the admin endpoints.
- Developers log in with their wallet and self-submit at `/submit` on the site (or
  `POST /api/apps/submit` with a wallet session cookie); new submissions land as
  `submitted` and appear publicly once approved in `/admin`. Once approved, the
  submitting wallet can request edits via `/apps/{slug}/update` (owner-only) or see
  all its apps at `/my-apps`.
- Backend is stateless (config via env, no local files) — Swarm-ready as-is.

### Optional backend env

| Variable | Default | Purpose |
|----------|---------|---------|
| `SITE_URL` | `https://nimiqminiapps.com` | Frontend origin used in `sitemap.xml` links |
| `API_PUBLIC_URL` | same as `SITE_URL` | Origin referenced in `robots.txt` sitemap line |
| `SUBMIT_WEBHOOK_URL` | _(empty)_ | POST JSON payload when a public submission is created |
| `DOMAIN_CHECK_ENABLED` | `true` | Periodic HTTPS probe of app domains |
| `DOMAIN_CHECK_INTERVAL` | `1h` | Re-check interval for reachable domains |
| `DOMAIN_CHECK_OFFLINE_INTERVAL` | `15m` | Re-check interval for unreachable domains (checked sooner) |
| `DOMAIN_CHECK_TICK` | `5m` | How often the worker looks for domains due for a check |
| `DOMAIN_CHECK_TIMEOUT` | `10s` | Per-domain HTTP timeout |
| `WALLET_AUTH_SECRET` | _(empty)_ | HMAC secret for wallet login session cookies; rotating it logs everyone out |
| `ADMIN_WALLET_ADDRESSES` | _(empty)_ | Comma-separated Nimiq addresses allowed to moderate via wallet session (also accepts `ADMIN_TOKEN` bearer) |

Pagination: `GET /api/apps?paginate=1&limit=20&offset=0` returns `{ items, total, limit, offset }`. Without `paginate` or `offset`, the response stays a plain JSON array (legacy).

Domain health: apps expose `domain_reachable` and `domain_checked_at` on admin listings. Trigger manually with `POST /api/admin/check-domains`.

OG prerender: `GET /og/apps/{slug}` returns HTML with Open Graph meta for social crawlers. Production nginx proxies crawler user agents on `/apps/:slug` to this endpoint.

Webhook events: `app.submitted`, `app.update_requested`.

Webhook payload shape (submission):

```json
{
  "event": "app.submitted",
  "submitted_at": "2026-07-09T12:00:00Z",
  "app": { "slug": "my-app", "name": "My App", "domain": "...", "category": "Games", "developer_name": "...", "tagline": "..." }
}
```

Public status check: `GET /api/apps/{slug}/status` or browse `/status/{slug}` on the frontend.
Status responses for live apps include `update_pending` when a change request is in the queue.

Author update requests: owners open `/apps/{slug}/update` from **My apps** (`/my-apps`), which calls `POST /api/apps/{slug}/request-update` (wallet login + ownership required). The live listing stays unchanged until an admin approves the revision in `/admin`. One pending request per app at a time.

Admin revision review:
```bash
curl $API/api/admin/revisions -H "Authorization: Bearer $TOKEN"
curl -X POST $API/api/admin/revisions/{id}/approve -H "Authorization: Bearer $TOKEN"
curl -X POST $API/api/admin/revisions/{id}/reject -H "Authorization: Bearer $TOKEN"
```

Collections: `GET /api/collections` and `GET /api/apps?collection=new-week|popular|rewards|games|usdt`.

SEO: `GET /sitemap.xml` and `GET /robots.txt` on the API host.

## Backend tests

```bash
cd backend && go test ./...
```
