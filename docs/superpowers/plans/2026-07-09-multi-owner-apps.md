# Multi-owner Apps Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let more than one wallet own an app (e.g. a developer's desktop wallet and their Nimiq Pay mobile wallet), any of which can manage it, replacing the single `developer_wallet_address` column with a join table.

**Architecture:** Add `app_owners(app_slug, wallet_address)` and drop `apps.developer_wallet_address`. Ownership checks everywhere become "is this wallet in the set" instead of an equality check. `App.owner_wallet_addresses []string` is computed via a correlated subquery appended to the existing `appColumns` snippet, so every endpoint that already returns `App` gets it for free. New self-service (`/api/apps/{slug}/owners`) and admin (`/api/admin/apps/{slug}/owners`) endpoints manage membership; `updateApp`/`createApp` stop touching ownership entirely.

**Tech Stack:** Go 1.x + `pgx/v5` (backend), Vue 3 + TypeScript (frontend), the `mcp/` TypeScript MCP server (Node, `@modelcontextprotocol/sdk`, Zod).

## Global Constraints

- Self-service add/remove requires the caller to already be a current owner (mutual trust — any owner can add or remove any other). Self-service removal is blocked (`409`) if it would leave zero owners; admin removal has no such block.
- Adding a co-owner is instant — no admin-approval queue. It only changes who can propose edits; the edits themselves still go through the existing `app_revisions` approval flow untouched.
- A wallet co-owning an app reuses that app's `developer_slug` if it later submits a new app of its own (`resolveDeveloperSlug` looks up via `app_owners`, not a single column).
- `developer_name`/`developer_slug` are untouched by this plan — they remain freely admin-editable regardless of ownership, per the prior fix.
- This repo has no Postgres integration-test harness (see prior plans) — verify DB-touching handlers manually against the local dev stack (`docker exec -i nimiqminiapps-postgres-1 psql ...` or `curl` per `docs/DEV.md`), and unit-test only pure functions.
- `docs/openapi.yaml` is the source of truth; regenerate `backend/openapi.yaml`/`.json` with `./scripts/gen-openapi.sh` — CI fails if they drift.

Spec: [`docs/superpowers/specs/2026-07-09-multi-owner-apps-design.md`](../specs/2026-07-09-multi-owner-apps-design.md)

---

### Task 1: Migration — `app_owners` table, backfill, drop old column

**Files:**
- Create: `backend/migrations/014_app_owners.sql`

**Interfaces:**
- Produces: table `app_owners(app_slug, wallet_address, added_at)`. Every later backend task depends on this.

- [ ] **Step 1: Write the migration**

```sql
CREATE TABLE IF NOT EXISTS app_owners (
    app_slug TEXT NOT NULL REFERENCES apps(slug) ON DELETE CASCADE,
    wallet_address TEXT NOT NULL REFERENCES users(wallet_address),
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (app_slug, wallet_address)
);

INSERT INTO app_owners (app_slug, wallet_address)
SELECT slug, developer_wallet_address FROM apps
WHERE developer_wallet_address IS NOT NULL
ON CONFLICT DO NOTHING;

DROP INDEX IF EXISTS apps_developer_wallet_address_idx;
ALTER TABLE apps DROP COLUMN IF EXISTS developer_wallet_address;
```

- [ ] **Step 2: Apply it against local dev Postgres and confirm the backfill**

```bash
cd backend && go run . &
sleep 2
curl -s localhost:8080/health
kill %1
```
Expected: `200`, and logs include `"applied migration" name=014_app_owners.sql`. Then:
```bash
docker exec -i nimiqminiapps-postgres-1 psql -U nimiq -d nimiq_miniapps -c \
  "SELECT app_slug, wallet_address FROM app_owners ORDER BY app_slug;"
docker exec -i nimiqminiapps-postgres-1 psql -U nimiq -d nimiq_miniapps -c "\d apps" | grep developer_wallet_address
```
Expected: one `app_owners` row per app that previously had a non-null
`developer_wallet_address`, and no `developer_wallet_address` column left
on `apps`.

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/014_app_owners.sql
git commit -m "Add app_owners table; backfill and drop developer_wallet_address"
```

---

### Task 2: `App.OwnerWalletAddresses` via correlated subquery

**Files:**
- Modify: `backend/handlers.go` (`App` struct, `appColumns`, `scanApp`)

**Interfaces:**
- Consumes: `app_owners` table (Task 1).
- Produces: `App.OwnerWalletAddresses []string` (JSON: `owner_wallet_addresses`), populated on every `App`-returning query automatically. Verified against a live Postgres that `RETURNING`/`SELECT` with this correlated subquery works for `INSERT`, `UPDATE`, and plain `SELECT`.

- [ ] **Step 1: Replace the field on the `App` struct**

In `backend/handlers.go`, replace:
```go
	DeveloperWalletAddress *string      `json:"developer_wallet_address"`
```
with:
```go
	OwnerWalletAddresses   []string     `json:"owner_wallet_addresses"`
```

- [ ] **Step 2: Replace the column in `appColumns` with the correlated subquery**

Change:
```go
const appColumns = `id, slug, name, domain, category, developer_slug, developer_name, tagline,
	description, long_description, tags, assets, status, release_stage, featured, featured_order,
	website_url, github_url, icon_url, discovered_icon_url, banner_url, media, socials, domain_reachable, domain_checked_at,
	submitter_contact, created_at, updated_at, developer_wallet_address`
```
to:
```go
const appColumns = `id, slug, name, domain, category, developer_slug, developer_name, tagline,
	description, long_description, tags, assets, status, release_stage, featured, featured_order,
	website_url, github_url, icon_url, discovered_icon_url, banner_url, media, socials, domain_reachable, domain_checked_at,
	submitter_contact, created_at, updated_at,
	(ARRAY(SELECT wallet_address FROM app_owners WHERE app_owners.app_slug = apps.slug ORDER BY added_at)) AS owner_wallet_addresses`
```
This is a plain, unaliased reference to the `apps` table's own name — confirmed working in `INSERT ... RETURNING`, `UPDATE ... RETURNING`, and `SELECT` against this schema. Every call site already does `FROM apps` / `INSERT INTO apps` / `UPDATE apps` without aliasing (checked across the codebase), so this is safe everywhere `appColumns` is used.

- [ ] **Step 3: Update `scanApp`'s `Scan` call**

Change the last line of the `Scan(...)` call in `scanApp` from:
```go
		&mediaJSON, &socialsJSON, &a.DomainReachable, &a.DomainCheckedAt, &a.SubmitterContact, &a.CreatedAt, &a.UpdatedAt,
		&a.DeveloperWalletAddress)
```
to:
```go
		&mediaJSON, &socialsJSON, &a.DomainReachable, &a.DomainCheckedAt, &a.SubmitterContact, &a.CreatedAt, &a.UpdatedAt,
		&a.OwnerWalletAddresses)
```
Add, right after the existing `if a.Socials == nil { a.Socials = []SocialLink{} }` block:
```go
	if a.OwnerWalletAddresses == nil {
		a.OwnerWalletAddresses = []string{}
	}
