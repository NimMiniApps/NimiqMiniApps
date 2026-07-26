---
name: Nimiq Mini Apps
description: A wallet-native directory of Nimiq Pay mini apps, rendered as a live departures board — every app a scheduled destination, its status a lamp you trust at a glance.
colors:
  nimiq-blue: "#0582ca"
  nimiq-blue-dark: "#0071c3"
  board-frame: "#eef0f9"
  board-frame-dark: "#14162b"
  board-flap: "#ffffff"
  board-flap-dark: "#1f2348"
  board-flap-hover: "#f1f2fa"
  board-flap-hover-dark: "#262b52"
  board-flap-ink: "rgba(31, 35, 72, 1)"
  board-flap-ink-dark: "#f4f6ff"
  board-flap-muted: "rgba(31, 35, 72, 0.6)"
  board-flap-muted-dark: "rgba(244, 246, 255, 0.62)"
  board-hairline: "rgba(31, 35, 72, 0.1)"
  board-hairline-dark: "rgba(244, 246, 255, 0.14)"
  lamp-live: "#0ca6fe"
  accent-ink: "#0582ca"
  accent-ink-dark: "#0ca6fe"
  lamp-pending: "#fc8702"
  lamp-cancelled: "#d94432"
  lamp-info: "#21bca5"
  page: "#fafafa"
  page-dark: "#151833"
  ink: "rgba(31, 35, 72, 1)"
  muted: "rgba(31, 35, 72, 0.6)"
typography:
  display:
    fontFamily: "Mulish, ui-sans-serif, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "clamp(1.875rem, 4vw, 3rem)"
    fontWeight: 800
    letterSpacing: "0"
    lineHeight: 1.15
  title:
    fontFamily: "Mulish, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 800
    letterSpacing: "0.02em"
  body:
    fontFamily: "Mulish, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "Mulish, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.625rem"
    fontWeight: 700
    letterSpacing: "0.04em"
  caption:
    fontFamily: "Mulish, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.6875rem"
    fontWeight: 700
    letterSpacing: "0.02em"
  mono:
    fontFamily: "Fira Mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace"
    fontWeight: 500
rounded:
  plate: "4px"
  badge: "3px"
  board: "8px"
  lamp: "999px"
spacing:
  xs: "0.5rem"
  sm: "0.75rem"
  md: "1rem"
  lg: "1.5rem"
  xl: "2rem"
components:
  board-plate-primary:
    backgroundColor: "{colors.nimiq-blue}"
    textColor: "#ffffff"
    rounded: "{rounded.plate}"
    padding: "0.5rem 0.75rem"
  board-plate-primary-hover:
    backgroundColor: "{colors.nimiq-blue-dark}"
  board-plate-ghost:
    backgroundColor: "{colors.board-flap-hover}"
    textColor: "{colors.board-flap-ink}"
    rounded: "{rounded.plate}"
    padding: "0.5rem 0.75rem"
  board-row:
    backgroundColor: "{colors.board-flap}"
    textColor: "{colors.board-flap-ink}"
    rounded: "{rounded.board}"
    padding: "1rem"
  badge:
    backgroundColor: "{colors.board-flap}"
    textColor: "{colors.board-flap-ink}"
    rounded: "{rounded.badge}"
    padding: "0.125rem 0.375rem"
---

# Design System: Nimiq Mini Apps

## Overview

**Creative North Star: "The Departures Board"**

Nimiq Mini Apps doesn't look like a bolted-on app store or a neon crypto dashboard — it reads like a live departures board. Every mini app is a scheduled destination: category is its track number, status is a lamp you read in half a second, and opening an app is "boarding" it. This directly refuses the category's two ruts: the dark-neon glassmorphic Web3 dashboard, and the flat generic App-Store clone. Neither appears here.

The board's material — casing, flap rows, badges, panels — re-tints with the app's light/dark toggle exactly like every other surface (light: near-white flaps on a pale casing; dark: navy flaps on near-black casing). An earlier version of this system kept the board permanently dark in both themes as a "physical object" conceit; real usage showed that reads as a stuck/broken theme toggle, not a deliberate material choice, so it was reversed. Only the **lamp/status colors** (live, pending, cancelled, info) stay constant across themes — those are signal colors, not surface material, the same way a real lamp doesn't change hue depending on room lighting.

Nimiq Blue keeps its role as the brand's signature hue, now cast as the board's "on-time" lamp — every live, healthy, reachable app glows the same blue that anchors the wallet brand. Amber and red lamps read pending and broken states with the same instant legibility as a real concourse board.

