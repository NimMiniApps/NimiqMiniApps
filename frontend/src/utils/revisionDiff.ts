import type { App, AppRevision } from '../api'
import { formatMediaLines } from './media'
import { formatSocialLines } from './socials'

export interface FieldChange {
  field: string
  label: string
  before: string
  after: string
}

function str(v: unknown): string {
  if (v == null || v === '') return '—'
  if (Array.isArray(v)) return v.length ? v.join(', ') : '—'
  return String(v)
}

function mediaStr(media: App['media']) {
  return formatMediaLines(media || []) || '—'
}

function socialsStr(socials: App['socials']) {
  return formatSocialLines(socials || []) || '—'
}

export function diffRevision(current: App, revision: AppRevision): FieldChange[] {
  const pairs: [string, string, string, string][] = [
    ['name', 'Name', str(current.name), str(revision.name)],
    ['domain', 'Domain', str(current.domain), str(revision.domain)],
    ['category', 'Category', str(current.category), str(revision.category)],
    ['developer_name', 'Developer', str(current.developer_name), str(revision.developer_name)],
    ['tagline', 'Tagline', str(current.tagline), str(revision.tagline)],
    ['description', 'Short description', str(current.description), str(revision.description)],
    ['long_description', 'Full description', str(current.long_description), str(revision.long_description)],
    ['release_stage', 'Release stage', str(current.release_stage), str(revision.release_stage)],
    ['tags', 'Tags', str(current.tags), str(revision.tags)],
    ['assets', 'Assets', str(current.assets), str(revision.assets)],
    ['reward_assets', 'Reward assets', str(current.reward_assets), str(revision.reward_assets)],
    ['website_url', 'Website', str(current.website_url), str(revision.website_url)],
    ['github_url', 'GitHub', str(current.github_url), str(revision.github_url)],
    ['icon_url', 'Icon URL', str(current.icon_url), str(revision.icon_url)],
    ['banner_url', 'Banner URL', str(current.banner_url), str(revision.banner_url)],
    ['media', 'Media', mediaStr(current.media), mediaStr(revision.media)],
    ['socials', 'Social links', socialsStr(current.socials), socialsStr(revision.socials)],
  ]

  return pairs
    .filter(([, , before, after]) => before !== after)
    .map(([field, label, before, after]) => ({ field, label, before, after }))
}
