import { describe, expect, it } from 'vitest'
import type { AnalyticsAppRow, AnalyticsDailyPoint } from '../api'
import {
  buildTrendPoints,
  filterAnalyticsApps,
  formatAnalyticsChange,
  sortAnalyticsApps,
} from './adminAnalytics'

const sampleApps: AnalyticsAppRow[] = [
  {
    slug: 'zeta', name: 'Zeta', views: 10, unique_viewers: 5, opens: 2, link_copies: 1,
    activations: 3, unique_activators: 2, conversion: 40,
  },
  {
    slug: 'alpha', name: 'Alpha', views: 20, unique_viewers: 10, opens: 8, link_copies: 0,
    activations: 8, unique_activators: 4, conversion: 40,
  },
  {
    slug: 'beta', name: 'Beta', views: 0, unique_viewers: 0, opens: 0, link_copies: 0,
    activations: 0, unique_activators: 0, conversion: null,
  },
]

describe('formatAnalyticsChange', () => {
  it('formats positive and negative deltas', () => {
    expect(formatAnalyticsChange(150, 100)).toBe('+50%')
    expect(formatAnalyticsChange(50, 100)).toBe('-50%')
  })

  it('returns null for zero or missing previous', () => {
    expect(formatAnalyticsChange(10, 0)).toBeNull()
    expect(formatAnalyticsChange(10, null)).toBeNull()
    expect(formatAnalyticsChange(10, undefined)).toBeNull()
  })
})

describe('buildTrendPoints', () => {
  it('returns empty string for empty series', () => {
    expect(buildTrendPoints([], 'visits', 100, 40)).toBe('')
  })

  it('builds a flat line for constant values', () => {
    const daily: AnalyticsDailyPoint[] = [
      { day: '2026-01-01', visits: 5, wallet_logins: 0, app_views: 0, app_opens: 0, link_copies: 0, activations: 0, unique_visitors: 0, unique_app_viewers: 0, unique_activators: 0 },
      { day: '2026-01-02', visits: 5, wallet_logins: 0, app_views: 0, app_opens: 0, link_copies: 0, activations: 0, unique_visitors: 0, unique_app_viewers: 0, unique_activators: 0 },
    ]
    const points = buildTrendPoints(daily, 'visits', 100, 40)
    expect(points.split(' ')).toHaveLength(2)
  })
})

describe('sortAnalyticsApps', () => {
  it('sorts by activations descending with name tie-break', () => {
    const sorted = sortAnalyticsApps(sampleApps, 'activations', false)
    expect(sorted.map((a) => a.slug)).toEqual(['alpha', 'zeta', 'beta'])
  })

  it('stable-ties by app name', () => {
    const sorted = sortAnalyticsApps(sampleApps, 'conversion', false)
    expect(sorted[0].slug).toBe('alpha')
    expect(sorted[1].slug).toBe('zeta')
  })
})

describe('filterAnalyticsApps', () => {
  it('filters case-insensitively by name or slug', () => {
    expect(filterAnalyticsApps(sampleApps, 'ALP').map((a) => a.slug)).toEqual(['alpha'])
    expect(filterAnalyticsApps(sampleApps, 'ZETA').map((a) => a.slug)).toEqual(['zeta'])
  })
})
