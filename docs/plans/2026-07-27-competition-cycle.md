# Competition Cycle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add optional competition-cycle metadata, backfill current Cycle 1 entries, show public cycle badges, and let visitors filter apps by cycle.

**Architecture:** Store `competition_cycle` as a nullable positive integer on apps and revisions. Carry it through every existing app write/read path, expose `competition_cycle=<number|none>` on the list endpoint, and normalize missing frontend values to `null`. Reuse a small frontend utility and badge component across cards/details; forms use an optional positive-number control so future cycles need no schema or UI release.

**Tech Stack:** PostgreSQL migrations, Go 1.24 HTTP API, OpenAPI 3.1, TypeScript MCP server with Zod, Vue 3, Vitest.

---

### Task 1: Backend model, validation, persistence, and filter

**Files:**
- Create: `backend/migrations/020_competition_cycle.sql`
- Create: `backend/competition_cycle_test.go`
- Modify: `backend/handlers.go`
- Modify: `backend/revisions.go`
- Modify: `backend/validate.go`
- Modify: `backend/validate_test.go`

**Step 1: Write failing validation and query-parser tests**

Add cases proving that `nil` and `1` are accepted, while `0` and `-1` are
rejected. Add table tests for a pure
`parseCompetitionCycleFilter(string) (cycle *int, withoutCycle bool, err error)`
helper:

- `""` returns no filter.
- `"1"` returns cycle 1.
- `"none"` requests `competition_cycle IS NULL`.
- `"0"`, `"-1"`, and `"abc"` return an error.

Also test `revisionToApp` to prove that a revision's cycle replaces the current
app cycle.

**Step 2: Run tests and verify they fail**

Run:

```bash
cd backend
go test ./... -run 'TestValidateApp|TestParseCompetitionCycleFilter|TestRevisionToAppCompetitionCycle'
```

Expected: compilation failures because the field and parser do not exist.

**Step 3: Add the migration and backend field**

Create migration 020:

```sql
ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS competition_cycle INTEGER
    CHECK (competition_cycle IS NULL OR competition_cycle > 0);

ALTER TABLE app_revisions
    ADD COLUMN IF NOT EXISTS competition_cycle INTEGER
    CHECK (competition_cycle IS NULL OR competition_cycle > 0);

UPDATE apps
SET competition_cycle = 1
WHERE slug IN (
    'xcrowhub', 'nimagent', 'nimhunt', 'nimiq-bazar', 'roundtrip',
    'nimjump', 'unlockmedia', 'nimiq-invoice-pay', 'nimiqstake',
    'nimconnect', 'tipwall', 'nimga', 'nimiq-radio', 'verilock'
);
```

Add `CompetitionCycle *int` with JSON name `competition_cycle` to `App` and
`AppRevision`. Add each column in the matching select/scan, insert, update,
revision insert, revision approval, and `revisionToApp` positions. Pointer
semantics distinguish an omitted cycle from Cycle 1.

In `validateApp`, reject non-null values below 1 with
`competition_cycle must be a positive integer`.

**Step 4: Implement list filtering**

Parse the query before constructing SQL. For a positive number append
`competition_cycle = $N`; for `none` append `competition_cycle IS NULL`.
Return HTTP 400 with
`competition_cycle must be a positive integer or none` on malformed input.

**Step 5: Format and rerun focused tests**

Run:

```bash
cd backend
gofmt -w handlers.go revisions.go validate.go validate_test.go competition_cycle_test.go
go test ./... -run 'TestValidateApp|TestParseCompetitionCycleFilter|TestRevisionToAppCompetitionCycle'
```

Expected: PASS.

**Step 6: Commit backend behavior**

```bash
git add backend/handlers.go backend/revisions.go backend/validate.go \
  backend/validate_test.go backend/competition_cycle_test.go \
  backend/migrations/020_competition_cycle.sql
git commit -m "Add competition cycle to catalog apps"
```

### Task 2: OpenAPI and project documentation

