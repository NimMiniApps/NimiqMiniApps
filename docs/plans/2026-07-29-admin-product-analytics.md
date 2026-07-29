# Admin Product Analytics Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a privacy-minimal admin analytics page that measures public visits, successful wallet logins, app-detail views, Nimiq Pay opens, successful link copies, unique browsers, unique wallets, and per-app conversion.

**Architecture:** The Vue frontend sends a strict set of catalog/app events with random browser and session identifiers. The Go API HMACs identifiers before an append-only event write, permanent daily rollup, and compact uniqueness upsert; wallet-login events are emitted only by the successful signature-verification path. A separate admin endpoint returns fixed-range summaries, comparisons, daily trends, a funnel, and sortable per-app aggregates to a new `/admin/analytics` page.

**Tech Stack:** Go 1.25, PostgreSQL 17, pgx, Vue 3, TypeScript, Vue Router, Tailwind CSS, Vitest, OpenAPI.

---

## Preconditions

- Use an isolated worktree because `backend/domaincheck.go` and `backend/domaincheck_test.go` have unrelated user changes in the original checkout.
- Invoke `@database-safety-guard` before running migrations or integration tests against PostgreSQL.
- Preserve the current owner-facing `app_stats_daily` behavior and the lightweight `/api/admin/stats` moderation contract.
- Use `docs/openapi.yaml` as the API source of truth and regenerate both embedded copies.
- Stage exact paths in every commit; never stage the unrelated domain-check files.

### Task 1: Add the analytics persistence model

**Files:**

- Create: `backend/migrations/021_product_analytics.sql`
- Create: `backend/analytics.go`
- Create: `backend/analytics_test.go`

**Step 1: Write the failing migration/store tests**

Add integration tests covering:

```go
func TestRecordAnalyticsEventHashesIdentifiers(t *testing.T)
func TestRecordAnalyticsEventRejectsRawWalletInput(t *testing.T)
func TestRecordAnalyticsEventDeduplicatesCatalogVisitPerSession(t *testing.T)
func TestRecordAnalyticsEventDeduplicatesAppViewPerAppSession(t *testing.T)
func TestRecordAnalyticsEventCountsEveryOpenAndCopy(t *testing.T)
func TestRecordAnalyticsEventUpdatesLegacyAppStats(t *testing.T)
```

Use isolated test identifiers and register cleanup before inserts. Verify stored `visitor_hash`, `session_hash`, and `wallet_hash` differ from their raw inputs and that no raw value appears in any analytics column.

**Step 2: Run the focused tests and verify failure**

Run:

```bash
cd backend
go test ./... -run 'TestRecordAnalyticsEvent' -count=1
```

Expected: FAIL because the migration and `recordAnalyticsEvent` do not exist.

**Step 3: Add the migration**

Create:

```sql
CREATE TABLE analytics_events (
    id BIGSERIAL PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    event_type TEXT NOT NULL CHECK (event_type IN (
        'catalog_visit', 'wallet_login', 'app_view', 'app_open', 'app_link_copy'
    )),
    visitor_hash TEXT NOT NULL,
    session_hash TEXT NOT NULL,
    wallet_hash TEXT,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE
);

CREATE INDEX analytics_events_occurred_at_idx ON analytics_events (occurred_at);
CREATE INDEX analytics_events_app_time_idx ON analytics_events (app_id, occurred_at)
    WHERE app_id IS NOT NULL;
CREATE INDEX analytics_events_type_time_idx ON analytics_events (event_type, occurred_at);

CREATE UNIQUE INDEX analytics_events_catalog_session_unique
    ON analytics_events (event_type, session_hash)
    WHERE event_type = 'catalog_visit';
CREATE UNIQUE INDEX analytics_events_app_view_session_unique
    ON analytics_events (event_type, app_id, session_hash)
    WHERE event_type = 'app_view';

CREATE TABLE analytics_daily (
    day DATE NOT NULL,
    event_type TEXT NOT NULL,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    total_events BIGINT NOT NULL DEFAULT 0,
    UNIQUE NULLS NOT DISTINCT (day, event_type, app_id)
);

CREATE TABLE analytics_uniques (
    subject_kind TEXT NOT NULL CHECK (subject_kind IN ('visitor', 'wallet')),
    event_type TEXT NOT NULL,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    subject_hash TEXT NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    UNIQUE NULLS NOT DISTINCT (subject_kind, event_type, app_id, subject_hash)
);

CREATE INDEX analytics_uniques_range_idx
    ON analytics_uniques (event_type, app_id, first_seen_at, last_seen_at);
```

