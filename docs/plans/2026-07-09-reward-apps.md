# Reward Apps Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add developer-declared reward asset metadata and an "Apps with rewards" discovery path.

**Architecture:** Store rewards as `reward_assets TEXT[] NOT NULL DEFAULT '{}'` on apps and app revisions. Expose it through public/admin API responses, submission/update bodies, collection filtering, frontend badges, and submit/admin/update forms.

**Tech Stack:** Go backend with PostgreSQL migrations, OpenAPI YAML generation, Vue 3 frontend, Vitest.

---

### Task 1: Backend Contract And Validation

**Files:**
- Modify: `backend/handlers.go`
- Modify: `backend/revisions.go`
- Modify: `backend/validate.go`
- Create: `backend/migrations/014_reward_assets.sql`
- Test: `backend/validate_test.go`
- Test: `backend/collections_test.go`

**Steps:**
1. Add failing tests for accepted and rejected `reward_assets`.
2. Add failing tests for `collection=rewards` and `?rewards=true` SQL helpers.
3. Add `RewardAssets []string` to app and revision structs.
4. Include `reward_assets` in SQL columns, inserts, updates, scans, and revision approval.
5. Add the migration.
6. Run `go test ./...`.

### Task 2: OpenAPI And Docs

**Files:**
- Modify: `docs/openapi.yaml`
- Generate: `backend/openapi.yaml`
- Generate: `backend/openapi.json`
- Modify: `README.md`
- Modify: `AGENTS.md`

**Steps:**
1. Document `reward_assets`, `rewards` query filter, and `collection=rewards`.
2. Run `./scripts/gen-openapi.sh`.
3. Keep README/AGENTS notes short and usage-focused.

### Task 3: Frontend Types And UI

**Files:**
- Modify: `frontend/src/api.ts`
- Create: `frontend/src/utils/rewards.ts`
- Test: `frontend/src/utils/rewards.test.ts`
- Create: `frontend/src/components/RewardBadge.vue`
- Modify: `frontend/src/components/AppCard.vue`
- Modify: `frontend/src/views/AppDetailView.vue`
- Modify: `frontend/src/views/AppsView.vue`
- Modify: `frontend/src/views/HomeView.vue`
- Modify: `frontend/src/i18n/messages.ts`

**Steps:**
1. Add a failing utility test for `Earn NIM` and multiple reward labels.
2. Add frontend normalization for missing `reward_assets`.
3. Add the reusable badge component and place it on cards/details.
4. Add collection label text for "Apps with rewards".
5. Run targeted Vitest, then `npm run build`.

### Task 4: Submit, Update, And Admin Forms

**Files:**
- Modify: `frontend/src/views/SubmitView.vue`
- Modify: `frontend/src/views/RequestUpdateView.vue`
- Modify: `frontend/src/views/AdminView.vue`
- Modify: `frontend/src/utils/revisionDiff.ts`

**Steps:**
1. Add `reward_assets` inputs next to existing `assets`.
2. Include `reward_assets` in submit/update/admin payloads.
3. Show reward assets in revision diffs.
4. Re-run frontend build.
