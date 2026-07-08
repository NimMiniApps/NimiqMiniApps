<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { listApps, listCategories, type App, type Category } from '../api'
import AppCard from '../components/AppCard.vue'
import StoreBadges from '../components/StoreBadges.vue'

const featured = ref<App[]>([])
const newest = ref<App[]>([])
const categories = ref<Category[]>([])
const searchResults = ref<App[]>([])
const homeQuery = ref('')
const error = ref('')
const searchError = ref('')
const searchLoading = ref(false)
const isSearching = computed(() => homeQuery.value.trim().length > 0)

let searchTimer: ReturnType<typeof setTimeout>

const homeCategoryThemes: Record<string, { accent: string; soft: string; ink: string }> = {
  games: { accent: '#1f74ff', soft: 'rgba(31, 116, 255, 0.13)', ink: '#7fd8ff' },
  utilities: { accent: '#14b8a6', soft: 'rgba(20, 184, 166, 0.14)', ink: '#5eead4' },
  finance: { accent: '#22c55e', soft: 'rgba(34, 197, 94, 0.14)', ink: '#86efac' },
  maps: { accent: '#f59e0b', soft: 'rgba(245, 158, 11, 0.16)', ink: '#fbbf24' },
  social: { accent: '#f43f5e', soft: 'rgba(244, 63, 94, 0.14)', ink: '#fb7185' },
  experiments: { accent: '#a855f7', soft: 'rgba(168, 85, 247, 0.15)', ink: '#c084fc' },
}

function categoryStyle(name: string) {
  const theme = homeCategoryThemes[name.toLowerCase()] ?? { accent: '#64748b', soft: 'rgba(100, 116, 139, 0.14)', ink: '#cbd5e1' }
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
  try {
    ;[featured.value, newest.value, categories.value] = await Promise.all([
      listApps({ featured: 'true' }),
      listApps({ sort: 'newest' }),
      listCategories(),
    ])
  } catch (e) {
    error.value = (e as Error).message
  }
})
</script>

<template>
  <div class="space-y-10">
    <section class="relative overflow-hidden rounded-3xl border border-line bg-surface p-6 shadow-xl shadow-blue-950/5 md:p-12 dark:shadow-black/20">
      <div class="absolute inset-y-0 right-0 hidden w-[32%] bg-nq-blue md:block" aria-hidden="true"></div>
      <div class="absolute bottom-0 right-0 hidden h-28 w-[32%] bg-accent-2/35 md:block" aria-hidden="true"></div>
      <div class="relative max-w-xl">
        <p class="text-sm font-bold uppercase tracking-widest text-accent-ink">Community curated</p>
        <h1 class="mt-2 max-w-xl text-3xl font-extrabold leading-tight md:text-5xl">
          Mini apps that live inside your <span class="text-accent-ink">Nimiq Pay</span> wallet
        </h1>
        <p class="mt-3 max-w-xl text-muted md:text-lg">
          Games, maps, tools and experiments — hand-picked by the community and open with one tap, no installs, no accounts.
        </p>
        <div class="mt-6 max-w-xl rounded-2xl border border-line bg-page/80 p-2 shadow-sm shadow-slate-950/5 dark:bg-surface-2/60">
          <label for="home-app-search" class="sr-only">Search mini apps</label>
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
                placeholder="Search mini apps"
                class="h-12 w-full rounded-xl border border-line bg-surface pl-10 pr-4 font-semibold outline-none transition-colors duration-200 placeholder:text-muted focus:border-accent"
              />
            </div>
            <RouterLink
              to="/apps"
              class="grid h-12 cursor-pointer place-items-center rounded-xl border border-line bg-surface px-5 text-sm font-bold text-ink transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
            >
              All apps
            </RouterLink>
          </div>
          <div v-if="categories.length" class="mt-2 border-t border-line pt-2">
            <p class="mb-2 text-xs font-bold uppercase tracking-wide text-muted">Browse by category</p>
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
            class="cursor-pointer rounded-xl bg-nq-blue px-6 py-3 font-bold text-white shadow-sm shadow-blue-700/25 transition duration-200 hover:bg-nq-blue-dark">
            Browse all apps
          </RouterLink>
          <RouterLink to="/submit"
            class="cursor-pointer rounded-xl border border-line bg-surface px-6 py-3 font-bold text-ink transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
            Submit your app
          </RouterLink>
        </div>
        <div class="mt-8 border-t border-line pt-5">
          <p class="mb-3 text-sm font-semibold text-muted">New here? Grab the free Nimiq Pay wallet:</p>
          <StoreBadges />
        </div>
      </div>
    </section>

    <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>

    <section v-if="isSearching">
      <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold">
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-accent-ink" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-3.5-3.5" />
        </svg>
        Search results
      </h2>
      <p v-if="searchError" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ searchError }}</p>
      <div v-else-if="searchLoading" class="grid gap-4 sm:grid-cols-2" aria-hidden="true">
        <div v-for="i in 2" :key="i" class="h-44 animate-pulse rounded-2xl border border-line bg-surface"></div>
      </div>
      <p v-else-if="!searchResults.length" class="rounded-2xl border border-line bg-surface p-8 text-center text-muted">
        No mini apps found for "{{ homeQuery }}".
      </p>
      <div v-else class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in searchResults" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && featured.length">
      <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold">
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-accent-ink" aria-hidden="true">
          <path d="M12 2l2.9 6.26L21 9.27l-4.5 4.38L17.8 21 12 17.77 6.2 21l1.3-7.35L3 9.27l6.1-1.01z" />
        </svg>
        Featured
      </h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in featured" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="!isSearching && newest.length">
      <h2 class="mb-4 flex items-center gap-2 text-xl font-extrabold">
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-accent-ink" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M12 3v3M12 18v3M3 12h3M18 12h3M5.6 5.6l2.1 2.1M16.3 16.3l2.1 2.1M5.6 18.4l2.1-2.1M16.3 7.7l2.1-2.1" />
        </svg>
        Newest
      </h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in newest.slice(0, 6)" :key="app.id" :app="app" />
      </div>
    </section>
  </div>
</template>