```
(pgx scans a Postgres empty array as an empty Go slice already in most cases, but this guards against a `NULL` array the same way `Media`/`Socials` already do.)

- [ ] **Step 4: Build**

```bash
cd backend && go build ./... 2>&1 | head -60
```
Expected: compile errors at every remaining use of `DeveloperWalletAddress` — that's the checklist for Tasks 3–7. Do not fix them yet; this step just confirms the struct/column change itself is syntactically correct in isolation. Note the error sites for reference.

- [ ] **Step 5: Commit**

```bash
git add backend/handlers.go
git commit -m "Replace App.DeveloperWalletAddress with OwnerWalletAddresses"
```

---

### Task 3: Owner management functions in `developer.go`

**Files:**
- Modify: `backend/developer.go`

**Interfaces:**
- Consumes: `app_owners` table.
- Produces: `(s *server) isOwner(ctx, slug, wallet string) (bool, error)`, `(s *server) addOwner(ctx, slug, wallet string) error`, `(s *server) removeOwner(ctx, slug, wallet string, allowEmpty bool) error`, sentinel `errLastOwner`. Tasks 6 and 7 (request-update, owner endpoints) depend on these exact signatures. `resolveDeveloperSlug` and `myApps` are rewritten to query through `app_owners`.

- [ ] **Step 1: Remove `validateDeveloperWallet`**

Delete the whole `validateDeveloperWallet` function from `backend/developer.go` (it operated on the now-removed `App.DeveloperWalletAddress` field; its validation logic moves into `addOwner` in Step 3 below).

- [ ] **Step 2: Rewrite `resolveDeveloperSlug` to query through `app_owners`**

Replace:
```go
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
with:
```go
func (s *server) resolveDeveloperSlug(ctx context.Context, address, displayName string) (string, error) {
	var existing string
	err := s.pool.QueryRow(ctx, `
		SELECT a.developer_slug FROM apps a
		JOIN app_owners o ON o.app_slug = a.slug
		WHERE o.wallet_address = $1 LIMIT 1`, address).
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
		if err := s.pool.QueryRow(ctx, `
			SELECT count(*) FROM apps a
			WHERE a.developer_slug = $1
			  AND NOT EXISTS (SELECT 1 FROM app_owners o WHERE o.app_slug = a.slug AND o.wallet_address = $2)`,
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
(A wallet that already co-owns *any* app reuses that app's slug — including apps it was added to as a co-owner, not just ones it originally submitted. The collision check treats an unclaimed app or one owned only by other wallets as a collision, same as before.)

- [ ] **Step 3: Add `isOwner`, `addOwner`, `removeOwner`, `errLastOwner`**

Add to `backend/developer.go`:
```go
var errLastOwner = errors.New("can't remove the last owner")

func (s *server) isOwner(ctx context.Context, slug, wallet string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM app_owners WHERE app_slug=$1 AND wallet_address=$2)`,
		slug, wallet).Scan(&ok)
	return ok, err
}

// addOwner links wallet to slug's ownership set. The wallet must already have a
// profile with a display name set (same bar the old single-owner link enforced).
// Adding an already-current owner again is a no-op.
func (s *server) addOwner(ctx context.Context, slug, wallet string) error {
	wallet = strings.TrimSpace(wallet)
	if wallet == "" {
		return errors.New("wallet_address is required")
	}
	var displayName *string
	err := s.pool.QueryRow(ctx,
		`SELECT display_name FROM users WHERE wallet_address=$1`, wallet).
		Scan(&displayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("wallet must have logged in at least once")
	}
	if err != nil {
		return err
	}
	if displayName == nil || strings.TrimSpace(*displayName) == "" {
		return errors.New("wallet must set a display name on their profile first")
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO app_owners (app_slug, wallet_address) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		slug, wallet)
	return err
}

// removeOwner unlinks wallet from slug's ownership set. When allowEmpty is false
// (self-service), it refuses to remove the last remaining owner.
func (s *server) removeOwner(ctx context.Context, slug, wallet string, allowEmpty bool) error {
	if !allowEmpty {
		var count int
		if err := s.pool.QueryRow(ctx,
			`SELECT count(*) FROM app_owners WHERE app_slug=$1`, slug).Scan(&count); err != nil {
			return err
		}
		if count <= 1 {
			return errLastOwner
		}
	}
	_, err := s.pool.Exec(ctx,
		`DELETE FROM app_owners WHERE app_slug=$1 AND wallet_address=$2`, slug, wallet)
	return err
}
```

- [ ] **Step 4: Rewrite `myApps` to query through `app_owners`**

Change:
```go
	rows, err := s.pool.Query(r.Context(),
		"SELECT "+appColumns+" FROM apps WHERE developer_wallet_address=$1 ORDER BY created_at DESC", address)
```
to:
```go
	rows, err := s.pool.Query(r.Context(),
		`SELECT `+appColumns+` FROM apps
		 WHERE EXISTS (SELECT 1 FROM app_owners o WHERE o.app_slug = apps.slug AND o.wallet_address = $1)
		 ORDER BY created_at DESC`, address)
```

- [ ] **Step 5: Build**

```bash
cd backend && go build ./... 2>&1 | head -60
```
Expected: remaining errors only in `handlers.go` (`decodeAndInsert`, `updateApp`) and `submit.go`/`revisions.go` — addressed in Tasks 4–6.

- [ ] **Step 6: Commit**

```bash
cd backend && gofmt -w developer.go
git add backend/developer.go
git commit -m "Add owner management functions; rewrite slug/my-apps queries through app_owners"
```

---

### Task 4: Transactional owner insert on self-submission

**Files:**
- Modify: `backend/handlers.go` (`decodeAndInsert`, `createApp`)
- Modify: `backend/submit.go` (`submitApp`)

