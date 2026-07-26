<script setup lang="ts">
/*
 * THESIS: Nimiq Mini Apps reads as a live departures board, not a bolted-on
 *   app store — refuses the neon-glass crypto-dashboard default.
 * OWN-WORLD: dark board casing (#14162b) and flap rows (#1f2348) in both
 *   themes, white flap type, Nimiq blue as the "on-time" lamp, amber/red
 *   lamps for pending/broken, rectangular board plates (not pills).
 * STORY: a wallet user reads the board in one glance — what's live, new,
 *   pending — then "boards" an app by tapping its row.
 * FIRST VIEWPORT: a literal board header (terminal search, track chips),
 *   collections rendered as rows of board plates below.
 * FORM: fused challenger "signals-instruments-split-flap-concourse" (won
 *   both audience-identification and product-clarity axes against the
 *   assigned grounded direction, seed key fb1484f2, mode operate, index 5).
 */
import { computed, ref, watch, onMounted } from 'vue'
import { listApps, listCategories, type App, type Category } from '../api'
import AppCard from '../components/AppCard.vue'
import EmptyState from '../components/EmptyState.vue'
import StoreBadges from '../components/StoreBadges.vue'
import FlapText from '../components/FlapText.vue'
import { useI18n } from '../composables/useI18n'
import { categoryTheme } from '../utils/appIcon'

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

