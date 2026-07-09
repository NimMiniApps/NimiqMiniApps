import { describe, expect, it } from 'vitest'
import {
  compileSocialRows,
  detectPlatformFromUrl,
  normalizeSocialUrl,
  newSocialRow,
} from './socials'

describe('normalizeSocialUrl', () => {
  it('normalizes X handles and URLs', () => {
    expect(normalizeSocialUrl('twitter', '@myapp')).toBe('https://x.com/myapp')
    expect(normalizeSocialUrl('twitter', 'https://twitter.com/myapp')).toBe('https://twitter.com/myapp')
  })

  it('normalizes Discord invites', () => {
    expect(normalizeSocialUrl('discord', 'abc123')).toBe('https://discord.gg/abc123')
    expect(normalizeSocialUrl('discord', 'https://discord.gg/abc123')).toBe('https://discord.gg/abc123')
  })

  it('normalizes Telegram handles', () => {
    expect(normalizeSocialUrl('telegram', '@nimiq')).toBe('https://t.me/nimiq')
  })

  it('normalizes Bluesky handles', () => {
    expect(normalizeSocialUrl('bluesky', '@user.bsky.social')).toBe('https://bsky.app/profile/user.bsky.social')
  })

  it('normalizes Reddit paths', () => {
    expect(normalizeSocialUrl('reddit', 'r/nimiq')).toBe('https://reddit.com/r/nimiq')
    expect(normalizeSocialUrl('reddit', 'u/builder')).toBe('https://reddit.com/u/builder')
  })

  it('normalizes Mastodon handles', () => {
    expect(normalizeSocialUrl('mastodon', '@user@mastodon.social')).toBe('https://mastodon.social/@user')
  })
})

describe('detectPlatformFromUrl', () => {
  it('detects platform from host', () => {
    expect(detectPlatformFromUrl('https://discord.gg/abc')).toBe('discord')
    expect(detectPlatformFromUrl('https://x.com/nimiq')).toBe('twitter')
  })
})

describe('compileSocialRows', () => {
  it('skips empty rows and dedupes by platform', () => {
    const rows = [
      newSocialRow({ platform: 'twitter', url: '@a' }),
      newSocialRow({ platform: 'twitter', url: '@b' }),
      newSocialRow({ platform: 'discord', url: 'invite' }),
      newSocialRow(),
    ]
    const compiled = compileSocialRows(rows)
    expect(compiled).toHaveLength(2)
    expect(compiled.find((item) => item.platform === 'twitter')?.url).toBe('https://x.com/b')
    expect(compiled.find((item) => item.platform === 'discord')?.url).toBe('https://discord.gg/invite')
  })
})
