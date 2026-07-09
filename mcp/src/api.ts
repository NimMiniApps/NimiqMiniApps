const baseUrl = () => (process.env.MINIAPPS_API_URL ?? 'http://localhost:8080').replace(/\/$/, '')
const adminToken = () => process.env.MINIAPPS_ADMIN_TOKEN ?? ''

export function apiInfo() {
  return {
    api_url: baseUrl(),
    admin_configured: adminToken().length > 0,
  }
}

function requireAdmin(): string {
  const token = adminToken()
  if (!token) {
    throw new Error('MINIAPPS_ADMIN_TOKEN is not set — admin tools are unavailable')
  }
  return token
}

async function request(path: string, init: RequestInit = {}): Promise<unknown> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch(`${baseUrl()}${path}`, { ...init, headers })
  if (res.status === 204) {
    return { ok: true, status: 204 }
  }

  const body = await res.json().catch(() => ({}))
  if (!res.ok) {
    const msg = typeof body === 'object' && body && 'error' in body ? String(body.error) : `HTTP ${res.status}`
    throw new Error(msg)
  }
  return body
}

function adminHeaders(): HeadersInit {
  return {
    Authorization: `Bearer ${requireAdmin()}`,
    'Content-Type': 'application/json',
  }
}

export function buildQuery(params: Record<string, string | boolean | undefined>): string {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === '') continue
    qs.set(key, String(value))
  }
  const s = qs.toString()
  return s ? `?${s}` : ''
}

// --- public ---

export async function healthCheck() {
  return request('/health')
}

export async function listApps(params: {
  q?: string
  category?: string
  developer?: string
  tag?: string
  asset?: string
  status?: string
  featured?: boolean
  sort?: 'featured' | 'newest' | 'name'
  limit?: number
  offset?: number
  paginate?: boolean
}) {
  const query: Record<string, string> = {}
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === '') continue
    query[key] = String(value)
  }
  return request(`/api/apps${buildQuery(query)}`)
}

export async function getApp(slug: string) {
  return request(`/api/apps/${encodeURIComponent(slug)}`)
}

export async function listCategories() {
  return request('/api/categories')
}

export async function getDeveloper(slug: string) {
  return request(`/api/developers/${encodeURIComponent(slug)}`)
}

export async function listDevelopers() {
  return request('/api/developers')
}

export async function getRelatedApps(slug: string) {
  return request(`/api/apps/${encodeURIComponent(slug)}/related`)
}

// --- admin ---

export async function adminListApps() {
  return request('/api/admin/apps', { headers: adminHeaders() })
}

export async function adminCreateApp(app: Record<string, unknown>) {
  return request('/api/admin/apps', {
    method: 'POST',
    headers: adminHeaders(),
    body: JSON.stringify(app),
  })
}

export async function adminUpdateApp(slug: string, app: Record<string, unknown>) {
  const all = (await adminListApps()) as Array<Record<string, unknown>>
  const current = all.find((a) => a.slug === slug)
  if (!current) throw new Error('app not found')
  return request(`/api/admin/apps/${encodeURIComponent(slug)}`, {
    method: 'PATCH',
    headers: adminHeaders(),
    body: JSON.stringify({ ...current, ...app }),
  })
}

export async function adminDeleteApp(slug: string) {
  return request(`/api/admin/apps/${encodeURIComponent(slug)}`, {
    method: 'DELETE',
    headers: adminHeaders(),
  })
}

export async function adminSetStatus(slug: string, action: 'verify' | 'approve' | 'reject') {
  return request(`/api/admin/apps/${encodeURIComponent(slug)}/${action}`, {
    method: 'POST',
    headers: adminHeaders(),
  })
}

export function asToolResult(data: unknown) {
  return {
    content: [{ type: 'text' as const, text: JSON.stringify(data, null, 2) }],
  }
}
