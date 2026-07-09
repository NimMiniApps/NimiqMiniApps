const NIMIQ_MINI_APPS_HOST = 'nimiqminiapps.com'

export function isNimiqMiniAppsHosted(domain: string): boolean {
  const trimmed = domain.trim()
  if (!trimmed) return false

  try {
    const url = new URL(/^https?:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`)
    const host = url.hostname.toLowerCase().replace(/\.$/, '')
    return host === NIMIQ_MINI_APPS_HOST || host.endsWith(`.${NIMIQ_MINI_APPS_HOST}`)
  } catch {
    return false
  }
}
