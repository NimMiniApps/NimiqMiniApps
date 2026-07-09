import { describe, expect, it } from 'vitest'
import { buildSparklinePoints } from './sparkline'

describe('buildSparklinePoints', () => {
  it('returns empty string for no data', () => {
    expect(buildSparklinePoints([], 100, 30)).toBe('')
  })

  it('returns a flat line for a single point', () => {
    const pts = buildSparklinePoints([{ date: '2026-01-01', value: 5 }], 100, 30)
    expect(pts).toMatch(/^2,\d+ 98,\d+$/)
    const [, y1] = pts.split(' ')[0].split(',')
    const [, y2] = pts.split(' ')[1].split(',')
    expect(y1).toBe(y2)
  })

  it('scales multiple points across the width', () => {
    const pts = buildSparklinePoints(
      [
        { date: '2026-01-01', value: 0 },
        { date: '2026-01-02', value: 10 },
      ],
      100,
      30,
    )
    expect(pts).toContain('2,')
    expect(pts).toContain('98,')
  })
})
