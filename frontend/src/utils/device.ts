import { ref } from 'vue'

const MOBILE_UA =
  /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini|CriOS|FxiOS|EdgiOS|Mobile/i

export function isMobileDevice(): boolean {
  if (typeof navigator === 'undefined' || typeof window === 'undefined') return false

  const uaData = navigator as Navigator & { userAgentData?: { mobile?: boolean } }
  if (uaData.userAgentData?.mobile === true) return true

  if (MOBILE_UA.test(navigator.userAgent)) return true

  // Touch-first devices (covers iOS "Request Desktop Site" in Chrome/Safari)
  if (window.matchMedia('(hover: none) and (pointer: coarse)').matches) return true

  // iPadOS often reports as MacIntel but has multi-touch
  if (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1) return true

  return false
}

export function useIsMobileDevice() {
  return ref(isMobileDevice())
}
