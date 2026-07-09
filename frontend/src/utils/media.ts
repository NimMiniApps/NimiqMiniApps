import type { MediaItem } from '../api'
import { isYoutubeUrl, youtubeVideoId } from './youtube'

export interface MediaRow {
  id: string
  url: string
  type: MediaItem['type'] | null
}

let rowId = 0

export function newMediaRow(item?: MediaItem): MediaRow {
  rowId += 1
  return {
    id: `media-${rowId}`,
    url: item?.url ?? '',
    type: item?.type ?? null,
  }
}

export function mediaItemsToRows(items: MediaItem[] | undefined): MediaRow[] {
  const list = items ?? []
  if (!list.length) return []
  return list.map((item) => newMediaRow(item))
}

function ensureHttps(value: string): string {
  const trimmed = value.trim()
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  return `https://${trimmed.replace(/^\/+/, '')}`
}

export function mediaTypeLabel(type: MediaItem['type'] | null): string {
  if (type === 'youtube') return 'YouTube video'
  if (type === 'image') return 'Screenshot'
  return ''
}

export function normalizeMediaItem(raw: string): MediaItem {
  const value = raw.trim()
  if (!value) throw new Error('URL is required')

  if (/^[a-zA-Z0-9_-]{11}$/.test(value)) {
    return { type: 'youtube', url: `https://www.youtube.com/watch?v=${value}` }
  }

  if (isYoutubeUrl(value)) {
    const id = youtubeVideoId(value)
    if (!id) throw new Error('Enter a valid YouTube link')
    return { type: 'youtube', url: `https://www.youtube.com/watch?v=${id}` }
  }

  let url: string
  try {
    url = new URL(ensureHttps(value)).toString().replace(/\/$/, '')
    const parsed = new URL(url)
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
      throw new Error('invalid')
    }
    if (!parsed.hostname) throw new Error('invalid')
    if (!parsed.hostname.includes('.')) throw new Error('invalid')
  } catch {
    throw new Error('Enter a direct image URL (https://…) or a YouTube link')
  }

  return { type: 'image', url }
}

export function compileMediaRows(rows: MediaRow[]): MediaItem[] {
  const out: MediaItem[] = []
  for (const row of rows) {
    if (!row.url.trim()) continue
    out.push(normalizeMediaItem(row.url))
  }
  return out
}

export function parseMediaLines(text: string): MediaItem[] {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((url) => normalizeMediaItem(url))
}

export function formatMediaLines(items: MediaItem[] | undefined): string {
  return (items ?? []).map((item) => item.url).join('\n')
}

export function youtubeThumbnail(url: string): string | null {
  const id = youtubeVideoId(url)
  return id ? `https://img.youtube.com/vi/${id}/hqdefault.jpg` : null
}