function categoryStyle(name: string) {
  const theme = categoryTheme(name)
  return { backgroundColor: theme.accent }
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
    <section class="board relative p-6 md:p-12">
      <div class="relative max-w-xl">
        <p class="font-mono text-xs font-bold uppercase tracking-[0.2em] text-accent-ink">
          <span class="board-lamp board-lamp-live mr-1.5 align-middle" aria-hidden="true"></span>{{ t('home.eyebrow') }}
        </p>
        <h1 class="mt-3 max-w-xl text-3xl font-extrabold uppercase leading-tight text-board-flap-ink md:text-5xl">
          <FlapText :text="t('home.title')" />
        </h1>
        <p class="mt-3 max-w-xl text-board-flap-muted md:text-lg">
          {{ t('home.subtitle') }}
        </p>
        <div class="mt-6 max-w-xl rounded-[4px] border border-board-hairline bg-board-frame p-2">
          <label for="home-app-search" class="sr-only">{{ t('home.searchLabel') }}</label>
          <div class="flex flex-col gap-2 sm:flex-row">
            <div class="relative flex-1">
              <svg viewBox="0 0 24 24" class="pointer-events-none absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-board-flap-muted" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                <circle cx="11" cy="11" r="7" />
                <path d="M20 20l-3.5-3.5" />
              </svg>
              <input
                id="home-app-search"
                v-model="homeQuery"
                type="search"
                :placeholder="t('home.searchPlaceholder')"
                class="h-12 w-full rounded-[4px] border border-board-hairline bg-board-flap pl-10 pr-4 font-mono font-semibold text-board-flap-ink outline-none transition-colors duration-200 placeholder:text-board-flap-muted focus:border-lamp-live focus-visible:ring-2 focus-visible:ring-lamp-live/40"
              />
            </div>
            <RouterLink
              to="/apps"
              class="board-plate board-plate-ghost grid h-12 cursor-pointer place-items-center px-5 text-xs"
            >
              {{ t('home.allApps') }}
            </RouterLink>
          </div>
          <div v-if="categories.length" class="mt-2 border-t border-board-hairline pt-2">
            <p class="mb-2 font-mono text-[11px] font-bold uppercase tracking-wide text-board-flap-muted">{{ t('home.browseCategories') }}</p>
            <div class="flex flex-wrap gap-1.5">
              <RouterLink
                v-for="category in categories"
                :key="category.name"
                :to="`/apps?category=${encodeURIComponent(category.name)}`"
                class="rounded-[3px] px-2.5 py-1 font-mono text-[11px] font-bold uppercase text-white transition-opacity duration-200 hover:opacity-85"
                :style="categoryStyle(category.name)"
              >
                {{ category.name }} <span class="opacity-70">{{ category.count }}</span>
              </RouterLink>
            </div>
          </div>
        </div>
        <div class="mt-6 flex flex-wrap gap-2.5">
          <RouterLink to="/apps"
            class="board-plate board-plate-primary cursor-pointer px-6 py-3 text-sm">
            {{ t('home.browseAll') }}
          </RouterLink>
          <RouterLink to="/submit"
            class="board-plate board-plate-ghost cursor-pointer px-6 py-3 text-sm">
            {{ t('home.submitApp') }}
          </RouterLink>
          <RouterLink to="/build"
            class="board-plate board-plate-ghost cursor-pointer px-6 py-3 text-sm">
            {{ t('home.buildApp') }}
          </RouterLink>
        </div>
        <div class="mt-8 border-t border-board-hairline pt-5">
          <p class="mb-3 text-sm font-semibold text-board-flap-muted">{{ t('home.walletPrompt') }}</p>
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
      <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold uppercase tracking-wide text-ink">
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
        <div v-for="i in 2" :key="i" class="h-44 animate-pulse rounded-[4px] border border-board-hairline bg-board-flap"></div>
      </div>
      <EmptyState
        v-else-if="!searchResults.length"
        :title="t('home.emptySearchTitle')"
        :description="t('home.emptySearchBody', { query: homeQuery.trim() })"
      >
        <template #actions>
          <RouterLink
            to="/apps"
            class="board-plate board-plate-primary cursor-pointer px-5 py-2.5 text-sm"
          >
            {{ t('common.browseAll') }}
          </RouterLink>
        </template>
      </EmptyState>
      <div v-else class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in searchResults" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && loading" class="space-y-10" aria-hidden="true">
      <div v-for="group in 3" :key="group" class="space-y-4">
        <div class="h-6 w-32 animate-pulse rounded-[3px] bg-surface-2"></div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div v-for="card in 2" :key="card" class="h-44 animate-pulse rounded-[4px] border border-board-hairline bg-board-flap"></div>
        </div>
      </div>
    </section>

    <template v-else-if="!isSearching">
      <section v-if="featured.length">
        <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold uppercase tracking-wide text-ink">
          <svg viewBox="0 0 24 24" class="h-5 w-5 fill-accent-ink" aria-hidden="true">
            <path d="M12 2l2.9 6.26L21 9.27l-4.5 4.38L17.8 21 12 17.77 6.2 21l1.3-7.35L3 9.27l6.1-1.01z" />
          </svg>
          {{ t('home.featured') }}
        </h2>
        <div class="grid gap-4 sm:grid-cols-2">
          <AppCard v-for="app in featured" :key="app.id" :app="app" />
        </div>
      </section>

      <section v-if="newest.length">
        <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold uppercase tracking-wide text-ink">
          <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-accent-ink" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M5.6 18.4l2.1-2.1M16.3 7.7l2.1-2.1" />
          </svg>
          {{ t('home.newest') }}
        </h2>
        <div class="grid gap-4 sm:grid-cols-2">
          <AppCard v-for="app in newest" :key="app.id" :app="app" />
        </div>
      </section>

      <section v-if="trending.length" class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-xl font-extrabold uppercase tracking-wide text-ink">{{ t('collections.popular') }}</h2>
          <RouterLink to="/apps?collection=popular" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <AppCard v-for="app in trending" :key="app.id" :app="app" />
        </div>
      </section>

      <section v-if="newWeek.length" class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-xl font-extrabold uppercase tracking-wide text-ink">{{ t('collections.newWeek') }}</h2>
          <RouterLink to="/apps?collection=new-week" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <AppCard v-for="app in newWeek" :key="app.id" :app="app" />
        </div>
      </section>

      <section v-if="rewardApps.length" class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-xl font-extrabold uppercase tracking-wide text-ink">{{ t('collections.rewards') }}</h2>
          <RouterLink to="/apps?collection=rewards" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <AppCard v-for="app in rewardApps" :key="app.id" :app="app" />
        </div>
      </section>

      <section v-if="games.length" class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-xl font-extrabold uppercase tracking-wide text-ink">{{ t('collections.games') }}</h2>
          <RouterLink to="/apps?collection=games" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <AppCard v-for="app in games" :key="app.id" :app="app" />
        </div>
      </section>

      <section v-if="usdtApps.length" class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-xl font-extrabold uppercase tracking-wide text-ink">{{ t('collections.usdt') }}</h2>
          <RouterLink to="/apps?collection=usdt" class="text-sm font-semibold text-accent-ink hover:underline">{{ t('common.viewAll') }}</RouterLink>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <AppCard v-for="app in usdtApps" :key="app.id" :app="app" />
        </div>
      </section>
    </template>
  </div>
</template>
