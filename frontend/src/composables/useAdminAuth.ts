import { computed, ref, watch } from 'vue'
import { adminStats, hasAdminToken } from '../api'
import { useWalletAuth } from './useWalletAuth'

export function useAdminAuth() {
  const { isAdmin: walletIsAdmin, checking: walletChecking, walletAddress } = useWalletAuth()
  const tokenIsAdmin = ref(false)
  const pendingCount = ref(0)
  const checking = ref(true)

  async function loadStats(): Promise<boolean> {
    try {
      const stats = await adminStats()
      pendingCount.value = stats.pending + (stats.pending_updates ?? 0)
      return true
    } catch {
      pendingCount.value = 0
      return false
    }
  }

  async function refresh() {
    if (walletChecking.value) return
    checking.value = true
    if (walletIsAdmin.value) {
      tokenIsAdmin.value = false
      await loadStats()
      checking.value = false
      return
    }
    if (!hasAdminToken()) {
      tokenIsAdmin.value = false
      pendingCount.value = 0
      checking.value = false
      return
    }
    tokenIsAdmin.value = await loadStats()
    checking.value = false
  }

  watch([walletIsAdmin, walletAddress, walletChecking], refresh, { immediate: true })

  const isAdmin = computed(() => walletIsAdmin.value || tokenIsAdmin.value)

  return { isAdmin, pendingCount, checking }
}
