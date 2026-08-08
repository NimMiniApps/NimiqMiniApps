# Admin App List Filters Design

## Goal

Make `/admin` manageable as the catalog grows by filtering the main app list by category, status, and text search, with filters persisted in the URL.

## Approach

Client-side only. Keep `GET /api/admin/apps` as a full list; filter and sort in `AdminView.vue`. No backend, OpenAPI, or MCP changes.

## UX

Above the main app list (alongside existing Sort controls):

- **Search** — matches name, slug, or domain (case-insensitive)
- **Category** — All + `APP_CATEGORIES`
- **Status** — All + `submitted` / `approved` / `verified` / `experimental` / `rejected`
- Keep Sort by name / opens / views
- Show “N of M apps” when any filter is active
- Empty state: “No apps match these filters”
- Show category on each list row
- Pending revisions, featured order, and offline sections stay unfiltered

## URL

| Param | Example |
|-------|---------|
| `q` | `?q=swap` |
| `category` | `?category=Games` |
| `status` | `?status=submitted` |

Omit empty params. Coexist with `?edit=`. Invalid category/status → treat as All. Debounce search ~200ms before writing `q` to the URL.

## Data flow

1. Read `q` / `category` / `status` from `route.query` into local refs
2. On change: `router.replace` with updated query (drop empty keys)
3. Pipeline: `apps` → filter → existing sort → render

## Out of scope

Server-side filtering, shared filter components, filtering moderation queues.