Use the existing migration runner; do not edit prior migrations.

**Step 4: Implement strict types and HMAC helpers**

In `backend/analytics.go`, define:

```go
type analyticsEventType string

const (
    analyticsCatalogVisit analyticsEventType = "catalog_visit"
    analyticsWalletLogin  analyticsEventType = "wallet_login"
    analyticsAppView      analyticsEventType = "app_view"
    analyticsAppOpen      analyticsEventType = "app_open"
    analyticsAppLinkCopy  analyticsEventType = "app_link_copy"
)

type analyticsEventInput struct {
    EventType analyticsEventType
    VisitorID string
    SessionID string
    WalletAddress string
    AppID string
    OccurredAt time.Time
}

func hashAnalyticsIdentifier(secret, kind, value string) string
func validateAnalyticsIdentifier(value string) error
func (s *server) recordAnalyticsEvent(ctx context.Context, input analyticsEventInput) (bool, error)
```

Accept UUID-shaped opaque identifiers with a maximum length of 64. HMAC `kind + "\x00" + value` with SHA-256 and encode as lowercase hex so visitor, session, and wallet namespaces cannot collide.

**Step 5: Implement one transactional write**

`recordAnalyticsEvent` must:

1. Validate the strict event/app rules.
2. Resolve hashes before starting the transaction.
3. Insert the detailed event with `ON CONFLICT DO NOTHING`.
4. Return `recorded=false` for a deduplicated catalog visit/app view.
5. Increment `analytics_daily` only when a detailed event was inserted.
6. Upsert visitor uniqueness for the event and wallet uniqueness only for a server-supplied wallet address.
7. Increment the existing `app_stats_daily.views` for `app_view` and `.opens` for `app_open`, only when the new event was recorded.
8. Commit all applicable writes together.

Do not add arbitrary metadata storage.

**Step 6: Run focused and backend tests**

Run:

```bash
cd backend
go test ./... -run 'TestRecordAnalyticsEvent' -count=1
go test ./... -run 'TestTrackEvent|TestRecordAnalyticsEvent' -count=1
```

Expected: PASS, including existing app-stat tests.

**Step 7: Commit**

```bash
git add backend/migrations/021_product_analytics.sql backend/analytics.go backend/analytics_test.go
git commit -m "feat: add product analytics storage"
```

### Task 2: Add public ingestion and successful-login tracking

**Files:**

- Modify: `backend/analytics.go`
- Modify: `backend/analytics_test.go`
- Modify: `backend/handlers.go`
- Modify: `backend/main.go`
- Modify: `backend/walletauth.go`
- Modify: `backend/walletauth_test.go`

**Step 1: Write failing handler tests**

Add:

```go
func TestTrackAnalyticsEventAcceptsStrictPublicEvents(t *testing.T)
func TestTrackAnalyticsEventRequiresAppForAppEvents(t *testing.T)
func TestTrackAnalyticsEventRejectsWalletLoginFromPublicPayload(t *testing.T)
func TestTrackAnalyticsEventRejectsUnknownEventAndIdentifier(t *testing.T)
func TestTrackAnalyticsEventRateLimitsByClientAndVisitor(t *testing.T)
func TestAuthVerifyRecordsWalletLoginOnlyAfterSuccess(t *testing.T)
func TestAuthVerifyStillSucceedsWhenAnalyticsUnavailable(t *testing.T)
```

Assert that a public payload cannot provide or influence `wallet_hash`, and that invalid signature/address verification creates no login event.

**Step 2: Run tests and verify failure**

```bash
cd backend
go test ./... -run 'TestTrackAnalyticsEvent|TestAuthVerifyRecordsWalletLogin' -count=1
```

Expected: FAIL because the route and auth integration are absent.

**Step 3: Add server configuration**

Extend `server` with:

