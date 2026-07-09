# Nimiq wallet login + app ratings/reviews

## Purpose

Let visitors connect their Nimiq wallet (desktop via Hub, or the injected
signer when embedded in Nimiq Pay) and leave a star rating + written review
on a directory-listed app. Reuses the proven signature-verification pattern
from `nimiq-2048`, but swaps its DB-backed session stack for a stateless
signed cookie to match this backend's stdlib-only style (plain `net/http` +
pgx, no framework, no session table today).

## Data model

New table `app_reviews`:

| column       | type        | notes                                    |
|--------------|-------------|-------------------------------------------|
| id           | uuid pk     | `gen_random_uuid()`                        |
| app_id       | uuid fk     | references `apps(id)`                      |
| wallet_address | text      | Nimiq address, not a separate users table  |
| rating       | smallint    | CHECK 1-5                                  |
| body         | text        | 0–1000 chars, empty allowed (rating-only)  |
| created_at   | timestamptz | default now()                              |
| updated_at   | timestamptz | default now()                              |

`UNIQUE (app_id, wallet_address)` — one review per wallet per app; a second
write from the same wallet is an edit (upsert), not a new row. Mirrors
nimiq-2048's `(wallet_address, puzzle_date)` unique-constraint pattern.

Average rating is computed on read (`AVG(rating) GROUP BY app_id`), not
cached — add a cached column later only if this becomes a measured perf
issue. No new `users` table: wallet_address is the entire identity.

## Auth flow (challenge → verify → cookie)

Crypto core borrowed from nimiq-2048 (`backend/platform/handlers/auth_handler.go`
in that repo); session layer replaced with a stateless signed cookie instead
of DB-backed sessions, since this backend has no session infra today (only a
single bearer-token check for `/api/admin/*`).

1. `POST /api/auth/challenge` — client sends `wallet_address`. Server
   generates a nonce, builds the same signable message format nimiq-2048
   uses (address, nonce, timestamp, domain, purpose, expiry), returns it.
   Nonce kept in an in-memory map with TTL (single backend instance; move to
   a table only if the backend ever runs multiple replicas).
2. Client signs the message — via `@nimiq/hub-api` popup/iframe on desktop,
   or the injected Nimiq Pay signer when embedded, chosen by the same
   `isNimiqPayAvailable()` detection nimiq-2048 uses.
3. `POST /api/auth/verify` — server re-applies Nimiq's signed-message prefix
   (`\x16Nimiq Signed Message:\n` + length + message), SHA-256 hashes,
   `ed25519.Verify`s, confirms the pubkey derives to the claimed address,
   checks the nonce hasn't expired or been reused. On success, issues an
   HMAC-signed cookie: `base64(wallet_address + expiry) +
   hmac-sha256(secret)`, `HttpOnly; Secure; SameSite=Lax`.
4. `walletAuthMiddleware(next)` (same shape as the existing `authMiddleware`
   used for `/api/admin/*`) validates the cookie signature + expiry on
   protected routes, extracts `wallet_address` — no DB lookup needed.
5. The cookie is refreshed (re-issued with a new expiry) on any authenticated
   request within its validity window, so active users stay logged in;
   idle sessions expire after ~7 days and require re-signing.

No instant server-side revocation (acceptable per user: short-lived
auto-expiry is fine, including for a possible future developer portal built
on the same cookie + middleware pattern).

## Frontend

- `WalletLoginButton` in the header — "Connect Wallet" → truncated address +
  logout (logout just clears the cookie).
- Login button routes through the same Hub-popup-or-Nimiq-Pay-signer
  detection as nimiq-2048's `walletLogin()`.
- App detail page gets a `ReviewForm` (star picker + textarea), rendered
  only when logged in ("Connect wallet to leave a review" otherwise), and a
  `ReviewList` showing existing reviews with edit/delete controls visible
  only on the review matching the logged-in wallet.
- No separate account/profile page.

## Moderation (auto-publish + basic filters)

Enforced server-side on write (client-side checks are trivially bypassed):

- Rate limit: max N review writes per wallet per time window (e.g. 5/hour),
  in-memory token bucket keyed by wallet.
- Validation: rating required 1-5; body 0–1000 chars.
- No profanity/spam ML — out of scope.
- Admin can remove any review via `DELETE /api/admin/apps/{slug}/reviews/{id}`,
  reusing the existing admin-token `authMiddleware`.

## API surface

```
POST   /api/auth/challenge                   { wallet_address } -> { message, nonce }
POST   /api/auth/verify                      { wallet_address, signature, public_key, nonce, source } -> sets cookie
POST   /api/apps/{slug}/reviews              { rating, body }  [wallet cookie] -> upsert own review
DELETE /api/apps/{slug}/reviews              [wallet cookie]   -> delete own review
GET    /api/apps/{slug}/reviews              -> paginated list + average (public, reuses pagination.go)
DELETE /api/admin/apps/{slug}/reviews/{id}   [admin token]     -> moderation delete
```

Errors follow the existing `writeError(w, status, msg)` convention in
`handlers.go`. Signature/nonce failures → 401; rate limit → 429; validation
failures (rating out of range, body too long) → 400.

## Testing

Following the repo's existing table-driven `*_test.go` stdlib `testing`
pattern (see `validate_test.go`, `pagination_test.go`):

- Signature verification: valid signature, invalid signature, replayed
  nonce, expired nonce.
- HMAC cookie encode/decode round-trip, including tampered/expired cases.
- Review upsert: second write from the same wallet edits rather than
  duplicates; unique constraint holds.

## Explicitly out of scope

- Comment threads / replies (rating + single review body only).
- Profanity/spam ML filtering.
- Instant session revocation / logout-everywhere.
- Cached/denormalized average rating column.
- Any changes to `backend/domaincheck.go` / `backend/icondiscovery.go`
  (unrelated SSRF finding surfaced during this session, tracked separately).
