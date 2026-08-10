import { createProfileClient, type DisplayIdentity } from '@nimconnect/profile-client'

export const NIMCONNECT_PUBLIC_ORIGIN = 'https://nimconnect.nimiqminiapps.com'

type HandleLookup = (address: string) => Promise<{ handle: string } | null>
type IdentityLookup = (address: string) => Promise<DisplayIdentity>

const cache = new Map<string, string | null>()
const identityCache = new Map<string, DisplayIdentity>()
const defaultClient = createProfileClient()
let lookup: HandleLookup = (address) => defaultClient.getHandleByAddress(address)
let identityLookup: IdentityLookup = (address) => defaultClient.getDisplayIdentity(address)

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

/**
 * Best display label from NimConnect: profile display_name, else claimed handle.
 * Returns null when neither is available.
 */
export async function resolveNimConnectDisplayName(
  address: string | undefined | null,
): Promise<string | null> {
  if (!address?.trim()) return null
  const key = cacheKey(address)
  let identity = identityCache.get(key)
  if (!identity) {
    try {
      identity = await identityLookup(address)
      identityCache.set(key, identity)
      if (identity.handle) cache.set(key, identity.handle)
    } catch {
      return null
    }
  }
  const name = identity.displayName?.trim() || identity.handle?.trim()
  return name || null
}

/** Test-only: swap the lookup implementation. Pass null to restore default. */
export function setNimConnectHandleLookupForTests(fn: HandleLookup | null): void {
  lookup = fn ?? ((address) => defaultClient.getHandleByAddress(address))
}

/** Test-only: swap identity lookup. Pass null to restore default. */
export function setNimConnectIdentityLookupForTests(fn: IdentityLookup | null): void {
  identityLookup = fn ?? ((address) => defaultClient.getDisplayIdentity(address))
}

/** Test-only: clear the session cache. */
export function clearNimConnectHandleCache(): void {
  cache.clear()
  identityCache.clear()
}
