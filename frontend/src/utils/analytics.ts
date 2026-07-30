export type ProductAnalyticsEvent =
  | 'catalog_visit'
  | 'app_view'
  | 'app_open'
  | 'app_link_copy'

const VISITOR_KEY = 'nma_analytics_visitor_id'
const SESSION_KEY = 'nma_analytics_session_id'
const CATALOG_VISIT_KEY = 'nma_analytics_catalog_visit'

const BASE = import.meta.env.VITE_API_BASE_URL ?? ''

function randomId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

function readOrCreate(storage: Storage | undefined, key: string): string {
  if (!storage) return randomId()
  try {
    const existing = storage.getItem(key)
    if (existing) return existing
    const id = randomId()
    storage.setItem(key, id)
    return id
  } catch {
    return randomId()
  }
}

export function getAnalyticsVisitorId(): string {
  return readOrCreate(typeof localStorage !== 'undefined' ? localStorage : undefined, VISITOR_KEY)
}

export function getAnalyticsSessionId(): string {
  return readOrCreate(typeof sessionStorage !== 'undefined' ? sessionStorage : undefined, SESSION_KEY)
}

export function analyticsIdentity(): {
  analytics_visitor_id: string
  analytics_session_id: string
} {
  return {
    analytics_visitor_id: getAnalyticsVisitorId(),
    analytics_session_id: getAnalyticsSessionId(),
  }
}

function deliver(body: string): void {
  const url = `${BASE}/api/analytics/events`
  try {
    if (typeof navigator !== 'undefined' && typeof navigator.sendBeacon === 'function') {
      const ok = navigator.sendBeacon(url, new Blob([body], { type: 'application/json' }))
      if (ok) return
    }
  } catch {
    /* fall through */
  }
  try {
    void fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body,
      keepalive: true,
    }).catch(() => {})
  } catch {
    /* never block product actions */
  }
}

export function trackProductEvent(event: ProductAnalyticsEvent, appSlug?: string): void {
  const payload: Record<string, string> = {
    event,
    visitor_id: getAnalyticsVisitorId(),
    session_id: getAnalyticsSessionId(),
  }
  if (appSlug) payload.app_slug = appSlug
  deliver(JSON.stringify(payload))
}

export function trackCatalogVisit(): void {
  if (typeof sessionStorage !== 'undefined') {
    try {
      if (sessionStorage.getItem(CATALOG_VISIT_KEY)) return
      sessionStorage.setItem(CATALOG_VISIT_KEY, '1')
    } catch {
      /* continue and send once best-effort */
    }
  }
  trackProductEvent('catalog_visit')
}
