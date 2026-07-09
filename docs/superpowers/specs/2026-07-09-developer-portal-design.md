# Developer portal: wallet-owned apps + self-service updates

## Purpose

Right now a developer who submitted an app has no way to change it
themselves — the only path is filing an issue and waiting on manual admin
work. This spec ties app ownership to the wallet identity introduced in
[2026-07-09-profile-identicon-design.md](2026-07-09-profile-identicon-design.md)
(the `users` table, `display_name`), so a logged-in developer can see the
apps they own and submit edits to them directly, subject to the existing
admin approval flow.

Two backend pieces already exist and this spec builds on top of them rather
than replacing them:
- `POST /api/apps/{slug}/request-update` + the `app_revisions`
  pending/approve/reject pipeline (`revisions.go`) — already does
  "propose a change, admin approves it."
- `adminAuthMiddleware` — admin actions can already be authorized by a
  wallet allowlist (`ADMIN_WALLET_ADDRESSES`), not just the static token.

What's missing is ownership: nothing currently ties a wallet to the app(s)
it submitted, so `request-update` is wide open (any caller, any app) and
there's no "my apps" view.

## Data model

Add one nullable column to `apps`:

| column                    | type | notes                                                        |
|---------------------------|------|---------------------------------------------------------------|
| developer_wallet_address  | text | `REFERENCES users(wallet_address)`, nullable                  |

Existing apps get `NULL` (unclaimed). `developer_name` / `developer_slug`
columns are unchanged in shape — they're still what's displayed and grouped
on — but for wallet-owned apps they're populated from the owner's
`display_name` at submit time rather than typed into a free-text field.
Admins can still hand-edit them afterward (rebrands, anonymous/legacy apps).

Migration `013_developer_wallet.sql`: add the column + index.

## API changes

**`POST /api/apps/submit`** (existing route, unchanged path) — now wrapped
in `walletAuthMiddleware`.
- Request body drops `developer_name` / `developer_slug`; the server
  fills them from the caller's `users.display_name` (see slug derivation
  below).
- If the caller has no `display_name` set yet, reject with a message
  telling them to set one first (links to `/profile`, no new endpoint —
  reuses `PUT /api/profile` from the profile spec).
- Sets `developer_wallet_address` to the caller's address on insert.
- This flips submit from unauthenticated to wallet-required — a
  developer-facing breaking change, not just an internal detail. See
  "Docs to update" below.

**`POST /api/apps/{slug}/request-update`** — now wrapped in
`walletAuthMiddleware`. After loading the current app, compare
`app.developer_wallet_address` to the authenticated address:
- mismatch (including `NULL` owner) → `403`, "you don't own this app."
- match → proceeds, but the server ignores any `developer_name` /
  `developer_slug` in the request body and carries forward the app's
  current values into the created `app_revisions` row instead. Owners can
  change everything else about their app through this flow, but not their
  own developer identity — that stays an admin-only change (via the admin
  update endpoint below), so a compromised or careless edit request can't
  quietly rebrand who an app is attributed to.

### Developer slug derivation