```go
analyticsHashSecret string
analyticsLimiter    *rateLimiter
```

Read `ANALYTICS_HASH_SECRET` in `main.go`. Log a warning when absent. Initialize a conservative public limiter, for example 120 events per minute per in-memory key.

Register:

```go
mux.HandleFunc("POST /api/analytics/events", s.trackAnalyticsEvent)
```

**Step 4: Implement the public handler**

Accept only:

```json
{
  "event": "catalog_visit",
  "visitor_id": "browser UUID",
  "session_id": "session UUID",
  "app_slug": "required only for app events"
}
```

Return `204` for a recorded or deduplicated valid event, `400` for malformed/invalid input, `429` when limited, and `503` when analytics is not configured. Use `clientIP(r)` only in the in-memory limiter key; never persist it.

Explicitly reject `wallet_login` and any wallet field on this endpoint.

**Step 5: Record login after verification**

Extend the auth-verify request with optional `analytics_visitor_id` and `analytics_session_id`. After signature verification, claimed-address verification, nonce consumption, and session-cookie creation succeed, call `recordAnalyticsEvent` with `analyticsWalletLogin` and the verified `address`.

Analytics failure must be logged but must not change a successful login response.

**Step 6: Run focused and full backend tests**

```bash
cd backend
go test ./... -run 'TestTrackAnalyticsEvent|TestAuthVerify' -count=1
go test ./...
```

Expected: PASS.

**Step 7: Commit**

```bash
git add backend/analytics.go backend/analytics_test.go backend/handlers.go backend/main.go backend/walletauth.go backend/walletauth_test.go
git commit -m "feat: ingest anonymous product events"
```

### Task 3: Add retention and admin analytics queries

**Files:**

- Modify: `backend/analytics.go`
- Modify: `backend/analytics_test.go`
- Modify: `backend/main.go`

**Step 1: Write failing query and retention tests**

Add table-driven tests:

```go
func TestParseAnalyticsRange(t *testing.T)
func TestAdminAnalyticsReturnsRangeAndPreviousPeriod(t *testing.T)
func TestAdminAnalyticsCalculatesUniqueVisitorFunnel(t *testing.T)
func TestAdminAnalyticsCalculatesPerAppConversion(t *testing.T)
func TestAdminAnalyticsAllTimeUsesRollupsAndUniqueness(t *testing.T)
func TestAdminAnalyticsOmitsAllTimeComparison(t *testing.T)
func TestPruneAnalyticsEventsKeepsComparisonWindow(t *testing.T)
```

Use an injected/reference `now` in query helpers so UTC day boundaries and exact preceding windows are deterministic.

**Step 2: Run tests and verify failure**

```bash
cd backend
go test ./... -run 'TestParseAnalyticsRange|TestAdminAnalytics|TestPruneAnalytics' -count=1
```

Expected: FAIL because query/retention functions are missing.

**Step 3: Define the response contract**

Use explicit structs:

```go
type analyticsMetrics struct {
    Visits            int64 `json:"visits"`
    UniqueVisitors    int64 `json:"unique_visitors"`
    WalletLogins      int64 `json:"wallet_logins"`
    UniqueWallets     int64 `json:"unique_wallets"`
    AppViews          int64 `json:"app_views"`
    UniqueAppViewers  int64 `json:"unique_app_viewers"`
    AppOpens          int64 `json:"app_opens"`
    LinkCopies        int64 `json:"link_copies"`
    Activations       int64 `json:"activations"`
    UniqueActivators  int64 `json:"unique_activators"`
}

type adminAnalyticsResponse struct {
    Range              string                 `json:"range"`
    CollectionStarted  *time.Time             `json:"collection_started_at"`
    Current            analyticsMetrics       `json:"current"`
    Previous           *analyticsMetrics      `json:"previous,omitempty"`
    Changes            map[string]*float64    `json:"changes,omitempty"`
    Daily              []analyticsDailyPoint  `json:"daily"`
    Funnel             []analyticsFunnelStage `json:"funnel"`
    Apps               []analyticsAppRow      `json:"apps"`
}
```

Per-app rows contain slug/name plus views, unique viewers, opens, link copies, activations, unique activators, and nullable conversion percentage.

