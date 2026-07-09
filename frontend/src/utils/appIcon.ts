import type { App } from '../api'

export function displayIconUrl(app: Pick<App, 'icon_url' | 'discovered_icon_url'>): string | null {
  const icon = app.icon_url?.trim()
  if (icon) return icon
  const discovered = app.discovered_icon_url?.trim()
  if (discovered) return discovered
  return null
}

export function appIdentityAccent(app: Pick<App, 'slug' | 'name'>): string {
  const themes = ['#1f74ff', '#14b8a6', '#f59e0b', '#f43f5e', '#a855f7', '#22c55e']
  const source = app.slug || app.name
  const index = [...source].reduce((sum, char) => sum + char.charCodeAt(0), 0) % themes.length
  return themes[index]
}
