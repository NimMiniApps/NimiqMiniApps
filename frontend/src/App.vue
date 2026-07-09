<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import StoreBadges from './components/StoreBadges.vue'
import { useAdminAuth } from './composables/useAdminAuth'
import { useI18n } from './composables/useI18n'
import { CATALOG_ISSUES_URL } from './utils/catalogLinks'

const route = useRoute()
const { isAdmin, pendingCount } = useAdminAuth()
const { t } = useI18n()

const navItems = [
  { to: '/', key: 'nav.home' as const, icon: 'M3 12l9-9 9 9M5 10v10h5v-6h4v6h5V10' },
  { to: '/apps', key: 'nav.apps' as const, icon: 'M4 4h7v7H4zM13 4h7v7h-7zM4 13h7v7H4zM13 13h7v7h-7z' },
  { to: '/build', key: 'nav.build' as const, icon: 'M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z' },
  { to: '/submit', key: 'nav.submit' as const, icon: 'M12 5v14M5 12h14' },
]
const desktopNavItems = [
  ...navItems,
]
const adminNavItem = { to: '/admin', key: 'nav.admin' as const }
const isActive = (to: string) => {
  if (to === '/') return route.path === '/'
  if (to === '/apps' && route.path === '/apps') return true
  return route.path.startsWith(to) && to !== '/apps'
}

const isDark = ref(document.documentElement.classList.contains('dark'))
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.theme = isDark.value ? 'dark' : 'light'
}

const moreOpen = ref(false)
const moreMenuRef = ref<HTMLElement | null>(null)

function toggleMore() {
  moreOpen.value = !moreOpen.value
}

function closeMore() {
  moreOpen.value = false
}

