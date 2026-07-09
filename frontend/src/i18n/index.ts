import { messages, SUPPORTED_LOCALES, type Locale, type MessageTree } from './messages'

declare global {
  interface Window {
    nimiqPay?: { readonly language?: string }
  }
}

export function resolveLocale(): Locale {
  const host = window.nimiqPay?.language?.toLowerCase().slice(0, 2)
  if (host && SUPPORTED_LOCALES.includes(host as Locale)) return host as Locale

  const browser = navigator.language?.toLowerCase().slice(0, 2)
  if (browser && SUPPORTED_LOCALES.includes(browser as Locale)) return browser as Locale

  return 'en'
}

type Path = {
  [K in keyof MessageTree]: `${K & string}.${keyof MessageTree[K] & string}`
}[keyof MessageTree]

export type MessageKey = Path

function getNested(tree: MessageTree, key: string): string | undefined {
  return key.split('.').reduce<unknown>((node, part) => {
    if (node && typeof node === 'object' && part in node) {
      return (node as Record<string, unknown>)[part]
    }
    return undefined
  }, tree) as string | undefined
}

export function createTranslator(locale: Locale) {
  const tree = messages[locale] ?? messages.en

  return function t(key: MessageKey, vars?: Record<string, string>): string {
    const template = getNested(tree, key) ?? getNested(messages.en, key) ?? key
    if (!vars) return template
    return template.replace(/\{(\w+)\}/g, (_, name: string) => vars[name] ?? `{${name}}`)
  }
}

export const locale = resolveLocale()
export const t = createTranslator(locale)
