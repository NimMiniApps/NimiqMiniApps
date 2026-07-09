/** Normalize Nimiq addresses for comparison (strip spaces, uppercase). */
export function normalizeWalletAddress(address: string): string {
  return address.replace(/\s+/g, '').toUpperCase()
}

export function walletOwnsApp(walletAddress: string | null | undefined, appWallet: string | null | undefined): boolean {
  if (!walletAddress || !appWallet) return false
  return normalizeWalletAddress(walletAddress) === normalizeWalletAddress(appWallet)
}
