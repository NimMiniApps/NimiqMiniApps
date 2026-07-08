// API client. Dev can use Vite's /api proxy; production bakes the Traefik API origin.
const BASE = import.meta.env.VITE_API_BASE_URL ?? ''

export interface MediaItem {
  type: 'image' | 'youtube'
  url: string
}

export interface App {
  id: string
  slug: string
  name: string
  domain: string
  category: string
  developer_slug: string
  developer_name: string
  tagline: string
  description: string
  long_description: string
  tags: string[]
  assets: string[]
  status: string
  release_stage: string
  featured: boolean
  featured_order: number
  website_url: string | null
  github_url: string | null
  icon_url: string | null
  banner_url: string | null
  media: MediaItem[]
  created_at: string
  updated_at: string
  open_url: string
}

export const APP_RELEASE_STAGES = ['concept', 'alpha', 'beta', 'released'] as const

export interface Category {
  name: string
  count: number
}

export const APP_CATEGORIES = ['Games', 'Utilities', 'Finance', 'Maps', 'Social', 'Experiments'] as const

export interface Developer {
  slug: string
  name: string
  apps: App[]
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, init)
  if (res.status === 204) return undefined as T
  const body = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(body.error || `HTTP ${res.status}`)
  return body as T
}

type RawApp = App & { screenshots?: string[] }

function normalizeApp(raw: RawApp): App {
  let media = raw.media ?? []
  if (!media.length && raw.screenshots?.length) {
    media = raw.screenshots.map((url) => ({ type: 'image' as const, url }))
  }
  return {
    ...raw,
    long_description: raw.long_description ?? '',
    release_stage: raw.release_stage ?? 'released',
    featured_order: raw.featured_order ?? 0,
    tags: raw.tags ?? [],
    assets: raw.assets ?? [],
    media,
  }
}

function normalizeApps(apps: RawApp[]): App[] {
  return apps.map(normalizeApp)
}

export function listApps(params: Record<string, string> = {}): Promise<App[]> {
  const qs = new URLSearchParams(Object.entries(params).filter(([, v]) => v !== ''))
  const s = qs.toString()
  return request<RawApp[]>(`/api/apps${s ? '?' + s : ''}`).then(normalizeApps)
}

export const getApp = (slug: string) =>
  request<RawApp>(`/api/apps/${slug}`).then(normalizeApp)
export const listCategories = () => request<Category[]>('/api/categories')
export const getDeveloper = (slug: string) =>
  request<{ slug: string; name: string; apps: RawApp[] }>(`/api/developers/${slug}`).then((dev) => ({
    ...dev,
    apps: normalizeApps(dev.apps),
  }))

export const submitApp = (app: Partial<App>) =>
  request<RawApp>('/api/apps/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(app),
  }).then(normalizeApp)

// --- admin ---

function adminHeaders(): HeadersInit {
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${localStorage.getItem('admin_token') || ''}`,
  }
}

export const adminListApps = () =>
  request<RawApp[]>('/api/admin/apps', { headers: adminHeaders() }).then(normalizeApps)

export const adminCreateApp = (app: Partial<App>) =>
  request<RawApp>('/api/admin/apps', { method: 'POST', headers: adminHeaders(), body: JSON.stringify(app) }).then(normalizeApp)

export const adminUpdateApp = (slug: string, app: Partial<App>) =>
  request<RawApp>(`/api/admin/apps/${slug}`, { method: 'PUT', headers: adminHeaders(), body: JSON.stringify(app) }).then(normalizeApp)

export const adminDeleteApp = (slug: string) =>
  request<void>(`/api/admin/apps/${slug}`, { method: 'DELETE', headers: adminHeaders() })

export const adminSetStatus = (slug: string, action: 'verify' | 'approve' | 'reject') =>
  request<RawApp>(`/api/admin/apps/${slug}/${action}`, { method: 'POST', headers: adminHeaders() }).then(normalizeApp)
