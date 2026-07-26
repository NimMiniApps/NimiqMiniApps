import type { App } from '../api'

export function displayIconUrl(app: Pick<App, 'icon_url' | 'discovered_icon_url'>): string | null {
  const icon = app.icon_url?.trim()
  if (icon) return icon
  const discovered = app.discovered_icon_url?.trim()
  if (discovered) return discovered
  return null
}

export interface AppTheme {
  accent: string
  soft: string
  ink: string
}

// No gold/yellow: the Nimiq Mini Apps palette excludes it, even though
// --nimiq-gold still exists as a legacy CSS variable elsewhere.
const CATEGORY_THEMES: Record<string, AppTheme> = {
  games: { accent: '#0582ca', soft: 'rgba(5, 130, 202, 0.1)', ink: '#0582ca' },
  utilities: { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)', ink: '#168f80' },
  finance: { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)', ink: '#168f80' },
  maps: { accent: '#d94432', soft: 'rgba(217, 68, 50, 0.13)', ink: '#a3341f' },
  social: { accent: '#fa7268', soft: 'rgba(250, 114, 104, 0.13)', ink: '#c44941' },
  experiments: { accent: '#5f4b8b', soft: 'rgba(95, 75, 139, 0.13)', ink: '#5f4b8b' },
}

const FALLBACK_CATEGORY_THEME: AppTheme = { accent: '#1f2348', soft: 'rgba(31, 35, 72, 0.06)', ink: '#1f2348' }

export function categoryTheme(name: string): AppTheme {
  return CATEGORY_THEMES[name.toLowerCase()] ?? FALLBACK_CATEGORY_THEME
}

const IDENTITY_THEMES: AppTheme[] = [
  { accent: '#0582ca', soft: 'rgba(5, 130, 202, 0.1)', ink: '#0582ca' },
  { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)', ink: '#168f80' },
  { accent: '#d94432', soft: 'rgba(217, 68, 50, 0.13)', ink: '#a3341f' },
  { accent: '#fa7268', soft: 'rgba(250, 114, 104, 0.13)', ink: '#c44941' },
  { accent: '#5f4b8b', soft: 'rgba(95, 75, 139, 0.13)', ink: '#5f4b8b' },
  { accent: '#fc8702', soft: 'rgba(252, 135, 2, 0.14)', ink: '#a15a00' },
]

export function appIdentityTheme(app: Pick<App, 'slug' | 'name'>): AppTheme {
  const source = app.slug || app.name
  const index = [...source].reduce((sum, char) => sum + char.charCodeAt(0), 0) % IDENTITY_THEMES.length
  return IDENTITY_THEMES[index]
}

export function appIdentityAccent(app: Pick<App, 'slug' | 'name'>): string {
  return appIdentityTheme(app).accent
}