`developer_slug` is derived from `display_name` at submit time using the
same slugification `submit.go`'s domain/app-slug path already needs
(lowercase, non-alphanumeric runs collapsed to `-`, trimmed), then checked
against `slugRe` (`^[a-z0-9]+(-[a-z0-9]+)*$` in `validate.go`). If the
result is empty (e.g. a display name that's entirely non-ASCII/symbols),
reject the submission and tell the user to pick a different display name.

Collisions: if the derived slug already belongs to a *different* wallet's
`developer_slug`, append `-2`, `-3`, etc. until unique — mirrors how
`app` slugs already handle collisions. A given wallet reuses its own
existing `developer_slug` across multiple apps rather than re-deriving
(look up any existing app row for this `developer_wallet_address` first).

Changing `display_name` later does **not** retroactively rename
`developer_slug` on already-submitted apps — slugs are assigned once, at
first submission, to avoid breaking existing app URLs/links that key off
`developer_slug`. `developer_name` (the display label, not the slug) can
still be refreshed by an admin if a developer rebrands.

**`GET /api/my/apps`** (new, wallet auth) — apps where
`developer_wallet_address = <caller>`. Each item includes `status` and a
`has_pending_revision` bool (reuses `hasPendingRevision` from
`revisions.go`). This powers the "My Apps" list.

**`GET /api/admin/users?q=`** (new, admin auth) — search `users` by
`display_name`/`wallet_address` prefix, for the admin developer picker.
Small, read-only, mirrors the existing `adminListApps` query shape.

**`PUT`/`PATCH /api/admin/apps/{slug}`** (existing, unchanged handler) —
extend the accepted body to include `developer_wallet_address` so admins
can assign/reassign ownership through the same generic update path.

## Frontend

- `frontend/src/api.ts`: `submitApp` and `requestAppUpdate` currently
  don't send `credentials: 'include'` (unlike the wallet-auth calls, e.g.
  `getProfile`, `submitAppReview`). Both need it added now that the
  endpoints they hit check the wallet session cookie.
- `SubmitAppView` (or equivalent submit form): gated behind
  `WalletLoginButton` login before the form is usable. Remove the
  free-text developer name input. Add a line of copy: "This app will be
  linked to your wallet (`{address}`) as the developer of record — admins
  can reassign it later." If `display_name` isn't set, show a prompt
  linking to `/profile` instead of the form.
- New `MyAppsView.vue` behind route `/my-apps` (wallet-gated, same pattern
  as `/profile`): lists apps from `GET /api/my/apps` with status badges
  and a "Request update" link into `RequestUpdateView` per app.
- `RequestUpdateView`: no UI restructuring — now requires login (reuses
  existing wallet-gated-page pattern); a `403` response surfaces as
  "this isn't your app to edit."
- Admin app form (`AdminView.vue` or wherever apps are created/edited):
  add a developer search/picker backed by `GET /api/admin/users?q=` that
  sets `developer_wallet_address`; falls back to today's free-text
  `developer_name` field for unclaimed/anonymous apps.

## Docs to update

Submit going from unauthenticated to wallet-required is a developer-facing
API contract change, not just an implementation detail, so it needs to
land alongside doc/spec updates rather than after:
- `docs/openapi.yaml` (+ regenerate `backend/openapi.yaml`/`.json`):
  `POST /api/apps/submit` gains the wallet-cookie security requirement and
  drops `developer_name`/`developer_slug` from the request schema; add
  `developer_wallet_address` to the `App` schema; document the new
  `GET /api/my/apps` and `GET /api/admin/users` routes.
- `README.md`, `docs/DEV.md`, `AGENTS.md`: update wherever they currently
  describe app submission as not requiring login.

## Testing

- Backend: `submitApp` rejects when caller has no `display_name` set;
  succeeds and stamps `developer_wallet_address` + derived
  `developer_name`/`developer_slug` when it does.
- Backend: `requestAppUpdate` returns `403` for a non-owner and for an
  unowned (`NULL`) app; succeeds for the matching owner — table-driven,
  following the existing test style in `revisions_test.go`.
- Backend: `GET /api/my/apps` returns only the caller's apps, with correct
  `has_pending_revision`.
- Backend: `GET /api/admin/users` search matches by prefix, admin-auth
  gated (reuses existing admin auth test helpers).
- Frontend: no new test infra beyond what the profile spec already
  established.

## Explicitly out of scope

- Claim flow for unclaimed legacy apps (developer proves ownership of an
  existing `NULL`-owner app). Flagged as a good follow-up; for now
  unclaimed apps are assigned manually by an admin via the picker.
- "Report a problem" button for non-owners to flag issues on apps they
  don't own — separate feature, not part of this spec.
- Multiple owners per app, or transferring ownership between two
  developer wallets without going through admin.
