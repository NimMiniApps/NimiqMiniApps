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

**Backend (`backend/stats.go`):**

- `POST /api/apps/{slug}/track` — public, unauthenticated. Body `{"event": "open" | "view"}`. Upserts today's row, returns 204. Fire-and-forget by design: bad/missing slug or malformed body just no-ops with a 204, since a lost stat ping should never surface as user-facing breakage.
- `GET /api/apps/{slug}/stats` — requires wallet auth + app ownership (same pattern as `addAppOwnerSelf` in `developer.go`), or admin auth. Returns totals plus a 30-day daily series:
  ```json
  { "totals": { "opens": 812, "views": 2044 },
    "daily": [{ "date": "2026-06-10", "opens": 12, "views": 30 }, ...] }
  ```
- `GET /api/admin/apps` (existing endpoint) gains `total_opens` / `total_views` columns via a subquery, so the admin apps table can show and sort by popularity without a separate endpoint.

**Frontend:**

- `AppCard.vue` and `AppDetailView.vue`: the `open_url` anchor's click handler fires `navigator.sendBeacon('/api/apps/{slug}/track', JSON.stringify({event: 'open'}))` — beacon is used because the tab is about to navigate away and a normal fetch could get cancelled.
- `AppDetailView.vue`: fires a `view` beacon once on mount.
- `MyAppsView.vue`: each owned app gets a stats block — total opens/views, last-7-days number, and a small hand-rolled inline SVG sparkline of the last 30 days (no new chart dependency; none of the current `dependencies` cover charting, and a single sparkline doesn't justify adding one).
- `AdminView.vue`: existing apps table gains sortable opens/views columns.

## Testing

Backend: table-driven tests for the upsert logic (new day creates a row, same day increments), the ownership/admin auth gate on `GET .../stats`, and that malformed track bodies no-op cleanly. Frontend: a unit test for the sparkline's path-generation math (given daily data, produces expected SVG points). Full verification includes `go test ./...`, frontend `vitest`, frontend build, and OpenAPI regeneration for the two new/changed endpoints.
