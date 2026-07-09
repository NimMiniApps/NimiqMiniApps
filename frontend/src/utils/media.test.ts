import { describe, expect, it } from 'vitest'
import { compileMediaRows, normalizeMediaItem, newMediaRow } from './media'

describe('normalizeMediaItem', () => {
  it('detects and normalizes YouTube URLs', () => {
    expect(normalizeMediaItem('https://youtu.be/dQw4w9WgXcQ')).toEqual({
      type: 'youtube',
      url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
    })
    expect(normalizeMediaItem('dQw4w9WgXcQ')).toEqual({
      type: 'youtube',
      url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
    })
  })

  it('accepts direct image URLs', () => {
    expect(normalizeMediaItem('https://example.com/shot.png')).toEqual({
      type: 'image',
      url: 'https://example.com/shot.png',
    })
  })

  it('rejects invalid values', () => {
    expect(() => normalizeMediaItem('not-a-url')).toThrow(/Enter a direct image URL/)
  })
})

describe('compileMediaRows', () => {
  it('skips empty rows and preserves order', () => {
    const rows = [
      newMediaRow({ type: 'image', url: 'https://example.com/a.png' }),
      newMediaRow(),
      newMediaRow({ type: 'youtube', url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' }),
    ]
    expect(compileMediaRows(rows)).toEqual([
      { type: 'image', url: 'https://example.com/a.png' },
      { type: 'youtube', url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' },
    ])
  })
})
