# App Credentials Design

Date: 2026-08-10
Status: proposed

## Goal

Let an approved app's owner issue its own API credential from the catalog, so
integrating with ecosystem services stops requiring a human to mint a key and
redeploy an env var.

Today every consumer keeps its own hand-maintained key list — NimConnect for
award posting, NimWorld's `APP_KEYS` for signed plaza events. Both are edited by
one person and shipped as configuration. That is the reason "a few apps want to
integrate" has not become "a few apps did": integration is gated on a favour,
not on a self-serve flow.

The registry is the right owner because it already knows everything the decision
needs: who owns the app (`app_owners`), whether the domain checks out
(`domaincheck.go`), what the app declares it needs (`declared_scopes`), and what
status review put it in.

## What already exists

Migration 022 and `registry.go` shipped the identity half of this:

- `registryServiceAuth` — a shared-secret service channel; NimConnect already
  authenticates to the registry with `nimconnectServiceToken`.
- `GET` app client record → `app_id`, `slug`, `name`, `icon_url`, `verified`,
  `declared_scopes`, `launch_origins`.
- `listAppClientChanges(since)` → a sync feed with typed reasons
  (`owner_transfer`, `scopes_changed`, `status_changed`, `metadata_changed`).

So the registry already publishes **who an app is and what it may ask for**. It
does not publish **how an app proves it is that app**. This design adds only
that, and reuses the two channels above for everything else.

## Storage

```sql
CREATE TABLE app_credentials (
    id          BIGSERIAL PRIMARY KEY,
    app_id      UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    key_prefix  TEXT NOT NULL UNIQUE,
    secret_hash BYTEA NOT NULL,
    label       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  TEXT NOT NULL,           -- wallet address of the issuing owner
    last_used_at TIMESTAMPTZ,
    revoked_at  TIMESTAMPTZ
);

CREATE INDEX app_credentials_app_idx ON app_credentials (app_id) WHERE revoked_at IS NULL;
```

Only the hash is stored. The plaintext secret is returned once, at creation, and
is unrecoverable afterwards — rotation is "issue a new one, revoke the old one",
which the multi-credential-per-app shape supports without downtime.

`last_used_at` is deliberately included: it is the only honest answer to "how
many apps are actually integrated", which is a number worth having.

### Key format

```
nma_<slug>_<32 bytes base64url>
```

The prefix is for humans reading logs and for the `key_prefix` index lookup. It
is **not** an authenticator — verification always compares the hash of the full
presented string in constant time.

## Issuance gate

**Any app with an approved catalog status can self-issue credentials.** Approval
already requires domain verification and a review pass, so there is still a gate
— it just is not a per-integration favour.

`verified` remains a *separate* trust signal, surfaced next to awards and app
identity, not a precondition for having a key. Making verification the gate
would put a human back in the loop for every integration, which is the exact
problem being solved.

Revoked, rejected, or unlisted apps cannot issue, and existing credentials stop
verifying when status leaves the approved set (see invalidation below).

## Endpoints

### Owner-facing (existing wallet auth + `isOwner`)

| Method | Path | Notes |
|---|---|---|
| `POST` | `/api/apps/{slug}/credentials` | Issue. Returns the plaintext secret **once**. |
| `GET` | `/api/apps/{slug}/credentials` | List metadata only — prefix, label, created, last used, revoked. Never the secret. |
| `DELETE` | `/api/apps/{slug}/credentials/{id}` | Revoke. Idempotent. |

These sit beside the owner and revision management the developer page already
does, and reuse `walletAuthMiddleware` + `isOwner` unchanged.

### Service-facing (existing `registryServiceAuth`)

| Method | Path | Notes |
|---|---|---|
| `POST` | `/registry/credentials/verify` | Body `{key}`. Returns the app client record + `declared_scopes`, or 401. |

Consumers never receive secrets or hashes — they present what the app presented
and get back an identity. Rate-limited, and it touches `last_used_at` at most
once a minute per credential to keep the write cheap.

## Verification, caching, invalidation

Consumers verify per request against a short-lived cache (~5 min), rather than
replicating hashed secrets into every service. Rationale:

- No secret material ever leaves the registry, so a compromised consumer cannot
  mint or replay another app's identity.
- Revocation lag is bounded by the cache TTL, and cut further by adding
  `credential_revoked` to the existing `app_client_changes` reasons — consumers
  already poll that feed, so early invalidation is free.
- Status changes already emit `status_changed`; a consumer seeing it drops the
  app's cached identity, which covers "app got unlisted" without new machinery.

The cost is one HTTP call per app per TTL window. At current ecosystem size that
is negligible, and the cache makes it independent of request volume.

## Scope enforcement

`declared_scopes` becomes load-bearing. Verify returns them; the consumer
enforces them:

- Posting an achievement requires `achievements:write`.
- Posting a plaza event requires `events:write`.

An app that never declared a scope cannot use it, and changing declared scopes
goes through review and emits `scopes_changed`. This is what makes "awarded by
X" trustworthy: X is the store's verified identity for a credential that was
allowed to write, not a display name the caller chose.

## Consumers

Two exist today and both migrate:

1. **NimConnect awards** — replaces its manual key set. Attribution
   (`name`, `icon_url`, `verified`) comes from the verify response rather than
   from app-supplied fields.
2. **NimWorld signed events** — `APP_KEYS` becomes a registry lookup. NimWorld
   is paused, so this lands whenever it is next touched; the HMAC scheme over
   the raw body is unchanged, only the secret's provenance moves.

A third (PlayNimiq, currently holding a NimConnect app key by hand) becomes a
normal registry consumer with no special casing.

Because there are two real consumers with a third in sight, the verification
contract is written once as a documented service endpoint rather than being
inlined into NimConnect. That is the extent of the generalisation — no plugin
system, no per-consumer configuration surface.

## Security notes

- Secrets are high-entropy random, hashed at rest, compared in constant time,
  never logged (log `key_prefix` only).
- Issue and verify are both rate-limited; verify failures are counted per
  service token.
- `created_by` records which owner issued a credential, so a later ownership
  dispute has an audit trail.
- Revocation is a timestamp, not a delete: history survives.

## Out of scope

- Per-credential scope narrowing (a credential inherits the app's declared
  scopes). Add when an app actually wants a read-only key.
- OAuth-style user-delegated app tokens. NimConnect already owns user
  authorization; this is app→service identity only.
- Automatic rotation policies.

## Open question

NimWorld's `/events` uses HMAC over the raw body; NimConnect's `/api/awards`
uses a bearer key. Unifying them is tempting but is a behaviour change to two
live schemes for no user-visible gain — this design keeps both and only moves
where the secret comes from. Revisit if a third scheme appears.
