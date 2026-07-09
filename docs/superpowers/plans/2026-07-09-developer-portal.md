# Developer Portal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a developer log in with their wallet, see the apps they own, and submit edits to them directly — replacing "file an issue and wait for manual admin work."

**Architecture:** Add one nullable `developer_wallet_address` column to `apps`. Reuse the existing wallet-auth cookie (`walletAuthMiddleware`) and admin-wallet-allowlist auth (`s.adminAuth`) — no new auth mechanism. Reuse the existing `app_revisions` pending/approve/reject pipeline for edits — only add an ownership check in front of it. Add two small read endpoints (`GET /api/my/apps`, `GET /api/admin/users`) and wire up the frontend forms that already mostly exist (`SubmitView.vue`, `RequestUpdateView.vue`).

**Tech Stack:** Go 1.x + `pgx/v5` (backend), Vue 3 + `<script setup>` + TypeScript (frontend), Postgres migrations (embedded, auto-applied on boot via `migrate()` in `backend/main.go`).

## Global Constraints

- No new auth mechanism: wallet-owned actions use `walletAuthMiddleware`; admin actions use `s.adminAuth` — both already exist and are unchanged by this plan.
- `developer_slug` is assigned once, at first submission, from the derived value of the user's `display_name` — it is never re-derived on later display-name changes (would break existing app URLs/links keyed on it).
- `request-update` remains owner-only; the server ignores any `developer_name`/`developer_slug` in that endpoint's request body and always carries forward the app's current values — developer identity changes are admin-only.
- Reuse the existing 5-requests-per-hour-per-IP rate limit (`allowSubmit`/`submitLimit`/`submitWindow` in `backend/submit.go`) for `submitApp` — do not introduce a second limiter.
- Existing apps get `developer_wallet_address = NULL` (unclaimed); backfilling ownership for them is a manual admin action outside this plan (claim flow and "report a problem" are explicitly out of scope — see the spec).
- `docs/openapi.yaml` is the single source of truth; `backend/openapi.yaml`/`backend/openapi.json` are generated via `./scripts/gen-openapi.sh` and CI fails if they drift — never hand-edit the generated copies.
- This repo has no Postgres integration-test harness — existing backend tests (`profile_test.go`, `walletauth_test.go`, `submit_test.go`) are pure unit tests only. Follow that pattern: unit-test pure functions, verify DB-touching handlers manually with `curl` against a local dev stack (see `docs/DEV.md`).

Spec: [`docs/superpowers/specs/2026-07-09-developer-portal-design.md`](../specs/2026-07-09-developer-portal-design.md)

---

### Task 1: Migration — add `developer_wallet_address` to `apps`

**Files:**
- Create: `backend/migrations/013_developer_wallet.sql`

**Interfaces:**
- Produces: column `apps.developer_wallet_address TEXT REFERENCES users(wallet_address)`, nullable, indexed. Every later task that reads/writes this column depends on this migration having run.

- [ ] **Step 1: Write the migration**

```sql
ALTER TABLE apps ADD COLUMN IF NOT EXISTS developer_wallet_address TEXT REFERENCES users(wallet_address);

CREATE INDEX IF NOT EXISTS apps_developer_wallet_address_idx
    ON apps (developer_wallet_address)
    WHERE developer_wallet_address IS NOT NULL;
```

- [ ] **Step 2: Apply it against local dev Postgres and confirm it's tracked**

Run (from `backend/`, with local Postgres up per `docs/DEV.md`):
```bash
go run . &
sleep 2
curl -s localhost:8080/health
kill %1
```
Expected: `/health` returns `200`, and server logs include `"applied migration" name=013_developer_wallet.sql`. Then confirm the column exists:
```bash
psql "$DATABASE_URL" -c "\d apps" | grep developer_wallet_address
```
Expected: a row showing `developer_wallet_address | text`.

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/013_developer_wallet.sql
git commit -m "Add developer_wallet_address column to apps"
```

---

### Task 2: Wire `developer_wallet_address` through the `App` model

**Files:**
- Modify: `backend/handlers.go` (`App` struct, `appColumns`, `scanApp`, `decodeAndInsert`'s INSERT, `updateApp`'s UPDATE)

**Interfaces:**
- Consumes: column from Task 1.
- Produces: `App.DeveloperWalletAddress *string` (JSON: `developer_wallet_address`), included in every `App` read/write path. Later tasks (3, 6, 7, 8) read/write this field.

- [ ] **Step 1: Add the field to the `App` struct**

In `backend/handlers.go`, in the `App` struct (starts at line 27), add after `DeveloperName`:

```go
	DeveloperSlug   string      `json:"developer_slug"`
	DeveloperName   string      `json:"developer_name"`
	DeveloperWalletAddress *string `json:"developer_wallet_address"`
```

- [ ] **Step 2: Add the column to `appColumns` and `scanApp`**

Change:
```go
const appColumns = `id, slug, name, domain, category, developer_slug, developer_name, tagline,
	description, long_description, tags, assets, status, release_stage, featured, featured_order,
	website_url, github_url, icon_url, discovered_icon_url, banner_url, media, socials, domain_reachable, domain_checked_at,
	submitter_contact, created_at, updated_at`
```
to:
```go
const appColumns = `id, slug, name, domain, category, developer_slug, developer_name, tagline,
	description, long_description, tags, assets, status, release_stage, featured, featured_order,
	website_url, github_url, icon_url, discovered_icon_url, banner_url, media, socials, domain_reachable, domain_checked_at,
	submitter_contact, created_at, updated_at, developer_wallet_address`
```

In `scanApp`, change the `Scan(...)` call to add `&a.DeveloperWalletAddress` as the last argument (matching the column appended last above):
```go
	err := row.Scan(&a.ID, &a.Slug, &a.Name, &a.Domain, &a.Category, &a.DeveloperSlug,
		&a.DeveloperName, &a.Tagline, &a.Description, &a.LongDescription, &a.Tags, &a.Assets, &a.Status,
		&a.ReleaseStage, &a.Featured, &a.FeaturedOrder, &a.WebsiteURL, &a.GithubURL, &a.IconURL, &a.DiscoveredIconURL, &a.BannerURL,
		&mediaJSON, &socialsJSON, &a.DomainReachable, &a.DomainCheckedAt, &a.SubmitterContact, &a.CreatedAt, &a.UpdatedAt,
		&a.DeveloperWalletAddress)
