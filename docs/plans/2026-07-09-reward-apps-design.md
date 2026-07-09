# Reward Apps Design

## Goal

Make it clear which mini apps can reward users with crypto assets, without treating every app that uses a token as an earning app.

## Design

Apps get a new developer-declared `reward_assets` field. It is an array of supported catalog assets such as `NIM`, `USDT`, or `USDC`, and it means users can receive those assets from the app through rewards, payouts, prizes, tips, or similar mechanics. The existing `assets` field remains the broader "uses/supports this asset" signal.

Developers can include reward assets when submitting or requesting updates. Admins can review, add, remove, or correct the field during moderation. The API validates reward assets against the same token allowlist as `assets`.

Public discovery gets two paths:

- `GET /api/apps?rewards=true` filters to apps with at least one reward asset.
- `collection=rewards` powers an "Apps with rewards" collection.

The frontend shows compact `Earn NIM` / `Earn USDT` style badges on cards and detail pages. The badge links to the rewards filter. The wording stays factual and token-specific instead of saying "earn money".

## Testing

Backend tests cover valid and invalid `reward_assets`, plus collection SQL behavior. Frontend tests cover reward label output. Full verification includes backend tests, frontend tests, frontend build, and OpenAPI regeneration.
