# Admin Product Analytics Design

## Goal

Give administrators a clear, privacy-minimal view of whether the catalog is attracting visitors and helping them discover and open mini apps.

The dashboard measures the catalog-controlled journey:

`Catalog visit -> wallet login -> app detail view -> Nimiq Pay open or link copy`

An activation means either opening an app in Nimiq Pay or successfully copying its link. It does not claim to measure activity inside independently operated mini apps.

## Product Surface

Add a dedicated `/admin/analytics` route rather than expanding the existing moderation page. Both pages use the current admin wallet-session or bearer-token authentication.

The moderation page links to Analytics, and the analytics page links back to moderation.

The analytics page provides:

- Date ranges for 7, 30, and 90 days, plus all time.
- Change against the immediately preceding equivalent period for 7, 30, and 90 days.
- Summary cards for visits, unique visitors, wallet logins, unique wallets, app-detail views, and activations.
- A daily trend chart with selectable metrics.
- A unique-visitor funnel from catalog visitor to app viewer to activation.
- A searchable, sortable per-app table with views, unique viewers, Nimiq Pay opens, link copies, combined activations, and view-to-activation conversion.
- Clear metric definitions, the analytics collection start date, and partial-history messaging.
- Responsive loading, empty, error, and retry states.

All-time results have no previous-period comparison because there is no meaningful preceding equivalent period.

## Metric Definitions

- **Visit:** one catalog visit per browser session.
- **Unique visitor:** one pseudonymous browser identifier in the selected period.
- **Wallet login:** each successfully verified signed wallet login.
- **Unique wallet:** one server-generated wallet hash in the selected period.
- **App view:** an app-detail visit, deduplicated once per app per browser session.
- **Nimiq Pay open:** every deliberate action that opens an app in Nimiq Pay.
- **Link copy:** every successful copy-link action.
- **Activation:** a Nimiq Pay open or successful link copy.
- **App conversion:** unique visitors who activated the app divided by unique visitors who viewed it.

The funnel uses unique browser visitors at every step. Unique-wallet counts are a separate supporting metric and never expose wallet identities.

## Privacy Model

The browser creates a random visitor identifier in `localStorage` and a random session identifier in `sessionStorage`.

The API transforms these values with a dedicated, stable `ANALYTICS_HASH_SECRET` before persistence. Raw browser identifiers are never stored. A wallet hash is computed server-side only after successful wallet authentication; clients cannot submit a wallet address as analytics metadata.

Analytics stores no IP addresses, names, user-agent strings, referrers, raw wallet addresses, or arbitrary event metadata. Admin responses contain aggregate numbers only.

Detailed pseudonymous events are retained for 90 days so the fixed 7-, 30-, and 90-day ranges can calculate exact funnels and unique counts. Daily aggregate totals are retained permanently. Compact first-seen and last-seen uniqueness records are retained so all-time unique visitor, wallet, and per-app numbers remain available without retaining detailed browsing history indefinitely.

## Data Model

Introduce three related analytics stores:

1. A short-lived event table containing the event time, strict event type, hashed visitor and session identifiers, optional server-generated wallet hash, and optional app ID.
2. A permanent daily aggregate table keyed by day, event type, and optional app ID.
3. A permanent compact uniqueness table keyed by subject kind, event type, optional app ID, and hash, with first-seen and last-seen timestamps.

Only the event types `catalog_visit`, `wallet_login`, `app_view`, `app_open`, and `app_link_copy` are accepted. App events require a valid public app slug. A wallet-login event is written from the successful authentication path rather than trusted from a public event payload.

Catalog visits are deduplicated once per session. App views are deduplicated once per app and session. Opens and successful link copies record every deliberate action.

The existing `app_stats_daily` data and owner-facing app stats contract remain intact. New app-view and app-open events continue updating those counters so current lifetime totals and owner stats do not regress. Existing totals predate unique analytics, so the dashboard clearly identifies when unique and funnel collection began.

