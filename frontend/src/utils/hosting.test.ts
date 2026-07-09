import { describe, expect, it } from 'vitest'
import { isNimiqMiniAppsHosted } from './hosting'

describe('isNimiqMiniAppsHosted', () => {
  it('detects the apex domain and exact subdomains', () => {
    expect(isNimiqMiniAppsHosted('nimiqminiapps.com')).toBe(true)
    expect(isNimiqMiniAppsHosted('nimfeed.nimiqminiapps.com')).toBe(true)
    expect(isNimiqMiniAppsHosted('nimiqminiapps.com/nimfeed')).toBe(true)
  })

  it('rejects lookalike domains', () => {
    expect(isNimiqMiniAppsHosted('fake-nimiqminiapps.com')).toBe(false)
    expect(isNimiqMiniAppsHosted('nimiqminiapps.com.evil.example')).toBe(false)
  })
})
