# Admin App List Filters Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add category, status, and text search filters to `/admin`, synced to the URL, so moderators can manage a larger catalog.

**Architecture:** Pure filter helpers in `frontend/src/utils/adminAppFilters.ts` (unit-tested). Wire them into `AdminView.vue` with client-side filter → existing sort pipeline, URL query sync modeled on `AppsView.vue`. No API changes.

**Tech Stack:** Vue 3, vue-router, Vitest, existing `APP_CATEGORIES` / app status strings

**Design:** @docs/plans/2026-08-09-admin-app-list-filters-design.md

---

### Task 1: Filter helper + tests

**Files:**
- Create: `frontend/src/utils/adminAppFilters.ts`
- Create: `frontend/src/utils/adminAppFilters.test.ts`

**Step 1: Write the failing tests**

```ts
import { describe, expect, it } from 'vitest'
import {
  ADMIN_APP_STATUSES,
  filterAdminApps,
  normalizeAdminCategoryFilter,
  normalizeAdminStatusFilter,
  type AdminAppFilterable,
} from './adminAppFilters'

const apps: AdminAppFilterable[] = [
  { name: 'Swap Desk', slug: 'swap-desk', domain: 'swap.example.com', category: 'Finance', status: 'verified' },
  { name: 'Nim Race', slug: 'nim-race', domain: 'race.example.com', category: 'Games', status: 'submitted' },
  { name: 'Map Pin', slug: 'map-pin', domain: 'maps.example.com', category: 'Maps', status: 'approved' },
]

describe('admin app filters', () => {
  it('exposes known statuses', () => {
    expect(ADMIN_APP_STATUSES).toEqual([
      'submitted', 'approved', 'verified', 'experimental', 'rejected',
    ])
  })

  it('normalizes category and status from the route', () => {
    expect(normalizeAdminCategoryFilter('Games')).toBe('Games')
    expect(normalizeAdminCategoryFilter('Nope')).toBe('')
    expect(normalizeAdminCategoryFilter(['Games'])).toBe('')
    expect(normalizeAdminStatusFilter('submitted')).toBe('submitted')
    expect(normalizeAdminStatusFilter('bogus')).toBe('')
    expect(normalizeAdminStatusFilter(undefined)).toBe('')
  })

  it('filters by category, status, and case-insensitive q on name/slug/domain', () => {
    expect(filterAdminApps(apps, { category: 'Games', status: '', q: '' }).map((a) => a.slug))
      .toEqual(['nim-race'])
    expect(filterAdminApps(apps, { category: '', status: 'verified', q: '' }).map((a) => a.slug))
      .toEqual(['swap-desk'])
    expect(filterAdminApps(apps, { category: '', status: '', q: 'SWAP' }).map((a) => a.slug))
      .toEqual(['swap-desk'])
    expect(filterAdminApps(apps, { category: '', status: '', q: 'maps.example' }).map((a) => a.slug))
      .toEqual(['map-pin'])
    expect(filterAdminApps(apps, { category: 'Finance', status: 'submitted', q: '' })).toEqual([])
  })
})
```

**Step 2: Run tests to verify they fail**

Run: `cd frontend && npm test -- src/utils/adminAppFilters.test.ts`

Expected: FAIL (module not found)

**Step 3: Implement helpers**

```ts
import { APP_CATEGORIES } from '../api'

export const ADMIN_APP_STATUSES = [
  'submitted', 'approved', 'verified', 'experimental', 'rejected',
] as const

export type AdminAppStatus = (typeof ADMIN_APP_STATUSES)[number]

export type AdminAppFilterable = {
  name: string
  slug: string
  domain: string
  category: string
  status: string
}

export type AdminAppFilters = {
  q: string
  category: string
  status: string
}

function asSingleString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

export function normalizeAdminCategoryFilter(value: unknown): string {
  const raw = asSingleString(value)
  return (APP_CATEGORIES as readonly string[]).includes(raw) ? raw : ''
}

export function normalizeAdminStatusFilter(value: unknown): string {
  const raw = asSingleString(value)
  return (ADMIN_APP_STATUSES as readonly string[]).includes(raw) ? raw : ''
}

export function filterAdminApps<T extends AdminAppFilterable>(
  apps: T[],
  filters: AdminAppFilters,
): T[] {
  const q = filters.q.trim().toLowerCase()
  return apps.filter((app) => {
    if (filters.category && app.category !== filters.category) return false
    if (filters.status && app.status !== filters.status) return false
    if (!q) return true
    return (
      app.name.toLowerCase().includes(q) ||
      app.slug.toLowerCase().includes(q) ||
      app.domain.toLowerCase().includes(q)
    )
  })
}
```

**Step 4: Run tests to verify they pass**

Run: `cd frontend && npm test -- src/utils/adminAppFilters.test.ts`

Expected: PASS

**Step 5: Commit**

```bash
git add frontend/src/utils/adminAppFilters.ts frontend/src/utils/adminAppFilters.test.ts
git commit -m "feat(web): add admin app list filter helpers"
```

---

### Task 2: Wire filters into AdminView

**Files:**
- Modify: `frontend/src/views/AdminView.vue`

**Step 1: Imports and filter state**

Add imports for `watch` (already imported? check — currently `ref, reactive, computed, onMounted, nextTick`; add `watch`), and:

```ts
import {
  ADMIN_APP_STATUSES,
  filterAdminApps,
  normalizeAdminCategoryFilter,
  normalizeAdminStatusFilter,
} from '../utils/adminAppFilters'
```

After `sortAsc` refs, add:

