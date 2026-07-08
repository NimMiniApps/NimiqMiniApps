import type { MediaItem } from '../api'
import { isYoutubeUrl } from './youtube'

export function parseMediaLines(text: string): MediaItem[] {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((url) => ({ type: isYoutubeUrl(url) ? 'youtube' : 'image', url }))
}

export function formatMediaLines(items: MediaItem[] | undefined): string {
  return (items ?? []).map((item) => item.url).join('\n')
}
