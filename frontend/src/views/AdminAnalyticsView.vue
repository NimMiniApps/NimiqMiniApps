<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  adminAnalytics,
  type AdminAnalyticsResponse,
  type AnalyticsRange,
  type AnalyticsMetrics,
} from '../api'
import { useAdminAuth } from '../composables/useAdminAuth'
import AnalyticsTrendChart from '../components/AnalyticsTrendChart.vue'
import {
  filterAnalyticsApps,
  formatAnalyticsChange,
  sortAnalyticsApps,
  type AppSortKey,
  type TrendMetric,
} from '../utils/adminAnalytics'

const { isAdmin, checking } = useAdminAuth()

const ranges: { key: AnalyticsRange; label: string }[] = [
  { key: '7d', label: '7 days' },
  { key: '30d', label: '30 days' },
  { key: '90d', label: '90 days' },
  { key: 'all', label: 'All time' },
]

const range = ref<AnalyticsRange>('30d')
const data = ref<AdminAnalyticsResponse | null>(null)
const loading = ref(false)
const error = ref('')
const query = ref('')
const sortKey = ref<AppSortKey>('activations')
const sortAsc = ref(false)
const trendMetric = ref<TrendMetric>('visits')

const trendOptions: { key: TrendMetric; label: string }[] = [
  { key: 'visits', label: 'Visits' },
  { key: 'unique_visitors', label: 'Unique visitors' },
  { key: 'wallet_logins', label: 'Wallet logins' },
  { key: 'app_views', label: 'App views' },
  { key: 'activations', label: 'Activations' },
]

const summaryCards: { key: keyof AnalyticsMetrics; label: string; hint: string }[] = [
  { key: 'visits', label: 'Visits', hint: 'One catalog visit per browser session' },
  { key: 'unique_visitors', label: 'Unique visitors', hint: 'Pseudonymous browser identifiers' },
  { key: 'wallet_logins', label: 'Wallet logins', hint: 'Successful signed logins' },
  { key: 'unique_wallets', label: 'Unique wallets', hint: 'Server-hashed wallets only' },
  { key: 'app_views', label: 'App views', hint: 'Detail views, once per app per session' },
  { key: 'activations', label: 'Activations', hint: 'Nimiq Pay opens + successful link copies' },
]

async function load() {
  if (!isAdmin.value) return
  loading.value = true
  error.value = ''
  try {
    data.value = await adminAnalytics(range.value)
  } catch (e) {
    data.value = null
    error.value = (e as Error).message || 'Failed to load analytics'
  } finally {
    loading.value = false
  }
}

watch(isAdmin, (ok) => {
  if (ok) void load()
})
watch(range, () => {
  void load()
})
onMounted(() => {
  if (isAdmin.value) void load()
})

function changeFor(key: keyof AnalyticsMetrics): string | null {
  if (!data.value?.previous) return null
  return formatAnalyticsChange(data.value.current[key], data.value.previous[key])
}

const filteredApps = computed(() => {
  if (!data.value) return []
  return sortAnalyticsApps(filterAnalyticsApps(data.value.apps, query.value), sortKey.value, sortAsc.value)
})

function toggleSort(key: AppSortKey) {
  if (sortKey.value === key) {
    sortAsc.value = !sortAsc.value
  } else {
    sortKey.value = key
    sortAsc.value = key === 'name'
  }
}

const collectionLabel = computed(() => {
  if (!data.value?.collection_started_at) return null
  return new Date(data.value.collection_started_at).toLocaleDateString()
})

const isEmpty = computed(() => {
  if (!data.value) return false
  return data.value.current.visits === 0
    && data.value.current.app_views === 0
    && data.value.current.activations === 0
    && data.value.apps.length === 0
})

