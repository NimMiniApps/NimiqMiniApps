# Nimiq Mini Apps — Roadmap

Improvement ideas for the community directory, grouped by priority.  
**Done** items are checked off so we can track what shipped vs. what’s next.

---

## Done

### Quick wins
- [x] Hide `submitted` / `rejected` apps on public `GET /api/apps/{slug}`
- [x] Validate optional URL fields (`website_url`, `github_url`, `icon_url`, `banner_url`, image media)
- [x] Broader search (tags, assets, developer name, `long_description`)
- [x] Filter by `?tag=` and `?asset=` on browse
- [x] Clickable tags and assets → filtered browse
- [x] Home page loading skeletons (featured / newest)
- [x] Icon and banner URLs on submit form
- [x] Copy open link + inline QR on desktop app detail
- [x] Admin “Pending review” queue
- [x] Sync `q`, `sort`, `tag`, `asset` to URL on `/apps`
- [x] Add USDC to allowed assets
- [x] Status badge tooltips (verified / approved / experimental)
- [x] Fix NimLens placeholder domain in seed migration

### High impact
- [x] Developers directory (`GET /api/developers`, `/developers` page)
- [x] Related apps (`GET /api/apps/{slug}/related`, detail section)
- [x] Build page (`/build`) with human-readable nimiq.dev doc links
- [x] Per-page Open Graph / meta tags (app detail, developers, static routes)
- [x] MCP tools: `list_developers`, `get_related_apps`

---

## Next — high value, moderate effort

| # | Item | Status |
|---|------|--------|
| 1 | **Clickable category on detail page** | Done |
| 2 | **Submission status lookup** (`/status/{slug}`, `GET /api/apps/{slug}/status`) | Done |
| 3 | **Collections / curated lists** (`?collection=`, home sections) | Done |
| 4 | **Sitemap + `robots.txt`** | Done |
| 5 | **Admin pending badge in nav** | Done |
| 6 | **Submission notifications** (`SUBMIT_WEBHOOK_URL` env) | Done |
| 7 | **Default OG image** (`/og-default.svg`) | Done |

---

## Next — ecosystem & UX

| # | Item | Why |
|---|------|-----|
| 8 | **i18n** | Read `window.nimiqPay.language` when embedded in Nimiq Pay; match wallet language for UI strings |
| 9 | **Share button on app detail** | Web Share API on mobile, copy URL on desktop |
| 10 | **Developers in mobile nav** | Footer link exists; optional 5th tab or “More” menu on small screens |
| 11 | **Empty / error states polish** | Richer copy on “No apps found”, failed load, and 404 app slug |
| 12 | **App detail breadcrumb** | e.g. Apps → Games → NimDoom |

---

## Infrastructure & scale

| # | Item | Why |
|---|------|-----|
| 13 | **Distributed rate limiting** | In-memory submit limit breaks with multiple backend replicas; move to Postgres or Redis |
| 14 | **Pagination on `GET /api/apps`** | Done |
| 15 | **Automated domain health check** | Done |
| 16 | **SSR or prerender for OG crawlers** | Done |

---

## Bigger bets

| # | Item | Why |
|---|------|-----|
| 17 | **Directory as a mini app** | Load this catalog inside Nimiq Pay natively (meta: dogfooding) |
| 18 | **Privacy-friendly open analytics** | Count “Open in Nimiq Pay” clicks to surface popular apps |
| 19 | **Developer accounts** | Let builders edit / resubmit their listing without full admin access (partial: public update requests + admin review) |
| 20 | **Review notes / rejection reason** | Admin can leave a message; optional email to submitter on reject |
| 21 | **RSS / Atom feed** | `GET /api/feed` of newly approved apps for community bots and newsletters |
| 22 | **Chain metadata** | First-class `chains` field (Polygon, Base, …) alongside `assets` for dual-chain discovery |

---

## Code quality (low urgency)

| # | Item | Why |
|---|------|-----|
| 23 | **Shared category theme tokens** | Duplicated color maps in `HomeView.vue` and `AppCard.vue` |
| 24 | **Frontend tests** | Smoke tests for API client normalization and route query sync |
| 25 | **Integration tests** | API tests for list filters, related apps, public visibility |

---

## Suggested order for the next sprint

1. Share button on app detail (Web Share API / copy URL)
2. i18n via `window.nimiqPay.language`
3. Empty / error states polish
4. Pagination on `GET /api/apps`
5. Distributed rate limiting for submit endpoint

---

## References

- [Development guide](./DEV.md)
- [Swarm deploy](./SWARM.md)
- [Nimiq mini apps docs](https://nimiq.dev/mini-apps/)
- [MCP server](../mcp/README.md)
