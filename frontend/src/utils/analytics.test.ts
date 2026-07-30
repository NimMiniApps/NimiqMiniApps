import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  analyticsIdentity,
  getAnalyticsSessionId,
  getAnalyticsVisitorId,
  trackCatalogVisit,
  trackProductEvent,
} from './analytics'

function memoryStorage(): Storage {
  const map = new Map<string, string>()
  return {
    get length() { return map.size },
    clear() { map.clear() },
    getItem(key: string) { return map.has(key) ? map.get(key)! : null },
    key(index: number) { return [...map.keys()][index] ?? null },
    removeItem(key: string) { map.delete(key) },
    setItem(key: string, value: string) { map.set(key, value) },
  }
}

describe('analytics identifiers', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', memoryStorage())
    vi.stubGlobal('sessionStorage', memoryStorage())
  })

  it('reuses the visitor id across sessions', () => {
    const first = getAnalyticsVisitorId()
    vi.stubGlobal('sessionStorage', memoryStorage())
    expect(getAnalyticsVisitorId()).toBe(first)
  })

  it('reuses the session id within one session only', () => {
    const first = getAnalyticsSessionId()
    expect(getAnalyticsSessionId()).toBe(first)
    vi.stubGlobal('sessionStorage', memoryStorage())
    expect(getAnalyticsSessionId()).not.toBe(first)
  })
})

describe('analytics delivery', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', memoryStorage())
    vi.stubGlobal('sessionStorage', memoryStorage())
    vi.restoreAllMocks()
    vi.stubGlobal('localStorage', memoryStorage())
    vi.stubGlobal('sessionStorage', memoryStorage())
  })

  it('sends only the strict event payload', async () => {
    const beacon = vi.fn().mockReturnValue(true)
    vi.stubGlobal('navigator', { sendBeacon: beacon })

    trackProductEvent('app_view', 'demo-app')

    expect(beacon).toHaveBeenCalledTimes(1)
    const [url, blob] = beacon.mock.calls[0] as [string, Blob]
    expect(url).toContain('/api/analytics/events')
    expect(blob).toBeInstanceOf(Blob)
    const text = await blob.text()
    expect(JSON.parse(text)).toMatchObject({
      event: 'app_view',
      app_slug: 'demo-app',
    })
    expect(JSON.parse(text)).not.toHaveProperty('wallet')
  })

  it('records a catalog visit once per session', () => {
    const beacon = vi.fn().mockReturnValue(true)
    vi.stubGlobal('navigator', { sendBeacon: beacon })

    trackCatalogVisit()
    trackCatalogVisit()
    expect(beacon).toHaveBeenCalledTimes(1)
  })

  it('does not throw when sendBeacon or fetch fails', () => {
    vi.stubGlobal('navigator', {
      sendBeacon: () => {
        throw new Error('beacon down')
      },
    })
    vi.stubGlobal('fetch', vi.fn(() => Promise.reject(new Error('network'))))
    expect(() => trackProductEvent('app_open', 'demo-app')).not.toThrow()
    expect(analyticsIdentity().analytics_visitor_id).toBeTruthy()
  })
})
