<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listAppsPaginated, listCategories, listDevelopers, type App, type Category, type DeveloperSummary } from '../api'
import AppCard from '../components/AppCard.vue'
import EmptyState from '../components/EmptyState.vue'
import { useI18n } from '../composables/useI18n'
import { useWalletAuth } from '../composables/useWalletAuth'
import { walletOwnsApp } from '../utils/wallet'
import { setPageMeta, resetPageMeta } from '../utils/meta'
import { nimConnectPublicUrl, resolveNimConnectHandle } from '../utils/nimconnect'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { walletAddress } = useWalletAuth()
const apps = ref<App[]>([])
const categories = ref<Category[]>([])
const developers = ref<DeveloperSummary[]>([])
const q = ref((route.query.q as string) || '')
const category = ref((route.query.category as string) || '')
const developer = ref((route.query.developer as string) || '')
const collection = ref((route.query.collection as string) || '')
const tag = ref((route.query.tag as string) || '')
const asset = ref((route.query.asset as string) || '')
const sort = ref((route.query.sort as string) || 'featured')
const error = ref('')
const loading = ref(true)
const loadingMore = ref(false)
const total = ref(0)
const nimConnectHandle = ref<string | null>(null)
const PAGE_SIZE = 20

const nimConnectUrl = computed(() =>
  nimConnectHandle.value ? nimConnectPublicUrl(nimConnectHandle.value) : null,
)

const hasMore = computed(() => apps.value.length < total.value)

const collectionLabels = computed<Record<string, string>>(() => ({
  'new-week': t('collections.newWeek'),
  popular: t('collections.popular'),
  rewards: t('collections.rewards'),
  games: t('collections.games'),
  usdt: t('collections.usdt'),
}))

const activeFilter = computed(() => {
  if (developer.value) {
    const name = developerLabel.value || developer.value
    return { type: t('apps.filterDeveloper'), value: name }
  }
  if (collection.value) return { type: t('apps.filterCollection'), value: collectionLabels.value[collection.value] || collection.value }
  if (tag.value) return { type: t('apps.filterTag'), value: tag.value }
  if (asset.value) return { type: t('apps.filterAsset'), value: asset.value }
  return null
})

const developerLabel = computed(() =>
  developers.value.find((d) => d.slug === developer.value)?.name
    ?? apps.value[0]?.developer_name
    ?? '',
)

const pageTitle = computed(() =>
  developer.value && developerLabel.value
    ? t('apps.developerTitle', { name: developerLabel.value })
    : t('apps.title'),
)

const hasFilters = computed(() =>
  !!(q.value.trim() || category.value || developer.value || collection.value || tag.value || asset.value),
)

const emptyDescription = computed(() => {
  if (q.value.trim()) return t('apps.emptySearchBody', { query: q.value.trim() })
  if (hasFilters.value) return t('apps.emptyFilteredBody')
  return t('apps.emptyBody')
})

function listParams(offset: number) {
  return {
    q: q.value,
    category: category.value,
    developer: developer.value,
    collection: collection.value,
    tag: tag.value,
    asset: asset.value,
    sort: sort.value,
    limit: String(PAGE_SIZE),
    offset: String(offset),
  }
}

