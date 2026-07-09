# User profile: display name + identicon

## Purpose

Give a connected wallet a lightweight identity beyond the raw address:
an optional display name and a deterministic identicon (Nimiq's official
wallet-style avatar), shown wherever a wallet address currently appears
(the reviews list, the header wallet button). This is the first of two
planned extensions to wallet login — a separate spec will cover linking a
wallet to a `developer_slug` for app ownership/management, reusing the same
`users` table this spec introduces.

## Data model

New table `users`:

| column       | type        | notes                                          |
|--------------|-------------|-------------------------------------------------|
| wallet_address | text pk   | same identity used by the wallet-auth cookie     |
| display_name | text        | nullable, max 50 chars                          |
| created_at   | timestamptz | default now()                                   |
| updated_at   | timestamptz | default now()                                   |

No explicit "register" step — the row is created (or updated) the first
time a wallet saves a display name, via an upsert (`INSERT ... ON CONFLICT
(wallet_address) DO UPDATE`), matching the `app_reviews` upsert pattern
already in the backend.

`GET /api/apps/{slug}/reviews` changes to a `LEFT JOIN users` on
`wallet_address` so each review item gains a nullable `display_name` field.
A `LEFT JOIN` (not inner) ensures reviews from wallets that never set a
name still show up, just without one.

## API

```
GET /api/profile   [wallet cookie]  -> { wallet_address, display_name }
PUT /api/profile   [wallet cookie]  { display_name } -> upsert, returns { wallet_address, display_name }
```

Both reuse the existing `walletAuthMiddleware` (same as the review
endpoints). `display_name` validation (empty allowed = clears the name,
max 50 chars) follows the same pure-function-returns-error-string pattern
as `validateReviewInput`.

`GET /api/apps/{slug}/reviews` response items gain `display_name: string |
null`.

## Frontend

- Add `@nimiq/identicons` (already proven in the sibling `NimFeed` project:
  `Identicons.toDataUrl(address)` → data URL, with
  `Identicons.placeholderToDataUrl(...)` as a fallback). Port
  `AddressIdenticon.vue` from `NimFeed` near-verbatim: an `<img>` bound to
  an address prop, rendering the data URL, rounded.
- `ReviewList.vue`: replace the raw truncated-address `<span>` with
  `<AddressIdenticon>` + the review's `display_name` (falling back to the
  truncated address when `display_name` is null).
- Header `WalletLoginButton.vue`: show the identicon next to the truncated
  address/name.
- New route `/profile` (wallet-gated: shows "Connect your wallet" in place
  of the form when logged out). Shows the identicon, full address, and an
  editable display-name text input with a save button, calling `PUT
  /api/profile`. This page becomes the home for developer-account fields
  in the next spec.

## Testing

- Backend: table-driven test for display-name validation (empty allowed,
  >50 chars rejected), following `TestValidateReviewInput`'s shape.
- Backend: reviews list test confirms a review from a wallet with no
  `users` row still returns with `display_name: null` (left join, not
  inner).
- Frontend: no new test infra — `AddressIdenticon.vue` is a working,
  already-shipped pattern ported as-is.

## Explicitly out of scope

- Uploaded/custom avatars (identicon is deterministic from the address
  only).
- Unique display names / collision handling.
- Developer-account linking, app ownership, or permissions (separate
  follow-up spec; the `/profile` page and `users` table are built to be
  extended by it, not to include it now).
