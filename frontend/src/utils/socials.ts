import type { SocialLink } from '../api'

export const SOCIAL_PLATFORMS = [
  'twitter', 'discord', 'telegram', 'bluesky', 'instagram',
  'youtube', 'linkedin', 'mastodon', 'reddit', 'tiktok',
] as const

export type SocialPlatform = (typeof SOCIAL_PLATFORMS)[number]

const platformLabels: Record<SocialPlatform, string> = {
  twitter: 'X',
  discord: 'Discord',
  telegram: 'Telegram',
  bluesky: 'Bluesky',
  instagram: 'Instagram',
  youtube: 'YouTube',
  linkedin: 'LinkedIn',
  mastodon: 'Mastodon',
  reddit: 'Reddit',
  tiktok: 'TikTok',
}

const platformPlaceholders: Record<SocialPlatform, string> = {
  twitter: '@yourapp or https://x.com/yourapp',
  discord: 'https://discord.gg/invite or invite code',
  telegram: '@channel or https://t.me/channel',
  bluesky: '@handle.bsky.social',
  instagram: '@yourapp',
  youtube: '@channel or YouTube URL',
  linkedin: 'company/name or full URL',
  mastodon: '@user@instance.social or profile URL',
  reddit: 'r/subreddit or u/username',
  tiktok: '@yourapp',
}

export function socialLabel(platform: string): string {
  return platformLabels[platform as SocialPlatform] ?? platform
}

export function socialPlaceholder(platform: SocialPlatform): string {
  return platformPlaceholders[platform]
}

export interface SocialLinkRow {
  id: string
  platform: SocialPlatform | ''
  value: string
}

let rowId = 0
export function newSocialRow(link?: SocialLink): SocialLinkRow {
  rowId += 1
  return {
    id: `social-${rowId}`,
    platform: (link?.platform as SocialPlatform) || '',
    value: link?.url || '',
  }
}

export function socialLinksToRows(items: SocialLink[] | undefined): SocialLinkRow[] {
  const links = items ?? []
  if (!links.length) return []
  return links.map((item) => newSocialRow(item))
}

function stripAt(value: string): string {
  return value.replace(/^@+/, '').trim()
}

function ensureHttps(value: string): string {
  const trimmed = value.trim()
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  return `https://${trimmed.replace(/^\/+/, '')}`
}

function tryParseUrl(value: string): URL | null {
  try {
    return new URL(ensureHttps(value))
  } catch {
    return null
  }
}

const hostPlatformRules: [RegExp, SocialPlatform][] = [
  [/^(www\.)?(x\.com|twitter\.com)$/i, 'twitter'],
  [/^(www\.)?discord\.(gg|com)$/i, 'discord'],
  [/^(www\.)?t\.me$/i, 'telegram'],
  [/^(www\.)?bsky\.app$/i, 'bluesky'],
  [/^(www\.)?instagram\.com$/i, 'instagram'],
  [/^(www\.)?(youtube\.com|youtu\.be)$/i, 'youtube'],
  [/^(www\.)?linkedin\.com$/i, 'linkedin'],
  [/^(www\.)?reddit\.com$/i, 'reddit'],
  [/^(www\.)?tiktok\.com$/i, 'tiktok'],
]

export function detectPlatformFromUrl(value: string): SocialPlatform | null {
  const parsed = tryParseUrl(value)
  if (!parsed) return null
  const host = parsed.hostname.toLowerCase()
  for (const [pattern, platform] of hostPlatformRules) {
    if (pattern.test(host)) return platform
  }
  if (parsed.pathname.includes('@')) return 'mastodon'
  return null
}

