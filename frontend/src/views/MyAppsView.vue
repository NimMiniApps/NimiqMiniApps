<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { useI18n } from '../composables/useI18n'
import { getMyApps, getAppStats, addAppOwner, removeAppOwner, type App, type AppStats } from '../api'
import AppCard from '../components/AppCard.vue'
import EmptyState from '../components/EmptyState.vue'
import StatsSparkline from '../components/StatsSparkline.vue'

const { walletAddress, checking } = useWalletAuth()
const { t } = useI18n()

const apps = ref<(App & { has_pending_revision: boolean })[]>([])
type StatsState =
  | { status: 'loading' }
  | { status: 'error' }
  | { status: 'ready'; data: AppStats }
const statsBySlug = reactive<Record<string, StatsState>>({})
const loading = ref(true)
const error = ref('')

const expandedSlug = ref('')
const newOwnerInput = reactive<Record<string, string>>({})
const ownerError = reactive<Record<string, string>>({})
const ownerBusy = reactive<Record<string, boolean>>({})

function last7DaysSum(daily: AppStats['daily'], metric: 'opens' | 'views'): number {
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - 6)
  const cutoffStr = cutoff.toISOString().slice(0, 10)
  return daily
    .filter((d) => d.date >= cutoffStr)
    .reduce((sum, d) => sum + (metric === 'opens' ? d.opens : d.views), 0)
}

async function loadStats(slugs: string[]) {
  for (const slug of slugs) {
    statsBySlug[slug] = { status: 'loading' }
  }
  await Promise.all(slugs.map(async (slug) => {
    try {
      statsBySlug[slug] = { status: 'ready', data: await getAppStats(slug) }
    } catch {
      statsBySlug[slug] = { status: 'error' }
    }
  }))
}

async function load() {
  if (!walletAddress.value) {
    loading.value = false
    return
  }
  loading.value = true
  error.value = ''
  try {
    apps.value = await getMyApps()
    for (const app of apps.value) {
      statsBySlug[app.slug] = { status: 'loading' }
    }
    void loadStats(apps.value.map((a) => a.slug))
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('myApps.errorBody')
  } finally {
    loading.value = false
  }
}

function toggleManageOwners(slug: string) {
  expandedSlug.value = expandedSlug.value === slug ? '' : slug
}

async function handleAddOwner(slug: string) {
  const wallet = (newOwnerInput[slug] || '').trim()
  if (!wallet) return
  ownerBusy[slug] = true
  ownerError[slug] = ''
  try {
    await addAppOwner(slug, wallet)
    newOwnerInput[slug] = ''
    await load()
  } catch (err) {
    ownerError[slug] = err instanceof Error ? err.message : 'Failed to add owner'
  } finally {
    ownerBusy[slug] = false
  }
}