**Step 4: Implement fixed-range queries**

- Accept only `7d`, `30d`, `90d`, and `all`; default to `30d`.
- Use half-open UTC intervals `[start, end)`.
- Query detailed events for current and prior fixed windows.
- Build funnel stages from distinct visitor hashes: visited, viewed any app, activated any app.
- Calculate app conversion as distinct activators divided by distinct viewers.
- Return `nil` conversion/change when the denominator is zero.
- Fill missing daily dates with zeroes.

For `all`, use `analytics_daily` for totals and `analytics_uniques` for exact ever-seen distinct counts; omit `previous` and `changes`.

**Step 5: Add retention**

Implement:

```go
func pruneAnalyticsEvents(ctx context.Context, pool *pgxpool.Pool, now time.Time) error
func (s *server) startAnalyticsRetentionWorker(ctx context.Context)
```

Delete only detailed events older than 180 days. Run once at startup and then daily. Never delete permanent daily or uniqueness rows.

**Step 6: Register the admin route**

```go
mux.HandleFunc("GET /api/admin/analytics", s.adminAuth(s.adminAnalytics))
```

Leave `/api/admin/stats` unchanged.

**Step 7: Run focused and full backend tests**

```bash
cd backend
go test ./... -run 'TestParseAnalyticsRange|TestAdminAnalytics|TestPruneAnalytics' -count=1
go test ./...
go build ./...
```

Expected: PASS.

**Step 8: Commit**

```bash
git add backend/analytics.go backend/analytics_test.go backend/main.go
git commit -m "feat: expose admin product analytics"
```

### Task 4: Document and generate the API contract

**Files:**

- Modify: `docs/openapi.yaml`
- Modify: `backend/openapi.yaml` (generated)
- Modify: `backend/openapi.json` (generated)

**Step 1: Add the OpenAPI paths and schemas**

Document:

- `POST /api/analytics/events`, including the strict enum, identifier limits, conditional `app_slug`, and `204`/`400`/`429`/`503`.
- Optional analytics identifiers on `POST /api/auth/verify`.
- `GET /api/admin/analytics` with the fixed range enum and full response schemas.
- Aggregate-only privacy semantics and “since collection began” behavior.

Do not change the existing `/api/admin/stats` or app-owner stats shapes.

**Step 2: Regenerate embedded copies**

```bash
./scripts/gen-openapi.sh
git diff --check -- docs/openapi.yaml backend/openapi.yaml backend/openapi.json
```

Expected: generation succeeds and the three contracts are synchronized.

**Step 3: Run backend verification**

```bash
cd backend
go test ./...
```

Expected: PASS.

**Step 4: Commit**

```bash
git add docs/openapi.yaml backend/openapi.yaml backend/openapi.json
git commit -m "docs: define product analytics API"
```

### Task 5: Add the frontend analytics client and product hooks

**Files:**

- Create: `frontend/src/utils/analytics.ts`
- Create: `frontend/src/utils/analytics.test.ts`
- Modify: `frontend/src/api.ts`
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/composables/useWalletAuth.ts`
- Modify: `frontend/src/components/AppCard.vue`
- Modify: `frontend/src/components/OpenInWalletPanel.vue`
- Modify: `frontend/src/components/ShareButton.vue`
- Modify: `frontend/src/views/AppDetailView.vue`

**Step 1: Write failing utility tests**

Cover:

```ts
describe('analytics identifiers', () => {
  it('reuses the visitor id across sessions')
  it('reuses the session id within one session only')
})

describe('analytics delivery', () => {
  it('sends only the strict event payload')
  it('records a catalog visit once per session')
  it('does not throw when sendBeacon or fetch fails')
})
```

Also add source-contract assertions that copy tracking occurs only after `clipboard.writeText` resolves and that existing open actions still navigate when delivery fails.

**Step 2: Run tests and verify failure**

```bash
cd frontend
npm test -- src/utils/analytics.test.ts
```

Expected: FAIL because the analytics utility is missing.

**Step 3: Implement identifier and delivery helpers**

Export:

```ts
export type ProductAnalyticsEvent =
  | 'catalog_visit'
  | 'app_view'
  | 'app_open'
  | 'app_link_copy'

