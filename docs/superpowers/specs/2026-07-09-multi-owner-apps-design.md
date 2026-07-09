# Multi-owner apps: link more than one wallet to an app

## Purpose

The developer portal ([2026-07-09-developer-portal-design.md](2026-07-09-developer-portal-design.md))
ties app ownership to a single wallet address. In practice a developer often
has more than one wallet — a desktop wallet and Nimiq Pay on mobile, say —
and today only the one wallet that owns the app can request updates to it.
This spec lets an app have multiple owning wallets, any of which can manage
it, while keeping everything else about the developer portal (edit review
flow, identity fields, admin override) unchanged.

## Data model

Replace the single `apps.developer_wallet_address` column with a join table:

```sql
CREATE TABLE app_owners (
    app_slug TEXT NOT NULL REFERENCES apps(slug) ON DELETE CASCADE,
    wallet_address TEXT NOT NULL REFERENCES users(wallet_address),
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (app_slug, wallet_address)
);
```

Migration: create the table, backfill one row per app from the existing
`developer_wallet_address` (skipping `NULL`s), then drop that column and
its index from `apps`.

`App.developer_wallet_address *string` becomes `App.owner_wallet_addresses
[]string` (possibly empty = unclaimed). It's populated via a correlated
subquery appended to `appColumns` —
`ARRAY(SELECT wallet_address FROM app_owners WHERE app_owners.app_slug =
apps.slug ORDER BY added_at) AS owner_wallet_addresses` — so every existing
`App`-returning endpoint gets it automatically, without restructuring the
dozen call sites that already use `appColumns`, and without N+1 queries.

`developer_name`/`developer_slug` are unaffected — they're already
decoupled from any single wallet (per the earlier fix to
`validateDeveloperWallet`), so multi-owner only changes *who can edit*, not
identity display.

`resolveDeveloperSlug` (in `backend/developer.go`) changes its "does this
wallet already have a slug" lookup from an equality check on
`developer_wallet_address` to a join through `app_owners`:
```sql
SELECT a.developer_slug FROM apps a
JOIN app_owners o ON o.app_slug = a.slug
WHERE o.wallet_address = $1 LIMIT 1
```
This means a wallet added as a co-owner of an existing app reuses that
app's `developer_slug` if it later submits a brand-new app of its own —
co-ownership is treated as proof of being part of that team.

Self-submission (`submitApp`) now performs two writes — insert the `apps`
row, then insert the submitter as the first `app_owners` row — wrapped in
one transaction (`s.pool.Begin`), so a failure between the two can't leave
an orphaned, unclaimed "submitted" app with no owner able to manage it.

## API

Self-service, wallet-auth, caller must already be a current owner of the app:
```
POST   /api/apps/{slug}/owners        { wallet_address }  -> add a co-owner
DELETE /api/apps/{slug}/owners/{wallet_address}            -> remove a co-owner
```
- `POST`: the target wallet must exist in `users` and have a `display_name`
  set (same validation `validateDeveloperWallet` already does today) —
  `400` otherwise. `403` if the caller isn't a current owner. Adding an
  already-current owner again is a no-op `200`, not an error.
- `DELETE`: `403` if the caller isn't a current owner. `409` if removal
  would leave zero owners — self-service can never fully unclaim an app;
  that guards against someone locking everyone (including themselves) out
  by accident.

Admin, admin-auth, no membership requirement:
```
POST   /api/admin/apps/{slug}/owners        { wallet_address }
DELETE /api/admin/apps/{slug}/owners/{wallet_address}
```
Same target-wallet validation, but no caller-is-owner check and no
"last owner" guardrail — an admin can fully unclaim an app. Both endpoint
pairs share one internal `addOwner`/`removeOwner` implementation in
`backend/developer.go`; only the guardrails around them differ.

`GET /api/apps/{slug}`, `GET /api/admin/apps`, and `GET /api/my/apps` need
no changes beyond the `appColumns` change above — they already return
`owner_wallet_addresses`.

`requestAppUpdate`'s ownership check (`backend/revisions.go`) changes from
comparing a single address to a membership query:
```sql
SELECT EXISTS(SELECT 1 FROM app_owners WHERE app_slug=$1 AND wallet_address=$2)
```

`updateApp` (admin's generic `PUT`/`PATCH /api/admin/apps/{slug}`) drops
all involvement with ownership — it no longer accepts or touches
`developer_wallet_address` at all; ownership is managed exclusively
through the new owner endpoints.

## Frontend

- `frontend/src/api.ts`: `App.developer_wallet_address: string | null`
  becomes `App.owner_wallet_addresses: string[]`. Add `addAppOwner(slug,
  address)`, `removeAppOwner(slug, address)` (wallet-auth, credentials
  included), and `adminAddAppOwner(slug, address)` /
  `adminRemoveAppOwner(slug, address)` (admin-auth).
- `frontend/src/utils/wallet.ts`: `walletOwnsApp` becomes an
  array-membership check (`app.owner_wallet_addresses` includes the
  normalized wallet address) instead of an equality check.
- `MyAppsView.vue`: each app gets an expandable "Manage owners" section —
  current owners shown as identicon + truncated address with a remove
  button each (the last one disabled, with the 409 message surfaced if
  somehow attempted), and an input to add another wallet address. Server
  validation errors (target has no display name, etc.) surface inline
  the same way every other form on this page already does.
- `AdminView.vue`: the current single wallet-picker-with-clear becomes a
  multi-owner section — owner chips with remove buttons, plus the
  existing search-and-pick input (`developerQuery`/`developerResults`/
  `adminSearchUsers`) to add more. Add/remove calls fire immediately
  (mirroring how `adminSetStatus` already works), not deferred to the
  "Save" button, since ownership no longer travels through the generic
  app-update payload. This section only renders in edit mode (an
  app needs a slug to attach owners to) — matches today's pattern of
  assigning ownership to admin-created apps after the fact.
- `SubmitView.vue`: one copy update noting additional wallets (e.g. a
  second device) can be linked afterward from My Apps.

## Testing

- Backend: no new pure functions to unit-test beyond what already exists
  (`slugifyDisplayName` is untouched). Verify `resolveDeveloperSlug`'s new
  join, `addOwner`/`removeOwner`'s guardrails (last-owner block for
  self-service, no block for admin), and the transactional submit-then-own
  insert manually against a local dev stack, per this repo's existing
  no-DB-integration-test convention.
- Frontend: no new test infra.

## Explicitly out of scope

- Per-owner permission levels (e.g. "can edit" vs "can manage owners") —
  every owner has identical rights, matching the mutual-trust model agreed
  on for this spec.
- Owner invitations/acceptance flow — adding a wallet is immediate and
  one-sided (any current owner can add any wallet that has a profile),
  same trust level as the existing wallet-picker had for a single owner.
- Any change to how `developer_name`/`developer_slug` are chosen or edited.
