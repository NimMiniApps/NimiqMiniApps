type CompetitionEntry = {
  competition_cycle?: number | null
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
