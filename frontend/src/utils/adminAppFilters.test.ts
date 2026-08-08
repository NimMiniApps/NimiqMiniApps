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
