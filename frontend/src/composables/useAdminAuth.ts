import { ref, onMounted } from 'vue'
import { adminStats, hasAdminToken } from '../api'

export function useAdminAuth() {
  const isAdmin = ref(false)
  const pendingCount = ref(0)
  const checking = ref(true)

  onMounted(async () => {
    if (!hasAdminToken()) {
      checking.value = false
      return
    }
    try {
      const stats = await adminStats()
      isAdmin.value = true
      pendingCount.value = stats.pending + (stats.pending_updates ?? 0)
    } catch {
      isAdmin.value = false
      pendingCount.value = 0
    } finally {
      checking.value = false
    }
  })

  return { isAdmin, pendingCount, checking }
}
