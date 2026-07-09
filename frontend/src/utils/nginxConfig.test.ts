import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const nginxConfig = readFileSync(new URL('../../nginx.conf', import.meta.url), 'utf8')

describe('frontend nginx cache policy', () => {
  it('revalidates SPA entry responses to avoid stale Vite chunks after deploys', () => {
    expect(nginxConfig).toContain('Cache-Control "no-cache, no-store, must-revalidate"')
    expect(nginxConfig).toContain('location = /index.html')
    expect(nginxConfig).toContain('location /assets/')
    expect(nginxConfig).toContain('location @missing_asset')
    expect(nginxConfig).toContain('Cache-Control "no-store" always')
  })
})
