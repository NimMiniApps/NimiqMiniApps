export type GalleryDirection = 'previous' | 'next'

export function orderMediaItems<T extends { type: 'youtube' | 'image' }>(
  items: readonly T[],
): T[] {
  return [
    ...items.filter((item) => item.type === 'youtube'),
    ...items.filter((item) => item.type === 'image'),
  ]
}

export function clampMediaIndex(index: number, itemCount: number): number {
  if (itemCount <= 0) return 0
  return Math.min(Math.max(index, 0), itemCount - 1)
}

export function moveMediaIndex(
  index: number,
  itemCount: number,
  direction: GalleryDirection,
): number {
  if (itemCount <= 0) return 0

  const currentIndex = clampMediaIndex(index, itemCount)
  if (direction === 'previous') {
    return currentIndex === 0 ? itemCount - 1 : currentIndex - 1
  }

  return currentIndex === itemCount - 1 ? 0 : currentIndex + 1
}
