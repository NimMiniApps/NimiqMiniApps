import { updateProfile } from '../api'
import { resolveNimConnectDisplayName } from './nimconnect'

/**
 * Ensure the signed-in wallet has a local catalog display name.
 * If missing, seed from NimConnect (profile display_name, else @handle).
 * Returns the local name to use, or null when nothing could be resolved.
 */
export async function ensureLocalDisplayName(
  walletAddress: string | null | undefined,
  current: string | null | undefined,
  refreshSession: () => Promise<void>,
): Promise<string | null> {
  const existing = current?.trim()
  if (existing) return existing
  if (!walletAddress?.trim()) return null

  const seeded = await resolveNimConnectDisplayName(walletAddress)
  if (!seeded) return null

  try {
    await updateProfile(seeded)
    await refreshSession()
    return seeded
  } catch {
    // Unique conflict or network — leave caller to show a profile CTA.
    return null
  }
}
