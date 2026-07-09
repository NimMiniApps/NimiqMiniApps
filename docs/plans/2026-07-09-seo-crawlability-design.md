# SEO Crawlability Design

Date: 2026-07-09

## Goal

Make `nimiqminiapps.com` indexable as a useful Nimiq Pay mini app directory, not only as a JavaScript application. The first SEO target is branded and ecosystem traffic:

- `Nimiq Mini Apps`
- `Nimiq Pay apps`
- `Nimiq mini app directory`
- `Nimiq wallet apps`

This design keeps the current Vue SPA for users and adds a pragmatic server-rendered SEO layer for crawlers and link previews. It deliberately avoids a full SSR or static-generation migration for the first phase.

## Current Shape

The public frontend is a Vue 3 + Vite SPA served by Nginx. The Go backend owns catalog data, sitemap generation, and a limited Open Graph prerender endpoint at `/og/apps/{slug}`. Nginx already detects crawler user agents and proxies app detail crawler requests to the backend preview HTML.

That means the cheapest path to stronger SEO is to expand the existing backend prerender concept into a first-class SEO renderer.

## Canonical Public URLs

The canonical indexable pages should live on `https://nimiqminiapps.com`, not the API domain.

Initial canonical pages:

- `/`
- `/apps`
- `/apps/{slug}`
- `/apps/category/{category}`
- `/developers/{slug}`
- `/build`
- `/submit`

Query-filtered UI pages such as `/apps?developer=...` can remain useful for the SPA, but they should not be the primary SEO surface. Developer and category pages need clean path-based URLs so they can have stable titles, canonical tags, sitemap entries, and internal links.

## SEO HTML Renderer

Add backend handlers that return complete, lightweight HTML for public pages:

- `GET /seo/home`
- `GET /seo/apps`
- `GET /seo/apps/{slug}`
- `GET /seo/apps/category/{category}`
- `GET /seo/developers/{slug}`

Each response should include:

- `<title>`
- meta description
- canonical link
- Open Graph tags
- Twitter card tags
- visible fallback content in the body
- JSON-LD structured data
- internal links to related public pages
- a normal link to the canonical Vue route

The HTML should be simple and cacheable. It does not need to match the full app UI. Its job is to give crawlers and unfurled links accurate page identity, page text, and discoverable internal links.

## App Detail HTML

For `/apps/{slug}`, render enough content for the page to stand alone in search results:

- app name
- tagline
- description and long description when present
- category
- tags
- developer name and developer URL
- release stage
- public status label
- open in Nimiq Pay URL
- website and GitHub links when present
- icon or banner image when present
- related app links
- last updated date
- domain health or last checked date if available

If an app is not public, return `404` from the SEO renderer.

## Directory, Category, and Developer HTML

For `/apps`, render a summary of the catalog plus links to public app detail pages, category pages, and developer pages.

For `/apps/category/{category}`, render:

- category-specific title and description
- list of public apps in that category
- links to related categories

For `/developers/{slug}`, render:

- developer name
- app count
- list of public apps by that developer
- links to app detail pages

The Vue router should eventually support these same path-based routes so users and crawlers share canonical URLs.

## Nginx Routing

Keep normal user traffic on the Vue SPA.

Crawler traffic for indexable public routes should proxy to backend SEO HTML:

- `/`
- `/apps`
- `/apps/{slug}`
- `/apps/category/{category}`
- `/developers/{slug}`

Public SEO infrastructure should also be served from the canonical site domain:

- `https://nimiqminiapps.com/robots.txt`
- `https://nimiqminiapps.com/sitemap.xml`

The API-domain sitemap can exist for operational convenience, but it should not be the primary sitemap advertised to crawlers.

## Sitemap

Generate a canonical sitemap using `SITE_URL`, expected to be `https://nimiqminiapps.com` in production.

Include:

- static public pages
- public app detail pages
- category pages with at least one public app
- developer pages with at least one public app

Do not include admin, status, update-request, or rejected/pending app URLs.

## Structured Data

Use JSON-LD for machine-readable page context:

- homepage: `WebSite`
- app list: `ItemList`
- app detail: `SoftwareApplication` or `WebApplication`
- developer pages: `Organization` where appropriate
- detail and listing pages: `BreadcrumbList`

Structured data should be generated from the same backend app records used for the visible fallback HTML.

## Content Quality

Search performance will be limited if app records are thin. The admin workflow should eventually expose SEO quality hints for public apps:

- missing long description
- missing icon
- missing banner or media
- missing tags
- missing developer information
- unreachable domain
- duplicate or overly short tagline

These checks should be advisory, not blockers, during the first implementation.

## Measurement

After rollout, set up or update Google Search Console for `nimiqminiapps.com`.

Track:

- sitemap submitted and discovered URL count
- indexed app page count
- pages discovered but not indexed
- canonical warnings
- crawl errors
- top queries
- average position for branded and ecosystem terms
- Core Web Vitals

Use Search Console's URL inspection on the homepage, `/apps`, and several strong app detail pages immediately after deployment.

## Phases

### Phase 1: Foundation

- Serve canonical `robots.txt` and `sitemap.xml` from the public site domain.
- Update sitemap URLs to use `https://nimiqminiapps.com`.
- Add canonical URL helper functions.
- Keep the current Vue SPA behavior unchanged for normal users.

### Phase 2: App Detail SEO

- Replace the current preview-only app renderer with richer SEO HTML for public app detail pages.
- Add canonical tags, metadata, JSON-LD, and visible fallback content.
- Keep social preview behavior working.

### Phase 3: Directory SEO

- Add SEO HTML for `/apps`.
- Add path-based category pages.
- Add path-based developer pages.
- Add all new canonical pages to the sitemap.

### Phase 4: Structured Data Tests

- Add focused tests for sitemap entries, canonical links, metadata, and JSON-LD fields.
- Validate that non-public apps do not appear in sitemap or SEO HTML.

### Phase 5: Content Quality

- Add admin-facing SEO quality hints.
- Prioritize improving descriptions, media, and internal links for featured apps first.

### Phase 6: Search Console Rollout

- Submit the canonical sitemap.
- Inspect key URLs.
- Monitor crawl/indexing behavior and adjust routing or content based on evidence.

## Non-Goals

- Full SSR migration.
- Rebuilding the frontend in Nuxt or another framework.
- Keyword stuffing or artificial landing pages.
- Serving misleading crawler-only content that differs materially from the user-facing page.
- Guaranteeing a number-one ranking for competitive queries.

## Open Questions

- Should `api.nimiqminiapps.com` redirect `/sitemap.xml` to the public domain sitemap or keep serving a duplicate operational sitemap?
- Should `/developers/{slug}` become a full Vue route in the same phase as SEO HTML, or first redirect users to `/apps?developer=...` while crawlers get backend HTML?
- Which app fields should be considered mandatory for a "strong SEO" admin hint?