async function load(reset = true) {
  if (reset) {
    loading.value = true
    apps.value = []
  } else {
    loadingMore.value = true
  }
  error.value = ''
  try {
    const offset = reset ? 0 : apps.value.length
    const result = await listAppsPaginated(listParams(offset))
    total.value = result.total
    apps.value = reset ? result.items : [...apps.value, ...result.items]
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function setCategory(name: string) {
  category.value = name
}

function toggleRewards() {
  collection.value = collection.value === 'rewards' ? '' : 'rewards'
}

function clearFilter() {
  collection.value = ''
  tag.value = ''
  asset.value = ''
  developer.value = ''
}

function clearAll() {
  q.value = ''
  category.value = ''
  clearFilter()
  sort.value = 'featured'
}

function syncQuery() {
  router.replace({
    query: {
      ...(q.value ? { q: q.value } : {}),
      ...(category.value ? { category: category.value } : {}),
      ...(developer.value ? { developer: developer.value } : {}),
      ...(collection.value ? { collection: collection.value } : {}),
      ...(tag.value ? { tag: tag.value } : {}),
      ...(asset.value ? { asset: asset.value } : {}),
      ...(sort.value !== 'featured' ? { sort: sort.value } : {}),
    },
  })
}

let timer: ReturnType<typeof setTimeout>
watch(q, () => { clearTimeout(timer); timer = setTimeout(() => { syncQuery(); load(true) }, 250) })
watch([category, sort, tag, asset, collection, developer], () => { syncQuery(); load(true) })

watch(() => route.query, (query) => {
  const nextQ = (query.q as string) || ''
  const nextCategory = (query.category as string) || ''
  const nextDeveloper = (query.developer as string) || ''
  const nextCollection = (query.collection as string) || ''
  const nextTag = (query.tag as string) || ''
  const nextAsset = (query.asset as string) || ''
  const nextSort = (query.sort as string) || 'featured'
  const changed =
    q.value !== nextQ ||
    category.value !== nextCategory ||
    developer.value !== nextDeveloper ||
    collection.value !== nextCollection ||
    tag.value !== nextTag ||
    asset.value !== nextAsset ||
    sort.value !== nextSort
  q.value = nextQ
  category.value = nextCategory
  developer.value = nextDeveloper
  collection.value = nextCollection
  tag.value = nextTag
  asset.value = nextAsset
  sort.value = nextSort
  if (changed) load(true)
})

watch([developer, developerLabel, total], () => {
  if (developer.value && developerLabel.value) {
    setPageMeta({
      title: t('apps.developerTitle', { name: developerLabel.value }),
      description: t('apps.developerMeta', { count: total.value, name: developerLabel.value }),
      url: window.location.href,
    })
  } else if (!developer.value) {
    resetPageMeta()
  }
})

let nimConnectLookupSeq = 0
watch(
  () => (developer.value ? apps.value[0]?.owner_wallet_addresses?.[0] : undefined),
  async (address) => {
    const seq = ++nimConnectLookupSeq
    nimConnectHandle.value = null
    if (!address) return
    const handle = await resolveNimConnectHandle(address)
    if (seq === nimConnectLookupSeq) nimConnectHandle.value = handle
  },
)

onMounted(async () => {
  load(true)
  try {
    ;[categories.value, developers.value] = await Promise.all([
      listCategories(),
      listDevelopers(),
    ])
  } catch { /* filters stay empty */ }
})
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-2xl font-extrabold">{{ pageTitle }}</h1>
    <div v-if="developer && developerLabel" class="space-y-1">
      <p class="text-sm text-muted">
        {{ t('apps.developerMeta', { count: total, name: developerLabel }) }}
      </p>
      <a
        v-if="nimConnectUrl && nimConnectHandle"
        :href="nimConnectUrl"
        target="_blank"
        rel="noopener"
        :aria-label="t('common.nimconnectProfile')"
        :title="t('common.nimconnectProfile')"
        class="inline-block text-sm text-accent-ink-dark hover:underline dark:text-accent-ink"
      >@{{ nimConnectHandle }}</a>
    </div>

    <div v-if="activeFilter" class="flex items-center gap-2 text-sm">
      <span class="text-muted">{{ t('apps.filteredBy', { type: activeFilter.type }) }}</span>
      <span class="rounded-full bg-accent/10 px-2.5 py-1 font-semibold text-accent-ink">{{ activeFilter.value }}</span>
      <button type="button" @click="clearFilter"
        class="cursor-pointer rounded-lg border border-line px-2 py-1 text-xs font-semibold text-muted hover:text-ink">
        {{ t('common.clear') }}
      </button>
    </div>

    <div class="flex flex-col gap-2 sm:flex-row">
      <input v-model="q" type="search" :placeholder="t('apps.searchPlaceholder')"
        class="flex-1 rounded-xl border border-line bg-surface px-4 py-2.5 outline-none transition-colors duration-200 placeholder:text-muted focus:border-accent" />
      <div class="flex flex-wrap gap-2">
        <select v-model="category" class="flex-1 cursor-pointer rounded-xl border border-line bg-surface px-3 py-2.5">
          <option value="">{{ t('apps.allCategories') }}</option>
          <option v-for="c in categories" :key="c.name" :value="c.name">{{ c.name }} ({{ c.count }})</option>
        </select>
        <select v-model="developer" class="flex-1 cursor-pointer rounded-xl border border-line bg-surface px-3 py-2.5">
          <option value="">{{ t('apps.allDevelopers') }}</option>
          <option v-for="d in developers" :key="d.slug" :value="d.slug">{{ d.name }} ({{ d.app_count }})</option>
        </select>
        <select v-model="sort" class="cursor-pointer rounded-xl border border-line bg-surface px-3 py-2.5">
          <option value="featured">{{ t('apps.sortFeatured') }}</option>
          <option value="trending">{{ t('apps.sortTrending') }}</option>
          <option value="newest">{{ t('apps.sortNewest') }}</option>
          <option value="name">{{ t('apps.sortName') }}</option>
        </select>
      </div>
    </div>

    <div v-if="categories.length" class="flex flex-wrap gap-2">
      <button type="button" @click="toggleRewards"
        class="cursor-pointer rounded-full border px-3 py-1.5 text-xs font-bold transition duration-200"
        :class="collection === 'rewards' ? 'border-emerald-500 bg-emerald-500/15 text-emerald-800 dark:text-emerald-200' : 'border-line bg-surface-2 text-muted hover:border-emerald-500/50 hover:text-emerald-700 dark:hover:text-emerald-200'">
        {{ t('collections.rewards') }}
      </button>
      <button type="button" @click="setCategory('')"
        class="cursor-pointer rounded-full border px-3 py-1.5 text-xs font-bold transition duration-200"
        :class="category === '' ? 'border-accent bg-accent/10 text-accent-ink' : 'border-line bg-surface-2 text-muted hover:border-accent/50'">
        {{ t('common.all') }}
      </button>
      <button v-for="c in categories" :key="c.name" type="button" @click="setCategory(c.name)"
        class="cursor-pointer rounded-full border px-3 py-1.5 text-xs font-bold transition duration-200"
        :class="category === c.name ? 'border-accent bg-accent/10 text-accent-ink' : 'border-line bg-surface-2 text-muted hover:border-accent/50'">
        {{ c.name }} <span class="opacity-70">{{ c.count }}</span>
      </button>
    </div>

    <EmptyState
      v-if="error"
      :title="t('apps.errorTitle')"
      :description="t('apps.errorBody')"
      variant="error"
    >
      <template #actions>
        <button
          type="button"
          class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
          @click="load(true)"
        >
          {{ t('common.retry') }}
        </button>
      </template>
    </EmptyState>

    <div v-else-if="loading" class="grid gap-4 sm:grid-cols-2" aria-hidden="true">
      <div v-for="i in 4" :key="i" class="h-40 animate-pulse rounded-2xl border border-line bg-surface"></div>
    </div>

    <EmptyState
      v-else-if="!apps.length"
      :title="t('apps.emptyTitle')"
      :description="emptyDescription"
    >
      <template #actions>
        <button
          v-if="hasFilters"
          type="button"
          class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
          @click="clearAll"
        >
          {{ t('common.clear') }}
        </button>
        <RouterLink
          to="/apps"
          class="cursor-pointer rounded-[500px] nq-primary px-5 py-2.5 text-sm font-bold text-white transition duration-200"
        >
          {{ t('common.browseAll') }}
        </RouterLink>
        <RouterLink
          to="/submit"
          class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
        >
          {{ t('nav.submit') }}
        </RouterLink>
      </template>
    </EmptyState>

    <div v-else class="space-y-4">
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard
          v-for="app in apps"
          :key="app.id"
          :app="app"
          :owned="walletOwnsApp(walletAddress, app.owner_wallet_addresses)"
        />
      </div>
      <div class="flex flex-col items-center gap-2">
        <p class="text-sm text-muted">{{ t('apps.showingCount', { shown: apps.length, total }) }}</p>
        <button v-if="hasMore" type="button" @click="load(false)" :disabled="loadingMore"
          class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-bold hover:border-accent disabled:opacity-50">
          {{ loadingMore ? t('common.loading') : t('apps.loadMore') }}
        </button>
      </div>
    </div>
  </div>
</template>
