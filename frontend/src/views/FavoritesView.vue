<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { useI18n } from '../composables/useI18n'
import { getMyFavorites, type App } from '../api'
import AppCard from '../components/AppCard.vue'
import EmptyState from '../components/EmptyState.vue'

const { walletAddress, checking } = useWalletAuth()
const { t } = useI18n()

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
    error.value = err instanceof Error ? err.message : t('favorites.errorBody')
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
      <h1 class="text-2xl font-extrabold">{{ t('favorites.title') }}</h1>
      <p class="mt-1 text-sm text-muted">{{ t('favorites.subtitle') }}</p>
    </div>

    <p v-if="checking || loading" class="text-sm text-muted">{{ t('common.loading') }}</p>

    <EmptyState
      v-else-if="!walletAddress"
      :title="t('favorites.title')"
      :description="t('favorites.connectWallet')"
    />

    <EmptyState
      v-else-if="error"
      :title="t('favorites.errorTitle')"
      :description="t('favorites.errorBody')"
      variant="error"
    >
      <template #actions>
        <button
          type="button"
          class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
          @click="load"
        >
          {{ t('common.retry') }}
        </button>
      </template>
    </EmptyState>

    <EmptyState
      v-else-if="apps.length === 0"
      :title="t('favorites.emptyTitle')"
      :description="t('favorites.emptyBody')"
    >
      <template #actions>
        <RouterLink
          to="/apps"
          class="cursor-pointer rounded-[500px] nq-primary px-5 py-2.5 text-sm font-bold text-white transition duration-200"
        >
          {{ t('common.browseAll') }}
        </RouterLink>
      </template>
    </EmptyState>

    <div v-else class="grid items-stretch gap-4 sm:grid-cols-2">
      <AppCard v-for="app in apps" :key="app.id" class="h-full min-h-0" :app="app" />
    </div>
  </div>
</template>
