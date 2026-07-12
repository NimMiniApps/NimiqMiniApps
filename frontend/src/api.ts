// API client. Dev can use Vite's /api proxy; production bakes the Traefik API origin.
const BASE = import.meta.env.VITE_API_BASE_URL ?? ''

export interface MediaItem {
  type: 'image' | 'youtube'
  url: string
}

export interface SocialLink {
  platform: string
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
  owner_wallet_addresses: string[]
  tagline: string
  description: string
  long_description: string
  tags: string[]
  assets: string[]
  reward_assets: string[]
  status: string
  release_stage: string
  featured: boolean
  featured_order: number
  website_url: string | null
  github_url: string | null
  icon_url: string | null
  discovered_icon_url: string | null
  banner_url: string | null
  media: MediaItem[]
  socials: SocialLink[]
  created_at: string
  updated_at: string
  open_url: string
  domain_reachable: boolean | null
  domain_checked_at: string | null
  avg_rating: number
  review_count: number
  submitter_contact?: string
  total_opens?: number
  total_views?: number
}

export const APP_RELEASE_STAGES = ['concept', 'alpha', 'beta', 'released'] as const

export interface Category {
  name: string
  count: number
}

export const APP_CATEGORIES = ['Games', 'Utilities', 'Finance', 'Maps', 'Social', 'Experiments'] as const
export const APP_ASSETS = ['NIM', 'USDT', 'USDC', 'BTC', 'ETH'] as const

export interface Developer {
  slug: string
  name: string
  apps: App[]
}

export interface DeveloperSummary {
  slug: string
  name: string
  app_count: number
}

export interface SubmissionStatus {
  slug: string
  name: string
  status: 'pending' | 'live' | 'rejected' | string
  raw_status: string
  public: boolean
  updated_at: string
  update_pending?: boolean
  rejection_note?: string
}

export interface AppRevision {
  id: string
  app_slug: string
  status: string
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
  reward_assets: string[]
  release_stage: string
  website_url: string | null
  github_url: string | null
  icon_url: string | null
  banner_url: string | null
  media: MediaItem[]
  socials: SocialLink[]
  author_note: string
  created_at: string
  reviewed_at: string | null
}

export interface RevisionReviewItem {
  revision: AppRevision
  current: App
}

export interface CatalogCollection {
  id: string
  title: string
  description: string
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
    owner_wallet_addresses: raw.owner_wallet_addresses ?? [],
    long_description: raw.long_description ?? '',
    release_stage: raw.release_stage ?? 'released',
    featured_order: raw.featured_order ?? 0,
    tags: raw.tags ?? [],
    assets: raw.assets ?? [],
    reward_assets: raw.reward_assets ?? [],
    media,
    socials: raw.socials ?? [],
  }
}

function normalizeApps(apps: RawApp[]): App[] {
  return apps.map(normalizeApp)
}

export interface PaginatedApps {
  items: App[]
  total: number
  limit: number
  offset: number
}

export function listApps(params: Record<string, string> = {}): Promise<App[]> {
  const qs = new URLSearchParams(Object.entries(params).filter(([, v]) => v !== ''))
  const s = qs.toString()
  return request<RawApp[]>(`/api/apps${s ? '?' + s : ''}`).then(normalizeApps)
}

export function listAppsPaginated(params: Record<string, string> = {}): Promise<PaginatedApps> {
  const qs = new URLSearchParams(Object.entries({ paginate: '1', ...params }).filter(([, v]) => v !== ''))
  const s = qs.toString()
  return request<{ items: RawApp[]; total: number; limit: number; offset: number }>(`/api/apps?${s}`).then((body) => ({
    items: normalizeApps(body.items ?? []),
    total: body.total,
    limit: body.limit,
    offset: body.offset,
  }))
}

export const getApp = (slug: string) =>
  request<RawApp>(`/api/apps/${slug}`).then(normalizeApp)
