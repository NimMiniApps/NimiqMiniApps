import { describe, expect, it } from 'vitest'
import {
  competitionCycleLabel,
  competitionCycleValues,
  competitionResultLabel,
  displayCompetitionResult,
  normalizeCompetitionCycle,
  normalizeCompetitionCycleFilter,
} from './competition'

describe('competition cycle helpers', () => {
  it('formats cycle labels and omits non-competition apps', () => {
    expect(competitionCycleLabel(1)).toBe('Cycle 1')
    expect(competitionCycleLabel(null)).toBe('')
  })

  it('normalizes positive cycles only', () => {
    expect(normalizeCompetitionCycle(undefined)).toBe(null)
    expect(normalizeCompetitionCycle('')).toBe(null)
    expect(normalizeCompetitionCycle('2')).toBe(2)
    expect(normalizeCompetitionCycle(3)).toBe(3)
    expect(normalizeCompetitionCycle('0')).toBe(null)
    expect(normalizeCompetitionCycle('nope')).toBe(null)
  })

  it('lists unique sorted cycles from apps', () => {
    expect(competitionCycleValues([
      { competition_cycle: 2 },
      { competition_cycle: null },
      { competition_cycle: 1 },
      { competition_cycle: 2 },
    ])).toEqual([1, 2])
  })

  it('normalizes filter query values', () => {
    expect(normalizeCompetitionCycleFilter('none')).toBe('none')
    expect(normalizeCompetitionCycleFilter('2')).toBe('2')
    expect(normalizeCompetitionCycleFilter('0')).toBe('')
    expect(normalizeCompetitionCycleFilter('nope')).toBe('')
    expect(normalizeCompetitionCycleFilter(['1'])).toBe('')
  })

  it('picks the matching or highest competition result', () => {
    expect(displayCompetitionResult({
      competition_cycle: 2,
      competition_results: [
        { cycle: 1, score: 10, place: null },
        { cycle: 2, score: 0, place: 1 },
      ],
    })).toEqual({ cycle: 2, score: 0, place: 1 })

    expect(displayCompetitionResult({
      competition_cycle: null,
      competition_results: [
        { cycle: 1, score: 10, place: null },
        { cycle: 3, score: 5, place: null },
      ],
    })?.cycle).toBe(3)
  })

  it('formats result labels with optional place', () => {
    expect(competitionResultLabel({ cycle: 1, score: 0, place: null })).toBe('Cycle 1 · 0')
    expect(competitionResultLabel({ cycle: 1, score: 0, place: 1 })).toBe('1st place · Cycle 1')
    expect(competitionResultLabel({ cycle: 1, score: 0, place: 2 })).toBe('2nd place · Cycle 1')
    expect(competitionResultLabel({ cycle: 1, score: 0, place: 3 })).toBe('3rd place · Cycle 1')
  })
})