export function normalizeSocialUrl(platform: SocialPlatform, raw: string): string {
  const value = raw.trim()
  if (!value) return ''

  if (platform === 'mastodon') {
    const handle = value.startsWith('@') ? value.slice(1) : value
    const atParts = handle.match(/^([^@]+)@([^@]+\.[^@]+)$/)
    if (atParts) return `https://${atParts[2]}/@${atParts[1]}`
  }

  const parsed = tryParseUrl(value)
  if (parsed) {
    const detected = detectPlatformFromUrl(value)
    if (detected === platform || detected === 'mastodon' || platform === 'mastodon') {
      return parsed.toString().replace(/\/$/, '')
    }
  }

  switch (platform) {
    case 'twitter': {
      if (value.startsWith('@')) return `https://x.com/${stripAt(value)}`
      if (/^(x\.com|twitter\.com)\//i.test(value)) return ensureHttps(value)
      if (!value.includes('/') && !value.includes('.')) return `https://x.com/${stripAt(value)}`
      break
    }
    case 'discord': {
      if (/^discord\.(gg|com)\//i.test(value)) return ensureHttps(value)
      if (/^[\w-]{2,}$/i.test(value) && !value.includes('.')) return `https://discord.gg/${value}`
      break
    }
    case 'telegram': {
      if (value.startsWith('@')) return `https://t.me/${stripAt(value)}`
      if (/^t\.me\//i.test(value)) return ensureHttps(value)
      if (!value.includes('/') && !value.includes('.')) return `https://t.me/${stripAt(value)}`
      break
    }
    case 'bluesky': {
      if (value.startsWith('@')) return `https://bsky.app/profile/${stripAt(value)}`
      if (/^[\w.-]+\.bsky\.social$/i.test(stripAt(value))) {
        return `https://bsky.app/profile/${stripAt(value)}`
      }
      break
    }
    case 'instagram': {
      if (value.startsWith('@')) return `https://instagram.com/${stripAt(value)}`
      if (!value.includes('/') && !value.includes('.')) return `https://instagram.com/${stripAt(value)}`
      break
    }
    case 'youtube': {
      if (value.startsWith('@')) return `https://youtube.com/${value}`
      if (/^[\w-]{11}$/.test(value)) return `https://youtu.be/${value}`
      break
    }
    case 'linkedin': {
      if (/^(in|company)\//i.test(value)) return `https://linkedin.com/${value}`
      if (!value.includes('/') && !value.includes('.')) return `https://linkedin.com/in/${stripAt(value)}`
      break
    }
    case 'reddit': {
      if (/^r\/[\w-]+$/i.test(value)) return `https://reddit.com/${value.toLowerCase()}`
      if (/^u\/[\w-]+$/i.test(value)) return `https://reddit.com/${value.toLowerCase()}`
      if (value.startsWith('@')) return `https://reddit.com/u/${stripAt(value)}`
      break
    }
    case 'tiktok': {
      if (value.startsWith('@')) return `https://tiktok.com/${value}`
      if (!value.includes('/') && !value.includes('.')) return `https://tiktok.com/@${stripAt(value)}`
      break
    }
  }

  if (parsed) return parsed.toString().replace(/\/$/, '')
  throw new Error(`Enter a valid ${socialLabel(platform)} URL or handle`)
}

export function compileSocialRows(rows: SocialLinkRow[]): SocialLink[] {
  const out: SocialLink[] = []
  for (const row of rows) {
    if (!row.platform || !row.value.trim()) continue
    const url = normalizeSocialUrl(row.platform, row.value)
    if (!url) continue
    out.push({ platform: row.platform, url })
  }
  const byPlatform = new Map<string, SocialLink>()
  for (const item of out) byPlatform.set(item.platform, item)
  return [...byPlatform.values()]
}

export function parseSocialLines(text: string): SocialLink[] {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const colon = line.indexOf(':')
      const space = line.indexOf(' ')
      if (colon > 0 && (space === -1 || colon < space)) {
        return {
          platform: line.slice(0, colon).trim().toLowerCase(),
          url: line.slice(colon + 1).trim(),
        }
      }
      if (space === -1) return { platform: line.toLowerCase(), url: '' }
      return {
        platform: line.slice(0, space).toLowerCase(),
        url: line.slice(space + 1).trim(),
      }
    })
    .filter((item) => item.platform && item.url)
}

export function formatSocialLines(items: SocialLink[] | undefined): string {
  return (items ?? []).map((item) => `${item.platform} ${item.url}`).join('\n')
}