export const getSubmissionStatus = (slug: string) =>
  request<SubmissionStatus>(`/api/apps/${slug}/status`)
export const getRelatedApps = (slug: string) =>
  request<RawApp[]>(`/api/apps/${slug}/related`).then(normalizeApps)
export const listCategories = () => request<Category[]>('/api/categories')
export const listCatalogCollections = () => request<CatalogCollection[]>('/api/collections')
export const listDevelopers = () => request<DeveloperSummary[]>('/api/developers')
export const getDeveloper = (slug: string) =>
  request<{ slug: string; name: string; apps: RawApp[] }>(`/api/developers/${slug}`).then((dev) => ({
    ...dev,
    apps: normalizeApps(dev.apps),
  }))

export const submitApp = (app: Partial<App>) =>
  request<RawApp>('/api/apps/submit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(app),
  }).then(normalizeApp)

export const requestAppUpdate = (slug: string, app: Partial<App> & { author_note?: string }) =>
  request<{ revision_id: string; app_slug: string; status: string }>(
    `/api/apps/${encodeURIComponent(slug)}/request-update`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(app),
    },
  )

export const addAppOwner = (slug: string, walletAddress: string) =>
  request<{ status: string }>(`/api/apps/${encodeURIComponent(slug)}/owners`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ wallet_address: walletAddress }),
  })

export const removeAppOwner = (slug: string, walletAddress: string) =>
  request<{ status: string }>(
    `/api/apps/${encodeURIComponent(slug)}/owners/${encodeURIComponent(walletAddress)}`,
    { method: 'DELETE', credentials: 'include' },
  )

export const getMyApps = () =>
  request<(RawApp & { has_pending_revision: boolean })[]>('/api/my/apps', { credentials: 'include' })
    .then((items) => items.map((item) => ({ ...normalizeApp(item), has_pending_revision: item.has_pending_revision })))

export const getMyFavorites = () =>
  request<RawApp[]>('/api/my/favorites', { credentials: 'include' }).then(normalizeApps)

export const addFavorite = (slug: string) =>
  request<void>(`/api/apps/${encodeURIComponent(slug)}/favorite`, { method: 'POST', credentials: 'include' })

export const removeFavorite = (slug: string) =>
  request<void>(`/api/apps/${encodeURIComponent(slug)}/favorite`, { method: 'DELETE', credentials: 'include' })

export interface AppStats {
  totals: { opens: number; views: number }
  daily: { date: string; opens: number; views: number }[]
}

export const getAppStats = (slug: string) =>
  request<AppStats>(`/api/apps/${encodeURIComponent(slug)}/stats`, { credentials: 'include' })

export function trackAppEvent(slug: string, event: 'open' | 'view'): void {
  const url = BASE + `/api/apps/${encodeURIComponent(slug)}/track`
  const body = JSON.stringify({ event })
  navigator.sendBeacon(url, new Blob([body], { type: 'application/json' }))
}

// --- admin ---

function adminHeaders(): HeadersInit {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const token = localStorage.getItem('admin_token')
  if (token) headers.Authorization = `Bearer ${token}`
  return headers
}

function adminRequest<T>(path: string, init?: RequestInit): Promise<T> {
  return request<T>(path, {
    ...init,
    credentials: 'include',
    headers: { ...adminHeaders(), ...(init?.headers as Record<string, string> | undefined) },
  })
}

export const adminListApps = () =>
  adminRequest<RawApp[]>('/api/admin/apps').then(normalizeApps)

export interface AdminUserResult {
  wallet_address: string
  display_name: string | null
}

export const adminSearchUsers = (q: string) =>
  adminRequest<AdminUserResult[]>(`/api/admin/users?q=${encodeURIComponent(q)}`)

export const adminAddAppOwner = (slug: string, walletAddress: string) =>
  adminRequest<{ status: string }>(`/api/admin/apps/${slug}/owners`, {
    method: 'POST',
    body: JSON.stringify({ wallet_address: walletAddress }),
  })