function onDocumentClick(e: MouseEvent) {
  if (!moreOpen.value || !moreMenuRef.value) return
  if (!moreMenuRef.value.contains(e.target as Node)) closeMore()
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onUnmounted(() => document.removeEventListener('click', onDocumentClick))
</script>

<template>
  <div class="flex min-h-screen flex-col pb-16 md:pb-0">
    <!-- top bar -->
    <header class="sticky top-0 z-20 border-b border-line bg-surface/90 shadow-sm shadow-slate-950/5 backdrop-blur dark:shadow-black/20">
      <div class="mx-auto flex max-w-5xl items-center gap-3 px-4 py-3">
        <RouterLink to="/" class="flex items-center gap-2 text-lg font-extrabold">
          <img
            src="/brand/output-smallpngtools.png"
            alt=""
            class="h-8 w-8 rounded-lg object-cover shadow-sm shadow-blue-700/25"
          />
          <span>Nimiq <span class="text-accent-ink">Mini Apps</span></span>
        </RouterLink>
        <nav class="ml-auto hidden gap-1 md:flex">
          <RouterLink
            v-for="item in desktopNavItems" :key="item.to" :to="item.to"
            class="rounded-lg px-3 py-1.5 text-sm font-semibold transition-colors duration-200 hover:bg-surface-2"
            :class="isActive(item.to) ? 'bg-surface-2 text-accent-ink' : 'text-muted'"
          >{{ t(item.key) }}</RouterLink>
          <RouterLink
            v-if="isAdmin"
            :to="adminNavItem.to"
            class="relative rounded-lg px-3 py-1.5 text-sm font-semibold transition-colors duration-200 hover:bg-surface-2"
            :class="isActive(adminNavItem.to) ? 'bg-surface-2 text-accent-ink' : 'text-muted'"
          >
            {{ t(adminNavItem.key) }}
            <span v-if="pendingCount > 0"
              class="ml-1.5 inline-flex min-w-[1.25rem] items-center justify-center rounded-full bg-amber-500/20 px-1.5 py-0.5 text-[10px] font-bold text-amber-800 dark:text-amber-200">
              {{ pendingCount }}
            </span>
          </RouterLink>
        </nav>
        <button @click="toggleTheme" :aria-label="isDark ? t('theme.light') : t('theme.dark')"
          class="ml-auto grid h-9 w-9 cursor-pointer place-items-center rounded-lg border border-line bg-surface-2 text-muted transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink md:ml-0">
          <svg v-if="isDark" viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
            <circle cx="12" cy="12" r="4" />
            <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
          </svg>
          <svg v-else viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
          </svg>
        </button>
      </div>
    </header>

    <main class="mx-auto w-full max-w-5xl flex-1 px-4 py-6">
      <RouterView />
    </main>

    <!-- Nimiq Pay install banner -->
    <footer class="mt-10">
      <div class="mx-auto max-w-5xl px-4 pb-8">
        <div class="relative overflow-hidden rounded-3xl border border-line bg-nq-blue p-6 text-white shadow-lg shadow-blue-900/10 md:p-10">
          <div class="absolute inset-y-0 right-0 hidden w-1/3 bg-accent-2/20 md:block" aria-hidden="true"></div>
          <div class="relative flex flex-col gap-5 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 class="text-xl font-extrabold md:text-2xl">{{ t('footer.title') }}</h2>
              <p class="mt-1 max-w-md text-white/75">
                {{ t('footer.body') }}
              </p>
            </div>
            <StoreBadges />
          </div>
        </div>
        <p class="mt-4 text-center text-xs text-muted">
          {{ t('footer.curated') }}
          <a href="https://www.nimiq.com/nimiq-pay/" target="_blank" rel="noopener" class="text-accent-ink hover:underline">Nimiq Pay</a>
          mini apps ·
          <RouterLink to="/apps" class="text-accent-ink hover:underline">{{ t('footer.developers') }}</RouterLink>
          ·
          <a :href="CATALOG_ISSUES_URL" target="_blank" rel="noopener" class="text-accent-ink hover:underline">{{ t('footer.githubIssues') }}</a>
        </p>
      </div>
    </footer>

    <!-- bottom nav (mobile) -->
    <nav class="fixed inset-x-0 bottom-0 z-20 border-t border-line bg-surface/95 backdrop-blur md:hidden">
      <div class="grid grid-cols-5">
        <RouterLink
          v-for="item in navItems" :key="item.to" :to="item.to"
          class="flex flex-col items-center gap-1 py-2.5 text-[11px] font-semibold transition-colors duration-200"
          :class="isActive(item.to) ? 'bg-surface-2 text-accent-ink' : 'text-muted'"
        >
          <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path :d="item.icon" />
          </svg>
          {{ t(item.key) }}
        </RouterLink>
        <div ref="moreMenuRef" class="relative">
          <button
            type="button"
            class="flex w-full flex-col items-center gap-1 py-2.5 text-[11px] font-semibold transition-colors duration-200"
            :class="moreOpen || isActive('/admin') ? 'bg-surface-2 text-accent-ink' : 'text-muted'"
            @click.stop="toggleMore"
          >
            <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <circle cx="5" cy="12" r="1.5" fill="currentColor" stroke="none" />
              <circle cx="12" cy="12" r="1.5" fill="currentColor" stroke="none" />
              <circle cx="19" cy="12" r="1.5" fill="currentColor" stroke="none" />
            </svg>
            {{ t('nav.more') }}
          </button>
          <div
            v-if="moreOpen"
            class="absolute bottom-full right-0 mb-2 min-w-[10rem] overflow-hidden rounded-xl border border-line bg-surface shadow-lg shadow-slate-950/10 dark:shadow-black/30"
          >
            <RouterLink
              to="/apps"
              class="block px-4 py-3 text-sm font-semibold transition-colors hover:bg-surface-2"
              :class="route.path === '/apps' ? 'text-accent-ink' : 'text-ink'"
              @click="closeMore"
            >
              {{ t('nav.developers') }}
            </RouterLink>
            <RouterLink
              v-if="isAdmin"
              to="/admin"
              class="flex items-center justify-between px-4 py-3 text-sm font-semibold transition-colors hover:bg-surface-2"
              :class="isActive('/admin') ? 'text-accent-ink' : 'text-ink'"
              @click="closeMore"
            >
              {{ t('nav.admin') }}
              <span v-if="pendingCount > 0"
                class="ml-2 inline-flex min-w-[1.25rem] items-center justify-center rounded-full bg-amber-500/20 px-1.5 py-0.5 text-[10px] font-bold text-amber-800 dark:text-amber-200">
                {{ pendingCount }}
              </span>
            </RouterLink>
          </div>
        </div>
      </div>
    </nav>
  </div>
</template>
