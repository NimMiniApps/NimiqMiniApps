# App Detail Media Gallery Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Keep app-detail banners and screenshots visually useful without allowing portrait catalog media to dominate the page.

**Architecture:** `MediaGallery.vue` will own a selected-media state and render one bounded featured viewport plus an accessible thumbnail selector. A small, pure gallery-state utility will make selection clamping and next/previous navigation deterministic and testable. `AppDetailView.vue` will constrain the listing banner independently so it remains a header, not page-sized content.

**Tech Stack:** Vue 3 composition API, TypeScript, Tailwind CSS v4, Vitest.

---

### Task 1: Define tested gallery selection behavior

**Files:**
- Create: `frontend/src/utils/mediaGallery.ts`
- Create: `frontend/src/utils/mediaGallery.test.ts`

**Step 1: Write the failing tests**

Add tests for `clampMediaIndex(index, itemCount)` and `moveMediaIndex(index, itemCount, direction)`. Assert that an empty gallery returns `0`, an initial index is clamped when the media list shrinks, and previous/next navigation wraps from either end.

**Step 2: Run the test to verify it fails**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: FAIL because `./mediaGallery` does not exist.

**Step 3: Write minimal implementation**

Add pure helpers that return `0` for empty lists, clamp to `[0, itemCount - 1]`, and wrap one-step movement via modulo arithmetic. Export a narrow `GalleryDirection` union (`'previous' | 'next'`).

**Step 4: Run the focused test to verify it passes**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: all gallery-state assertions pass.

**Step 5: Commit**

```bash
git add frontend/src/utils/mediaGallery.ts frontend/src/utils/mediaGallery.test.ts
git commit -m "test: define media gallery navigation"
```

### Task 2: Build the featured-media gallery with thumbnail selection

**Files:**
- Modify: `frontend/src/components/MediaGallery.vue`

**Step 1: Write the failing component-level behavior target**

Use the Task 1 helpers as the contract: the component starts at media item zero, updates the featured viewport from an accessible thumbnail button, wraps its previous/next controls, and safely reclamps if the supplied media list changes.

**Step 2: Run the existing focused utility test before the component change**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: PASS, establishing the interaction contract.

**Step 3: Implement the minimal gallery**

Replace the responsive two-column media grid with:

- a dark, rounded featured panel with a bounded `aspect-video` viewport;
- the active image using `object-contain` so portrait screenshots are fully visible rather than stretched or cropped;
- the active YouTube item as the existing lazy iframe;
- previous/next buttons only when there is more than one item, with descriptive `aria-label`s;
- a horizontally scrollable, snap-aligned thumbnail row of buttons, each with a selected-state ring and `aria-current`/pressed state;
- thumbnail images cropped only inside their small previews, while video thumbnails use a neutral labelled preview instead of loading an iframe per card.

Keep the section title and continue using `youtubeEmbedUrl()` for embed safety.

**Step 4: Run focused tests**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts src/utils/media.test.ts`

Expected: all media utility tests pass.

**Step 5: Commit**

```bash
git add frontend/src/components/MediaGallery.vue frontend/src/utils/mediaGallery.ts frontend/src/utils/mediaGallery.test.ts
git commit -m "feat: add compact app media gallery"
```

### Task 3: Bound the app banner independently

**Files:**
- Modify: `frontend/src/views/AppDetailView.vue`

**Step 1: Confirm the media utility tests are green**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: PASS.

**Step 2: Implement the bounded banner**

Wrap the existing `banner_url` image in a rounded, bordered surface with a responsive `max-h` cap. Use `object-contain` and a neutral surface background so the full supplied artwork remains visible without dictating page height.

**Step 3: Run frontend tests and type/build checks**

Run: `cd frontend && npm test && npm run build`

Expected: Vitest passes and Vite emits a production bundle with exit code 0.

**Step 4: Manual visual verification**

Run the local frontend and inspect an app with portrait banner/media at desktop and narrow mobile widths. Verify: banner remains compact; featured item preserves full image; thumbnails remain reachable; next/previous wrap; video remains playable; keyboard focus is visible.

**Step 5: Commit**

```bash
git add frontend/src/views/AppDetailView.vue
git commit -m "fix: bound app detail banner height"
```

### Task 4: Document and final verification

**Files:**
- Modify: `README.md`
- Modify: `docs/plans/2026-07-28-app-detail-media-gallery.md`

**Step 1: Add a concise README note**

In the existing catalog feature description, add one short mention that app pages display submitted media in an interactive gallery. Do not duplicate UI implementation detail.

**Step 2: Run diff and full frontend verification**

Run: `git diff --check && cd frontend && npm test && npm run build`

Expected: no whitespace errors, all tests pass, and production build succeeds.

**Step 3: Commit**

```bash
git add README.md docs/plans/2026-07-28-app-detail-media-gallery.md
git commit -m "docs: describe app media gallery"
```