**Files:**
- Modify: `docs/openapi.yaml`
- Regenerate: `backend/openapi.yaml`
- Regenerate: `backend/openapi.json`
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `docs/DEV.md`

**Step 1: Add a failing OpenAPI contract assertion**

Extend `backend/openapi_test.go` to assert the generated JSON contains the
`competition_cycle` schema property and the list query parameter.

**Step 2: Run the OpenAPI test and verify it fails**

Run:

```bash
cd backend
go test ./... -run TestOpenAPI
```

Expected: FAIL because the source contract has no cycle field.

**Step 3: Update the source contract**

Add nullable integer `competition_cycle` with `minimum: 1` to app schemas,
submission bodies, and revision schemas. Add the string-form list parameter
documenting positive integers and `none`, with examples for both.

**Step 4: Regenerate embedded copies**

Run:

```bash
./scripts/gen-openapi.sh
```

Expected: `backend/openapi.yaml` and `backend/openapi.json` match
`docs/openapi.yaml`.

**Step 5: Update developer-facing docs**

Document:

- Optional `competition_cycle` in the README API/feature summary.
- The submission payload and pre-check meaning in `AGENTS.md`.
- A curl filter example in `docs/DEV.md`.

Keep the descriptions short and link to OpenAPI for details.

**Step 6: Verify and commit contracts**

Run:

```bash
cd backend
go test ./... -run TestOpenAPI
```

Expected: PASS.

```bash
git add docs/openapi.yaml backend/openapi.yaml backend/openapi.json \
  backend/openapi_test.go README.md AGENTS.md docs/DEV.md
git commit -m "Document competition cycle API"
```

### Task 3: MCP support

**Files:**
- Modify: `mcp/src/index.ts`
- Modify: `mcp/src/api.ts`
- Modify: `mcp/README.md`

**Step 1: Add the MCP field schema**

Define:

```ts
competition_cycle: z.number().int().positive().nullable().optional().describe(
  'Nimiq Mini Apps competition cycle; null or omitted means this is not a competition entry.',
),
```

Include it in `appFields`, which automatically covers `admin_create_app` and
the optional patch fields for `admin_update_app`.

Update API TypeScript types if they enumerate app request fields.

**Step 2: Document the MCP field**

Add `competition_cycle` to the field-notes table and mention that public
`list_apps` accepts the cycle filter if its current argument map is extended.
Extend `list_apps` with `competition_cycle` as a union of a positive integer
and `"none"` and serialize it to the REST query.

**Step 3: Build the MCP server**

Run:

```bash
cd mcp
npm run build
```

Expected: TypeScript build succeeds.

**Step 4: Commit MCP support**

```bash
git add mcp/src/index.ts mcp/src/api.ts mcp/README.md
git commit -m "Expose competition cycles through MCP"
```

### Task 4: Frontend types, cycle utility, and badge

**Files:**
- Create: `frontend/src/utils/competition.ts`
- Create: `frontend/src/utils/competition.test.ts`
- Create: `frontend/src/components/CompetitionCycleBadge.vue`
- Modify: `frontend/src/api.ts`
- Modify: `frontend/src/components/AppCard.vue`
- Modify: `frontend/src/views/AppDetailView.vue`
- Modify: `frontend/src/locales/en.ts`

**Step 1: Write failing utility tests**

Test:

```ts
expect(competitionCycleLabel(1)).toBe('Cycle 1')
expect(competitionCycleLabel(null)).toBe('')
expect(normalizeCompetitionCycle(undefined)).toBe(null)
expect(competitionCycleValues([{ competition_cycle: 2 }, { competition_cycle: 1 }, { competition_cycle: 2 }]))
  .toEqual([1, 2])
```

**Step 2: Run the focused test and verify it fails**

Run:

```bash
cd frontend
npm test -- src/utils/competition.test.ts
```

Expected: FAIL because the utility does not exist.

**Step 3: Implement types and normalization**

Add `competition_cycle: number | null` to `App` and `AppRevision`. Normalize
missing values to `null` in `normalizeApp`.

