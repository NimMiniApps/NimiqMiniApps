const DEFAULT_TITLE = 'Nimiq Mini Apps'
const DEFAULT_DESCRIPTION =
  'Discover community-curated mini apps for the Nimiq Pay wallet — games, tools, maps and more, open straight from your wallet.'

export const DEFAULT_OG_IMAGE = `${typeof window !== 'undefined' ? window.location.origin : ''}/og-default.svg`

function setMeta(attr: 'name' | 'property', key: string, content: string) {
  let el = document.querySelector(`meta[${attr}="${key}"]`) as HTMLMetaElement | null
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    document.head.appendChild(el)
  }
  el.content = content
}

function setCanonical(url: string) {
  let el = document.querySelector('link[rel="canonical"]') as HTMLLinkElement | null
  if (!el) {
    el = document.createElement('link')
    el.setAttribute('rel', 'canonical')
    document.head.appendChild(el)
  }
  el.href = url
}

function setJsonLd(data: object | null) {
  let el = document.querySelector('script#jsonld') as HTMLScriptElement | null
  if (!data) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('script')
    el.id = 'jsonld'
    el.type = 'application/ld+json'
    document.head.appendChild(el)
  }
  el.textContent = JSON.stringify(data)
}

export interface PageMeta {
  title?: string
  description?: string
  image?: string | null
  url?: string
  jsonLd?: object | null
}

export function setPageMeta(meta: PageMeta = {}) {
  const title = meta.title ? `${meta.title} · Nimiq Mini Apps` : DEFAULT_TITLE
  const description = meta.description ?? DEFAULT_DESCRIPTION
  const url = meta.url ?? window.location.href
  const image = meta.image || DEFAULT_OG_IMAGE

  document.title = title
  setMeta('name', 'description', description)
  setMeta('property', 'og:title', title)
  setMeta('property', 'og:description', description)
  setMeta('property', 'og:url', url)
  setMeta('property', 'og:type', 'website')
  setMeta('property', 'og:image', image)
  setMeta('name', 'twitter:card', 'summary_large_image')
  setMeta('name', 'twitter:title', title)
  setMeta('name', 'twitter:description', description)
  setMeta('name', 'twitter:image', image)
  setCanonical(url)
  setJsonLd(meta.jsonLd ?? null)
}

export function resetPageMeta() {
  setPageMeta({
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'WebSite',
      name: 'Nimiq Mini Apps',
      description: DEFAULT_DESCRIPTION,
      url: typeof window !== 'undefined' ? window.location.origin : '',
    },
  })
}
