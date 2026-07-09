<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getMyFavorites, type App } from '../api'
import AppCard from '../components/AppCard.vue'

const { walletAddress, checking } = useWalletAuth()

const apps = ref<App[]>([])
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
    apps.value = await getMyFavorites()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load favorites'
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
      <h1 class="text-2xl font-extrabold">Favorites</h1>
      <p class="mt-1 text-sm text-muted">Apps you've saved with the heart icon.</p>
    </div>

    <p v-if="checking || loading" class="text-sm text-muted">Loading…</p>
    <p v-else-if="!walletAddress" class="text-sm text-muted">Connect your wallet to see your favorites.</p>
    <p v-else-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>
    <p v-else-if="apps.length === 0" class="text-sm text-muted">
      No favorites yet. Tap the heart on any app to save it here.
    </p>

    <div v-else class="grid items-stretch gap-4 sm:grid-cols-2">
      <AppCard v-for="app in apps" :key="app.id" class="h-full min-h-0" :app="app" />
    </div>
  </div>
</template>