**Key Characteristics:**
- The board re-tints with the theme toggle like any other surface — only lamp/status colors stay constant.
- Status is always a lamp + word, never color alone.
- Rectangular board plates replace pill buttons; ruled rows replace shadow-heavy cards.
- Fira Mono carries track codes, tags, and terminal-style search — Mulish carries every heading and body line.
- One signature motion: a per-character split-flap cascade, played once per row on load, never looped or idle.

## Colors

A dark board-casing palette anchored by Nimiq Blue as the "live" signal; amber and red carry every other state.

### Primary
- **Board Frame** (#EEF0F9 light / #14162B dark): the casing color — global header, bottom nav, and the outer edge of every board container.
- **Board Flap** (#FFFFFF light / #1F2348 dark): the row/card surface inside the board — where app rows, badges, and menu panels sit. Lightens further to Board Flap Hover (#F1F2FA light / #262B52 dark) on hover/focus.

### Secondary
- **Nimiq Blue → Lamp Live** (#0CA6FE, theme-constant): the "on-time" signal. Marks live/verified/approved apps and the primary action fill. **For readable text** (not fills, not dots), use **Accent Ink** instead (#0582CA light / #0CA6FE dark) — the raw Lamp Live cyan fails contrast on the new light-mode white board surfaces.
- **Lamp Pending** (#FC8702, theme-constant): submitted/beta/alpha/awaiting-review states.
- **Lamp Cancelled** (#D94432, theme-constant): rejected apps and unreachable domains.
- **Lamp Info** (#21BCA5, theme-constant): experimental/concept, informational-only states.

### Neutral
- **Board Flap Ink** (`rgba(31,35,72,1)` light / #F4F6FF dark): primary text on the board.
- **Board Flap Muted** (`rgba(31,35,72,0.6)` light / `rgba(244,246,255,0.62)` dark): secondary text on the board.
- **Board Hairline** (`rgba(31,35,72,0.1)` light / `rgba(244,246,255,0.14)` dark): dividers and borders on the board.
- **Page** (#FAFAFA light / #151833 dark): the app background outside the board.

### Named Rules
**The Board Re-Tints Rule.** Board casing, flap rows, badges, and menu panels flip with the theme toggle exactly like page chrome — there is no "always-dark" surface anywhere in the system.
**The Lamp Color vs. Lamp Text Rule.** Lamp/status hues (`--color-lamp-*`) are theme-constant and safe for fills, dots, and borders — but never for body-sized readable text on a board surface, since they're tuned for dark backgrounds and fail contrast on light ones. Text uses the theme-aware `accent-ink` token instead.
**The No-Gold Rule.** Gold/yellow never appears in the palette; `--nimiq-gold` is a dead legacy CSS variable, not a usable color.
**The Lamp-Plus-Word Rule.** Status is never color alone — every lamp pairs with an uppercase status word for colorblind-safe reading.

## Typography

**Display/Body Font:** Mulish (with Muli, ui-sans-serif, system-ui fallback)
**Label/Mono Font:** Fira Mono — carries track codes, tag chips, and the terminal-style search input.

**Character:** Confident extrabold Mulish for every heading, uppercase and tracked when it appears on the board itself; Fira Mono steps in wherever the content is literally a code, count, or terminal-style value.

### Hierarchy
- **Display** (800, `text-3xl` → `md:text-5xl`, uppercase, animated flap-in): the hero headline only.
- **Title** (800, `text-xl`, uppercase tracked): section headings, board panel titles.
- **Body** (400, `text-sm`–`text-base`): descriptions, taglines — never uppercase.
- **Label** (700, 10px, uppercase tracked): badges, lamps, track chips.
- **Caption** (700, 11px, uppercase tracked): nav labels, wallet menu micro-copy — one step up from Label where 10px reads too small against denser chrome.
- **Mono** (500, `text-xs`/`text-[11px]`, Fira Mono): track codes, tag chips, search input, wallet addresses.

### Named Rules
**The Board Shouts, Content Doesn't Rule.** Uppercase tracking is reserved for board furniture — headings, nav, badges, plates. Body copy (taglines, descriptions) always stays sentence case for readability.

## Layout

Unchanged from the incumbent shell: mobile-first with a persistent bottom nav (4-column, safe-area aware) and a `max-w-5xl` centered content column. The header and bottom nav are board-frame casing, present on every route, re-tinting with the theme like the rest of the board. Content grids stay single-column on mobile, `sm:grid-cols-2` for card/row listings.

## Elevation & Depth

The board is one physical object holding many flat rows: rows carry no individual shadow, only a 1px hairline divider between them (`.board-row` in a stacked frame) or, standalone, a wide soft ambient shadow with no per-row lift (`.board`: `0 1.5rem 3rem rgba(3,4,12,0.28)`, darker in dark mode). Buttons and badges are flush plates — flat, distinguished by a hairline border, not elevation.

### Shadow Vocabulary
- **Board ambient** (`box-shadow: 0 1.5rem 3rem rgba(3,4,12,0.28)`, dark: `rgba(0,0,0,0.45)`): the board container/card as one object.
- **Header/nav casing** (`shadow-lg shadow-black/20`, bottom nav: `0 -0.5rem 1.5rem rgba(0,0,0,0.3)`): separates persistent chrome from scrolling content.

### Named Rules
**The One-Object Rule.** A board (or board-styled card) casts one shadow as a whole; its internal rows never cast their own.

## Shapes

Two radius families: **4px** rectangular "plates" for every interactive control (buttons, inputs, the outer board-plate wallet trigger) and **3px** for badges/lamp chips/tag chips — both far tighter than the old pill system. The board container itself (`.board`) uses **8px** for its outer casing edge only; internal rows are square (0 radius) so many rows read as one ruled sheet. The one full-circle exception is the **999px** lamp dot itself — a literal round bulb, not a button; it is never used for anything with a label or click target. No pill-shaped controls remain anywhere in the system.

### Named Rules
**The Plate-Not-Pill Rule.** Every control that used to be `rounded-[500px]` is now a 4px rectangular plate. A pill anywhere in new work is a regression to the retired world.

## Components

### Board Plates (buttons)
- **Shape:** 4px rectangle, uppercase bold label, tracked letter-spacing.
- **Primary:** Nimiq Blue fill, white text, blue glow shadow; darkens on hover.
- **Ghost:** Board Flap Hover fill, Board Flap Ink text, hairline border; border/text tint to Lamp Live on hover.
- **Padding:** `px-3 py-2` default, `px-6 py-3` for hero-scale actions.

### Badges & Lamps
- **Style:** 3px rectangle, Board Flap background, hairline border, 10–11px uppercase bold label.
- **Lamp:** a 7px glowing dot (`box-shadow: 0 0 6px currentColor`) always paired with a status word — colors per the Lamp-Plus-Word Rule.

### Board Rows / App Cards
- **Corner Style:** 8px on the outer container only; internal rows are square.
- **Background:** Board Flap, re-tinting with theme like every other board surface.
- **Category identity:** a solid category-colored label badge in the badge row (e.g. "GAMES"), not a separate icon or stripe — the app's own icon already carries visual identity, so category only needs a legible, named badge, not a second graphic.
- **Signature motion:** the app name flaps in via a per-character cascade on mount (`FlapText` component), once, never looped.

### Inputs / Search
- **Style:** Board Flap background, hairline border, Fira Mono text — reads as a terminal/station input rather than a soft rounded search box.
- **Focus:** border and ring shift to Lamp Live.

### Navigation
- **Header & bottom nav:** Board Frame casing, present on every route, re-tinting with theme. Active items read in Accent Ink (text) with a Lamp Live border/accent; bottom-nav active state adds a 2px Lamp Live top border.
- **Wallet menu:** a `.board` panel — same casing material as the rest of the system, not a separate light dropdown.

## Do's and Don'ts

### Do:
- **Do** let the board (rows, badges, plates, menus) re-tint with the theme toggle like every other surface.
- **Do** keep lamp/status hues theme-constant, but use `accent-ink` (not raw `lamp-live`) for any readable text.
- **Do** pair every status color with an uppercase word, never color alone.
- **Do** use 4px plates for controls and 3px for badges — no pill radius anywhere.
- **Do** reserve the flap-in motion for a row's own key value, played once on mount.

### Don't:
- **Don't** reintroduce `rounded-[500px]` pills — every control is a rectangular plate now.
- **Don't** hardcode a board surface to always-dark; every board token must resolve through the theme-aware CSS variables (`--nq-board-*`), never a bare hex.
- **Don't** use `lamp-live` as a text color on a board surface — it fails contrast in light mode. Use `accent-ink`.
- **Don't** duplicate an app's identity as both its real icon and a separate colored square/stripe — category gets one legible badge, not a second graphic.
- **Don't** loop or idle-animate the flap cascade; it fires once per row on mount, never continuously.
- **Don't** introduce gold/yellow — `--nimiq-gold` remains a dead legacy token.