```

- [ ] **Step 3: Add the column to the submit-path INSERT in `decodeAndInsert`**

Change the INSERT in `decodeAndInsert` (around line 465) from:
```go
	a, err = scanApp(s.pool.QueryRow(r.Context(), `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline,
			description, long_description, tags, assets, status, release_stage, featured, featured_order,
			website_url, github_url, icon_url, banner_url, media, socials, submitter_contact)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING `+appColumns,
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.LongDescription, a.Tags, a.Assets, a.Status, a.ReleaseStage, a.Featured, a.FeaturedOrder,
		a.WebsiteURL, a.GithubURL, a.IconURL, a.BannerURL, mediaJSON, socialsJSON, a.SubmitterContact))
```
to:
```go
	a, err = scanApp(s.pool.QueryRow(r.Context(), `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline,
			description, long_description, tags, assets, status, release_stage, featured, featured_order,
			website_url, github_url, icon_url, banner_url, media, socials, submitter_contact, developer_wallet_address)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
		RETURNING `+appColumns,
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.LongDescription, a.Tags, a.Assets, a.Status, a.ReleaseStage, a.Featured, a.FeaturedOrder,
		a.WebsiteURL, a.GithubURL, a.IconURL, a.BannerURL, mediaJSON, socialsJSON, a.SubmitterContact, a.DeveloperWalletAddress))
```

- [ ] **Step 4: Add the column to the admin UPDATE in `updateApp`**

Change the UPDATE in `updateApp` (around line 537) from:
```go
	a, err = scanApp(s.pool.QueryRow(r.Context(), `
		UPDATE apps SET slug=$1, name=$2, domain=$3, category=$4, developer_slug=$5,
			developer_name=$6, tagline=$7, description=$8, long_description=$9, tags=$10, assets=$11,
			status=$12, release_stage=$13, featured=$14, featured_order=$15, website_url=$16, github_url=$17,
			icon_url=$18, banner_url=$19, media=$20, socials=$21, submitter_contact=$22, updated_at=now()
		WHERE id=$23
		RETURNING `+appColumns,
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.LongDescription, a.Tags, a.Assets, a.Status, a.ReleaseStage, a.Featured, a.FeaturedOrder,
		a.WebsiteURL, a.GithubURL, a.IconURL, a.BannerURL, mediaJSON, socialsJSON, a.SubmitterContact, a.ID))
```
to:
```go
	a, err = scanApp(s.pool.QueryRow(r.Context(), `
		UPDATE apps SET slug=$1, name=$2, domain=$3, category=$4, developer_slug=$5,
			developer_name=$6, tagline=$7, description=$8, long_description=$9, tags=$10, assets=$11,
			status=$12, release_stage=$13, featured=$14, featured_order=$15, website_url=$16, github_url=$17,
			icon_url=$18, banner_url=$19, media=$20, socials=$21, submitter_contact=$22, developer_wallet_address=$23, updated_at=now()
		WHERE id=$24
		RETURNING `+appColumns,
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.LongDescription, a.Tags, a.Assets, a.Status, a.ReleaseStage, a.Featured, a.FeaturedOrder,
		a.WebsiteURL, a.GithubURL, a.IconURL, a.BannerURL, mediaJSON, socialsJSON, a.SubmitterContact, a.DeveloperWalletAddress, a.ID))
```

This is enough for admins to assign/reassign ownership through the existing `PUT`/`PATCH /api/admin/apps/{slug}` — the handler already decodes the full JSON body into `a` before this UPDATE runs, so `developer_wallet_address` just needs to be accepted in the schema (done) and persisted (this step).

- [ ] **Step 5: Build and run existing tests to confirm nothing broke**

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```
Expected: builds clean, all existing tests still pass (no test currently touches `App` field count, so this is a compile-correctness check).

- [ ] **Step 6: Format and commit**

```bash
cd backend && gofmt -w handlers.go
git add backend/handlers.go
git commit -m "Add developer_wallet_address to the App model"
```

---

### Task 3: Developer slug derivation

**Files:**
- Create: `backend/developer.go`
- Test: `backend/developer_test.go`

**Interfaces:**
- Consumes: `s.pool` (`*pgxpool.Pool`), `App.DeveloperSlug`/`DeveloperWalletAddress` from Task 2.
- Produces: `slugifyDisplayName(name string) string` (pure), `(s *server) resolveDeveloperSlug(ctx context.Context, address, displayName string) (string, error)`. Task 4 (submit) calls `resolveDeveloperSlug`.

- [ ] **Step 1: Write the failing test for the pure slugify function**

