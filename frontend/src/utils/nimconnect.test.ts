import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  clearNimConnectHandleCache,
  nimConnectPublicUrl,
  resolveNimConnectHandle,
  setNimConnectHandleLookupForTests,
} from './nimconnect'

afterEach(() => {
  clearNimConnectHandleCache()
  setNimConnectHandleLookupForTests(null)
})

describe('nimConnectPublicUrl', () => {
  it('builds the production public page URL', () => {
    expect(nimConnectPublicUrl('chuck')).toBe('https://nimconnect.nimiqminiapps.com/@chuck')
  })
})

describe('resolveNimConnectHandle', () => {
  it('returns null for empty address', async () => {
    expect(await resolveNimConnectHandle('')).toBeNull()
    expect(await resolveNimConnectHandle(null)).toBeNull()
  })

  it('returns the claimed handle', async () => {
    setNimConnectHandleLookupForTests(async () => ({ handle: 'chuck' }))
    expect(await resolveNimConnectHandle('NQ01 TEST')).toBe('chuck')
  })

  it('returns null on 404-style miss and caches the negative', async () => {
    const lookup = vi.fn(async () => null)
    setNimConnectHandleLookupForTests(lookup)
    expect(await resolveNimConnectHandle('NQ01 NONE')).toBeNull()
    expect(await resolveNimConnectHandle('NQ01 NONE')).toBeNull()
    expect(lookup).toHaveBeenCalledTimes(1)
  })

  it('caches successful lookups by compacted address', async () => {
    const lookup = vi.fn(async () => ({ handle: 'alice' }))
    setNimConnectHandleLookupForTests(lookup)
    expect(await resolveNimConnectHandle('NQ01  ADDR')).toBe('alice')
    expect(await resolveNimConnectHandle('nq01addr')).toBe('alice')
    expect(lookup).toHaveBeenCalledTimes(1)
  })

  it('does not cache thrown errors', async () => {
    const lookup = vi.fn()
      .mockRejectedValueOnce(new Error('network'))
      .mockResolvedValueOnce({ handle: 'bob' })
    setNimConnectHandleLookupForTests(lookup)
    expect(await resolveNimConnectHandle('NQ01 ERR')).toBeNull()
    expect(await resolveNimConnectHandle('NQ01 ERR')).toBe('bob')
    expect(lookup).toHaveBeenCalledTimes(2)
  })
})
