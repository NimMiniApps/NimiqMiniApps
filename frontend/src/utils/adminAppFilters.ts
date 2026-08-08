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
