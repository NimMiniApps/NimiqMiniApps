# README Branding Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Give the project README a visual, catalog-specific introduction consistent with the NimBomber README pattern.

**Architecture:** Reuse the checked-in Open Graph graphic as a wide visual header and the checked-in Mini Apps logo as a centered brand mark. Keep all existing technical documentation below the introductory block, changing no runtime code or deployment behavior.

**Tech Stack:** Markdown, repository static assets.

---

### Task 1: Add the branded README introduction

**Files:**
- Modify: `README.md:1`
- Verify: `frontend/public/og-default.svg`
- Verify: `frontend/public/brand/nimiq-mini-apps-logo.png`

**Step 1: Confirm the referenced assets exist**

Run: `test -f frontend/public/og-default.svg && test -f frontend/public/brand/nimiq-mini-apps-logo.png`

Expected: exit status 0.

**Step 2: Replace the plain title block**

Add a centered wide header image, centered Mini Apps logo, title, short catalog description, and direct links to the live catalog, Nimiq Pay, submission page, and developer guide. Leave the existing Stack section and all subsequent documentation intact.

**Step 3: Verify Markdown links and image paths**

Run: `rg -n 'og-default|nimiq-mini-apps-logo|nimiqminiapps.com|nimpay.app|docs/DEV.md' README.md`

Expected: the new visual references and all four links are present.

**Step 4: Commit**

```bash
git add README.md
git commit -m "Brand README header"
```
