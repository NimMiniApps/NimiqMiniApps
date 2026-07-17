import { createProfileClient } from '@nimconnect/profile-client'

export const NIMCONNECT_PUBLIC_ORIGIN = 'https://nimconnect.nimiqminiapps.com'

type HandleLookup = (address: string) => Promise<{ handle: string } | null>

const cache = new Map<string, string | null>()
const defaultClient = createProfileClient()
let lookup: HandleLookup = (address) => defaultClient.getHandleByAddress(address)

function cacheKey(address: string): string {
  return address.replace(/\s+/g, '').toUpperCase()
}

/** Public NimConnect page for a claimed handle. */
export function nimConnectPublicUrl(handle: string): string {
  return `${NIMCONNECT_PUBLIC_ORIGIN}/@${encodeURIComponent(handle)}`
}

/**
 * Resolve a wallet's claimed @handle. Returns null when missing or on failure.
 * Session-caches successes and 404s; does not cache thrown errors.
 */
export async function resolveNimConnectHandle(
  address: string | undefined | null,
): Promise<string | null> {
  if (!address?.trim()) return null
  const key = cacheKey(address)
  if (cache.has(key)) return cache.get(key) ?? null
  try {
    const claim = await lookup(address)
    const handle = claim?.handle ?? null
    cache.set(key, handle)
    return handle
  } catch {
    return null
  }
}

/** Test-only: swap the lookup implementation. Pass null to restore default. */
export function setNimConnectHandleLookupForTests(fn: HandleLookup | null): void {
  lookup = fn ?? ((address) => defaultClient.getHandleByAddress(address))
}

/** Test-only: clear the session cache. */
export function clearNimConnectHandleCache(): void {
  cache.clear()
}
