import { describe, expect, it } from 'vitest'
import { rewardLabel } from './rewards'

describe('rewardLabel', () => {
  it('formats a single token reward label', () => {
    expect(rewardLabel(['NIM'])).toBe('Earn NIM')
  })

  it('formats multiple token reward labels', () => {
    expect(rewardLabel(['NIM', 'USDT'])).toBe('Earn NIM / USDT')
  })
})
