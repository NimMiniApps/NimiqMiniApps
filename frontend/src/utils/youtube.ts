const YOUTUBE_RE =
  /(?:youtube\.com\/(?:watch\?v=|embed\/|shorts\/)|youtu\.be\/)([a-zA-Z0-9_-]{11})/

export function youtubeVideoId(url: string): string | null {
  const match = url.match(YOUTUBE_RE)
  return match?.[1] ?? null
}

export function isYoutubeUrl(url: string): boolean {
  return youtubeVideoId(url) !== null
}

export function youtubeEmbedUrl(url: string): string | null {
  const id = youtubeVideoId(url)
  return id ? `https://www.youtube-nocookie.com/embed/${id}` : null
}
