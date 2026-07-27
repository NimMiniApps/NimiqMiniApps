import { describe, expect, it } from 'vitest'
import {
  competitionCycleLabel,
  competitionCycleValues,
  normalizeCompetitionCycle,
  normalizeCompetitionCycleFilter,
} from './competition'

describe('competition cycle helpers', () => {
  it('formats cycle labels and omits non-competition apps', () => {
    expect(competitionCycleLabel(1)).toBe('Cycle 1')
    expect(competitionCycleLabel(null)).toBe('')
  })

  it('normalizes missing and form cycle values', () => {
    expect(normalizeCompetitionCycle(undefined)).toBe(null)
    expect(normalizeCompetitionCycle('')).toBe(null)
    expect(normalizeCompetitionCycle('2')).toBe(2)
    expect(normalizeCompetitionCycle(3)).toBe(3)
    expect(normalizeCompetitionCycle('0')).toBe(null)
    expect(normalizeCompetitionCycle('nope')).toBe(null)
  })

  it('extracts sorted unique cycles from apps', () => {
    expect(competitionCycleValues([
      { competition_cycle: 2 },
      { competition_cycle: null },
      { competition_cycle: 1 },
      { competition_cycle: 2 },
    ])).toEqual([1, 2])
  })

  it('normalizes route filter values', () => {
    expect(normalizeCompetitionCycleFilter('none')).toBe('none')
    expect(normalizeCompetitionCycleFilter('2')).toBe('2')
    expect(normalizeCompetitionCycleFilter('0')).toBe('')
    expect(normalizeCompetitionCycleFilter('nope')).toBe('')
    expect(normalizeCompetitionCycleFilter(['1'])).toBe('')
  })
})