```go
package main

import "testing"

func TestSlugifyDisplayName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Satoshi", "satoshi"},
		{"two words", "Satoshi Nakamoto", "satoshi-nakamoto"},
		{"extra whitespace and punctuation", "  Multi   Space! ", "multi-space"},
		{"mixed case with numbers", "Team42 Studio", "team42-studio"},
		{"all symbols", "💯💯", ""},
		{"leading/trailing separators collapse", "-Foo-Bar-", "foo-bar"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := slugifyDisplayName(tc.input)
			if got != tc.want {
				t.Fatalf("slugifyDisplayName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run it and confirm it fails to compile (function doesn't exist yet)**

```bash
cd backend && go test ./... -run TestSlugifyDisplayName -v
```
Expected: `FAIL` — `undefined: slugifyDisplayName`.

- [ ] **Step 3: Implement `backend/developer.go`**

```go
package main

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// slugifyDisplayName mirrors the app-slug generation already used on the submit
// form client-side: lowercase, collapse non-alphanumeric runs into single hyphens,
// trim leading/trailing hyphens. May return "" for a name with no ASCII letters/digits.
func slugifyDisplayName(name string) string {
	lower := strings.ToLower(name)
	var b strings.Builder
	prevSep := true // suppress a leading hyphen
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			prevSep = false
		case !prevSep:
			b.WriteByte('-')
			prevSep = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// resolveDeveloperSlug returns the developer_slug a wallet should submit under.
// A wallet that already owns an app reuses that app's developer_slug (identity is
// assigned once, at first submission — see the developer portal spec). Otherwise it
// derives one from displayName, appending -2, -3, ... on collision with a different
// wallet's developer_slug.
func (s *server) resolveDeveloperSlug(ctx context.Context, address, displayName string) (string, error) {
	var existing string
	err := s.pool.QueryRow(ctx,
		`SELECT developer_slug FROM apps WHERE developer_wallet_address=$1 LIMIT 1`, address).
		Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	base := slugifyDisplayName(displayName)
	if base == "" {
		return "", errors.New("display name must contain at least one letter or number")
	}
	slug := base
	for i := 2; ; i++ {
		var count int
		if err := s.pool.QueryRow(ctx,
			`SELECT count(*) FROM apps WHERE developer_slug=$1 AND developer_wallet_address IS DISTINCT FROM $2`,
			slug, address).Scan(&count); err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		slug = base + "-" + strconv.Itoa(i)
	}
}
```

- [ ] **Step 4: Run the test again to confirm it passes**

```bash
cd backend && go test ./... -run TestSlugifyDisplayName -v
```
Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
cd backend && gofmt -w developer.go developer_test.go
git add backend/developer.go backend/developer_test.go
git commit -m "Add developer slug derivation for wallet-owned submissions"
```

---

### Task 4: Gate `submitApp` behind wallet auth and derive developer identity

**Files:**
- Modify: `backend/submit.go`
- Modify: `backend/main.go` (route line 174)

**Interfaces:**
- Consumes: `walletAuthMiddleware` (existing, `backend/walletauth.go:172`), `s.resolveDeveloperSlug` (Task 3), `s.decodeAndInsert` (existing, `backend/handlers.go`).
- Produces: `submitApp` now has signature `func(w http.ResponseWriter, r *http.Request, address string)` — Task 9/frontend depend on the route requiring the wallet cookie.

- [ ] **Step 1: Change `submitApp`'s signature and add identity resolution**

In `backend/submit.go`, replace:
```go
func (s *server) submitApp(w http.ResponseWriter, r *http.Request) {
	if !allowSubmit(clientIP(r), time.Now()) {
		writeError(w, http.StatusTooManyRequests, "too many submissions, try again later")
		return
	}
	s.decodeAndInsert(w, r, func(a *App) {
		a.Status = "submitted"
		a.Featured = false
	}, true)
}
```
with:
```go
func (s *server) submitApp(w http.ResponseWriter, r *http.Request, address string) {
	if !allowSubmit(clientIP(r), time.Now()) {
		writeError(w, http.StatusTooManyRequests, "too many submissions, try again later")
		return
	}

	var displayName *string
	if err := s.pool.QueryRow(r.Context(),
		`SELECT display_name FROM users WHERE wallet_address=$1`, address).
		Scan(&displayName); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if displayName == nil || strings.TrimSpace(*displayName) == "" {
		writeError(w, http.StatusBadRequest, "set a display name on your profile before submitting an app")
		return
	}

	devSlug, err := s.resolveDeveloperSlug(r.Context(), address, *displayName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.decodeAndInsert(w, r, func(a *App) {
		a.Status = "submitted"
		a.Featured = false
		a.DeveloperWalletAddress = &address
		a.DeveloperSlug = devSlug
		a.DeveloperName = *displayName
	}, true)
}
```

Add the new imports to `backend/submit.go`'s import block (`errors`, `strings`, and `github.com/jackc/pgx/v5` for `pgx.ErrNoRows`):
```go
import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)
```

- [ ] **Step 2: Wrap the route in `walletAuthMiddleware`**

In `backend/main.go`, change:
```go
	mux.HandleFunc("POST /api/apps/submit", s.submitApp)
```
to:
```go
	mux.HandleFunc("POST /api/apps/submit", walletAuthMiddleware(walletAuthSecret, s.submitApp))
```

- [ ] **Step 3: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: clean build.

- [ ] **Step 4: Manually verify against local dev stack**

Per `docs/DEV.md`, start the backend + Postgres locally, then:
```bash
# without a wallet cookie — should be rejected
curl -i -X POST localhost:8080/api/apps/submit -H "Content-Type: application/json" -d '{}'
```
Expected: `401` `{"error":"wallet login required"}`.

Then complete a wallet login via the frontend at `/`, save a display name at `/profile`, and submit a test app via `/submit` in the browser (see Task 12). Confirm the created row has `developer_wallet_address` set and `developer_slug` derived from the display name:
```bash
psql "$DATABASE_URL" -c "SELECT slug, developer_slug, developer_name, developer_wallet_address FROM apps ORDER BY created_at DESC LIMIT 1;"
```

- [ ] **Step 5: Commit**

```bash
cd backend && gofmt -w submit.go main.go
git add backend/submit.go backend/main.go
git commit -m "Require wallet login to submit an app; derive developer identity from profile"
```

---

### Task 5: Owner-only `request-update`, preserving developer identity

**Files:**
- Modify: `backend/revisions.go`
- Modify: `backend/main.go` (route line 175)

**Interfaces:**
- Consumes: `walletAuthMiddleware`, `App.DeveloperWalletAddress` (Task 2).
- Produces: `requestAppUpdate` signature becomes `func(w http.ResponseWriter, r *http.Request, address string)`.

- [ ] **Step 1: Add the ownership check and signature change**

In `backend/revisions.go`, change the function signature from:
```go
func (s *server) requestAppUpdate(w http.ResponseWriter, r *http.Request) {
```
to:
```go
func (s *server) requestAppUpdate(w http.ResponseWriter, r *http.Request, address string) {
```

Immediately after the existing block that loads `current` and checks `isPublicStatus` (the block ending `writeError(w, http.StatusNotFound, "app not found"); return }`), insert the ownership check:
```go
	if current.DeveloperWalletAddress == nil || *current.DeveloperWalletAddress != address {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}
```

- [ ] **Step 2: Stop trusting client-supplied developer identity in the revision insert**

In the `INSERT INTO app_revisions` call, change the two `body.DeveloperSlug, body.DeveloperName` arguments to `current.DeveloperSlug, current.DeveloperName`:
```go
	rev, err := scanRevision(s.pool.QueryRow(r.Context(), `
		INSERT INTO app_revisions (
			app_slug, name, domain, category, developer_slug, developer_name, tagline,
			description, long_description, tags, assets, release_stage,
			website_url, github_url, icon_url, banner_url, media, socials, author_note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING `+revisionColumns,
		slug, body.Name, body.Domain, body.Category, current.DeveloperSlug, current.DeveloperName,
		body.Tagline, body.Description, body.LongDescription, body.Tags, body.Assets, body.ReleaseStage,
		body.WebsiteURL, body.GithubURL, body.IconURL, body.BannerURL, mediaJSON, socialsJSON, body.AuthorNote))
```

(`revisionToApp`/`approveRevision` are unchanged — they already just copy whatever landed in the `app_revisions` row onto the app, which is now always the preserved identity.)

- [ ] **Step 3: Wrap the route in `walletAuthMiddleware`**

In `backend/main.go`, change:
```go
	mux.HandleFunc("POST /api/apps/{slug}/request-update", s.requestAppUpdate)
```
to:
```go
	mux.HandleFunc("POST /api/apps/{slug}/request-update", walletAuthMiddleware(walletAuthSecret, s.requestAppUpdate))
```

- [ ] **Step 4: Build**

```bash
cd backend && go build ./... && go vet ./...
```
Expected: clean build.

- [ ] **Step 5: Manually verify**

```bash
# no cookie
curl -i -X POST localhost:8080/api/apps/some-slug/request-update -H "Content-Type: application/json" -d '{}'
```
Expected: `401`.

With a logged-in wallet cookie that does **not** own `some-slug` (or where the app is unowned/`NULL`):
Expected: `403` `{"error":"you don't own this app"}`.

With the owning wallet's cookie, submit an update via `/apps/{slug}/update` in the browser (Task 13) with a changed `tagline`; confirm `developer_name`/`developer_slug` in the resulting `app_revisions` row still match the app's current values even if the request tried to change them:
```bash
psql "$DATABASE_URL" -c "SELECT app_slug, developer_slug, developer_name, tagline, status FROM app_revisions ORDER BY created_at DESC LIMIT 1;"
```

- [ ] **Step 6: Commit**

```bash
cd backend && gofmt -w revisions.go main.go
git add backend/revisions.go backend/main.go
git commit -m "Restrict request-update to the owning wallet; lock developer identity"
```

---

### Task 6: `GET /api/my/apps`

**Files:**
- Modify: `backend/developer.go` (add handler)
- Modify: `backend/main.go` (add route)

**Interfaces:**
- Consumes: `App`, `appColumns`, `scanApp` (Task 2), `s.hasPendingRevision` (existing, `backend/revisions.go:323`).
- Produces: `(s *server) myApps(w http.ResponseWriter, r *http.Request, address string)`. Task 14 (frontend `MyAppsView`) depends on this route and response shape.

- [ ] **Step 1: Add the handler to `backend/developer.go`**

```go
func (s *server) myApps(w http.ResponseWriter, r *http.Request, address string) {
	rows, err := s.pool.Query(r.Context(),
		"SELECT "+appColumns+" FROM apps WHERE developer_wallet_address=$1 ORDER BY created_at DESC", address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type myApp struct {
		App
		HasPendingRevision bool `json:"has_pending_revision"`
	}
	items := []myApp{}
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		pending, err := s.hasPendingRevision(r.Context(), a.Slug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, myApp{App: a, HasPendingRevision: pending})
	}
	writeJSON(w, http.StatusOK, items)
}
```

Add `"net/http"` to `backend/developer.go`'s imports if not already present (it is, from Task 3's `context`/`errors`/`strconv`/`strings` list — add `"net/http"` alongside them).

- [ ] **Step 2: Add the route**

In `backend/main.go`, add near the other wallet-auth routes (after the reviews routes, line 184):
```go
	mux.HandleFunc("GET /api/my/apps", walletAuthMiddleware(walletAuthSecret, s.myApps))
```

- [ ] **Step 3: Build**

```bash
cd backend && go build ./... && go vet ./...
```

- [ ] **Step 4: Manually verify**

```bash
curl -i localhost:8080/api/my/apps
```
Expected: `401` (no cookie). With a logged-in wallet cookie that owns at least one app (from Task 4's manual test), expect `200` with a JSON array containing that app plus `"has_pending_revision"`.

- [ ] **Step 5: Commit**

```bash
cd backend && gofmt -w developer.go main.go
git add backend/developer.go backend/main.go
git commit -m "Add GET /api/my/apps for the developer portal"
```

---

### Task 7: `GET /api/admin/users` search

**Files:**
- Modify: `backend/developer.go` (add handler)
- Modify: `backend/main.go` (add route)

**Interfaces:**
- Consumes: `users` table (`wallet_address`, `display_name`), `s.adminAuth` (existing, `backend/adminauth.go:51`).
- Produces: `(s *server) adminSearchUsers(w http.ResponseWriter, r *http.Request)`. Task 15 (admin developer picker) depends on this route and response shape.

- [ ] **Step 1: Add the handler to `backend/developer.go`**

```go
func (s *server) adminSearchUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, []struct{}{})
		return
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT wallet_address, display_name FROM users
		WHERE display_name ILIKE $1 OR wallet_address ILIKE $1
		ORDER BY display_name ASC NULLS LAST LIMIT 20`, q+"%")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type userResult struct {
		WalletAddress string  `json:"wallet_address"`
		DisplayName   *string `json:"display_name"`
	}
	items := []userResult{}
	for rows.Next() {
		var it userResult
		if err := rows.Scan(&it.WalletAddress, &it.DisplayName); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, items)
}
```

- [ ] **Step 2: Add the route**

In `backend/main.go`, add alongside the other admin routes (after line 189, `GET /api/admin/apps`):
```go
	mux.HandleFunc("GET /api/admin/users", s.adminAuth(s.adminSearchUsers))