export function getAnalyticsVisitorId(): string
export function getAnalyticsSessionId(): string
export function analyticsIdentity(): {
  analytics_visitor_id: string
  analytics_session_id: string
}
export function trackProductEvent(event: ProductAnalyticsEvent, appSlug?: string): void
export function trackCatalogVisit(): void
```

Use `crypto.randomUUID()`, `localStorage`, and `sessionStorage`. Send only `event`, `visitor_id`, `session_id`, and optional `app_slug`. Prefer `navigator.sendBeacon`; fall back to `fetch(..., {method: 'POST', keepalive: true})`. Swallow delivery failures.

**Step 4: Add API types and login identifiers**

Update `authVerify` so the frontend includes `analyticsIdentity()` in the successful verification request. Add typed `adminAnalytics(range)` later without duplicating the base request helper.

**Step 5: Wire product events**

- In `router.afterEach`, call `trackCatalogVisit()` for non-admin routes; the helper deduplicates once per session.
- Replace current `trackAppEvent(..., 'view'|'open')` calls with `trackProductEvent`.
- Record `app_view` after a valid app loads.
- Record `app_open` for every deliberate Nimiq Pay open.
- In `OpenInWalletPanel.vue`, move tracking after successful `clipboard.writeText` and emit `app_link_copy`.
- Give `ShareButton.vue` an optional `appSlug`; emit `app_link_copy` only after successful clipboard copy. Pass the detail app slug from `AppDetailView.vue`.
- Never await analytics before navigation/open actions.

**Step 6: Run focused tests and build**

```bash
cd frontend
npm test -- src/utils/analytics.test.ts
npm run build
```

Expected: PASS.

**Step 7: Commit**

```bash
git add frontend/src/utils/analytics.ts frontend/src/utils/analytics.test.ts frontend/src/api.ts frontend/src/main.ts frontend/src/composables/useWalletAuth.ts frontend/src/components/AppCard.vue frontend/src/components/OpenInWalletPanel.vue frontend/src/components/ShareButton.vue frontend/src/views/AppDetailView.vue
git commit -m "feat: track catalog product funnel"
```

### Task 6: Build the admin analytics page

**Files:**

- Create: `frontend/src/views/AdminAnalyticsView.vue`
- Create: `frontend/src/components/AnalyticsTrendChart.vue`
- Create: `frontend/src/utils/adminAnalytics.ts`
- Create: `frontend/src/utils/adminAnalytics.test.ts`
- Modify: `frontend/src/api.ts`
- Modify: `frontend/src/main.ts`
- Modify: `frontend/src/views/AdminView.vue`
- Modify: `frontend/src/i18n/messages.ts`

**Step 1: Write failing presentation-logic tests**

Test pure helpers for:

```ts
formatAnalyticsChange(current, previous)
buildTrendPoints(daily, metric, width, height)
sortAnalyticsApps(rows, key, ascending)
filterAnalyticsApps(rows, query)
```

Cover zero denominators, missing comparisons, flat/empty charts, stable tie-breaking by app name, and case-insensitive search.

**Step 2: Run tests and verify failure**

```bash
cd frontend
npm test -- src/utils/adminAnalytics.test.ts
```

Expected: FAIL because the helpers do not exist.

**Step 3: Add typed API support**

Define `AdminAnalyticsResponse`, metric, daily-point, funnel-stage, and app-row interfaces in `frontend/src/api.ts`, then add:

```ts
export const adminAnalytics = (range: '7d' | '30d' | '90d' | 'all') =>
  adminRequest<AdminAnalyticsResponse>(`/api/admin/analytics?range=${range}`)