const trendLabel = computed(() => trendOptions.find((o) => o.key === trendMetric.value)?.label ?? 'Metric')
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 class="text-2xl font-extrabold">Product analytics</h1>
        <p class="mt-1 text-sm text-muted">
          Catalog visit → login → app view → Nimiq Pay open or link copy
        </p>
      </div>
      <RouterLink to="/admin" class="text-sm font-semibold text-accent-ink hover:underline">
        Back to moderation
      </RouterLink>
    </div>

    <p v-if="checking" class="text-sm text-muted">Checking admin access…</p>
    <div v-else-if="!isAdmin" class="rounded-xl border border-line bg-surface p-5 text-sm text-muted">
      Admin access required. Sign in with an admin wallet or token on the
      <RouterLink to="/admin" class="font-semibold text-accent-ink hover:underline">moderation page</RouterLink>.
    </div>

    <template v-else>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="item in ranges"
          :key="item.key"
          type="button"
          class="rounded-xl border px-3 py-1.5 text-sm font-semibold transition-colors"
          :class="range === item.key
            ? 'border-accent bg-accent-soft text-accent-ink'
            : 'border-line bg-surface text-muted hover:border-accent/40'"
          @click="range = item.key"
        >
          {{ item.label }}
        </button>
      </div>

      <p v-if="collectionLabel" class="text-xs text-muted">
        Unique visitor and funnel collection started {{ collectionLabel }}.
        Earlier lifetime open/view totals may predate this history.
      </p>

      <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-sm text-red-600 dark:text-red-300">
        {{ error }}
        <button type="button" class="ml-3 font-semibold underline" @click="load">Retry</button>
      </p>

      <div v-if="loading && !data" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="n in 6" :key="n" class="h-24 animate-pulse rounded-xl border border-line bg-surface-2" />
      </div>

      <template v-else-if="data">
        <p v-if="isEmpty" class="rounded-xl border border-line bg-surface p-5 text-sm text-muted">
          No analytics events in this period yet.
        </p>

        <section class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <article
            v-for="card in summaryCards"
            :key="card.key"
            class="rounded-xl border border-line bg-surface p-4"
            :title="card.hint"
          >
            <p class="text-xs font-semibold uppercase tracking-wide text-muted">{{ card.label }}</p>
            <p class="mt-1 text-2xl font-extrabold tabular-nums">{{ data.current[card.key] }}</p>
            <p
              v-if="changeFor(card.key)"
              class="mt-1 text-xs font-semibold tabular-nums"
              :class="(changeFor(card.key) || '').startsWith('-') ? 'text-red-500' : 'text-emerald-600'"
            >
              {{ changeFor(card.key) }} vs prior period
            </p>
            <p v-else-if="range === 'all'" class="mt-1 text-xs text-muted">All-time · no comparison</p>
          </article>
        </section>

        <section class="rounded-xl border border-line bg-surface p-4 space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <h2 class="font-bold">Daily trend</h2>
            <div class="flex flex-wrap gap-1">
              <button
                v-for="opt in trendOptions"
                :key="opt.key"
                type="button"
                class="rounded-lg border px-2.5 py-1 text-xs font-semibold"
                :class="trendMetric === opt.key
                  ? 'border-accent bg-accent-soft text-accent-ink'
                  : 'border-line text-muted'"
                @click="trendMetric = opt.key"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>
          <AnalyticsTrendChart :daily="data.daily" :metric="trendMetric" :label="trendLabel" />
        </section>

        <section class="rounded-xl border border-line bg-surface p-4 space-y-3">
          <h2 class="font-bold">Unique visitor funnel</h2>
          <ol class="grid gap-3 sm:grid-cols-3">
            <li
              v-for="stage in data.funnel"
              :key="stage.key"
              class="rounded-lg border border-line bg-surface-2 px-3 py-3"
            >
              <p class="text-xs font-semibold uppercase tracking-wide text-muted">{{ stage.label }}</p>
              <p class="mt-1 text-xl font-extrabold tabular-nums">{{ stage.count }}</p>
              <p v-if="stage.rate_from_prior != null" class="mt-1 text-xs text-muted">
                {{ Math.round(stage.rate_from_prior * 10) / 10 }}% of prior stage
              </p>
            </li>
          </ol>
        </section>

        <section class="space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <h2 class="font-bold">Per-app</h2>
            <input
              v-model="query"
              type="search"
              placeholder="Search apps"
              class="w-full max-w-xs rounded-xl border border-line bg-surface px-3 py-2 text-sm outline-none focus:border-accent sm:w-64"
            />
          </div>

          <div class="hidden overflow-x-auto rounded-xl border border-line md:block">
            <table class="min-w-full text-left text-sm">
              <thead class="bg-surface-2 text-xs uppercase tracking-wide text-muted">
                <tr>
                  <th class="px-3 py-2"><button type="button" @click="toggleSort('name')">App</button></th>
                  <th class="px-3 py-2"><button type="button" @click="toggleSort('views')">Views</button></th>
                  <th class="px-3 py-2"><button type="button" @click="toggleSort('unique_viewers')">Unique</button></th>
                  <th class="px-3 py-2"><button type="button" @click="toggleSort('opens')">Opens</button></th>
                  <th class="px-3 py-2"><button type="button" @click="toggleSort('link_copies')">Copies</button></th>
                  <th class="px-3 py-2"><button type="button" @click="toggleSort('activations')">Activations</button></th>
                  <th class="px-3 py-2"><button type="button" @click="toggleSort('conversion')">Conversion</button></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in filteredApps" :key="row.slug" class="border-t border-line">
                  <td class="px-3 py-2 font-semibold">
                    <RouterLink :to="`/apps/${row.slug}`" class="hover:text-accent-ink">{{ row.name }}</RouterLink>
                  </td>
                  <td class="px-3 py-2 tabular-nums">{{ row.views }}</td>
                  <td class="px-3 py-2 tabular-nums">{{ row.unique_viewers }}</td>
                  <td class="px-3 py-2 tabular-nums">{{ row.opens }}</td>
                  <td class="px-3 py-2 tabular-nums">{{ row.link_copies }}</td>
                  <td class="px-3 py-2 tabular-nums font-semibold">{{ row.activations }}</td>
                  <td class="px-3 py-2 tabular-nums">
                    {{ row.conversion == null ? '—' : `${Math.round(row.conversion * 10) / 10}%` }}
                  </td>
                </tr>
                <tr v-if="filteredApps.length === 0">
                  <td colspan="7" class="px-3 py-6 text-center text-muted">No apps match this search.</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="space-y-2 md:hidden">
            <article
              v-for="row in filteredApps"
              :key="row.slug"
              class="rounded-xl border border-line bg-surface p-3"
            >
              <RouterLink :to="`/apps/${row.slug}`" class="font-semibold hover:text-accent-ink">{{ row.name }}</RouterLink>
              <dl class="mt-2 grid grid-cols-2 gap-1 text-xs text-muted">
                <div>Views <span class="float-right font-semibold text-ink tabular-nums">{{ row.views }}</span></div>
                <div>Unique <span class="float-right font-semibold text-ink tabular-nums">{{ row.unique_viewers }}</span></div>
                <div>Opens <span class="float-right font-semibold text-ink tabular-nums">{{ row.opens }}</span></div>
                <div>Copies <span class="float-right font-semibold text-ink tabular-nums">{{ row.link_copies }}</span></div>
                <div>Activations <span class="float-right font-semibold text-ink tabular-nums">{{ row.activations }}</span></div>
                <div>Conversion
                  <span class="float-right font-semibold text-ink tabular-nums">
                    {{ row.conversion == null ? '—' : `${Math.round(row.conversion * 10) / 10}%` }}
                  </span>
                </div>
              </dl>
            </article>
          </div>
        </section>

        <section class="rounded-xl border border-line bg-surface-2 p-4 text-xs text-muted space-y-1">
          <p><strong class="text-ink">Unique visitor</strong> is a browser identifier, not a guaranteed person.</p>
          <p><strong class="text-ink">Activation</strong> means opening in Nimiq Pay or successfully copying a link — not in-app engagement.</p>
        </section>
      </template>
    </template>
  </div>
</template>
