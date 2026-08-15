# Competition Results & Cycle 1 Announcement Design

## Goal

Record per-cycle competition scores and places for catalog apps, surface them
on app cards and detail pages, and update the `/build` competition section for
Cycle 1 results (closed) plus Cycle 2 dates. Official council totals were not
published for every entry, so store real scores when known and use `0` as the
public placeholder until then.

## Decisions

- Store history in a join table (not a single score column, not frontend-only JSON).
- Embed `competition_results` on public app payloads so cards/detail need no
  extra fetch.
- Keep `competition_cycle` as membership (“entered this cycle”); results are
  scored history and can outlive a later cycle change.
- Seed Cycle 1: score `0` for every catalog app with `competition_cycle = 1`,
  places `1` / `2` / `3` for Nimiq Space / NimJump / NimQuest.
- Surfaces: app cards, app detail, `/build`. No dedicated scoreboard page, no
  home hero change, no import of proxy CSV scores as official results.

## Data model and API

Add table `app_competition_results`:

| Column   | Type                         | Notes                                      |
|----------|------------------------------|--------------------------------------------|
| `id`     | serial / uuid PK             | Existing project convention                |
| `app_id` | FK → `apps` ON DELETE CASCADE|                                            |
| `cycle`  | positive integer               |                                            |
| `score`  | integer ≥ 0                  | Placeholder `0` until official totals land |
| `place`  | nullable positive integer    | `1`–`3` for podium; null otherwise         |

Constraints:

- Unique `(app_id, cycle)`.
- Soft uniqueness for places per cycle (admin validation / clear error), not a
  hard DB unique on place, so corrections and edge cases stay possible.

Public and admin app responses include:

```json
"competition_results": [
  { "cycle": 1, "score": 0, "place": 1 }
]
```

Normalize a missing field to `[]` for deploy compatibility. Sort results by
`cycle` ascending in responses.

Admin write path (Bearer `ADMIN_TOKEN`): accept `competition_results` on
existing admin create/update of an app and replace that app’s result set for
the cycles included in the payload (upsert by cycle; omit a cycle to leave it
unchanged, or use an empty list only if we explicitly document “clear all” —
default is upsert-by-cycle, never wipe unspecified cycles). No separate
endpoint in v1. OpenAPI source (`docs/openapi.yaml`) is updated and
regenerated; MCP `admin_update_app` / `admin_create_app` document the field.

Validation: `cycle > 0`, `score >= 0` integer, `place` null or positive integer;
duplicate `(app_id, cycle)` → 400.

### Seed migration

1. Set `competition_cycle = 1` on `nimiq-space` if still null (winner is in the
   catalog without membership today).
2. Insert `{cycle: 1, score: 0, place: null}` for every app with
   `competition_cycle = 1` (idempotent / `ON CONFLICT DO NOTHING`).
3. Update places: `nimiq-space` → 1, `nimjump` → 2, `nimquest` → 3 for cycle 1.

Do not treat `frontend/miniapps_competition_scored.csv` as official scores.

## User interface

### App cards and detail

Extend the competition badge (or a sibling) to show cycle and score, e.g.
`Cycle 1 · 0`. When `place` is 1–3, show a compact place mark consistent with
the board visual language (prefer text/lamp treatment over emoji clutter).

Display selection:

1. Prefer the result whose `cycle` matches the app’s current `competition_cycle`.
2. Else the highest-`cycle` result.
3. After seed, Cycle 1 members always have a row; do not invent client-only
   zeros for unrelated apps.

On detail, list all `competition_results` when more than one cycle exists.

### `/build` competition section

Replace the generic “open competition” pitch with:

1. Cycle 1 closed: 62 submissions, Community Council scored, results final.
2. Winners strip: 1st Nimiq Space, 2nd NimJump, 3rd NimQuest — links to catalog
   detail pages and
   https://www.nimiq.com/blog/mini-apps-competition-cycle-1-winner-announcement
3. Cycle 2: opens August 24, submissions close September 18 23:59 UTC, same
   $17,000 USDT pool, four Sip & Ship calls; CTAs to register / Skool / submit.
4. Non-winners may refine and resubmit; prize winners cannot resubmit the same
   winning app (per competition rules; short reminder only).

No new route.

### Admin UI

Thin section on the existing admin app edit form: cycle, score, place for
results (enough to replace curl). Full multi-row editor is acceptable if cheap;
otherwise one “primary” cycle row plus API for bulk history.

## Data flow

List/get app handlers join or batch-load results and attach
`competition_results`. Catalog cards use the embedded array. Admins upsert
scores when official numbers exist; public reads remain unauthenticated.

Developer submit/update does **not** set scores or places — moderators only.

## Testing

- Backend: seed leaves Cycle 1 apps at score 0 with correct places; app JSON
  includes results; admin upsert; invalid cycle/score/place rejected.
- Frontend: badge/detail rendering; api normalizer `[]` default; `/build`
  winners and Cycle 2 copy.
- OpenAPI sync check; regenerate via `./scripts/gen-openapi.sh`.

## Documentation

Update README, AGENTS.md, and MCP README briefly for `competition_results` and
the admin write path. Link deeper detail to this design and OpenAPI.

## Out of scope

- Dedicated public scoreboard / `/competition` page
- Home hero announcement
- Importing proxy CSV totals as official scores
- Automatic score sync from external competition systems