```

- [ ] **Step 3: Build**

```bash
cd backend && go build ./... && go vet ./...
```

- [ ] **Step 4: Manually verify**

```bash
curl -i localhost:8080/api/admin/users?q=sat
```
Expected: `401` without admin auth. With `Authorization: Bearer $ADMIN_TOKEN` (or an admin wallet cookie), expect `200` and a JSON array of matching users (empty array if none, and empty array immediately for `q=`).

- [ ] **Step 5: Commit**

```bash
cd backend && gofmt -w developer.go main.go
git add backend/developer.go backend/main.go
git commit -m "Add GET /api/admin/users search for the admin developer picker"
```

---

### Task 8: OpenAPI spec

**Files:**
- Modify: `docs/openapi.yaml`
- Generated (via script, not by hand): `backend/openapi.yaml`, `backend/openapi.json`

**Interfaces:**
- Consumes: routes/behavior from Tasks 4–7.
- Produces: accurate, CI-validated API docs. `openapi_test.go` (existing) fails if generated files drift from `docs/openapi.yaml`.

- [ ] **Step 1: Add a `walletCookie` security scheme**

In `docs/openapi.yaml`, under `components.securitySchemes` (currently only `adminBearer`), add:
```yaml
    walletCookie:
      type: apiKey
      in: cookie
      name: wallet_session
      description: "Signed wallet-session cookie set by POST /api/auth/verify"
