# App registry for NimConnect authorization

Status: approved
Date: 2026-08-07

Companion docs, one per repo:
- `NimConnect/docs/plans/2026-08-07-ecosystem-awards-and-app-registry-design.md` — the authorization server
- `NimWorld/docs/plans/2026-08-07-cross-app-platform-design.md` — the first consumer

## Context

NimConnect issues audience-scoped authorizations, but only to "apps owned by
the project" — a fixed list. For an ecosystem, any listed app should be able to
ask a user for scopes, and a user should be able to see and revoke what they
granted.

This catalog already owns everything an authorization server needs to know
about an app: identity, ownership, verification, name and icon, and a review
flow. That makes it the natural **client registry** in the standard split —
NimiqMiniApps identifies apps, NimConnect authorizes users.

We already store part of the picture: NimWorld's app manifest carries
`nimconnect.requestedScopes`, and `/api/apps/{slug}` serves app metadata. What
is missing is a contract NimConnect can rely on, and the events that keep it
honest.

## Locked decisions

- **No `can_request_scopes` flag (1A).** Status plus `declared_scopes` encodes
  everything a flag would. A second switch is a second place for the two to
  disagree. Fail closed on every axis: unknown app, `submitted`/`rejected`, or
  empty `declared_scopes` → no scopes, no exceptions.
- **`declared_scopes` on `app_revisions`.** Scope changes go through the existing
  review flow; review is the gate, not a flag.
- **Resolvable statuses for `/client`:** `approved`, `verified`, `experimental`.
  Experimental apps resolve, but the consent screen must visually separate
  unverified from verified (NimConnect UI; we expose raw `status` and
  `verified: status == "verified"`).
- **Poll first, webhook later (2B).** Ship
  `GET /api/apps/changed?since=<ts>` as the convergence guarantee. Webhook-only
  is ruled out — a dropped delivery leaves the mirror stale indefinitely.
  Webhook may come later as a latency optimization; the poll stays the backstop.
- **Award-posting credentials** are issued by NimConnect separately, not by
  catalog registration.
- **`app_id`** is the existing immutable `apps.id` UUID. Slug stays a label.

## What the registry exposes

### 1. App records for the authorization server

`GET /api/apps/{slug}/client` — server-to-server, `NIMCONNECT_SERVICE_TOKEN`.

Returns exactly what a consent screen and a scope check need, nothing else:

- `app_id` (stable UUID), `slug`, `name`, `icon_url`
- `verified` and `status`
- `declared_scopes` — the scopes this app may ever request
- `launch_origins` — exact-match origins derived from `https://{domain}` and
  the origin of `website_url` when set (not the NimPay host `open_url`)
- `updated_at`

NimConnect mirrors these records so consent keeps working when we are down; it
must not depend on our uptime to authorize a user.

### 2. Change events

A mirrored record goes stale silently, and two kinds of staleness are security
bugs rather than cosmetic ones:

- **Ownership transfer** (`/api/apps/{slug}/owners`) — a grant is user→app. If
  an app is sold or transferred, NimConnect must force re-consent rather than
  hand the new owner everyone's existing grants.
- **Scope changes** in an approved revision — an app that adds
  `marketplace:trade` must not inherit consent granted for something narrower.
- **Status changes** — delist / reject / re-approve.

v1 ships a single poll request, not N:

`GET /api/apps/changed?since=<RFC3339>` → `{ app_id, slug, updated_at, reasons[] }`
where reasons are `owner_transfer` / `scopes_changed` / `status_changed` /
`metadata_changed`.

### 3. Stable app id

Grants and awards are keyed on app id ecosystem-wide. The catalog already has
an immutable UUID (`apps.id` from migration 001); expose it as `app_id` and
treat slug as a label.

## What this changes for the review flow

Approving a listing today grants visibility. Under this, approval also
determines whether an app can ask users for identity scopes.

- The existing `verify` flow stops being cosmetic and becomes the signal shown
  on consent screens.
- `declared_scopes` is reviewable via `app_revisions`; an app requesting
  `friends:read` or future trade scopes deserves more scrutiny than one
  requesting nothing, and the listing page should show the ask **before**
  install.
- Rejected or unlisted apps must not resolve through `/client`.

## Later: achievements on app pages

Once NimConnect serves awards keyed by app id, an app page can show its
achievements, and a profile can show which of them a player has earned. That is
the payoff of one stable id across all three services — it needs no work here
beyond the id contract above, and is listed only so the id decision is made
with it in mind.

## Non-goals

- The catalog does not store grants, tokens or scopes a user has approved.
  Those are NimConnect's, and duplicating them would create two answers to
  "what did I authorize".
- `/api/apps/{slug}/track` stays what it is: an anonymous, no-auth engagement
  counter. It cannot identify users and must not be extended to try — a
  client-supplied wallet field is exactly the pattern `/api/auth/verify`
  already refuses.
- No per-user library or favourites here yet. "Connected" comes from
  NimConnect's grants; a curated library is a separate, later idea and should
  not be conflated with it.
- Webhooks (v1). Poll is the guarantee.
- Award-posting credentials.

## Rider for NimConnect (not implemented here)

App status, ownership, and declared scopes must be checked **at token
validation**, not only at grant. Tokens live seven days. Blocking `/client`
only stops new grants — every existing token keeps working until it expires
unless NimConnect re-checks on use. See the companion NimConnect design doc.
