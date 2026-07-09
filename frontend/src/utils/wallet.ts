/** Normalize Nimiq addresses for comparison (strip spaces, uppercase). */
export function normalizeWalletAddress(address: string): string {
  return address.replace(/\s+/g, '').toUpperCase()
}

export function walletOwnsApp(walletAddress: string | null | undefined, ownerWalletAddresses: string[] | null | undefined): boolean {
  if (!walletAddress || !ownerWalletAddresses?.length) return false
  const normalized = normalizeWalletAddress(walletAddress)
  return ownerWalletAddresses.some((owner) => normalizeWalletAddress(owner) === normalized)
}
