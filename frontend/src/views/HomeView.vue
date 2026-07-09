<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { listApps, listCategories, type App, type Category } from '../api'
import AppCard from '../components/AppCard.vue'
import EmptyState from '../components/EmptyState.vue'
import StoreBadges from '../components/StoreBadges.vue'
import { useI18n } from '../composables/useI18n'

const { t } = useI18n()

const featured = ref<App[]>([])
const newest = ref<App[]>([])
const newWeek = ref<App[]>([])
const trending = ref<App[]>([])
const rewardApps = ref<App[]>([])
const games = ref<App[]>([])
const usdtApps = ref<App[]>([])
const categories = ref<Category[]>([])
const searchResults = ref<App[]>([])
const homeQuery = ref('')
const error = ref('')
const searchError = ref('')
const searchLoading = ref(false)
const loading = ref(true)
const isSearching = computed(() => homeQuery.value.trim().length > 0)

let searchTimer: ReturnType<typeof setTimeout>

const homeCategoryThemes: Record<string, { accent: string; soft: string; ink: string }> = {
  games: { accent: '#0582ca', soft: 'rgba(5, 130, 202, 0.1)', ink: '#0582ca' },
  utilities: { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)', ink: '#168f80' },
  finance: { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)', ink: '#168f80' },
  maps: { accent: '#e9b213', soft: 'rgba(233, 178, 19, 0.16)', ink: '#9c7300' },
  social: { accent: '#fa7268', soft: 'rgba(250, 114, 104, 0.13)', ink: '#c44941' },
  experiments: { accent: '#5f4b8b', soft: 'rgba(95, 75, 139, 0.13)', ink: '#5f4b8b' },
}

function categoryStyle(name: string) {
  const theme = homeCategoryThemes[name.toLowerCase()] ?? { accent: '#1f2348', soft: 'rgba(31, 35, 72, 0.06)', ink: '#1f2348' }
  return {
    borderColor: `${theme.accent}66`,
    backgroundColor: theme.soft,
    color: theme.ink,
  }
}

async function searchHomeApps() {
  const query = homeQuery.value.trim()
  searchError.value = ''
  if (!query) {
    searchResults.value = []
    searchLoading.value = false
    return
  }

  searchLoading.value = true
  try {
    searchResults.value = await listApps({ q: homeQuery.value })
  } catch (e) {
    searchError.value = (e as Error).message
  } finally {
    searchLoading.value = false
  }
}

watch(homeQuery, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(searchHomeApps, 220)
})

