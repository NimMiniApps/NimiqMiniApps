# App Credentials Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Let approved-app owners self-issue hashed API credentials, and let registry consumers verify those keys to receive the app client record.

**Architecture:** New `app_credentials` table (hash-at-rest). Owner wallet endpoints for issue/list/revoke. Service endpoint `POST /registry/credentials/verify` behind existing `registryServiceAuth`. Revocation emits `credential_revoked` on the existing `app_client_changes` feed. Issuance/verification gated by `isPublicStatus` (approved/verified/experimental).

**Tech Stack:** Go + pgx, existing walletAuthMiddleware / registryServiceAuth / rateLimiter, OpenAPI YAML + gen script, Vue My Apps UI.

**Design:** `docs/plans/2026-08-10-app-credentials-design.md`

---

### Task 1: Migration — `app_credentials` + change reason

**Files:**
- Create: `backend/migrations/023_app_credentials.sql`
- Modify: (none yet)

**Step 1: Write migration**

```sql
CREATE TABLE IF NOT EXISTS app_credentials (
    id           BIGSERIAL PRIMARY KEY,
    app_id       UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    key_prefix   TEXT NOT NULL UNIQUE,
    secret_hash  BYTEA NOT NULL,
    label        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by   TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS app_credentials_app_idx
    ON app_credentials (app_id) WHERE revoked_at IS NULL;

ALTER TABLE app_client_changes DROP CONSTRAINT IF EXISTS app_client_changes_reason_check;
ALTER TABLE app_client_changes ADD CONSTRAINT app_client_changes_reason_check
    CHECK (reason IN (
        'owner_transfer', 'scopes_changed', 'status_changed',
        'metadata_changed', 'credential_revoked'
    ));
```

**Step 2: Verify migrate runs**

Run: `cd backend && go test -count=1 -run TestLaunchOrigins ./...`
Expected: PASS (or skip if no DB; if DB present, migrate succeeds via testPool)

**Step 3: Commit** (only if user asked)

---

### Task 2: Credential crypto helpers + unit tests

**Files:**
- Create: `backend/credentials.go` (helpers only first)
- Create: `backend/credentials_test.go`

**Step 1: Failing tests for key format / hash / prefix**

- `generateAppCredential(slug)` → `nma_<slug>_<base64url(32 bytes)>`, prefix = `nma_<slug>_` + first 8 chars of secret part (or whole prefix string used for lookup — see design: key_prefix for index lookup)
- Design: key is `nma_<slug>_<32 bytes base64url>`; `key_prefix` is for humans/logs/lookup — store a stable unique prefix extracted from the key (recommend: `nma_<slug>_<first 12 chars of random part>` unique enough + UNIQUE constraint)
- `hashAppCredential(secret string) []byte` → SHA-256 of full secret
- Verify uses constant-time compare of hashes

**Step 2: Implement helpers**

```go
const (
	clientChangeCredentialRevoked = "credential_revoked"
	appCredentialPrefix           = "nma_"
)

func generateAppCredential(slug string) (plaintext, keyPrefix string, hash []byte, err error)
func hashAppCredential(plaintext string) []byte
func parseCredentialPrefix(plaintext string) (keyPrefix string, ok bool)
```

Key format: `nma_<slug>_<base64.RawURLEncoding of 32 random bytes>`
`key_prefix`: full string up to and including first 12 chars of the random part (unique, loggable). Lookup: extract prefix from presented key the same way, load row by `key_prefix`, compare full hash.

**Step 3: Tests pass**

Run: `cd backend && go test -count=1 -run 'TestGenerateAppCredential|TestHashAppCredential|TestParseCredentialPrefix' ./...`

---

### Task 3: Issue / list / revoke handlers (owner)

**Files:**
- Modify: `backend/credentials.go`
- Modify: `backend/credentials_test.go`
- Modify: `backend/handlers.go` (add limiters on server struct)
- Modify: `backend/main.go` (wire routes + limiters)

