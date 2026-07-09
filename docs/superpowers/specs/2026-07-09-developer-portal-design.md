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

**`POST /api/apps` (submit)** — now wrapped in `walletAuthMiddleware`.
- Request body drops `developer_name` / `developer_slug`; the server
  fills them from the caller's `users.display_name`.
- If the caller has no `display_name` set yet, reject with a message
  telling them to set one first (links to `/profile`, no new endpoint —
  reuses `PUT /api/profile` from the profile spec).
- Sets `developer_wallet_address` to the caller's address on insert.

**`POST /api/apps/{slug}/request-update`** — now wrapped in
`walletAuthMiddleware`. After loading the current app, compare
`app.developer_wallet_address` to the authenticated address:
- mismatch (including `NULL` owner) → `403`, "you don't own this app."
- match → proceeds exactly as today (creates a pending `app_revisions` row).

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