## API

Add:

- `POST /api/analytics/events` for catalog visits and app view/open/copy events.
- `GET /api/admin/analytics?range=7d|30d|90d|all` for summary, previous-period comparison, daily series, funnel, and per-app aggregates.

Keep:

- `GET /api/admin/stats` as the lightweight moderation-count endpoint used by current admin authorization and badges.
- `POST /api/apps/{slug}/track` during compatibility migration.
- `GET /api/apps/{slug}/stats` for owner and admin per-app daily totals.

The new public event endpoint accepts only the strict event type, optional app slug where required, and opaque visitor/session identifiers. It is rate-limited and returns no visitor information.

The admin response includes the selected range, collection start time, summary metrics, optional previous-period metrics and percentage changes, daily points, funnel stages, and per-app rows.

The OpenAPI source at `docs/openapi.yaml` is updated first, followed by `./scripts/gen-openapi.sh` so `backend/openapi.yaml` and `backend/openapi.json` remain generated copies.

## Frontend Data Flow

Analytics helpers own visitor/session identifier creation and non-blocking event delivery.

- The catalog records one visit after the application reaches the catalog surface.
- The app-detail route records one view after a valid app loads.
- Every existing Nimiq Pay open surface records an open before navigating.
- The desktop share/copy surface records a copy only after the clipboard operation succeeds.
- The backend records wallet login only after signature and claimed-address verification succeeds.

Analytics failures never block navigation, login, opening Nimiq Pay, or copying a link. Delivery uses beacon or keepalive semantics where appropriate. Invalid or rate-limited analytics requests fail independently of the product action.

## Admin Experience

The page follows the existing Nimiq Mini Apps visual system instead of introducing a generic analytics theme.

Summary cards lead with current values and restrained comparison indicators. The chart remains a simple product trend view rather than a general-purpose chart builder. The funnel labels both counts and conversion percentages. The per-app table defaults to activations descending and supports search plus sorting by each useful metric.

Metric definitions explain that a unique visitor is a browser, not guaranteed to be a person, and that activation measures a catalog action rather than in-app engagement.

## Failure Handling

- Tracking calls are fire-and-forget from product interactions.
- Malformed, unknown, or incorrectly scoped events are rejected without creating partial data.
- The ingestion path updates detailed, aggregate, uniqueness, and legacy app counters transactionally where they apply.
- Retention cleanup is idempotent and safe to rerun.
- The admin page distinguishes authorization failures, unavailable analytics, empty periods, and partial pre-launch history.
- A failed comparison or daily-series query does not result in silently misleading zeroes.

## Testing

Backend coverage includes:

- Strict event validation and required app association.
- HMAC hashing without raw identifier persistence.
- Successful-login-only wallet events.
- Catalog and app-view deduplication.
- Deliberate open and copy counting.
- Aggregate, uniqueness, and legacy app-stat updates.
- Fixed range boundaries and previous-period calculations.
- All-time behavior and collection start time.
- Retention aggregation and cleanup.
- Public rate limiting and admin authorization.

Frontend coverage includes:

- Stable visitor and session identifier handling.
- Correct event emission from catalog, detail, open, and copy flows.
- Successful-copy-only tracking.
- Product actions continuing when analytics delivery fails.
- Range selection, comparison rendering, chart data, funnel calculations, table sorting/search, partial-history messaging, and error states.

Verification runs the focused backend and frontend tests first, then the full backend test suite, frontend test suite, production frontend build, OpenAPI generation check, and any repository-wide verification gate.

## Documentation

Update:

- `docs/openapi.yaml` and both generated backend copies.
- `README.md` with the admin analytics surface and tracking behavior.
- `docs/DEV.md` with `ANALYTICS_HASH_SECRET` and local analytics notes.
- `AGENTS.md` with the new admin analytics endpoint and event semantics.

No external analytics service or third-party tracking dependency is introduced.
