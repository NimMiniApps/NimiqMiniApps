import type { AnalyticsAppRow, AnalyticsDailyPoint } from '../api'

export type TrendMetric =
  | 'visits'
  | 'unique_visitors'
  | 'wallet_logins'
  | 'app_views'
  | 'activations'

export type AppSortKey =
  | 'name'
  | 'views'
  | 'unique_viewers'
  | 'opens'
  | 'link_copies'
  | 'activations'
  | 'unique_activators'
  | 'conversion'

export function formatAnalyticsChange(current: number, previous: number | null | undefined): string | null {
  if (previous == null || previous === 0) return null
  const pct = ((current - previous) / previous) * 100
  const rounded = Math.round(pct * 10) / 10
  const sign = rounded > 0 ? '+' : ''
  return `${sign}${rounded}%`
}

export function buildTrendPoints(
  daily: AnalyticsDailyPoint[],
  metric: TrendMetric,
  width: number,
  height: number,
): string {
  if (daily.length === 0) return ''
  const values = daily.map((d) => d[metric] ?? 0)
  const max = Math.max(...values, 1)
  const pad = 2
  const innerW = width - pad * 2
  const innerH = height - pad * 2
  if (daily.length === 1) {
    const y = pad + innerH / 2
    return `${pad},${y} ${pad + innerW},${y}`
  }
  return daily
    .map((_, i) => {
      const x = pad + (i / (daily.length - 1)) * innerW
      const y = pad + innerH - (values[i] / max) * innerH
      return `${x},${y}`
    })
    .join(' ')
}

export function sortAnalyticsApps(
  rows: AnalyticsAppRow[],
  key: AppSortKey,
  ascending: boolean,
): AnalyticsAppRow[] {
  const dir = ascending ? 1 : -1
  return [...rows].sort((a, b) => {
    if (key === 'name') return dir * a.name.localeCompare(b.name)
    const av = a[key] ?? -1
    const bv = b[key] ?? -1
    if (av !== bv) return dir * (Number(av) - Number(bv))
    return a.name.localeCompare(b.name)
  })
}

export function filterAnalyticsApps(rows: AnalyticsAppRow[], query: string): AnalyticsAppRow[] {
  const q = query.trim().toLowerCase()
  if (!q) return rows
  return rows.filter((row) => row.name.toLowerCase().includes(q) || row.slug.toLowerCase().includes(q))
}
