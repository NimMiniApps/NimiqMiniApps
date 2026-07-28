import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { clampMediaIndex, moveMediaIndex } from './mediaGallery'

const mediaGallery = readFileSync(new URL('../components/MediaGallery.vue', import.meta.url), 'utf8')

describe('clampMediaIndex', () => {
  it('returns zero for an empty media list', () => {
    expect(clampMediaIndex(4, 0)).toBe(0)
  })

  it('keeps an index within the media list bounds', () => {
    expect(clampMediaIndex(-1, 3)).toBe(0)
    expect(clampMediaIndex(1, 3)).toBe(1)
    expect(clampMediaIndex(3, 3)).toBe(2)
  })
})

describe('moveMediaIndex', () => {
  it('returns zero for an empty media list', () => {
    expect(moveMediaIndex(0, 0, 'next')).toBe(0)
  })

  it('wraps to the final item when moving previous from the first item', () => {
    expect(moveMediaIndex(0, 3, 'previous')).toBe(2)
  })

  it('wraps to the first item when moving next from the final item', () => {
    expect(moveMediaIndex(2, 3, 'next')).toBe(0)
  })
})

describe('MediaGallery controls', () => {
  it('keeps navigation controls out of the featured media viewport', () => {
    expect(mediaGallery).not.toMatch(/<div class="relative overflow-hidden[^>]*>[\s\S]*?absolute left-3/)
  })

  it('insets thumbnail previews inside their selected frame', () => {
    expect(mediaGallery).toMatch(/class="[^"]*h-16[^"]*p-1/)
  })
})