```

**Step 4: Implement pure helpers and the SVG trend chart**

Use a dependency-free responsive SVG consistent with `StatsSparkline.vue`. Provide an accessible text summary/table fallback or `aria-label`; do not rely on color alone.

**Step 5: Implement `/admin/analytics`**

Build:

- Admin-auth loading and unauthorized states using `useAdminAuth`.
- Range tabs for 7/30/90/all.
- Summary cards with restrained prior-period deltas.
- Metric-selectable daily trend.
- Three-stage unique-visitor funnel with counts and rates.
- Searchable/sortable per-app table, defaulting to activations descending.
- Collection-start and partial-history notices.
- Loading skeleton, empty period, API error, and retry states.
- Mobile stacked app rows while retaining a readable desktop table.

Follow the existing board visual language and use current CSS tokens. Invoke `@nimiq-ui-kit` during implementation for local design-system guidance.

**Step 6: Add navigation**

Register:

```ts
{ path: '/admin/analytics', component: () => import('./views/AdminAnalyticsView.vue'), meta: { title: 'Admin Analytics' } }
```

Place it before or alongside `/admin`. Add an **Analytics** link near the top of `AdminView.vue` and a **Back to moderation** link in analytics.

**Step 7: Run tests and build**

```bash
cd frontend
npm test -- src/utils/adminAnalytics.test.ts src/utils/analytics.test.ts
npm test
npm run build
```

Expected: PASS.

**Step 8: Commit**

```bash
git add frontend/src/views/AdminAnalyticsView.vue frontend/src/components/AnalyticsTrendChart.vue frontend/src/utils/adminAnalytics.ts frontend/src/utils/adminAnalytics.test.ts frontend/src/api.ts frontend/src/main.ts frontend/src/views/AdminView.vue frontend/src/i18n/messages.ts
git commit -m "feat: add admin product analytics page"
```

### Task 7: Add configuration and user-facing documentation

**Files:**

- Modify: `docker-compose.yml`
- Modify: `docker-stack.yml`
- Modify: `README.md`
- Modify: `docs/DEV.md`
- Modify: `AGENTS.md`

**Step 1: Add the stable hash secret configuration**

- Local Compose may use an explicit development-only fallback.
- Swarm configuration must require `ANALYTICS_HASH_SECRET` and must not silently generate or rotate it.
- Document that changing the secret breaks continuity of unique counts.

Do not expose the secret through the frontend.

**Step 2: Update concise project documentation**

- README: add the admin analytics route and aggregate tracking behavior.
- `docs/DEV.md`: add the environment variable, local testing instructions, retention, and privacy notes.
- `AGENTS.md`: add public event/admin analytics endpoints and precise definitions for view/open/copy.

**Step 3: Run configuration checks**

```bash
docker compose config >/dev/null
git diff --check -- docker-compose.yml docker-stack.yml README.md docs/DEV.md AGENTS.md
```

Expected: Compose configuration parses and diffs have no whitespace errors.

**Step 4: Commit**

```bash
git add docker-compose.yml docker-stack.yml README.md docs/DEV.md AGENTS.md
git commit -m "docs: configure product analytics"
```

### Task 8: Run full verification and review

**Files:**

- Modify only files required by verified failures attributable to this feature.

**Step 1: Verify generated OpenAPI is clean**

```bash
./scripts/gen-openapi.sh
git diff --exit-code -- docs/openapi.yaml backend/openapi.yaml backend/openapi.json
```

Expected: no generated drift.

**Step 2: Run backend gates**

After invoking `@database-safety-guard`:

```bash
cd backend
go test ./...
go build ./...
```

Expected: PASS.

**Step 3: Run frontend gates**

```bash
cd frontend
npm test
npm run build
```

Expected: PASS.

**Step 4: Review privacy and behavior manually**

Confirm:

- No raw visitor/session/wallet identifiers are stored.
- No IP, user-agent, referrer, or arbitrary metadata columns exist.
- Failed analytics calls do not block login, open, navigation, or copy.
- Wallet-login events are created only by successful verification.
- All-time omits comparison.
- Unique/funnel history clearly starts at feature launch.
- `/admin` moderation behavior and owner app stats are unchanged.

**Step 5: Inspect the final diff and commits**

```bash
git status --short
git diff --check HEAD~7..HEAD
git log --oneline --decorate -10
```

Expected: only planned feature changes; the original checkout's unrelated domain-check changes are absent from this worktree/branch.

**Step 6: Request code review**

Invoke `@superpowers:requesting-code-review` and resolve confirmed findings using `@superpowers:receiving-code-review`.

**Step 7: Final verification**

Invoke `@superpowers:verification-before-completion`, rerun any affected focused test plus the full gates above, and record exact results. Do not push or deploy unless the user separately requests it.