```

- [ ] **Step 2: Mark `submitApp` as wallet-authenticated and drop the free-text developer fields from its example**

Change the description and add `security` to the `/api/apps/submit` operation (around line 114-158):
```yaml
  /api/apps/submit:
    post:
      tags: [Submit]
      summary: Submit a new app for review
      description: |
        Creates an app with `status=submitted` and `featured=false`. Requires a wallet
        session (see `POST /api/auth/verify`); the caller must have a `display_name` set
        on their profile first. `developer_slug`/`developer_name` are derived from the
        caller's wallet identity, not taken from the request body.
        Rate limited to 5 submissions per hour per client IP (HTTP 429).
      operationId: submitApp
      security:
        - walletCookie: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AppSubmitRequest'
            example:
              slug: my-mini-app
              name: My Mini App
              domain: myapp.example.com
              category: Games
              tagline: One sentence pitch
              submitter_contact: '@telegram or you@example.com'
              description: Short plain-text summary for listings
              long_description: "## Features\n\n- **Bold** and lists work here"
              tags: [games]
              assets: [NIM]
              release_stage: beta
      responses:
        '201':
          description: App submitted
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AppPublic'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '409':
          description: Slug already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '429':
          $ref: '#/components/responses/TooManyRequests'
```

Also remove `developer_slug`/`developer_name` from the `AppSubmitRequest` schema's `required` list and `properties` (find `AppSubmitRequest:` under `components.schemas` and delete those two entries — the server now derives them).

Update the top-level description note (line 8) from:
```yaml
    **Submitting an app** — `POST /api/apps/submit` (no auth). Rate limit: **5 requests per hour per IP**.
```
to:
```yaml
    **Submitting an app** — `POST /api/apps/submit` (wallet login required). Rate limit: **5 requests per hour per IP**.
```

- [ ] **Step 3: Mark `requestAppUpdate` as owner-only**

Add `security` and a `403` response to `/api/apps/{slug}/request-update` (around line 213-247):
```yaml
  /api/apps/{slug}/request-update:
    post:
      tags: [Submit]
      summary: Request a listing update (reviewed by admins)
      description: |
        Proposes changes to an existing public listing. Requires a wallet session matching
        the app's `developer_wallet_address`. `developer_name`/`developer_slug` in the
        request body are ignored — identity changes are admin-only. Rate limited (same
        5/hour per IP as submit). Only one pending revision per app at a time.
      operationId: requestAppUpdate
      security:
        - walletCookie: []
      parameters:
        - $ref: '#/components/parameters/slug'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/AppUpdateRequest'
      responses:
        '201':
          description: Update request queued
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UpdateRequestCreated'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          description: The wallet session does not own this app
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          description: A pending update already exists for this app
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '429':
          $ref: '#/components/responses/TooManyRequests'
```

- [ ] **Step 4: Document the two new endpoints**

Add near the other wallet-scoped paths (after `/api/apps/{slug}/request-update`):
```yaml
  /api/my/apps:
    get:
      tags: [Submit]
      summary: List apps owned by the logged-in wallet
      operationId: myApps
      security:
        - walletCookie: []
      responses:
        '200':
          description: Apps owned by the caller, each with has_pending_revision
          content:
            application/json:
              schema:
                type: array
                items:
                  allOf:
                    - $ref: '#/components/schemas/AppAdmin'
                    - type: object
                      properties:
                        has_pending_revision:
                          type: boolean
        '401':
          $ref: '#/components/responses/Unauthorized'
```

Add near the other admin paths (after `/api/admin/apps`):
```yaml
  /api/admin/users:
    get:
      tags: [Admin]
      summary: Search users by display name or wallet address (for the developer picker)
      operationId: adminSearchUsers
      security:
        - adminBearer: []
      parameters:
        - name: q
          in: query
          description: Prefix match on display_name or wallet_address; empty returns []
          schema:
            type: string
      responses:
        '200':
          description: Matching users
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    wallet_address:
                      type: string
                    display_name:
                      type: string
                      nullable: true
        '401':
          $ref: '#/components/responses/Unauthorized'
```

- [ ] **Step 5: Add `developer_wallet_address` to the `AppPublic`/`AppAdmin` schemas**

Find `AppPublic:` and `AppAdmin:` under `components.schemas` and add to each `properties`:
```yaml
      developer_wallet_address:
        type: string
        nullable: true
```

- [ ] **Step 6: Regenerate the embedded copies and verify**

```bash
./scripts/gen-openapi.sh
cd backend && go test ./... -run TestOpenAPI -v
```
Expected: script runs without error, `TestOpenAPI...` (in `openapi_test.go`) passes, confirming `backend/openapi.yaml`/`.json` match `docs/openapi.yaml`.

- [ ] **Step 7: Commit**

```bash
git add docs/openapi.yaml backend/openapi.yaml backend/openapi.json
git commit -m "Document wallet-auth requirements and new endpoints in OpenAPI"
```

---

### Task 9: README / DEV.md / AGENTS.md updates

**Files:**
- Modify: `README.md`
- Modify: `docs/DEV.md`
- Modify: `AGENTS.md`

**Interfaces:**
- Consumes: behavior from Tasks 4–7. Pure documentation, no code interfaces produced.

- [ ] **Step 1: Update `README.md`**

Change:
```
- **Submit** — developers submit apps at `/submit` (rate-limited, no account needed).
```
to:
```
- **Submit** — developers log in with their Nimiq wallet and submit apps at `/submit`
  (rate-limited). Once approved, the submitting wallet can request edits to its own
  apps via `/apps/{slug}/update` or manage them from `/my-apps`.
