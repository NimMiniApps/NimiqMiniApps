<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getMyApps, type App } from '../api'

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
  <div class="mx-auto max-w-2xl space-y-5">
    <h1 class="text-xl font-extrabold">My apps</h1>

    <p v-if="checking || loading" class="text-sm text-muted">Loading…</p>
    <p v-else-if="!walletAddress" class="text-sm text-muted">Connect your wallet to see the apps you own.</p>
    <p v-else-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>
    <p v-else-if="apps.length === 0" class="text-sm text-muted">
      No apps linked to this wallet yet. <RouterLink to="/submit" class="font-semibold text-accent-ink hover:underline">Submit one</RouterLink>.
    </p>

    <ul v-else class="space-y-3">
      <li v-for="app in apps" :key="app.slug" class="rounded-2xl border border-line bg-surface p-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <RouterLink :to="`/apps/${app.slug}`" class="font-bold hover:underline">{{ app.name }}</RouterLink>
            <p class="text-sm text-muted">{{ app.status }}<span v-if="app.has_pending_revision"> · update pending review</span></p>
          </div>
          <RouterLink :to="`/apps/${app.slug}/update`"
            class="shrink-0 rounded-xl border border-line bg-surface-2 px-3 py-1.5 text-sm font-semibold hover:border-accent/50">
            Request update
          </RouterLink>
        </div>
      </li>
    </ul>
  </div>
</template>
