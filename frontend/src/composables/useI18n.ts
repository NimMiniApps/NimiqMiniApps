import { t, locale, type MessageKey } from '../i18n'

export function useI18n() {
  return { t, locale }
}

export type { MessageKey }