```ts
const filterQ = ref(typeof route.query.q === 'string' ? route.query.q : '')
const filterCategory = ref(normalizeAdminCategoryFilter(route.query.category))
const filterStatus = ref(normalizeAdminStatusFilter(route.query.status))

const hasListFilters = computed(() =>
  !!(filterQ.value.trim() || filterCategory.value || filterStatus.value),
)

const filteredApps = computed(() =>
  filterAdminApps(apps.value, {
    q: filterQ.value,
    category: filterCategory.value,
    status: filterStatus.value,
  }),
)
```

Change `sortedApps` to sort `filteredApps.value` instead of `apps.value`.

**Step 2: URL sync (preserve `edit` and other keys)**

```ts
function syncFilterQuery() {
  const query: Record<string, string | string[] | undefined | null> = { ...route.query }
  if (filterQ.value.trim()) query.q = filterQ.value.trim()
  else delete query.q
  if (filterCategory.value) query.category = filterCategory.value
  else delete query.category
  if (filterStatus.value) query.status = filterStatus.value
  else delete query.status
  router.replace({ query })
}

let filterQTimer: ReturnType<typeof setTimeout> | undefined
watch(filterQ, () => {
  clearTimeout(filterQTimer)
  filterQTimer = setTimeout(() => syncFilterQuery(), 200)
})
watch([filterCategory, filterStatus], () => syncFilterQuery())

watch(() => route.query, (query) => {
  const nextQ = typeof query.q === 'string' ? query.q : ''
  const nextCategory = normalizeAdminCategoryFilter(query.category)
  const nextStatus = normalizeAdminStatusFilter(query.status)
  if (filterQ.value !== nextQ) filterQ.value = nextQ
  if (filterCategory.value !== nextCategory) filterCategory.value = nextCategory
  if (filterStatus.value !== nextStatus) filterStatus.value = nextStatus
})
```

Note: when writing back from `route.query` into `filterQ`, the `watch(filterQ)` debounce will fire — acceptable; `syncFilterQuery` should be a no-op if query already matches. Optionally skip sync when values equal current query to avoid loops:

```ts
function syncFilterQuery() {
  const nextQ = filterQ.value.trim()
  const curQ = typeof route.query.q === 'string' ? route.query.q : ''
  const curCat = normalizeAdminCategoryFilter(route.query.category)
  const curStatus = normalizeAdminStatusFilter(route.query.status)
  if (
    nextQ === curQ &&
    filterCategory.value === curCat &&
    filterStatus.value === curStatus
  ) return
  // ... replace as above
}
```

**Step 3: Template — filter bar + row category + empty state**

Replace the Sort-only block (~lines 718–733) with a filter+sort bar:

```html
<div v-if="!showForm && apps.length" class="space-y-3">
  <div class="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-end">
    <label class="min-w-[12rem] flex-1 text-xs">
      <span class="mb-1 block font-semibold text-muted">Search</span>
      <input v-model="filterQ" type="search" placeholder="Name, slug, or domain"
        class="w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm outline-none focus:border-accent" />
    </label>
    <label class="text-xs">
      <span class="mb-1 block font-semibold text-muted">Category</span>
      <select v-model="filterCategory"
        class="w-full cursor-pointer rounded-lg border border-line bg-surface px-3 py-2 text-sm outline-none focus:border-accent sm:w-40">
        <option value="">All</option>
        <option v-for="category in APP_CATEGORIES" :key="category" :value="category">{{ category }}</option>
      </select>
    </label>
    <label class="text-xs">
      <span class="mb-1 block font-semibold text-muted">Status</span>
      <select v-model="filterStatus"
        class="w-full cursor-pointer rounded-lg border border-line bg-surface px-3 py-2 text-sm outline-none focus:border-accent sm:w-40">
        <option value="">All</option>
        <option v-for="s in ADMIN_APP_STATUSES" :key="s" :value="s">{{ s }}</option>
      </select>
    </label>
  </div>
  <div class="flex flex-wrap items-center gap-2 text-xs font-semibold text-muted">
    <span>Sort:</span>
    <!-- existing sort buttons unchanged -->
    <span v-if="hasListFilters" class="ml-auto font-medium">
      {{ sortedApps.length }} of {{ apps.length }} apps
    </span>
  </div>
</div>
```

In each list row subtitle (currently `{{ app.slug }} · {{ app.domain }}`), add category:

```html
<p class="truncate text-sm text-muted">{{ app.slug }} · {{ app.category }} · {{ app.domain }}</p>
```

After the `v-for` list (or wrap list), add empty filtered state:

```html
<p v-if="!showForm && apps.length && !sortedApps.length" class="text-sm text-muted">
  No apps match these filters
</p>
```

Keep `v-if="!showForm && apps.length"` for the filter bar so it only shows when there is a catalog to filter. List loop stays on `sortedApps`.

**Step 4: Manual check**

Run: `cd frontend && npm test && npm run build`

Expected: tests pass; build succeeds.

In browser (dev or deployed preview): open `/admin`, confirm filters narrow the list, refresh keeps `?q=&category=&status=`, `?edit=slug` still opens the form without wiping filters, moderation queues unchanged.

**Step 5: Commit**

```bash
git add frontend/src/views/AdminView.vue
git commit -m "feat(web): filter admin app list by category, status, and search"
```

---

### Task 3: Docs touch-up (only if user-facing)

**Files:**
- Modify: `README.md` only if admin docs mention list management — skip if nothing user/dev-facing needs updating (internal admin UX; AGENTS.md already covers admin). Prefer skip.

**Step 1:** Confirm no README change required → done.

No commit if nothing changed.