```

- [ ] **Step 2: Update `docs/DEV.md`**

Change:
```
# public submission (no auth; forced to status=submitted, featured=false; 5/hour per IP)
curl -X POST $API/api/apps/submit -H "Content-Type: application/json" \
  -d '{"slug":"my-app","name":"My App","domain":"myapp.example.com","category":"Utilities","developer_slug":"me","developer_name":"Me","tagline":"Does a thing.","submitter_contact":"you@example.com"}'
```
to:
```
# public submission (wallet login required; forced to status=submitted, featured=false; 5/hour per IP)
# developer_slug/developer_name are derived from the caller's profile display_name, not sent here
curl -X POST $API/api/apps/submit -H "Content-Type: application/json" \
  -H "Cookie: wallet_session=<value from POST /api/auth/verify>" \
  -d '{"slug":"my-app","name":"My App","domain":"myapp.example.com","category":"Utilities","tagline":"Does a thing.","submitter_contact":"you@example.com"}'
```

And change:
```
- Developers self-submit at `/submit` on the site (or `POST /api/apps/submit`); new
  submissions land as `submitted` and appear publicly once approved in `/admin`.
```
to:
```
- Developers log in with their wallet and self-submit at `/submit` on the site (or
  `POST /api/apps/submit` with a wallet session cookie); new submissions land as
  `submitted` and appear publicly once approved in `/admin`. Once approved, the
  submitting wallet can request edits via `/apps/{slug}/update` (owner-only) or see
  all its apps at `/my-apps`.
```

- [ ] **Step 3: Update `AGENTS.md`**

Change the submit section (around lines 8-52) to note wallet auth is required and drop `developer_slug`/`developer_name` from the required-fields list and curl example. Find:
```
### 3. Submit

curl -X POST https://api.nimiqminiapps.com/api/apps/submit \
```
and the surrounding curl body/required-fields text, and update:
- The curl example's `-H` list gains `-H "Cookie: wallet_session=<...>"` with a note: "Agents submitting on a developer's behalf cannot complete wallet login themselves — direct the developer to submit via `/submit` in the browser instead, or ask an admin to create the listing via the `admin_create_app` MCP tool (unaffected by this change — it's admin-token-authenticated, not wallet-authenticated)."
- The **Required** fields line: remove `developer_slug`, `developer_name` (now `slug`, `name`, `domain`, `category`, `tagline`, `submitter_contact`).

- [ ] **Step 4: Commit**

```bash
git add README.md docs/DEV.md AGENTS.md
git commit -m "Document wallet-login requirement for app submission"
```

---

### Task 10: Frontend API client updates

**Files:**
- Modify: `frontend/src/api.ts`

**Interfaces:**
- Consumes: `developer_wallet_address` field (Task 2), `GET /api/my/apps` (Task 6), `GET /api/admin/users` (Task 7).
- Produces: `App.developer_wallet_address: string | null`, `submitApp`/`requestAppUpdate` now send `credentials: 'include'`, `export const getMyApps`, `export const adminSearchUsers`. Tasks 12–15 depend on these.

- [ ] **Step 1: Add `developer_wallet_address` to the `App` interface**

In `frontend/src/api.ts`, in `interface App`, add after `developer_name: string`:
```ts
  developer_name: string
  developer_wallet_address: string | null
```

- [ ] **Step 2: Add `credentials: 'include'` to `submitApp` and `requestAppUpdate`**

Change:
```ts
export const submitApp = (app: Partial<App>) =>
  request<RawApp>('/api/apps/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(app),
  }).then(normalizeApp)

export const requestAppUpdate = (slug: string, app: Partial<App> & { author_note?: string }) =>
  request<{ revision_id: string; app_slug: string; status: string }>(
    `/api/apps/${encodeURIComponent(slug)}/request-update`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(app),
    },
  )
```
to:
```ts
export const submitApp = (app: Partial<App>) =>
  request<RawApp>('/api/apps/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(app),
  }).then(normalizeApp)

export const requestAppUpdate = (slug: string, app: Partial<App> & { author_note?: string }) =>
  request<{ revision_id: string; app_slug: string; status: string }>(
    `/api/apps/${encodeURIComponent(slug)}/request-update`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(app),
    },
  )
```

- [ ] **Step 3: Add `getMyApps`**

Add near `getSubmissionStatus`/`getRelatedApps`:
```ts
export const getMyApps = () =>
  request<(RawApp & { has_pending_revision: boolean })[]>('/api/my/apps', { credentials: 'include' })
    .then((items) => items.map((item) => ({ ...normalizeApp(item), has_pending_revision: item.has_pending_revision })))
```

- [ ] **Step 4: Add `adminSearchUsers`**

Add near the other `admin*` functions:
```ts
export interface AdminUserResult {
  wallet_address: string
  display_name: string | null
}

export const adminSearchUsers = (q: string) =>
  adminRequest<AdminUserResult[]>(`/api/admin/users?q=${encodeURIComponent(q)}`)
```

- [ ] **Step 5: Typecheck**

```bash
cd frontend && npm run type-check 2>/dev/null || npx vue-tsc --noEmit
```
Expected: no new type errors. (`normalizeApp`/`RawApp` already pass through unknown extra fields structurally, so the `has_pending_revision` spread in Step 3 needs `RawApp` to allow it — if `RawApp` is a strict type alias rather than `App`, adjust the intersection to `RawApp & { has_pending_revision: boolean }` as shown; if `normalizeApp` strips unknown keys, keep the explicit re-attachment of `has_pending_revision` as written.)

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api.ts
git commit -m "Add my-apps/admin-user-search API client functions; send wallet cookie on submit/request-update"
```

---

### Task 11: `SubmitView.vue` — require wallet login, drop free-text developer fields

**Files:**
- Modify: `frontend/src/views/SubmitView.vue`

**Interfaces:**
- Consumes: `useWalletAuth` (existing, `frontend/src/composables/useWalletAuth.ts`), `WalletLoginButton.vue` (existing), `submitApp` (Task 10).

- [ ] **Step 1: Gate the form behind wallet login and remove developer fields**