**Interfaces:**
- Consumes: `s.addOwner`... actually no — the *first* owner on submission is inserted directly (it's already validated by `submitApp` itself, which already checked the wallet has a display name), not through `addOwner`'s validation (which would just redundantly re-check the same thing). Depends on `app_owners` (Task 1).
- Produces: `decodeAndInsert(w, r, force func(*App), requireContact bool, ownerAddress *string)` — the new trailing parameter. `createApp` passes `nil`; `submitApp` passes `&address`.

- [ ] **Step 1: Change `decodeAndInsert`'s signature and remove the old wallet validation call**

In `backend/handlers.go`, change:
```go
func (s *server) decodeAndInsert(w http.ResponseWriter, r *http.Request, force func(*App), requireContact bool) {
	var a App
	a.Tags, a.Assets, a.Media = []string{}, []string{}, []MediaItem{}
	a.Status = "submitted"
	a.ReleaseStage = "released"
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	a.SubmitterContact = strings.TrimSpace(a.SubmitterContact)
	if force != nil {
		force(&a)
	}
	if err := s.validateDeveloperWallet(r.Context(), &a); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if requireContact {
```
to:
```go
func (s *server) decodeAndInsert(w http.ResponseWriter, r *http.Request, force func(*App), requireContact bool, ownerAddress *string) {
	var a App
	a.Tags, a.Assets, a.Media = []string{}, []string{}, []MediaItem{}
	a.Status = "submitted"
	a.ReleaseStage = "released"
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	a.SubmitterContact = strings.TrimSpace(a.SubmitterContact)
	if force != nil {
		force(&a)
	}
	if requireContact {
```

- [ ] **Step 2: Wrap the insert in a transaction when there's an owner to add**

Change the INSERT block from:
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
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		writeError(w, http.StatusConflict, "slug already exists")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
```
to:
```go
	insertSQL := `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline,
			description, long_description, tags, assets, status, release_stage, featured, featured_order,
			website_url, github_url, icon_url, banner_url, media, socials, submitter_contact)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING ` + appColumns
	insertArgs := []any{
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.LongDescription, a.Tags, a.Assets, a.Status, a.ReleaseStage, a.Featured, a.FeaturedOrder,
		a.WebsiteURL, a.GithubURL, a.IconURL, a.BannerURL, mediaJSON, socialsJSON, a.SubmitterContact,
	}

	var pgErr *pgconn.PgError
	if ownerAddress != nil {
		tx, err2 := s.pool.Begin(r.Context())
		if err2 != nil {
			writeError(w, http.StatusInternalServerError, err2.Error())
			return
		}
		defer tx.Rollback(r.Context())
		a, err = scanApp(tx.QueryRow(r.Context(), insertSQL, insertArgs...))
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if _, err = tx.Exec(r.Context(),
			`INSERT INTO app_owners (app_slug, wallet_address) VALUES ($1,$2)`, a.Slug, *ownerAddress); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err = tx.Commit(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.OwnerWalletAddresses = []string{*ownerAddress}
	} else {
		a, err = scanApp(s.pool.QueryRow(r.Context(), insertSQL, insertArgs...))
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
```
(The subquery in `appColumns` can't see the `app_owners` row until it's inserted — which must happen *after* the `apps` row exists, since `app_owners.app_slug` references it. Rather than re-`SELECT` the app a third time, the owner is known exactly, so `a.OwnerWalletAddresses` is set directly in Go for the response.)

- [ ] **Step 3: Update `createApp`'s call site**

Change:
```go
func (s *server) createApp(w http.ResponseWriter, r *http.Request) {
	s.decodeAndInsert(w, r, nil, false)
}
```
to:
```go
func (s *server) createApp(w http.ResponseWriter, r *http.Request) {
	s.decodeAndInsert(w, r, nil, false, nil)
}
```

- [ ] **Step 4: Update `submitApp`'s call site**

In `backend/submit.go`, change the final call from:
```go
	s.decodeAndInsert(w, r, func(a *App) {
		a.Status = "submitted"
		a.Featured = false
		a.DeveloperWalletAddress = &address
		a.DeveloperSlug = devSlug
		a.DeveloperName = *displayName
	}, true)
```
to:
```go
	s.decodeAndInsert(w, r, func(a *App) {
		a.Status = "submitted"
		a.Featured = false
		a.DeveloperSlug = devSlug
		a.DeveloperName = *displayName
	}, true, &address)
```

- [ ] **Step 5: Build**

```bash
cd backend && go build ./... 2>&1 | head -60
```
Expected: remaining errors only in `updateApp` (Task 5) and `revisions.go` (Task 6).

- [ ] **Step 6: Manually verify against local dev stack**

Log in with a wallet that has a display name set (per the developer-portal plan's Task 4 verification), submit a new app via `/submit`, then:
```bash
docker exec -i nimiqminiapps-postgres-1 psql -U nimiq -d nimiq_miniapps -c \
  "SELECT app_slug, wallet_address FROM app_owners WHERE app_slug='<the-new-slug>';"
```
Expected: exactly one row, matching the submitting wallet. Also confirm the JSON response from the submit call included `"owner_wallet_addresses": ["<address>"]`.

- [ ] **Step 7: Commit**

```bash
cd backend && gofmt -w handlers.go submit.go
git add backend/handlers.go backend/submit.go
git commit -m "Insert first owner transactionally on self-submission"
```

---

### Task 5: Drop ownership from `updateApp`

**Files:**
- Modify: `backend/handlers.go` (`updateApp`)

**Interfaces:**
- Consumes: nothing new.
- Produces: `updateApp` no longer reads or writes ownership at all — it's managed exclusively by Task 7's endpoints.

- [ ] **Step 1: Remove the wallet validation call**

Delete this block from `updateApp`:
```go
	if err := s.validateDeveloperWallet(r.Context(), &a); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
```

- [ ] **Step 2: Drop `developer_wallet_address` from the UPDATE statement**

Change:
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
to:
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
(`a.OwnerWalletAddresses`, if present in the request JSON, is simply never referenced in this SQL — it's read-only from the client's perspective, same as `open_url`.)

- [ ] **Step 2: Build and run existing tests**

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```
Expected: clean build; remaining errors only in `revisions.go` (Task 6).

- [ ] **Step 3: Commit**

```bash
cd backend && gofmt -w handlers.go
git add backend/handlers.go
git commit -m "Stop touching ownership in updateApp — managed via owner endpoints"
```

---

### Task 6: Membership-based ownership check in `request-update`

**Files:**
- Modify: `backend/revisions.go` (`requestAppUpdate`)

**Interfaces:**
- Consumes: `s.isOwner` (Task 3).

- [ ] **Step 1: Replace the equality check with a membership check**

Change:
```go
	if current.DeveloperWalletAddress == nil || *current.DeveloperWalletAddress != address {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}
```
to:
```go
	owner, err := s.isOwner(r.Context(), slug, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !owner {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}
```

- [ ] **Step 2: Build**

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```
Expected: clean build, all tests pass — this should be the last of the `DeveloperWalletAddress` compile errors from Task 2.

- [ ] **Step 3: Manually verify**

With two wallets both added as owners of the same app (once Task 7 ships you can set this up via the new endpoints; until then, insert a second row directly for a manual check):
```bash
docker exec -i nimiqminiapps-postgres-1 psql -U nimiq -d nimiq_miniapps -c \
  "INSERT INTO app_owners (app_slug, wallet_address) SELECT '<slug>', wallet_address FROM users LIMIT 1 OFFSET 1 ON CONFLICT DO NOTHING;"
```
Then confirm both wallets' sessions can successfully `POST /apps/{slug}/update`, and a third, non-owner wallet gets `403`.

- [ ] **Step 4: Commit**

```bash
cd backend && gofmt -w revisions.go
git add backend/revisions.go
git commit -m "Check app_owners membership instead of single-wallet equality in request-update"
```

---

### Task 7: Owner management endpoints (self-service + admin)

**Files:**
- Modify: `backend/developer.go` (add handlers)
- Modify: `backend/main.go` (add routes)

**Interfaces:**
- Consumes: `s.isOwner`, `s.addOwner`, `s.removeOwner`, `errLastOwner` (Task 3).
- Produces: the four endpoints the frontend (Tasks 12, 13) and MCP server (Task 9) call against.

- [ ] **Step 1: Add the four handlers to `backend/developer.go`**

```go
type ownerRequestBody struct {
	WalletAddress string `json:"wallet_address"`
}

func (s *server) addAppOwnerSelf(w http.ResponseWriter, r *http.Request, address string) {
	slug := r.PathValue("slug")
	owner, err := s.isOwner(r.Context(), slug, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !owner {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}
	var body ownerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := s.addOwner(r.Context(), slug, body.WalletAddress); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (s *server) removeAppOwnerSelf(w http.ResponseWriter, r *http.Request, address string) {
	slug := r.PathValue("slug")
	owner, err := s.isOwner(r.Context(), slug, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !owner {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}
	target := r.PathValue("wallet")
	if err := s.removeOwner(r.Context(), slug, target, false); err != nil {
		if errors.Is(err, errLastOwner) {
			writeError(w, http.StatusConflict, "can't remove the last owner — ask an admin to unclaim this app instead")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (s *server) adminAddAppOwner(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var body ownerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := s.addOwner(r.Context(), slug, body.WalletAddress); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (s *server) adminRemoveAppOwner(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	target := r.PathValue("wallet")
	if err := s.removeOwner(r.Context(), slug, target, true); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
```

- [ ] **Step 2: Add the routes**

In `backend/main.go`, add near the other wallet-auth app routes (after `GET /api/my/apps`):
```go
	mux.HandleFunc("POST /api/apps/{slug}/owners", walletAuthMiddleware(walletAuthSecret, s.addAppOwnerSelf))
	mux.HandleFunc("DELETE /api/apps/{slug}/owners/{wallet}", walletAuthMiddleware(walletAuthSecret, s.removeAppOwnerSelf))
```
And near the other admin app routes (after `PATCH /api/admin/apps/{slug}`):
```go
	mux.HandleFunc("POST /api/admin/apps/{slug}/owners", s.adminAuth(s.adminAddAppOwner))
	mux.HandleFunc("DELETE /api/admin/apps/{slug}/owners/{wallet}", s.adminAuth(s.adminRemoveAppOwner))
```

- [ ] **Step 3: Build**

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

- [ ] **Step 4: Manually verify all four endpoints**

```bash
# self-service, no cookie
curl -i -X POST localhost:8080/api/apps/some-slug/owners -d '{"wallet_address":"NQ..."}'
```
Expected `401`. With a cookie for a wallet that isn't a current owner: `403`. With the owning wallet's cookie and a valid second wallet address (one that has logged in and set a display name): `200`, and a follow-up `GET /api/apps/some-slug` shows both addresses in `owner_wallet_addresses`.

```bash
# self-service remove down to the last owner
curl -i -X DELETE "localhost:8080/api/apps/some-slug/owners/<last-remaining-wallet>" \
  -H "Cookie: wallet_session=<owning wallet's cookie>"
```
Expected `409` once only one owner remains.

```bash
# admin add/remove — no membership or last-owner restriction
curl -i -X POST localhost:8080/api/admin/apps/some-slug/owners \
  -H "Authorization: Bearer $ADMIN_TOKEN" -d '{"wallet_address":"NQ..."}'
curl -i -X DELETE localhost:8080/api/admin/apps/some-slug/owners/<wallet> \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```
Expected both succeed regardless of who currently owns the app, and admin can remove the last owner down to zero (unclaimed).

- [ ] **Step 5: Commit**

```bash
cd backend && gofmt -w developer.go main.go
git add backend/developer.go backend/main.go
git commit -m "Add self-service and admin owner management endpoints"
```

---

### Task 8: OpenAPI spec

**Files:**
- Modify: `docs/openapi.yaml`
- Generated: `backend/openapi.yaml`, `backend/openapi.json`

- [ ] **Step 1: Replace `developer_wallet_address` with `owner_wallet_addresses` in `AppPublic`**

In `docs/openapi.yaml`, find the `AppPublic` schema's property:
```yaml
        developer_wallet_address:
          type: string
          nullable: true
```
Replace with:
```yaml
        owner_wallet_addresses:
          type: array
          items:
            type: string
          description: Wallets that can manage this listing (My apps, request-update). Empty when unclaimed.
```

- [ ] **Step 2: Remove the now-redundant duplicate in `AppAdmin`**

Find the `AppAdmin` schema:
```yaml
    AppAdmin:
      allOf:
        - $ref: '#/components/schemas/AppPublic'
        - type: object
          properties:
            submitter_contact:
              type: string
            developer_wallet_address:
              type: string
              nullable: true
```
Change to:
```yaml
    AppAdmin:
      allOf:
        - $ref: '#/components/schemas/AppPublic'
        - type: object
          properties:
            submitter_contact:
              type: string
```

- [ ] **Step 3: Reword the `request-update` ownership description**

Change:
```yaml
        Proposes changes to an existing public listing. Requires a wallet session matching
        the app's `developer_wallet_address`. `developer_name`/`developer_slug` in the
```
to:
```yaml
        Proposes changes to an existing public listing. Requires a wallet session for one
        of the app's current owners (see `owner_wallet_addresses`). `developer_name`/`developer_slug` in the
```

- [ ] **Step 4: Document the four new owner endpoints**

Add after `/api/apps/{slug}/request-update`:
```yaml
  /api/apps/{slug}/owners:
    post:
      tags: [Submit]
      summary: Add a co-owner wallet (self-service, caller must be a current owner)
      operationId: addAppOwner
      security:
        - walletCookie: []
      parameters:
        - $ref: '#/components/parameters/slug'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [wallet_address]
              properties:
                wallet_address:
                  type: string
      responses:
        '200':
          description: Owner added (or already present)
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: added
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          description: The wallet session is not a current owner
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /api/apps/{slug}/owners/{wallet}:
    delete:
      tags: [Submit]
      summary: Remove a co-owner wallet (self-service, caller must be a current owner)
      operationId: removeAppOwner
      security:
        - walletCookie: []
      parameters:
        - $ref: '#/components/parameters/slug'
        - name: wallet
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Owner removed
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          description: The wallet session is not a current owner
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '409':
          description: Would remove the last remaining owner
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

Add after `/api/admin/apps/{slug}`:
```yaml
  /api/admin/apps/{slug}/owners:
    post:
      tags: [Admin]
      summary: Add a co-owner wallet (admin, no membership requirement)
      operationId: adminAddAppOwner
      security:
        - adminBearer: []
      parameters:
        - $ref: '#/components/parameters/slug'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [wallet_address]
              properties:
                wallet_address:
                  type: string
      responses:
        '200':
          description: Owner added (or already present)
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'

  /api/admin/apps/{slug}/owners/{wallet}:
    delete:
      tags: [Admin]
      summary: Remove a co-owner wallet (admin — can remove the last owner to fully unclaim)
      operationId: adminRemoveAppOwner
      security:
        - adminBearer: []
      parameters:
        - $ref: '#/components/parameters/slug'
        - name: wallet
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Owner removed
        '401':
          $ref: '#/components/responses/Unauthorized'
```

- [ ] **Step 5: Regenerate and verify**

```bash
./scripts/gen-openapi.sh
cd backend && go test ./... -run TestOpenAPI -v
```
Expected: `PASS`.

- [ ] **Step 6: Commit**

```bash
git add docs/openapi.yaml backend/openapi.yaml backend/openapi.json
git commit -m "Document owner_wallet_addresses and owner management endpoints in OpenAPI"
```

---

### Task 9: MCP server — drop `developer_wallet_address`, add owner tools

**Files:**
- Modify: `mcp/src/api.ts`
- Modify: `mcp/src/index.ts`
- Modify: `mcp/README.md`

**Interfaces:**
- Consumes: `POST /api/admin/apps/{slug}/owners`, `DELETE /api/admin/apps/{slug}/owners/{wallet}` (Task 7).
- Produces: `adminAddAppOwner(slug, wallet)`, `adminRemoveAppOwner(slug, wallet)` in `api.ts`; MCP tools `admin_add_app_owner`, `admin_remove_app_owner`.

- [ ] **Step 1: Add the API functions**

In `mcp/src/api.ts`, add near `adminSearchUsers`:
```ts
export async function adminAddAppOwner(slug: string, walletAddress: string) {
  return request(`/api/admin/apps/${encodeURIComponent(slug)}/owners`, {
    method: 'POST',
    headers: adminHeaders(),
    body: JSON.stringify({ wallet_address: walletAddress }),
  })
}

export async function adminRemoveAppOwner(slug: string, walletAddress: string) {
  return request(`/api/admin/apps/${encodeURIComponent(slug)}/owners/${encodeURIComponent(walletAddress)}`, {
    method: 'DELETE',
    headers: adminHeaders(),
  })
}
```

- [ ] **Step 2: Remove `developer_wallet_address` from `appFields` and reword the auto-fill descriptions**

In `mcp/src/index.ts`, change:
```ts
  developer_slug: z.string().optional().describe(
    'Public catalog developer slug. Required for unclaimed apps; auto-filled from the wallet owner profile when developer_wallet_address is set.',
  ),
  developer_name: z.string().optional().describe(
    'Public catalog developer name. Required for unclaimed apps; auto-filled from profile when developer_wallet_address is set.',
  ),
  developer_wallet_address: z.string().nullable().optional().describe(
    'Wallet address of the app owner (My apps + request-update). Null or omit for unclaimed/legacy listings.',
  ),
```
to:
```ts
  developer_slug: z.string().optional().describe(
    'Public catalog developer slug. Always required — set directly, or use admin_add_app_owner afterward to link a wallet (ownership no longer travels through create/update).',
  ),
  developer_name: z.string().optional().describe(
    'Public catalog developer name. Always required — set directly; unaffected by ownership.',
  ),
```

- [ ] **Step 3: Update `admin_search_users`, `admin_create_app`, `admin_update_app` descriptions**

Change:
```ts
    description:
      'Search users by display name or wallet address prefix (requires admin token). Use before assigning developer_wallet_address on an app.',
```
to:
```ts
    description:
      'Search users by display name or wallet address prefix (requires admin token). Use before admin_add_app_owner.',
```
Change:
```ts
    description:
      'Create a new app (requires admin token). Set developer_wallet_address to link an owner — name/slug are taken from their profile. Unclaimed apps need developer_name and developer_slug.',
```
to:
```ts
    description:
      'Create a new app (requires admin token). Always set developer_name and developer_slug directly. Use admin_add_app_owner afterward to link one or more wallets.',
```
Change:
```ts
    description:
      'Update an app by slug; merges with the current record (requires admin token). Set developer_wallet_address to assign/reassign ownership (name/slug derived from profile). Pass null to unclaim.',
```
to:
```ts
    description:
      'Update an app by slug; merges with the current record (requires admin token). Ownership is managed separately — use admin_add_app_owner / admin_remove_app_owner.',
```

- [ ] **Step 4: Register the two new tools**

Add after the `admin_search_users` tool registration in `mcp/src/index.ts`:
```ts
server.registerTool(
  'admin_add_app_owner',
  {
    description: 'Link a wallet as a co-owner of an app, granting My apps / request-update access (requires admin token). No effect if already an owner.',
    inputSchema: {
      slug: z.string().describe('App slug'),
      wallet_address: z.string().describe('Wallet to add — must have logged in and set a display name'),
    },
  },
  async ({ slug, wallet_address }) => {
    try {
      return api.asToolResult(await api.adminAddAppOwner(slug, wallet_address))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'admin_remove_app_owner',
  {
    description: 'Unlink a wallet from an app\'s ownership (requires admin token). Unlike the self-service endpoint, this can remove the last owner, fully unclaiming the app.',
    inputSchema: {
      slug: z.string().describe('App slug'),
      wallet_address: z.string().describe('Wallet to remove'),
    },
  },
  async ({ slug, wallet_address }) => {
    try {
      return api.asToolResult(await api.adminRemoveAppOwner(slug, wallet_address))
    } catch (error) {
      return toolError(error)
    }
  },
)
```

- [ ] **Step 5: Build the MCP server**

```bash
cd mcp && npm run build 2>&1 | tail -40
```
Expected: clean TypeScript build.

- [ ] **Step 6: Update `mcp/README.md`**

Update the tools table row:
```
| `admin_search_users` | Find a wallet to assign as `developer_wallet_address` |
```
to:
```
| `admin_search_users` | Find a wallet before `admin_add_app_owner` |
| `admin_add_app_owner`, `admin_remove_app_owner` | Manage which wallets can self-service edit an app |
```
And the two other lines mentioning `developer_wallet_address`:
```
There is **no** `submit_app` MCP tool. Public submission requires a **wallet session cookie** (`POST /api/apps/submit` after `POST /api/auth/verify`) — direct developers to `/submit` in the browser, or use `admin_create_app` / `admin_update_app` with `developer_wallet_address`.
```
to:
```
There is **no** `submit_app` MCP tool. Public submission requires a **wallet session cookie** (`POST /api/apps/submit` after `POST /api/auth/verify`) — direct developers to `/submit` in the browser, or use `admin_create_app` / `admin_update_app` plus `admin_add_app_owner` for agent workflows.
```
```
Public `get_app` / `list_apps` responses omit `submitter_contact`. Use `admin_list_apps` to see it. `developer_wallet_address` is included on public app objects (null when unclaimed).
```
to:
```
Public `get_app` / `list_apps` responses omit `submitter_contact`. Use `admin_list_apps` to see it. `owner_wallet_addresses` is included on public app objects (empty array when unclaimed).
```

- [ ] **Step 7: Commit**

```bash
git add mcp/src/api.ts mcp/src/index.ts mcp/README.md
git commit -m "MCP: replace developer_wallet_address with owner management tools"
```

---

### Task 10: `README.md` / `docs/DEV.md` / `AGENTS.md`

**Files:**
- Modify: `AGENTS.md`

- [ ] **Step 1: Update `AGENTS.md`'s admin field table and MCP tool references**

Change:
```
| `developer_wallet_address` | Nimiq address or `null` | Links app to an owner (My apps, request-update). Use `admin_search_users` to find wallets. |
```
to:
```
| `owner_wallet_addresses` | array of Nimiq addresses (read-only) | Wallets that can self-service manage this app (My apps, request-update). Manage via `admin_add_app_owner` / `admin_remove_app_owner`, not through create/update. |
```
Change:
```
| `admin_search_users` | Find a wallet to assign as `developer_wallet_address` |
```
to:
```
| `admin_search_users` | Find a wallet before `admin_add_app_owner` |
| `admin_add_app_owner`, `admin_remove_app_owner` | Link/unlink a wallet's self-service access to an app |
```
Change:
```
There is **no** `submit_app` MCP tool. Public submit needs a wallet session cookie (`POST /api/auth/verify` then `POST /api/apps/submit`); use `admin_create_app` / `admin_update_app` with `developer_wallet_address` for agent workflows.
```
to:
```
There is **no** `submit_app` MCP tool. Public submit needs a wallet session cookie (`POST /api/auth/verify` then `POST /api/apps/submit`); use `admin_create_app` / `admin_update_app` plus `admin_add_app_owner` for agent workflows.
```
Change:
```
Public `get_app` / `list_apps` responses omit `submitter_contact`. Use `admin_list_apps` to see it. `developer_wallet_address` is included on public app objects (null when unclaimed).
```
to:
```
Public `get_app` / `list_apps` responses omit `submitter_contact`. Use `admin_list_apps` to see it. `owner_wallet_addresses` is included on public app objects (empty array when unclaimed).
```

- [ ] **Step 2: Commit**

```bash
git add AGENTS.md
git commit -m "Document owner_wallet_addresses and owner tools in AGENTS.md"
```

---

### Task 11: Frontend API client

**Files:**
- Modify: `frontend/src/api.ts`

**Interfaces:**
- Produces: `App.owner_wallet_addresses: string[]`, `addAppOwner(slug, wallet)`, `removeAppOwner(slug, wallet)`, `adminAddAppOwner(slug, wallet)`, `adminRemoveAppOwner(slug, wallet)`. Tasks 12–15 depend on these.

- [ ] **Step 1: Replace the `App` interface field**

Change:
```ts
  developer_wallet_address: string | null
```
to:
```ts
  owner_wallet_addresses: string[]
```

- [ ] **Step 2: Update `normalizeApp`**

Change:
```ts
    developer_wallet_address: raw.developer_wallet_address ?? null,
```
to:
```ts
    owner_wallet_addresses: raw.owner_wallet_addresses ?? [],
```

- [ ] **Step 3: Add self-service owner functions**

Add near `requestAppUpdate`:
```ts
export const addAppOwner = (slug: string, walletAddress: string) =>
  request<{ status: string }>(`/api/apps/${encodeURIComponent(slug)}/owners`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ wallet_address: walletAddress }),
  })

export const removeAppOwner = (slug: string, walletAddress: string) =>
  request<{ status: string }>(
    `/api/apps/${encodeURIComponent(slug)}/owners/${encodeURIComponent(walletAddress)}`,
    { method: 'DELETE', credentials: 'include' },
  )
```

- [ ] **Step 4: Add admin owner functions**

Add near `adminSearchUsers`:
```ts
export const adminAddAppOwner = (slug: string, walletAddress: string) =>
  adminRequest<{ status: string }>(`/api/admin/apps/${slug}/owners`, {
    method: 'POST',
    body: JSON.stringify({ wallet_address: walletAddress }),
  })

export const adminRemoveAppOwner = (slug: string, walletAddress: string) =>
  adminRequest<{ status: string }>(
    `/api/admin/apps/${slug}/owners/${encodeURIComponent(walletAddress)}`,
    { method: 'DELETE' },
  )
```

- [ ] **Step 5: Typecheck**

```bash
cd frontend && npx vue-tsc --noEmit 2>&1 | head -60
```
Expected: errors at every remaining use of `developer_wallet_address` — that's the checklist for Tasks 12–15.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api.ts
git commit -m "Add owner management API client functions; owner_wallet_addresses replaces developer_wallet_address"
```

---

### Task 12: `utils/wallet.ts` — membership check

**Files:**
- Modify: `frontend/src/utils/wallet.ts`

- [ ] **Step 1: Change `walletOwnsApp` to an array-membership check**

Change:
```ts
export function walletOwnsApp(walletAddress: string | null | undefined, appWallet: string | null | undefined): boolean {
  if (!walletAddress || !appWallet) return false
  return normalizeWalletAddress(walletAddress) === normalizeWalletAddress(appWallet)
}
```
to:
```ts
export function walletOwnsApp(walletAddress: string | null | undefined, ownerWalletAddresses: string[] | null | undefined): boolean {
  if (!walletAddress || !ownerWalletAddresses?.length) return false
  const normalized = normalizeWalletAddress(walletAddress)
  return ownerWalletAddresses.some((owner) => normalizeWalletAddress(owner) === normalized)
}
```

- [ ] **Step 2: Update the one call site**

In `frontend/src/views/AppsView.vue`, change:
```ts
          :owned="walletOwnsApp(walletAddress, app.developer_wallet_address)"
```
to:
```ts
          :owned="walletOwnsApp(walletAddress, app.owner_wallet_addresses)"
```

- [ ] **Step 3: Typecheck**

```bash
cd frontend && npx vue-tsc --noEmit 2>&1 | head -60
```
Expected: `AppsView.vue`'s error from Task 11 Step 5 is gone; remaining errors only in `SubmitView.vue`, `MyAppsView.vue`, `AdminView.vue` (Tasks 13–15) plus any pre-existing unrelated ones.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/utils/wallet.ts frontend/src/views/AppsView.vue
git commit -m "walletOwnsApp: array-membership check for multi-owner apps"
```

---

### Task 13: `SubmitView.vue` copy update

**Files:**
- Modify: `frontend/src/views/SubmitView.vue`

- [ ] **Step 1: Update the "linked to your wallet" copy**

Change:
```html
      <p v-if="walletAddress" class="text-xs text-muted">
        Submitting as <span class="font-mono">{{ walletAddress }}</span> — this app will be linked to your wallet as the developer of record; admins can reassign it later.
      </p>
```
to:
```html
      <p v-if="walletAddress" class="text-xs text-muted">
        Submitting as <span class="font-mono">{{ walletAddress }}</span> — this app will be linked to your wallet as the developer of record. You can link additional wallets (like a second device) from
        <RouterLink to="/my-apps" class="font-semibold text-accent-ink hover:underline">My apps</RouterLink> afterward.
      </p>
```

- [ ] **Step 2: Manually verify**

Visit `/submit` logged in — the updated copy renders with a working link to `/my-apps`.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/SubmitView.vue
git commit -m "Mention linking additional wallets from My apps"
```

---

### Task 14: `MyAppsView.vue` — manage owners

**Files:**
- Modify: `frontend/src/views/MyAppsView.vue`

**Interfaces:**
- Consumes: `addAppOwner`, `removeAppOwner` (Task 11).

- [ ] **Step 1: Add owner-management state and functions**

In `frontend/src/views/MyAppsView.vue`, change the script block from:
```ts
<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getMyApps, type App } from '../api'
import AppCard from '../components/AppCard.vue'

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
```
to:
```ts
<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getMyApps, addAppOwner, removeAppOwner, type App } from '../api'
import AppCard from '../components/AppCard.vue'

const { walletAddress, checking } = useWalletAuth()

const apps = ref<(App & { has_pending_revision: boolean })[]>([])
const loading = ref(true)
const error = ref('')

const expandedSlug = ref('')
const newOwnerInput = reactive<Record<string, string>>({})
const ownerError = reactive<Record<string, string>>({})
const ownerBusy = reactive<Record<string, boolean>>({})

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

function toggleManageOwners(slug: string) {
  expandedSlug.value = expandedSlug.value === slug ? '' : slug
}

async function handleAddOwner(slug: string) {
  const wallet = (newOwnerInput[slug] || '').trim()
  if (!wallet) return
  ownerBusy[slug] = true
  ownerError[slug] = ''
  try {
    await addAppOwner(slug, wallet)
    newOwnerInput[slug] = ''
    await load()
  } catch (err) {
    ownerError[slug] = err instanceof Error ? err.message : 'Failed to add owner'
  } finally {
    ownerBusy[slug] = false
  }
}

async function handleRemoveOwner(slug: string, wallet: string) {
  ownerBusy[slug] = true
  ownerError[slug] = ''
  try {
    await removeAppOwner(slug, wallet)
    await load()
  } catch (err) {
    ownerError[slug] = err instanceof Error ? err.message : 'Failed to remove owner'
  } finally {
    ownerBusy[slug] = false
  }
}

watch([checking, walletAddress], () => {
  if (!checking.value) void load()
}, { immediate: true })
</script>
```

- [ ] **Step 2: Add the "Manage owners" UI below each `AppCard`**

Change:
```html
    <div v-else class="grid gap-4 sm:grid-cols-2">
      <AppCard
        v-for="app in apps"
        :key="app.id"
        :app="app"
        owned
        :pending-update="app.has_pending_revision"
        show-manage-actions
      />
    </div>
```
to:
```html
    <div v-else class="grid gap-4 sm:grid-cols-2">
      <div v-for="app in apps" :key="app.id" class="space-y-2">
        <AppCard
          :app="app"
          owned
          :pending-update="app.has_pending_revision"
          show-manage-actions
        />
        <button type="button" class="text-xs font-semibold text-accent-ink hover:underline"
          @click="toggleManageOwners(app.slug)">
          {{ expandedSlug === app.slug ? 'Hide owners' : `Manage owners (${app.owner_wallet_addresses.length})` }}
        </button>
        <div v-if="expandedSlug === app.slug" class="space-y-2 rounded-xl border border-line bg-surface-2/50 p-3 text-sm">
          <ul class="space-y-1">
            <li v-for="wallet in app.owner_wallet_addresses" :key="wallet" class="flex items-center justify-between gap-2">
              <span class="truncate font-mono text-xs">{{ wallet }}</span>
              <button type="button" :disabled="ownerBusy[app.slug] || app.owner_wallet_addresses.length <= 1"
                class="shrink-0 text-xs font-semibold text-red-600 hover:underline disabled:cursor-default disabled:opacity-40 dark:text-red-400"
                @click="handleRemoveOwner(app.slug, wallet)">
                Remove
              </button>
            </li>
          </ul>
          <div class="flex gap-2">
            <input v-model="newOwnerInput[app.slug]" placeholder="Wallet address (e.g. your other device)"
              class="min-w-0 flex-1 rounded-lg border border-line bg-surface px-2 py-1.5 text-xs outline-none focus:border-accent" />
            <button type="button" :disabled="ownerBusy[app.slug]"
              class="shrink-0 rounded-lg bg-accent px-3 py-1.5 text-xs font-semibold text-white disabled:opacity-50"
              @click="handleAddOwner(app.slug)">
              Add
            </button>
          </div>
          <p v-if="ownerError[app.slug]" class="text-xs text-red-600 dark:text-red-400">{{ ownerError[app.slug] }}</p>
        </div>
      </div>
    </div>
```

- [ ] **Step 3: Typecheck**

```bash
cd frontend && npx vue-tsc --noEmit 2>&1 | head -60
```
Expected: `MyAppsView.vue` clean; remaining errors only in `AdminView.vue` (Task 15) plus pre-existing unrelated ones.

- [ ] **Step 4: Manually verify**

Log in as an owner of an app with only one owner — "Remove" is disabled on that sole owner. Add a second wallet address (one that's logged in and has a display name) — appears in the list, "Remove" becomes enabled on both. Remove one — back to one owner, disabled again. Try adding a wallet with no display name — inline error from the server.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/views/MyAppsView.vue
git commit -m "Add owner management UI to My Apps"
```

---

### Task 15: `AdminView.vue` — multi-owner section

**Files:**
- Modify: `frontend/src/views/AdminView.vue`

**Interfaces:**
- Consumes: `adminAddAppOwner`, `adminRemoveAppOwner` (Task 11).

- [ ] **Step 1: Remove `developer_wallet_address` from form state**

Remove this line from `emptyForm`:
```ts
  developer_wallet_address: null as string | null,
```
Remove this line from `startEdit`'s `Object.assign(form, {...})`:
```ts
    developer_wallet_address: app.developer_wallet_address,
```
Remove this line from `submit()`'s `payload`:
```ts
    developer_wallet_address: form.developer_wallet_address || null,
```

- [ ] **Step 2: Replace the picker's local-state model with immediate owner management**

Change:
```ts
function onDeveloperQueryInput() {
  form.developer_wallet_address = null
  clearTimeout(developerSearchTimer)
  developerSearchTimer = setTimeout(async () => {
    const q = developerQuery.value.trim()
    if (!q) {
      developerResults.value = []
      return
    }
    try {
      developerResults.value = await adminSearchUsers(q)
    } catch {
      developerResults.value = []
    }
  }, 250)
}

function pickDeveloper(user: AdminUserResult) {
  if (!user.display_name?.trim()) {
    error.value = 'This user must set a display name on their profile before they can own an app.'
    return
  }
  error.value = ''
  form.developer_wallet_address = user.wallet_address
  developerQuery.value = user.display_name
  developerResults.value = []
  form.developer_name = user.display_name
  form.developer_slug = slugify(user.display_name)
}

const slugify = (s: string) =>
  s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')

const developerLinked = computed(() => !!form.developer_wallet_address)
const developerPickPending = computed(
  () => developerQuery.value.trim() !== '' && !form.developer_wallet_address,
)
```
to:
```ts
function onDeveloperQueryInput() {
  clearTimeout(developerSearchTimer)
  developerSearchTimer = setTimeout(async () => {
    const q = developerQuery.value.trim()
    if (!q) {
      developerResults.value = []
      return
    }
    try {
      developerResults.value = await adminSearchUsers(q)
    } catch {
      developerResults.value = []
    }
  }, 250)
}

const currentOwners = ref<string[]>([])
const ownerBusy = ref(false)

async function addOwnerFromPicker(user: AdminUserResult) {
  if (!user.display_name?.trim()) {
    error.value = 'This user must set a display name on their profile before they can own an app.'
    return
  }
  error.value = ''
  ownerBusy.value = true
  try {
    await adminAddAppOwner(editingSlug.value, user.wallet_address)
    if (!currentOwners.value.includes(user.wallet_address)) currentOwners.value.push(user.wallet_address)
    if (!form.developer_name) form.developer_name = user.display_name
    if (!form.developer_slug) form.developer_slug = slugify(user.display_name)
    developerQuery.value = ''
    developerResults.value = []
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    ownerBusy.value = false
  }
}

async function removeOwnerFromApp(wallet: string) {
  ownerBusy.value = true
  error.value = ''
  try {
    await adminRemoveAppOwner(editingSlug.value, wallet)
    currentOwners.value = currentOwners.value.filter((w) => w !== wallet)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    ownerBusy.value = false
  }
}

const slugify = (s: string) =>
  s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')
```
Add `adminAddAppOwner, adminRemoveAppOwner,` to the import list from `'../api'`.

- [ ] **Step 3: Populate/reset `currentOwners` in `startCreate`/`startEdit`**

In `startCreate`, add:
```ts
  currentOwners.value = []
```
In `startEdit`, remove `developer_wallet_address: app.developer_wallet_address,` from the `Object.assign` (done in Step 1) and add:
```ts
  currentOwners.value = [...app.owner_wallet_addresses]
```
Change the `developerQuery.value = ...` line right after from:
```ts
  developerQuery.value = app.developer_wallet_address
    ? (app.developer_name || app.developer_wallet_address)
    : ''
```
to:
```ts
  developerQuery.value = ''
```
(the query box is now purely a search-to-add field, not a display of the current link — the owners list below shows current state instead).

- [ ] **Step 4: Guard `submit()` against stray developer_wallet_address handling**

Confirm `submit()`'s `payload` object (after Step 1's removal) has no ownership field left — it should just be the remaining `...form` spread plus the existing non-ownership overrides. No code change needed here beyond Step 1's removal; this step is a read-through check, not an edit.

- [ ] **Step 5: Replace the template's owner section**

Change:
```html
          <label class="relative block text-sm">
            <span class="mb-1 block font-semibold text-muted">Wallet owner</span>
            <input v-model="developerQuery" @input="onDeveloperQueryInput"
              placeholder="Search by display name or wallet address — pick a result to link"
              class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 focus:border-accent"
              :class="developerPickPending ? 'border-amber-500/60' : developerLinked ? 'border-emerald-500/40' : ''" />
            <p v-if="developerPickPending" class="mt-1 text-xs text-amber-700 dark:text-amber-200">
              Choose a user from the list below (they must have connected their wallet on the site first).
            </p>
            <p v-else-if="!developerLinked" class="mt-1 text-xs text-muted">
              Unclaimed — only admins can edit this listing until a wallet is linked.
            </p>
            <ul v-if="developerResults.length" class="absolute z-10 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border border-line bg-surface shadow-lg">
              <li v-for="user in developerResults" :key="user.wallet_address"
                @click="pickDeveloper(user)"
                class="cursor-pointer px-3 py-2 text-sm hover:bg-surface-2"
                :class="{ 'opacity-50': !user.display_name?.trim() }">
                {{ user.display_name ?? 'No display name' }}
                <span class="block font-mono text-xs text-muted">{{ user.wallet_address }}</span>
              </li>
            </ul>
            <p v-else-if="developerQuery.trim() && !developerLinked" class="mt-1 text-xs text-muted">
              No matching wallets — the user must log in at least once before you can assign them.
            </p>
            <span v-if="form.developer_wallet_address" class="mt-1 block text-xs text-emerald-700 dark:text-emerald-300">
              Linked to <span class="font-mono">{{ form.developer_wallet_address }}</span>
              <button type="button" @click="form.developer_wallet_address = null; developerQuery = ''" class="ml-1 text-accent-ink hover:underline">clear</button>
            </span>
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="text-sm">
              <span class="mb-1 block text-muted">Public developer name *</span>
              <input v-model="form.developer_name" required
                placeholder="Shown on listings"
                class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
              <p v-if="developerLinked" class="mt-1 text-xs text-muted">Pre-filled from the linked wallet's profile — edit freely to rebrand.</p>
            </label>
            <label class="text-sm">
              <span class="mb-1 block text-muted">Public developer slug *</span>
              <input v-model="form.developer_slug" required
                placeholder="Used in developer URLs"
                class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
            </label>
          </div>
```
to:
```html
          <div v-if="editingSlug" class="space-y-2">
            <span class="mb-1 block text-sm font-semibold text-muted">Owner wallets</span>
            <ul v-if="currentOwners.length" class="space-y-1">
              <li v-for="wallet in currentOwners" :key="wallet" class="flex items-center justify-between gap-2 rounded-lg bg-surface px-2 py-1.5 text-sm">
                <span class="truncate font-mono text-xs">{{ wallet }}</span>
                <button type="button" :disabled="ownerBusy"
                  class="shrink-0 text-xs font-semibold text-red-600 hover:underline disabled:cursor-default disabled:opacity-40 dark:text-red-400"
                  @click="removeOwnerFromApp(wallet)">
                  Remove
                </button>
              </li>
            </ul>
            <p v-else class="text-xs text-muted">Unclaimed — only admins can edit this listing until a wallet is added.</p>
          </div>
          <label class="relative block text-sm">
            <span class="mb-1 block font-semibold text-muted">Add owner wallet</span>
            <input v-model="developerQuery" @input="onDeveloperQueryInput"
              :disabled="!editingSlug"
              placeholder="Search by display name or wallet address — pick a result to add"
              class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 focus:border-accent disabled:opacity-50" />
            <p v-if="!editingSlug" class="mt-1 text-xs text-muted">Save this app first, then add owner wallets.</p>
            <ul v-if="developerResults.length" class="absolute z-10 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border border-line bg-surface shadow-lg">
              <li v-for="user in developerResults" :key="user.wallet_address"
                @click="addOwnerFromPicker(user)"
                class="cursor-pointer px-3 py-2 text-sm hover:bg-surface-2"
                :class="{ 'opacity-50': !user.display_name?.trim() }">
                {{ user.display_name ?? 'No display name' }}
                <span class="block font-mono text-xs text-muted">{{ user.wallet_address }}</span>
              </li>
            </ul>
            <p v-else-if="developerQuery.trim()" class="mt-1 text-xs text-muted">
              No matching wallets — the user must log in at least once before you can add them.
            </p>
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="text-sm">
              <span class="mb-1 block text-muted">Public developer name *</span>
              <input v-model="form.developer_name" required
                placeholder="Shown on listings"
                class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
            </label>
            <label class="text-sm">
              <span class="mb-1 block text-muted">Public developer slug *</span>
              <input v-model="form.developer_slug" required
                placeholder="Used in developer URLs"
                class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
            </label>
          </div>
```

- [ ] **Step 6: Typecheck**

```bash
cd frontend && npx vue-tsc --noEmit 2>&1 | head -60
```
Expected: clean except the pre-existing unrelated `AppsView.vue` `number`/`string` errors confirmed before this feature started.

- [ ] **Step 7: Manually verify**

Create a new app in `/admin` — the "Add owner wallet" field is disabled with "Save this app first" until the app exists. Save, then edit it again — the field is now enabled; search and add a wallet — appears in "Owner wallets" immediately (no separate Save needed); remove it — disappears immediately. Confirm the main "Save" button no longer sends any ownership field (check the network request body).

- [ ] **Step 8: Commit**

```bash
git add frontend/src/views/AdminView.vue
git commit -m "Replace single-owner picker with multi-owner management in admin"
```

---

## Final check

- [ ] **Full backend build, vet, test**
```bash
cd backend && go build ./... && go vet ./... && go test ./...
```
- [ ] **MCP server build**
```bash
cd mcp && npm run build
```
- [ ] **Frontend typecheck**
```bash
cd frontend && npx vue-tsc --noEmit
```
- [ ] **End-to-end manual walkthrough**: submit a new app with wallet A → confirm A is sole owner → from My Apps, add wallet B as co-owner → log in as B, confirm B can see the app in My Apps and successfully request an update → as A, remove B → confirm B can no longer request updates (`403`) and A alone can't be removed (`409`) → as admin, force-remove A too, fully unclaiming the app → confirm `/my-apps` is now empty for both wallets and the app still displays normally on the public catalog with `owner_wallet_addresses: []`.
