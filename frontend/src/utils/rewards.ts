export function rewardLabel(assets: string[] | null | undefined): string {
  const tokens = (assets ?? []).filter(Boolean)
  return tokens.length ? `Earn ${tokens.join(' / ')}` : ''
}
