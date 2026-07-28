import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { clampMediaIndex, moveMediaIndex, orderMediaItems } from './mediaGallery'

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

describe('orderMediaItems', () => {
  it('puts videos first without changing order within each media type', () => {
    const items = [
      { type: 'image' as const, url: 'first-image' },
      { type: 'youtube' as const, url: 'first-video' },
      { type: 'image' as const, url: 'second-image' },
      { type: 'youtube' as const, url: 'second-video' },
    ]

    expect(orderMediaItems(items).map((item) => item.url)).toEqual([
      'first-video',
      'second-video',
      'first-image',
      'second-image',
    ])
    expect(items.map((item) => item.url)).toEqual([
      'first-image',
      'first-video',
      'second-image',
      'second-video',
    ])
  })
})

describe('MediaGallery controls', () => {
  it('renders and navigates the video-first media list', () => {
    expect(mediaGallery).toContain('const orderedItems = computed(() => orderMediaItems(props.items))')
    expect(mediaGallery).toContain('v-for="(item, index) in orderedItems"')
    expect(mediaGallery).not.toContain('v-for="(item, index) in props.items"')
  })

  it('keeps navigation controls out of the featured media viewport', () => {
    expect(mediaGallery).not.toMatch(/<div class="relative overflow-hidden[^>]*>[\s\S]*?absolute left-3/)
  })

  it('insets thumbnail previews inside their selected frame', () => {
    expect(mediaGallery).toMatch(/class="[^"]*h-16[^"]*p-1/)
  })

  it('keeps the selected thumbnail highlight inside its frame', () => {
    expect(mediaGallery).toContain("'ring-2 ring-inset ring-accent'")
    expect(mediaGallery).not.toContain('ring-offset-2 ring-offset-base')
  })

  it('keeps navigation controls adjacent to the thumbnail strip', () => {
    expect(mediaGallery).not.toContain('min-w-0 flex-1 snap-x')
  })

  it('opens featured images in an accessible native dialog', () => {
    expect(mediaGallery).toContain('ref="imageDialog"')
    expect(mediaGallery).toContain('<dialog')
    expect(mediaGallery).toContain('@click.self="closeImageDialog"')
    expect(mediaGallery).toContain(':aria-label="`View screenshot ${selectedIndex + 1} full screen`"')
    expect(mediaGallery).toContain('aria-label="Close full-screen image"')
  })
})
