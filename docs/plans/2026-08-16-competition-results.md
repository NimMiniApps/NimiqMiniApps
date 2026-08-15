# Competition Results Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Persist per-cycle competition scores/places, seed Cycle 1 placeholders (score 0 + podium places), show them on cards/detail, and update `/build` for Cycle 1 winners + Cycle 2 dates.

**Architecture:** New `app_competition_results` table; embed `competition_results` on app JSON; admin upsert via create/update body; public submit strips scores. Frontend badge + `/build` announcement; thin admin form fields.

**Tech Stack:** PostgreSQL migrations, Go HTTP API, OpenAPI 3.1, MCP Zod, Vue 3, Vitest.

**Design:** `docs/plans/2026-08-16-competition-results-design.md`

---

### Task 1: Migration + seed

**Files:**
- Create: `backend/migrations/024_competition_results.sql`

**Steps:** Create table with unique `(app_id, cycle)`, score ≥ 0, place null or > 0. Set `nimiq-space` cycle 1. Insert score 0 for all `competition_cycle = 1`. Set places 1/2/3 for `nimiq-space` / `nimjump` / `nimquest`.

### Task 2: Backend load + admin upsert + tests

**Files:**
- Create: `backend/competition_results.go`, `backend/competition_results_test.go`
- Modify: `backend/handlers.go`, `backend/submit.go`, `backend/validate.go` (if needed)

**Steps:** `CompetitionResult` type on `App`. After scan paths, attach results in batch. Validate results. Upsert on admin create/update. Strip on public submit. Tests for validation, attach, admin upsert, seed places.

### Task 3: OpenAPI + MCP

**Files:**
- Modify: `docs/openapi.yaml`, `mcp/src/index.ts`, `mcp/README.md`
- Run: `./scripts/gen-openapi.sh`

### Task 4: Frontend display + admin + `/build`

**Files:**
- Modify: `frontend/src/api.ts`, `frontend/src/utils/competition.ts`, `frontend/src/utils/competition.test.ts`, `frontend/src/components/CompetitionCycleBadge.vue`, `frontend/src/views/AppDetailView.vue`, `frontend/src/views/BuildView.vue`, `frontend/src/views/AdminView.vue`, `frontend/src/utils/catalogLinks.ts` (blog URL)

### Task 5: Docs + verify

**Files:** `README.md`, `AGENTS.md`

**Verify:** `go test ./...`, frontend tests/build, openapi sync.