onMounted(async () => {
  loading.value = true
  try {
    ;[featured.value, newest.value, trending.value, newWeek.value, rewardApps.value, games.value, usdtApps.value, categories.value] = await Promise.all([
      listApps({ featured: 'true' }),
      listApps({ sort: 'newest', limit: '6' }),
      listApps({ collection: 'popular', limit: '4' }),
      listApps({ collection: 'new-week', limit: '4' }),
      listApps({ collection: 'rewards', limit: '4' }),
      listApps({ collection: 'games', limit: '4' }),
      listApps({ collection: 'usdt', limit: '4' }),
      listCategories(),
    ])
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-10">
    <section class="nq-brand-surface nq-card-shadow relative overflow-hidden rounded-[10px] border border-line p-6 md:p-12 dark:border-white/10 dark:text-white dark:shadow-black/20">
      <div class="nq-hero-accent absolute inset-y-0 right-0 hidden w-[36%] md:block" aria-hidden="true"></div>
      <div class="relative max-w-xl">
        <p class="text-sm font-bold uppercase tracking-widest text-accent-ink dark:text-white/80">{{ t('home.eyebrow') }}</p>
        <h1 class="mt-2 max-w-xl text-3xl font-extrabold leading-tight md:text-5xl">
          {{ t('home.title') }}
        </h1>
        <p class="mt-3 max-w-xl text-muted md:text-lg dark:text-white/75">
          {{ t('home.subtitle') }}
        </p>
        <div class="mt-6 max-w-xl rounded-[10px] border border-line bg-white/90 p-2 shadow-sm dark:border-white/15 dark:bg-white/10">
          <label for="home-app-search" class="sr-only">{{ t('home.searchLabel') }}</label>
          <div class="flex flex-col gap-2 sm:flex-row">
            <div class="relative flex-1">
              <svg viewBox="0 0 24 24" class="pointer-events-none absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-muted" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                <circle cx="11" cy="11" r="7" />
                <path d="M20 20l-3.5-3.5" />
              </svg>
              <input
                id="home-app-search"
                v-model="homeQuery"
                type="search"
                :placeholder="t('home.searchPlaceholder')"
                class="h-12 w-full rounded-[10px] border border-line bg-surface pl-10 pr-4 font-semibold outline-none transition-colors duration-200 placeholder:text-muted focus:border-accent"
              />
            </div>
            <RouterLink
              to="/apps"
              class="grid h-12 cursor-pointer place-items-center rounded-[500px] border border-line bg-surface px-5 text-sm font-bold text-ink transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
            >
              {{ t('home.allApps') }}
            </RouterLink>
          </div>
          <div v-if="categories.length" class="mt-2 border-t border-line pt-2">
            <p class="mb-2 text-xs font-bold uppercase tracking-wide text-muted dark:text-white/60">{{ t('home.browseCategories') }}</p>
            <div class="flex flex-wrap gap-2">
              <RouterLink
                v-for="category in categories"
                :key="category.name"
                :to="`/apps?category=${encodeURIComponent(category.name)}`"
                class="rounded-full border px-3 py-1.5 text-xs font-extrabold transition duration-200 hover:-translate-y-0.5"
                :style="categoryStyle(category.name)"
              >
                {{ category.name }} <span class="opacity-70">{{ category.count }}</span>
              </RouterLink>
            </div>
          </div>
        </div>
        <div class="mt-6 flex flex-wrap gap-2.5">
          <RouterLink to="/apps"
            class="nq-primary cursor-pointer rounded-[500px] px-6 py-3 font-bold text-white transition duration-200">
            {{ t('home.browseAll') }}
          </RouterLink>
          <RouterLink to="/submit"
            class="cursor-pointer rounded-[500px] border border-line bg-white/90 px-6 py-3 font-bold text-ink transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink dark:border-white/15 dark:bg-white/10 dark:text-white dark:hover:text-white">
            {{ t('home.submitApp') }}
          </RouterLink>
          <RouterLink to="/build"
            class="cursor-pointer rounded-[500px] border border-line bg-white/90 px-6 py-3 font-bold text-ink transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink dark:border-white/15 dark:bg-white/10 dark:text-white dark:hover:text-white">
            {{ t('home.buildApp') }}
          </RouterLink>
        </div>
        <div class="mt-8 border-t border-line pt-5 dark:border-white/15">
          <p class="mb-3 text-sm font-semibold text-muted dark:text-white/70">{{ t('home.walletPrompt') }}</p>
          <StoreBadges />
        </div>
      </div>
    </section>

    <EmptyState
      v-if="error"
      :title="t('home.errorTitle')"
      :description="t('home.errorBody')"
      variant="error"
    />

    <section v-if="isSearching">
      <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold">
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-accent-ink" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3.5-3.5" />
        </svg>
        {{ t('home.searchResults') }}
      </h2>
      <EmptyState
        v-if="searchError"
        :title="t('apps.errorTitle')"
        :description="t('apps.errorBody')"
        variant="error"
      />
      <div v-else-if="searchLoading" class="grid gap-4 sm:grid-cols-2" aria-hidden="true">
        <div v-for="i in 2" :key="i" class="h-44 animate-pulse rounded-2xl border border-line bg-surface"></div>
      </div>
      <EmptyState
        v-else-if="!searchResults.length"
        :title="t('home.emptySearchTitle')"
        :description="t('home.emptySearchBody', { query: homeQuery.trim() })"
      >
        <template #actions>
          <RouterLink
            to="/apps"
            class="nq-primary cursor-pointer rounded-[500px] px-5 py-2.5 text-sm font-bold text-white transition duration-200"
          >
            {{ t('common.browseAll') }}
          </RouterLink>
        </template>
      </EmptyState>
      <div v-else class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in searchResults" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && loading">
      <h2 class="mb-4 text-xl font-extrabold">{{ t('home.featured') }}</h2>
      <div class="grid gap-4 sm:grid-cols-2" aria-hidden="true">
        <div v-for="i in 2" :key="i" class="h-44 animate-pulse rounded-2xl border border-line bg-surface"></div>
      </div>
    </section>

    <section v-else-if="!isSearching && featured.length">
      <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold">
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-accent-ink" aria-hidden="true">
          <path d="M12 2l2.9 6.26L21 9.27l-4.5 4.38L17.8 21 12 17.77 6.2 21l1.3-7.35L3 9.27l6.1-1.01z" />
        </svg>
        {{ t('home.featured') }}
      </h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in featured" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && loading">
      <h2 class="mb-4 text-xl font-extrabold">{{ t('home.newest') }}</h2>
      <div class="grid gap-4 sm:grid-cols-2" aria-hidden="true">
        <div v-for="i in 4" :key="i" class="h-44 animate-pulse rounded-2xl border border-line bg-surface"></div>
      </div>
    </section>

    <section v-else-if="!isSearching && newest.length">
      <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold">
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-accent-ink" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M5.6 18.4l2.1-2.1M16.3 7.7l2.1-2.1" />
        </svg>
        {{ t('home.newest') }}
      </h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in newest" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && trending.length" class="space-y-4">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-extrabold">{{ t('collections.popular') }}</h2>
        <RouterLink to="/apps?collection=popular" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in trending" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && newWeek.length" class="space-y-4">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-extrabold">{{ t('collections.newWeek') }}</h2>
        <RouterLink to="/apps?collection=new-week" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in newWeek" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && rewardApps.length" class="space-y-4">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-extrabold">{{ t('collections.rewards') }}</h2>
        <RouterLink to="/apps?collection=rewards" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in rewardApps" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && games.length" class="space-y-4">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-extrabold">{{ t('collections.games') }}</h2>
        <RouterLink to="/apps?collection=games" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in games" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && usdtApps.length" class="space-y-4">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-extrabold">{{ t('collections.usdt') }}</h2>
        <RouterLink to="/apps?collection=usdt" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in usdtApps" :key="app.id" :app="app" />
      </div>
    </section>
  </div>
</template>
