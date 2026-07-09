# App Stats Design

## Goal

Give developers visibility into how much their mini apps are being opened/viewed, and give admins the same data across all apps, without building a general analytics system.

## Design

Two events are tracked per app: **open** (user clicks the primary launch action — `open_url` on mobile, or the website link) and **view** (user lands on the app's detail page). Counts are raw, no dedup by visitor/IP — directional trends matter more than exact uniques here, and dedup would add session/IP handling for little benefit.

**Storage:** one daily rollup row per app, no event log:

```sql
CREATE TABLE app_stats_daily (
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    day DATE NOT NULL,
    opens INT NOT NULL DEFAULT 0,
    views INT NOT NULL DEFAULT 0,
    PRIMARY KEY (app_id, day)
);
```

Each event does `INSERT ... ON CONFLICT (app_id, day) DO UPDATE SET opens = opens + 1` (or `views`). No history is lost, but there's no per-visitor row to grow unbounded.

These are untrusted engagement counters, not popularity or reward data: the track endpoint is public and unauthenticated (see below), so counts can be inflated by anyone hitting it directly. They're good enough for developers/admins to eyeball trends, not for ranking or payouts.

**Backend (`backend/stats.go`):**

- `POST /api/apps/{slug}/track` — public, unauthenticated, rate-limited per IP+slug (a simple in-memory token bucket, e.g. 20/minute — enough to absorb real usage bursts while blocking naive spam scripts; no new dependency, just a `sync.Mutex`-guarded map). Body `{"event": "open" | "view"}`. Upserts today's row, returns 204. Bad/missing slug or malformed body just no-ops with a 204, since a lost stat ping should never surface as user-facing breakage.
- `GET /api/apps/{slug}/stats` — new `s.ownerOrAdminAuth(slug, next)` wrapper: checks the wallet cookie first (if the address is an admin wallet, or `s.isOwner(ctx, slug, address)` is true, allow), then falls back to the existing admin bearer-token check, else 401. This is a new helper alongside `walletAuthMiddleware`/`adminAuthMiddleware`, not a reuse of either alone. Returns totals plus a 30-day daily series:
  ```json
  { "totals": { "opens": 812, "views": 2044 },
    "daily": [{ "date": "2026-06-10", "opens": 12, "views": 30 }, ...] }
  ```
- `GET /api/admin/apps` (existing endpoint, `adminListApps`): add `TotalOpens`/`TotalViews int` to the `App` struct (admin-only fields, omitted from public responses like `SubmitterContact` already is via `stripPrivateAppFields`), a left-join subquery in the admin query path, and matching fields in the frontend `App` TS type / OpenAPI spec so the client and admin table stay in sync.

Stats only count public apps: `POST /track` upserts regardless (an app might go live moments after a click lands), but `GET /stats` and the admin listing only report apps that are or were public — `submitted`/`rejected` apps won't show meaningful stats since they're never linked to from the public site.

**Desktop opens:** the desktop flow (`OpenInWalletPanel.vue`) shows a QR code and a "copy link" button, not a clickable `open_url` anchor — a QR scan happens on a phone camera, invisible to page JS. Rather than add a third event type, `copyLink()` in `OpenInWalletPanel.vue` also fires the same `open` beacon: copying the link is the desktop-equivalent intent signal. QR scans themselves stay untracked (documented limitation, not silently miscounted as page views).

**Frontend:**

- New `trackAppEvent(slug: string, event: 'open' | 'view')` helper in `api.ts`, using the existing `BASE` (`VITE_API_BASE_URL`) constant and `navigator.sendBeacon(BASE + `/api/apps/${slug}/track`, ...)`, so beacon calls go through the same base-URL/proxy setup as every other request instead of a hardcoded path.
- `AppCard.vue` and `AppDetailView.vue`: the `open_url` anchor's click handler calls `trackAppEvent(slug, 'open')`.
- `OpenInWalletPanel.vue`: `copyLink()` also calls `trackAppEvent(slug, 'open')` (needs `slug` added as a prop).
- `AppDetailView.vue`: view tracking fires from the resolved-load path, not `onMounted`. Concretely, inside `loadApp(slug)` (the function called both on `onMounted` and on the `route.params.slug` watcher), call `trackAppEvent(slug, 'view')` only after `getApp(slug)` resolves successfully — this way in-app slug navigation (same component instance, new slug) is tracked, and failed/404 loads are not.
- `MyAppsView.vue`: each owned app gets a stats block — total opens/views, last-7-days number, and a small hand-rolled inline SVG sparkline of the last 30 days (no new chart dependency; none of the current `dependencies` cover charting, and a single sparkline doesn't justify adding one).
- `AdminView.vue`: existing apps table gains sortable opens/views columns, sourced from the new `TotalOpens`/`TotalViews` fields.

## Testing

Backend: table-driven tests for the upsert logic (new day creates a row, same day increments), the per-IP rate limit on `POST /track`, the `ownerOrAdminAuth` gate on `GET .../stats` (owner wallet, admin wallet, admin bearer token, and rejected cases), and that malformed track bodies no-op cleanly. Frontend: a unit test for the sparkline's path-generation math, and a test that `loadApp` only tracks a view after a successful (not failed) load. Full verification includes `go test ./...`, frontend `vitest`, frontend build, and OpenAPI regeneration for the two new/changed endpoints.