export const adminRemoveAppOwner = (slug: string, walletAddress: string) =>
  adminRequest<{ status: string }>(
    `/api/admin/apps/${slug}/owners/${encodeURIComponent(walletAddress)}`,
    { method: 'DELETE' },
  )

export const adminStats = () =>
  adminRequest<{ pending: number; unreachable: number; pending_updates: number }>('/api/admin/stats')

export const adminListRevisions = () =>
  adminRequest<RevisionReviewItem[]>('/api/admin/revisions')

export const adminApproveRevision = (id: string) =>
  adminRequest<RawApp>(`/api/admin/revisions/${id}/approve`, { method: 'POST' }).then(normalizeApp)

export const adminRejectRevision = (id: string) =>
  adminRequest<{ status: string }>(`/api/admin/revisions/${id}/reject`, { method: 'POST' })

export const adminCheckDomains = () =>
  adminRequest<{ status: string }>('/api/admin/check-domains', { method: 'POST' })

export const adminCreateApp = (app: Partial<App>) =>
  adminRequest<RawApp>('/api/admin/apps', { method: 'POST', body: JSON.stringify(app) }).then(normalizeApp)

export const adminUpdateApp = (slug: string, app: Partial<App>) =>
  adminRequest<RawApp>(`/api/admin/apps/${slug}`, { method: 'PUT', body: JSON.stringify(app) }).then(normalizeApp)

export const adminDeleteApp = (slug: string) =>
  adminRequest<void>(`/api/admin/apps/${slug}`, { method: 'DELETE' })

export const adminSetStatus = (slug: string, action: 'verify' | 'approve' | 'reject', note?: string) =>
  adminRequest<RawApp>(`/api/admin/apps/${slug}/${action}`, {
    method: 'POST',
    ...(action === 'reject' ? { body: JSON.stringify({ note: note ?? '' }) } : {}),
  }).then(normalizeApp)

export function hasAdminToken(): boolean {
  return !!localStorage.getItem('admin_token')
}

// --- wallet auth & reviews ---

export interface AppReview {
  id: string
  app_id: string
  wallet_address: string
  display_name: string | null
  rating: number
  body: string
  created_at: string
  updated_at: string
}

export interface AppReviewsResponse {
  items: AppReview[]
  average: number
  count: number
}

export const authChallenge = (wallet_address: string) =>
  request<{ nonce: string; message: string }>('/api/auth/challenge', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ wallet_address }),
  })

export const authVerify = (payload: {
  wallet_address: string
  nonce: string
  signature: string
  public_key: string
}) =>
  request<{ wallet_address: string }>('/api/auth/verify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(payload),
  })

export interface AuthMe {
  wallet_address: string
  display_name: string | null
  is_admin: boolean
}

export const authMe = () =>
  request<AuthMe>('/api/auth/me', { credentials: 'include' })

export const authLogout = () =>
  request<void>('/api/auth/logout', { method: 'POST', credentials: 'include' })

export const listAppReviews = (slug: string) =>
  request<AppReviewsResponse>(`/api/apps/${encodeURIComponent(slug)}/reviews`)

export const submitAppReview = (slug: string, rating: number, body: string) =>
  request<AppReview>(`/api/apps/${encodeURIComponent(slug)}/reviews`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ rating, body }),
  })

export const deleteOwnAppReview = (slug: string) =>
  request<void>(`/api/apps/${encodeURIComponent(slug)}/reviews`, {
    method: 'DELETE',
    credentials: 'include',
  })

export interface Profile {
  wallet_address: string
  display_name: string | null
}

export const getProfile = () =>
  request<Profile>('/api/profile', { credentials: 'include' })

export const updateProfile = (display_name: string) =>
  request<Profile>('/api/profile', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ display_name }),
  })