Implement label, normalization, and sorted-unique extraction helpers in
`competition.ts`.

**Step 4: Add the reusable badge**

Create `CompetitionCycleBadge.vue` with a required positive numeric `cycle`
prop and the same compact pill hierarchy used by release/reward badges. Render
it on public app cards and detail pages only when the cycle is non-null.

Add English strings for cycle labels and accessible naming to the existing
locale map rather than embedding user-facing filter text in logic.

**Step 5: Run tests and build**

Run:

```bash
cd frontend
npm test -- src/utils/competition.test.ts
npm run build
```

Expected: PASS.

**Step 6: Commit frontend primitives**

```bash
git add frontend/src/api.ts frontend/src/utils/competition.ts \
  frontend/src/utils/competition.test.ts \
  frontend/src/components/CompetitionCycleBadge.vue \
  frontend/src/components/AppCard.vue frontend/src/views/AppDetailView.vue \
  frontend/src/locales/en.ts
git commit -m "Show competition cycle badges"
```

### Task 5: Submission, update, admin, and catalog filter UI

**Files:**
- Modify: `frontend/src/views/SubmitView.vue`
- Modify: `frontend/src/views/RequestUpdateView.vue`
- Modify: `frontend/src/views/AdminView.vue`
- Modify: `frontend/src/views/AppsView.vue`
- Modify: `frontend/src/api.ts`
- Modify: `frontend/src/locales/en.ts`
- Test: `frontend/src/utils/competition.test.ts`

**Step 1: Extend failing query/value tests**

Add pure helper tests proving:

- Empty form values serialize to `null`.
- Positive numeric strings serialize to numbers.
- `"none"` is preserved as the public filter value.
- Cycle query state round-trips without treating `0` or junk as valid.

**Step 2: Run tests and verify they fail**

Run:

```bash
cd frontend
npm test -- src/utils/competition.test.ts
```

Expected: FAIL for the new helpers.

**Step 3: Add form controls**

Add an optional `type="number"`, `min="1"`, `step="1"` Competition cycle
field to submit, request-update, and admin forms. Store the editable value as
`number | ''`; serialize empty to `null`. Populate it from existing apps during
edit/update.

**Step 4: Add the public filter**

Add `competition_cycle` to `AppsView` route state, query synchronization,
request parameters, active-filter display, clear behavior, and watchers.
Populate a select with:

- All apps
- Not a competition entry (`none`)
- Sorted `Cycle N` values represented by the unfiltered public catalog

Load the distinct values once from the existing unpaginated `listApps()` call;
keep the currently selected value visible even if another filter produces no
matching rows.

**Step 5: Run frontend verification**

Run:

```bash
cd frontend
npm test
npm run build
```

Expected: all tests and the production build pass.

**Step 6: Commit UI flow**

```bash
git add frontend/src/views/SubmitView.vue \
  frontend/src/views/RequestUpdateView.vue frontend/src/views/AdminView.vue \
  frontend/src/views/AppsView.vue frontend/src/api.ts \
  frontend/src/locales/en.ts frontend/src/utils/competition.test.ts
git commit -m "Add competition cycle selection and filtering"
```

### Task 6: Full verification

**Files:**
- Verify all changed files

**Step 1: Check formatting and generated files**

Run:

```bash
git diff --check
./scripts/gen-openapi.sh
git diff --exit-code -- backend/openapi.yaml backend/openapi.json
```

Expected: no whitespace errors and no generated drift.

**Step 2: Run backend verification**

Run:

```bash
cd backend
go test ./...
go build ./...
```

Expected: PASS.

**Step 3: Run MCP verification**

Run:

```bash
cd mcp
npm run build
```

Expected: PASS.

**Step 4: Run frontend verification**

Run:

```bash
cd frontend
npm test
npm run build
```

Expected: PASS.

**Step 5: Inspect the final diff and repository state**

Run:

```bash
git status --short
git diff HEAD~5 --stat
```

Expected: only intentional competition-cycle changes are present and all
implementation commits are recorded.
