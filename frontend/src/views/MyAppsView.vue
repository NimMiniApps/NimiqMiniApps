<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getMyApps, type App } from '../api'
import AppCard from '../components/AppCard.vue'

const { walletAddress, checking } = useWalletAuth()

const apps = ref<(App & { has_pending_revision: boolean })[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  if (!walletAddress.value) {
    loading.value = false
    return
  }
  loading.value = true
  error.value = ''
  try {
    apps.value = await getMyApps()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load your apps'
  } finally {
    loading.value = false
  }
}

watch([checking, walletAddress], () => {
  if (!checking.value) void load()
}, { immediate: true })
</script>

<template>
  <div class="space-y-5">
    <div>
      <h1 class="text-2xl font-extrabold">My apps</h1>
      <p class="mt-1 text-sm text-muted">Apps linked to your wallet — edit listings or check review status.</p>
    </div>

    <p v-if="checking || loading" class="text-sm text-muted">Loading…</p>
    <p v-else-if="!walletAddress" class="text-sm text-muted">Connect your wallet to see the apps you own.</p>
    <p v-else-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>
    <p v-else-if="apps.length === 0" class="text-sm text-muted">
      No apps linked to this wallet yet. Ask a catalog admin to assign your listings in
      <RouterLink to="/admin" class="font-semibold text-accent-ink hover:underline">Admin</RouterLink>,
      or <RouterLink to="/submit" class="font-semibold text-accent-ink hover:underline">submit a new one</RouterLink>.
    </p>

    <div v-else class="grid gap-4 sm:grid-cols-2">
      <AppCard
        v-for="app in apps"
        :key="app.id"
        :app="app"
        owned
        :pending-update="app.has_pending_revision"
        show-manage-actions
      />
    </div>
  </div>
</template>
