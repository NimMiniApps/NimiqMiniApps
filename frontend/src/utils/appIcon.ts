import type { App } from '../api'

export function displayIconUrl(app: Pick<App, 'icon_url' | 'discovered_icon_url'>): string | null {
  const icon = app.icon_url?.trim()
  if (icon) return icon
  const discovered = app.discovered_icon_url?.trim()
  if (discovered) return discovered
  return null
}

export function appIdentityAccent(app: Pick<App, 'slug' | 'name'>): string {
  const themes = ['#0582ca', '#21bca5', '#e9b213', '#fa7268', '#5f4b8b', '#fc8702']
  const source = app.slug || app.name
  const index = [...source].reduce((sum, char) => sum + char.charCodeAt(0), 0) % themes.length
  return themes[index]
}