async function handleRemoveOwner(slug: string, wallet: string) {
  ownerBusy[slug] = true
  ownerError[slug] = ''
  try {
    await removeAppOwner(slug, wallet)
    await load()
  } catch (err) {
    ownerError[slug] = err instanceof Error ? err.message : 'Failed to remove owner'
  } finally {
    ownerBusy[slug] = false
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

    <p v-if="checking || loading" class="text-sm text-muted">{{ t('common.loading') }}</p>

    <EmptyState
      v-else-if="!walletAddress"
      title="My apps"
      :description="t('myApps.connectWallet')"
    />

    <EmptyState
      v-else-if="error"
      :title="t('myApps.errorTitle')"
      :description="t('myApps.errorBody')"
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
      :title="t('myApps.emptyTitle')"
      :description="t('myApps.emptyBody')"
    >
      <template #actions>
        <RouterLink
          to="/submit"
          class="cursor-pointer rounded-[500px] nq-primary px-5 py-2.5 text-sm font-bold text-white transition duration-200"
        >
          {{ t('nav.submit') }}
        </RouterLink>
        <RouterLink
          to="/admin"
          class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
        >
          {{ t('nav.admin') }}
        </RouterLink>
      </template>
    </EmptyState>

    <div v-else class="grid items-stretch gap-4 sm:grid-cols-2">
      <div v-for="app in apps" :key="app.id" class="flex min-h-0 flex-col gap-2">
        <AppCard
          class="h-full min-h-0 flex-1"
          :app="app"
          owned
          :pending-update="app.has_pending_revision"
          show-manage-actions
        />
        <div class="shrink-0 rounded-xl border border-line bg-surface-2/50 p-3 text-sm">
          <div v-if="statsBySlug[app.slug]?.status === 'loading'" class="grid min-h-[5.5rem] grid-cols-2 gap-3">
            <div v-for="n in 2" :key="n" class="space-y-1">
              <div class="h-3 w-12 animate-pulse rounded bg-line/60" />
              <div class="h-6 w-10 animate-pulse rounded bg-line/60" />
              <div class="h-3 w-20 animate-pulse rounded bg-line/60" />
              <div class="mt-1 h-8 w-full animate-pulse rounded bg-line/40" />
            </div>
          </div>
          <div
            v-else-if="statsBySlug[app.slug]?.status === 'ready'"
            class="grid min-h-[5.5rem] grid-cols-2 gap-3"
          >
            <div class="flex flex-col">
              <p class="text-xs font-semibold text-muted">{{ t('myApps.stats.opens') }}</p>
              <p class="text-lg font-bold">{{ statsBySlug[app.slug].data.totals.opens.toLocaleString() }}</p>
              <p class="text-xs text-muted">
                {{ t('myApps.stats.last7Days') }}: {{ last7DaysSum(statsBySlug[app.slug].data.daily, 'opens').toLocaleString() }}
              </p>
              <div class="mt-auto pt-1">
                <StatsSparkline :daily="statsBySlug[app.slug].data.daily" metric="opens" class="w-full" />
              </div>
            </div>
            <div class="flex flex-col">
              <p class="text-xs font-semibold text-muted">{{ t('myApps.stats.views') }}</p>
              <p class="text-lg font-bold">{{ statsBySlug[app.slug].data.totals.views.toLocaleString() }}</p>
              <p class="text-xs text-muted">
                {{ t('myApps.stats.last7Days') }}: {{ last7DaysSum(statsBySlug[app.slug].data.daily, 'views').toLocaleString() }}
              </p>
              <div class="mt-auto pt-1">
                <StatsSparkline :daily="statsBySlug[app.slug].data.daily" metric="views" class="w-full" />
              </div>
            </div>
          </div>
          <p v-else class="min-h-[5.5rem] text-xs text-muted">Stats unavailable</p>
        </div>
        <button type="button" class="shrink-0 text-left text-xs font-semibold text-accent-ink hover:underline"
          @click="toggleManageOwners(app.slug)">
          {{ expandedSlug === app.slug ? 'Hide owners' : `Manage owners (${app.owner_wallet_addresses.length})` }}
        </button>
        <div v-if="expandedSlug === app.slug" class="space-y-2 rounded-xl border border-line bg-surface-2/50 p-3 text-sm">
          <ul class="space-y-1">
            <li v-for="wallet in app.owner_wallet_addresses" :key="wallet" class="flex items-center justify-between gap-2">
              <span class="truncate font-mono text-xs">{{ wallet }}</span>
              <button type="button" :disabled="ownerBusy[app.slug] || app.owner_wallet_addresses.length <= 1"
                class="shrink-0 text-xs font-semibold text-red-600 hover:underline disabled:cursor-default disabled:opacity-40 dark:text-red-400"
                @click="handleRemoveOwner(app.slug, wallet)">
                Remove
              </button>
            </li>
          </ul>
          <div class="flex gap-2">
            <input v-model="newOwnerInput[app.slug]" placeholder="Wallet address (e.g. your other device)"
              class="min-w-0 flex-1 rounded-lg border border-line bg-surface px-2 py-1.5 text-xs outline-none focus:border-accent" />
            <button type="button" :disabled="ownerBusy[app.slug]"
              class="shrink-0 rounded-lg bg-accent px-3 py-1.5 text-xs font-semibold text-white disabled:opacity-50"
              @click="handleAddOwner(app.slug)">
              Add
            </button>
          </div>
          <p v-if="ownerError[app.slug]" class="text-xs text-red-600 dark:text-red-400">{{ ownerError[app.slug] }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
