import type { CompetitionResult } from '../api'

type CompetitionEntry = {
  competition_cycle?: number | null
  competition_results?: CompetitionResult[]
}

export function competitionCycleLabel(cycle: number | null | undefined): string {
  return cycle == null ? '' : `Cycle ${cycle}`
}

export function normalizeCompetitionCycle(value: unknown): number | null {
	if (value === '' || value == null) return null
	if (typeof value !== 'number' && typeof value !== 'string') return null
	const cycle = typeof value === 'number' ? value : Number(value)
  return Number.isInteger(cycle) && cycle > 0 ? cycle : null
}

export function competitionCycleValues(apps: CompetitionEntry[]): number[] {
  const values = apps
    .map((app) => normalizeCompetitionCycle(app.competition_cycle))
    .filter((cycle): cycle is number => cycle !== null)
  return [...new Set(values)].sort((left, right) => left - right)
}

export function normalizeCompetitionCycleFilter(value: unknown): string {
  if (value === 'none') return value
  const cycle = normalizeCompetitionCycle(value)
  return cycle === null ? '' : String(cycle)
}

/** Prefer the result for the app's current cycle; else highest cycle. */
export function displayCompetitionResult(app: CompetitionEntry): CompetitionResult | null {
  const results = app.competition_results ?? []
  if (!results.length) return null
  const current = normalizeCompetitionCycle(app.competition_cycle)
  if (current != null) {
    const match = results.find((r) => r.cycle === current)
    if (match) return match
  }
  return results.reduce((best, row) => (row.cycle > best.cycle ? row : best))
}

export function competitionResultLabel(result: CompetitionResult | null | undefined): string {
  if (!result) return ''
  const base = `${competitionCycleLabel(result.cycle)} · ${result.score}`
  if (result.place == null) return base
  const place =
    result.place === 1 ? '1st' : result.place === 2 ? '2nd' : result.place === 3 ? '3rd' : `#${result.place}`
  return `${base} · ${place}`
}