In `frontend/src/views/SubmitView.vue`:
- Import `useWalletAuth`:
```ts
import { useWalletAuth } from '../composables/useWalletAuth'
```
- Destructure it in `<script setup>`:
```ts
const { walletAddress, displayName, checking, login } = useWalletAuth()
```
- Delete this line from the `fields` array (the only developer-identity entry in it today):
```ts
  ['developer_name', 'Developer name', true, 'Your name or team'],
```
- In `submit()`, remove `developer_slug`/`developer_name` from the `submitApp` call payload:
```ts
    await submitApp({
      ...form,
      slug: slugify(form.slug),
      tags: csv(form.tags),
      assets: csv(form.assets),
      media: mediaEditor.value?.validate() ?? [],
      socials: socialEditor.value?.validate() ?? [],
      icon_url: form.icon_url || null,
      banner_url: form.banner_url || null,
      website_url: form.website_url || null,
      github_url: form.github_url || null,
    })
```
- Remove the now-unused `form.developer_slug`/`developer_name` initializers from `reactive({...})`.

- [ ] **Step 2: Add the login gate and explanatory copy to the template**

Wrap the existing `<form>` block: before `<form @submit.prevent="submit" ...>`, add:
```html
      <div v-if="checking" class="rounded-2xl border border-line bg-surface p-5 text-sm text-muted">
        Checking wallet session…
      </div>
      <div v-else-if="!walletAddress" class="rounded-2xl border border-line bg-surface p-5 text-center">
        <p class="text-sm text-muted">Connect your Nimiq wallet to submit an app. It will be linked to your wallet as the developer of record — admins can reassign it later.</p>
        <WalletLoginButton class="mt-3 inline-block" />
      </div>
      <div v-else-if="!displayName" class="rounded-2xl border border-line bg-surface p-5 text-center">
        <p class="text-sm text-muted">Set a display name on your profile before submitting — it becomes your public developer name.</p>
        <RouterLink to="/profile" class="mt-3 inline-block rounded-xl bg-nq-blue px-5 py-2.5 font-bold text-white">Go to profile</RouterLink>
      </div>
      <form v-else @submit.prevent="submit" ...>
```
and import `WalletLoginButton`:
```ts
import WalletLoginButton from '../components/WalletLoginButton.vue'
```
Add a line under the existing intro paragraph (`Get your Nimiq Pay mini app listed...`) confirming the tie to wallet identity: "This app will be linked to your wallet (`{{ walletAddress }}`) as the developer of record — admins can reassign it later." shown only when `walletAddress` is set, e.g. directly above the form:
```html
      <p v-if="walletAddress" class="text-xs text-muted">
        Submitting as <span class="font-mono">{{ walletAddress }}</span> — this app will be linked to your wallet as the developer of record; admins can reassign it later.
      </p>
```

- [ ] **Step 3: Manually verify in the browser**

```bash
cd frontend && npm run dev
```
Visit `/submit`: logged out → see "Connect your wallet" prompt, no form. Log in via the wallet button, no display name set → see "set a display name" prompt linking to `/profile`. Set a display name, return to `/submit` → form appears without a developer-name field, submits successfully, and the created app's `developer_name` matches the profile display name (check via `GET /api/apps/{slug}`).

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/SubmitView.vue
git commit -m "Gate app submission behind wallet login; derive developer identity from profile"
```

---

### Task 12: `RequestUpdateView.vue` — lock developer identity field

**Files:**
- Modify: `frontend/src/views/RequestUpdateView.vue`

**Interfaces:**
- Consumes: `requestAppUpdate` (Task 10, now cookie-authenticated and owner-checked server-side).

- [ ] **Step 1: Remove the editable `developer_name` field**

In `frontend/src/views/RequestUpdateView.vue`, delete this line from the `fields` array (developer identity is now locked to admin-only edits):
```ts
  ['developer_name', 'Developer name', true],
```
Also stop sending `developer_slug`/`developer_name` in the `requestAppUpdate(...)` payload inside `submit()`:
```ts
    await requestAppUpdate(slug.value, {
      name: form.name,
      domain: form.domain,
      category: form.category,
      tagline: form.tagline,
      description: form.description,
      long_description: form.long_description,
      release_stage: form.release_stage,
      tags: csv(form.tags),
      assets: csv(form.assets),
      media: mediaEditor.value?.validate() ?? [],
      socials: socialEditor.value?.validate() ?? [],
      icon_url: form.icon_url || null,
      banner_url: form.banner_url || null,
      website_url: form.website_url || null,
      github_url: form.github_url || null,
      author_note: form.author_note,
    })
```
Remove `developer_slug: ''`, `developer_name: ''` from the `form` initializer, and stop assigning them in `load()`'s `Object.assign(form, {...})` call.

- [ ] **Step 2: Surface the ownership 403 clearly**

The existing `catch (e) { error.value = (e as Error).message }` in `submit()` already displays whatever message the backend returns, so a `403` with `"you don't own this app"` will show as-is — no additional code needed. Confirm `request()` in `api.ts` surfaces the response body's `error` field as the thrown `Error`'s message (it already does, per the existing pattern used by every other view's error handling).

- [ ] **Step 3: Manually verify**

Log in as a wallet that owns an app, visit `/apps/{slug}/update` for that app, submit a change — succeeds, form has no developer-name field. Log in as a different wallet (or log out) and visit the same URL — submitting shows "you don't own this app".

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/RequestUpdateView.vue
git commit -m "Lock developer identity out of the request-update form"
```

---

### Task 13: `MyAppsView.vue` + route

**Files:**
- Create: `frontend/src/views/MyAppsView.vue`
- Modify: `frontend/src/main.ts` (add route)

**Interfaces:**
- Consumes: `useWalletAuth`, `getMyApps` (Task 10), `AddressIdenticon.vue` (existing).

- [ ] **Step 1: Create `MyAppsView.vue`**

