# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Dual audience:
- **Wallet users** browsing the directory to discover and open mini apps (games, tools, maps, and more) directly inside the Nimiq Pay wallet.
- **Developers** who build mini apps and use the site to submit new apps, request updates, and track their own apps (My Apps, submission status).
- **Admins** who review, approve, reject, and manage submitted apps (Admin view).

## Product Purpose

Nimiq Mini Apps is a community-curated directory of mini apps for the Nimiq Pay wallet. It lets wallet users find and launch apps straight from their wallet, and gives developers a path to submit and manage apps that reach that audience. Success is a trustworthy, current catalog that wallet users can browse confidently and developers can get their apps listed in.

## Positioning

Apps open directly inside Nimiq Pay via the Hub API / mini-app SDK — a native, one-tap launch experience that a generic app store or web directory (without wallet integration) cannot offer.

## Operating Context

- Browsing/discovery: home, all-apps listing, app detail, favorites.
- Developer workflow: build guide, submit form, submission status tracking, request-update flow, my-apps management.
- Identity/session: wallet login (Nimiq Hub), profile, address identicons.
- Admin workflow: review queue, approve/reject, app management (admin view).
- Sharing: per-app share, social links, "open in wallet" panel, store badges.

## Capabilities and Constraints

- Built with Vue 3 + Vue Router + Tailwind v4 + Vite.
- Integrates `@nimiq/hub-api` (wallet login/session), `@nimiq/mini-app-sdk`, `@nimiq/identicons`, `@nimconnect/profile-client` (developer public profiles).
- Supports light/dark theme, chosen pre-paint from `localStorage.theme` or system preference.
- Markdown rendering (sanitized via DOMPurify) for app descriptions/changelogs.
- QR code generation (`qrcode`) for wallet-side app access.

## Brand Commitments

- Product name: "Nimiq Mini Apps".
- Typefaces already in use: Mulish (UI text) and Fira Mono (monospace/code contexts).
- Theme color `#FAFAFA` (light).
- No gold/yellow in the palette; radiant blue is the accent (see design system notes).

## Evidence on Hand

No case studies, testimonials, or usage numbers on hand. Do not fabricate adoption figures, partner names, or press mentions in future work.

## Product Principles

1. Wallet-native first: every discovery and launch flow should feel like an extension of Nimiq Pay, not a separate web app store.
2. Trust through curation: admin review and status transparency (submission status, review stage) matter as much as catalog breadth.
3. Low-friction developer path: submitting, updating, and tracking an app should require minimal steps outside the normal build/submit/status loop.
4. Community-curated, not algorithmically ranked: favor legibility and honesty about an app's status over engagement-optimized presentation.