**Behavior:**
- `POST /api/apps/{slug}/credentials` — wallet auth + isOwner; body optional `{label}`; only if `isPublicStatus`; rate limit ~10/hour per wallet; return `{id, key, key_prefix, label, created_at, created_by}` once
- `GET /api/apps/{slug}/credentials` — wallet auth + isOwner; metadata only (no secret/hash)
- `DELETE /api/apps/{slug}/credentials/{id}` — wallet auth + isOwner; set `revoked_at=now()` if null; emit `credential_revoked`; 200 idempotent

**Step 1: Write integration tests** (issue success, forbidden non-owner, rejected status cannot issue, list hides secret, revoke + verify fails)

**Step 2: Implement handlers + routes**

```go
mux.HandleFunc("POST /api/apps/{slug}/credentials", walletAuthMiddleware(..., s.createAppCredential))
mux.HandleFunc("GET /api/apps/{slug}/credentials", walletAuthMiddleware(..., s.listAppCredentials))
mux.HandleFunc("DELETE /api/apps/{slug}/credentials/{id}", walletAuthMiddleware(..., s.revokeAppCredential))
```

**Step 3: Tests pass**

Run: `cd backend && go test -count=1 -run 'TestAppCredential' ./...`

---

### Task 4: Service verify endpoint

**Files:**
- Modify: `backend/credentials.go`
- Modify: `backend/credentials_test.go`
- Modify: `backend/main.go`
- Modify: `backend/registry.go` (export/use `clientChangeCredentialRevoked` constant if placed there)

**Behavior:**
- `POST /registry/credentials/verify` — `registryServiceAuth`; body `{ "key": "..." }`; rate limit failures per service token (~60/min); on success return `AppClientRecord` (+ app must still be public); touch `last_used_at` at most once/minute; 401 on bad/revoked/non-public

**Step 1: Tests** (valid key, revoked, unlisted status, bad key, missing auth)

**Step 2: Implement**

**Step 3: Tests pass**

Run: `cd backend && go test -count=1 -run 'TestVerifyAppCredential|TestAppCredential' ./...`

---

### Task 5: OpenAPI + docs

**Files:**
- Modify: `docs/openapi.yaml`
- Run: `./scripts/gen-openapi.sh`
- Modify: `README.md` (NimConnect registry bullet)
- Modify: `AGENTS.md` (quick reference)
- Modify: `docs/DEV.md` if curl examples section exists for registry
- Modify: `mcp/README.md` only if needed (no new MCP tool)

Schemas: `AppCredential`, `AppCredentialCreated`, `AppCredentialList`, `VerifyCredentialRequest`; extend change reason enum with `credential_revoked`.

**Step 1: Edit openapi + regenerate**

**Step 2: `go test -count=1 -run TestOpenAPI ./...`**

**Step 3: README/AGENTS one-liners**

---

### Task 6: Frontend — My Apps credentials UI

**Files:**
- Modify: `frontend/src/api.ts`
- Modify: `frontend/src/views/MyAppsView.vue`
- Modify: `frontend/src/i18n/messages.ts` (minimal strings)

**Behavior:** For public-status owned apps, expand panel to list credentials, issue (show plaintext once + copy), revoke. Match existing owners-panel styling.

**Step 1: API helpers**

**Step 2: UI section next to Manage owners**

**Step 3: Manual smoke / typecheck** `cd frontend && npm run build` (or `vue-tsc`/existing script)

---

### Task 7: Full verification

Run:
```bash
cd backend && go test -count=1 ./...
cd frontend && npm test 2>/dev/null || npm run build
./scripts/gen-openapi.sh && git diff --exit-code backend/openapi.yaml backend/openapi.json || true
```

Out of scope (per design): NimConnect/NimWorld consumer migration, per-credential scopes, OAuth user tokens.
