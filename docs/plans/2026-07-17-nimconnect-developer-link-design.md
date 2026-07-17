# NimConnect developer link — design

## Goal

When a catalog app’s first owner has a claimed NimConnect `@handle`, link to their public page from the app detail and developer-filter views.

## Decisions

- **Surfaces:** app detail (“by {developer}”) and `/apps?developer=` header
- **Wallet:** `owner_wallet_addresses[0]` only
- **Exists:** claimed `@handle` is enough (empty profile OK)
- **Approach:** frontend-only via `@nimconnect/profile-client` — no MiniApps API changes

## Data flow

1. Take `owner_wallet_addresses[0]` from the app payload (already public).
2. Call `getHandleByAddress` through `@nimconnect/profile-client`.
3. If a handle returns, link to `https://nimconnect.nimiqminiapps.com/@{handle}`.
4. On 404 / network / 5xx: hide the link (no error UI).
5. Session-memory cache by address (including negative results).

## UI

- **App detail:** keep the catalog “by {name}” filter link; add a secondary `@handle` outbound link beside it when resolved.
- **Developer filter:** under the developer title/meta, show the same `@handle` link when the first loaded app’s first owner resolves.
- Do not block render on the lookup; no “claim your handle” CTA.
- Visible text is `@handle`; aria/title via i18n (“NimConnect profile”).

## Testing

- Unit-test the helper: claim → URL, 404 → null, caches negatives.
- Smoke: link present when handle exists, absent otherwise.
