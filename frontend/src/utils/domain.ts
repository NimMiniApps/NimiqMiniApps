/** Strip scheme and trailing slashes from a mini-app domain/path field. */
export function normalizeDomain(domain: string): string {
  let d = domain.trim()
  for (;;) {
    const lower = d.toLowerCase()
    if (lower.startsWith('https://')) {
      d = d.slice(8)
    } else if (lower.startsWith('http://')) {
      d = d.slice(7)
    } else {
      break
    }
  }
  return d.replace(/\/+$/, '')
}
