// API client. In dev, vite proxies /api to the backend; in production nginx does.
// Set VITE_API_BASE_URL to hit a backend on another host directly.
const BASE = import.meta.env.VITE_API_BASE_URL ?? ''

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
  tags: string[]
  assets: string[]
  status: string
  featured: boolean
  website_url: string | null
  github_url: string | null
  icon_url: string | null
  banner_url: string | null
  screenshots: string[]
  created_at: string
  updated_at: string
  open_url: string
}

export interface Category {
  name: string
  count: number
}

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

export function listApps(params: Record<string, string> = {}): Promise<App[]> {
  const qs = new URLSearchParams(Object.entries(params).filter(([, v]) => v !== ''))
  const s = qs.toString()
  return request(`/api/apps${s ? '?' + s : ''}`)
}

export const getApp = (slug: string) => request<App>(`/api/apps/${slug}`)
export const listCategories = () => request<Category[]>('/api/categories')
export const getDeveloper = (slug: string) => request<Developer>(`/api/developers/${slug}`)

export const submitApp = (app: Partial<App>) =>
  request<App>('/api/apps/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(app),
  })

// --- admin ---

function adminHeaders(): HeadersInit {
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${localStorage.getItem('admin_token') || ''}`,
  }
}

export const adminListApps = () =>
  request<App[]>('/api/admin/apps', { headers: adminHeaders() })

export const adminCreateApp = (app: Partial<App>) =>
  request<App>('/api/admin/apps', { method: 'POST', headers: adminHeaders(), body: JSON.stringify(app) })

export const adminUpdateApp = (slug: string, app: Partial<App>) =>
  request<App>(`/api/admin/apps/${slug}`, { method: 'PUT', headers: adminHeaders(), body: JSON.stringify(app) })

export const adminDeleteApp = (slug: string) =>
  request<void>(`/api/admin/apps/${slug}`, { method: 'DELETE', headers: adminHeaders() })

export const adminSetStatus = (slug: string, action: 'verify' | 'approve' | 'reject') =>
  request<App>(`/api/admin/apps/${slug}/${action}`, { method: 'POST', headers: adminHeaders() })
