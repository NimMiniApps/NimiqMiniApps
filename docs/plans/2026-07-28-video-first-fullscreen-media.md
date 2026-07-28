# Video-First Fullscreen Media Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Put videos first in every app media gallery and open featured screenshots in an accessible full-resolution lightbox.

**Architecture:** A pure helper will stable-order media without mutating API data. `MediaGallery.vue` will render and navigate that ordered list, while a native `<dialog>` owned by the component provides the image lightbox without another dependency.

**Tech Stack:** Vue 3 composition API, TypeScript, native HTML dialog, Tailwind CSS v4, Vitest, GitHub Actions, Docker Swarm.

---

### Task 1: Stable video-first ordering

**Files:**
- Modify: `frontend/src/utils/mediaGallery.ts`
- Modify: `frontend/src/utils/mediaGallery.test.ts`

**Step 1: Write the failing test**

Import `orderMediaItems` and pass an interleaved list containing two images and two YouTube videos. Assert that both videos are first, that order within each type is unchanged, and that the input array remains unchanged.

**Step 2: Run the focused test to verify it fails**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: FAIL because `orderMediaItems` is not exported.

**Step 3: Write the minimal implementation**

Export a generic stable helper that returns a new array composed of the input's YouTube items followed by its image items.

**Step 4: Run the focused test to verify it passes**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: all gallery utility tests pass.

### Task 2: Use ordered media and add the image dialog

**Files:**
- Modify: `frontend/src/components/MediaGallery.vue`
- Modify: `frontend/src/utils/mediaGallery.test.ts`

**Step 1: Write failing component-contract assertions**

Assert that the component derives `orderedItems` through `orderMediaItems`, renders and navigates that list, wraps a featured image in a descriptive button, and contains a native `<dialog>` with the original selected image and an accessible close button.

**Step 2: Run the focused test to verify it fails**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: FAIL because the ordered display list and dialog markup do not exist.

**Step 3: Implement the minimal component behavior**

- Derive `orderedItems` with a computed property.
- Replace prop-list references in selection, rendering, and keys with the ordered list.
- Add a dialog ref plus `openImageDialog` and `closeImageDialog`.
- Render featured images as buttons.
- Render a fixed, backdrop-styled native dialog containing the original image URL, close button, and backdrop-click handling.
- Leave the YouTube iframe unchanged so its native fullscreen remains available.

**Step 4: Run focused and full frontend verification**

Run: `cd frontend && npm test -- src/utils/mediaGallery.test.ts`

Expected: focused tests pass.

Run: `cd frontend && npm test && npm run build`

Expected: all frontend tests pass and Vite builds successfully.

### Task 3: Commit, publish, and deploy

**Files:**
- Commit only the two plan documents and intended frontend source/test changes.

**Step 1: Audit and commit**

Run: `git diff --check`, inspect `git status --short --branch`, and review the outgoing diff for secrets or unrelated files.

Commit the implementation with: `git commit -m "feat: prioritize videos and expand gallery images"`

**Step 2: Push and wait for CI**

Push `main`, capture the full SHA, locate the matching CI run, and run `gh run watch <run-id> --exit-status`.

Expected: backend, frontend, and both image-publish jobs pass.

**Step 3: Deploy the exact frontend image**

On the production manager, pull `ghcr.io/nimminiapps/nimiq-mini-apps-frontend:<full-sha>` and update only `nimiqminiapps_frontend`.

Expected: the service converges at `1/1`.

**Step 4: Verify production**

Verify the exact service image and completed update, frontend/API health endpoints, an app-detail route, referenced bundle assets, and the live app-detail bundle's video-first ordering and native dialog code.
