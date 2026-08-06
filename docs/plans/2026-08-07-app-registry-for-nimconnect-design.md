# App registry for NimConnect authorization

Status: draft — needs approval
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

## What the registry must expose

### 1. App records for the authorization server

`GET /api/apps/{slug}/client` — server-to-server, NimConnect's service key.

Returns exactly what a consent screen and a scope check need, nothing else:

- `app_id` (stable — see below), `name`, `icon_url`
- `verified` and `status` (approved / rejected / pending)
- `declared_scopes` — the scopes this app may ever request
- `launch_origins` — exact-match origins derived from `website_url` /
  `open_url`, not a fuzzy field

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

Either a webhook to NimConnect or a pollable `updated_at` on the client record;
webhook preferred, since the security value is in the promptness.

### 3. Stable app id

Grants and awards are keyed on app id ecosystem-wide, so it must never be
reused or repointed. `slug` is user-facing and may change; the registry should
expose an immutable `app_id` and treat slug as a label. If slugs are currently
the de-facto key, this is the moment to fix it — before grants and achievements
reference them.

## What this changes for the review flow

Approving a listing today grants visibility. Under this, approval also
determines whether an app can ask users for identity scopes.

- The existing `verify` flow stops being cosmetic and becomes the signal shown
  on consent screens.
- `declared_scopes` should be reviewable: an app requesting `friends:read` or
  future trade scopes deserves more scrutiny than one requesting nothing, and
  the listing page should show the ask **before** install.
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

## Open questions

- Does app registration also issue the award-posting credential, or does
  NimConnect issue that separately? One registration is friendlier; two keeps
  the blast radius separate.
- Do we need a per-app "this app may request scopes at all" flag, distinct from
  approval — i.e. can an app be listed but not authorized?