```vue
<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getMyApps, type App } from '../api'

const { walletAddress, checking } = useWalletAuth()

const apps = ref<(App & { has_pending_revision: boolean })[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  if (!walletAddress.value) {
    loading.value = false
    return
  }
  loading.value = true
  error.value = ''
  try {
    apps.value = await getMyApps()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load your apps'
  } finally {
    loading.value = false
  }
}

watch([checking, walletAddress], () => {
  if (!checking.value) void load()
}, { immediate: true })
</script>

<template>
  <div class="mx-auto max-w-2xl space-y-5">
    <h1 class="text-xl font-extrabold">My apps</h1>

    <p v-if="checking || loading" class="text-sm text-muted">Loading…</p>
    <p v-else-if="!walletAddress" class="text-sm text-muted">Connect your wallet to see the apps you own.</p>
    <p v-else-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>
    <p v-else-if="apps.length === 0" class="text-sm text-muted">
      No apps linked to this wallet yet. <RouterLink to="/submit" class="font-semibold text-accent-ink hover:underline">Submit one</RouterLink>.
    </p>

    <ul v-else class="space-y-3">
      <li v-for="app in apps" :key="app.slug" class="rounded-2xl border border-line bg-surface p-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <RouterLink :to="`/apps/${app.slug}`" class="font-bold hover:underline">{{ app.name }}</RouterLink>
            <p class="text-sm text-muted">{{ app.status }}<span v-if="app.has_pending_revision"> · update pending review</span></p>
          </div>
          <RouterLink :to="`/apps/${app.slug}/update`"
            class="shrink-0 rounded-xl border border-line bg-surface-2 px-3 py-1.5 text-sm font-semibold hover:border-accent/50">
            Request update
          </RouterLink>
        </div>
      </li>
    </ul>
  </div>
</template>
```

- [ ] **Step 2: Add the route**

In `frontend/src/main.ts`, add near `/profile`:
```ts
    { path: '/my-apps', component: () => import('./views/MyAppsView.vue'), meta: { title: 'My Apps' } },
```

- [ ] **Step 3: Manually verify**

Visit `/my-apps` logged out → "Connect your wallet" message. Log in as a wallet owning apps (from Task 4/11's manual test) → list shows them with status and a working "Request update" link into `/apps/{slug}/update`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/MyAppsView.vue frontend/src/main.ts
git commit -m "Add My Apps developer dashboard"
```

---

### Task 14: Admin developer picker

**Files:**
- Modify: `frontend/src/views/AdminView.vue`

**Interfaces:**
- Consumes: `adminSearchUsers` (Task 10).

- [ ] **Step 1: Add a debounced search input bound to `developer_wallet_address`**

In `frontend/src/views/AdminView.vue`, import `adminSearchUsers` and its result type:
```ts
import { /* existing imports */, adminSearchUsers, type AdminUserResult } from '../api'
```
Add state for the picker (near the existing `form`/`editingSlug` refs):
```ts
const developerQuery = ref('')
const developerResults = ref<AdminUserResult[]>([])
let developerSearchTimer: ReturnType<typeof setTimeout> | undefined

function onDeveloperQueryInput() {
  clearTimeout(developerSearchTimer)
  developerSearchTimer = setTimeout(async () => {
    developerResults.value = developerQuery.value.trim()
      ? await adminSearchUsers(developerQuery.value.trim())
      : []
  }, 250)
}

function pickDeveloper(user: AdminUserResult) {
  form.developer_wallet_address = user.wallet_address
  developerQuery.value = user.display_name ?? user.wallet_address
  developerResults.value = []
}
```
Add `developer_wallet_address: '' as string | null` to `form`'s `reactive({...})` initializer.

- [ ] **Step 2: Add the picker to the template**

Near the existing `developer_slug`/`developer_name` fields in the admin app form, add:
```html
          <label class="relative text-sm sm:col-span-2">
            <span class="mb-1 block font-semibold text-muted">Owning developer (optional)</span>
            <input v-model="developerQuery" @input="onDeveloperQueryInput"
              placeholder="Search by display name or wallet address — leave blank for an unclaimed/anonymous app"
              class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 focus:border-accent" />
            <ul v-if="developerResults.length" class="absolute z-10 mt-1 w-full rounded-lg border border-line bg-surface shadow-lg">
              <li v-for="user in developerResults" :key="user.wallet_address"
                @click="pickDeveloper(user)"
                class="cursor-pointer px-3 py-2 text-sm hover:bg-surface-2">
                {{ user.display_name ?? 'No display name' }}
                <span class="block font-mono text-xs text-muted">{{ user.wallet_address }}</span>
              </li>
            </ul>
            <span v-if="form.developer_wallet_address" class="mt-1 block text-xs text-muted">
              Linked to <span class="font-mono">{{ form.developer_wallet_address }}</span>
              <button type="button" @click="form.developer_wallet_address = ''; developerQuery = ''" class="ml-1 text-accent-ink hover:underline">clear</button>
            </span>
          </label>
```

This sits alongside the existing free-text `developer_slug`/`developer_name` fields (unchanged, still admin-editable for anonymous/legacy apps) — the picker only sets `developer_wallet_address`; it doesn't overwrite the name/slug fields, matching the spec's "falls back to today's free-text field for unclaimed/anonymous apps."

- [ ] **Step 3: Ensure edit-mode prefills the picker**

Wherever the admin form is populated for editing an existing app (the function that sets `form` from a selected app, mentioned around line 76 with `developer_slug: app.developer_slug, developer_name: app.developer_name,`), add:
```ts
    developer_wallet_address: app.developer_wallet_address,
```
and set `developerQuery.value = app.developer_name` there too so the picker shows something sensible when editing.

- [ ] **Step 4: Manually verify**

In `/admin`, open an existing app for edit, search a wallet's display name in the new field, pick a result, save — `GET /api/admin/apps` afterward shows `developer_wallet_address` set. Clear it and save — confirm it goes back to `null` (app becomes unclaimed/admin-managed again, matching "leave blank for anonymous").

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/AdminView.vue
git commit -m "Add developer wallet picker to the admin app form"
```

---

## Final check

- [ ] **Full backend test + build**
```bash
cd backend && go build ./... && go vet ./... && go test ./...
```
- [ ] **Frontend typecheck**
```bash
cd frontend && npx vue-tsc --noEmit
```
- [ ] **End-to-end manual walkthrough**: submit an app while logged in with no prior apps (identity gets derived + assigned) → approve it in `/admin` → see it in `/my-apps` → request an update from `/my-apps` → approve the revision in `/admin` → confirm the live listing changed but `developer_name`/`developer_slug` did not, even if the update request tried to change them.
