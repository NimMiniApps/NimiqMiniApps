<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import StoreBadges from './components/StoreBadges.vue'
import WalletLoginButton from './components/WalletLoginButton.vue'
import { useI18n } from './composables/useI18n'
import { CATALOG_ISSUES_URL } from './utils/catalogLinks'
import { hasInjectedNimiqPayHost, waitForNimiqPayHost } from './utils/nimiqWallet'

const route = useRoute()
const { t } = useI18n()
const insideNimiqPay = ref(hasInjectedNimiqPayHost())
if (!insideNimiqPay.value) {
  waitForNimiqPayHost().then((result) => { insideNimiqPay.value = result })
}

const navItems = [
  { to: '/', key: 'nav.home' as const, icon: 'M3 12l9-9 9 9M5 10v10h5v-6h4v6h5V10' },
  { to: '/apps', key: 'nav.apps' as const, icon: 'M4 4h7v7H4zM13 4h7v7h-7zM4 13h7v7H4zM13 13h7v7h-7z' },
  { to: '/build', key: 'nav.build' as const, icon: 'M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z' },
  { to: '/submit', key: 'nav.submit' as const, icon: 'M12 5v14M5 12h14' },
]
const desktopNavItems = [
  ...navItems,
]
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
</script>

<template>
  <div class="flex min-h-dvh flex-col pb-mobile-nav md:min-h-screen md:pb-0">
    <!-- top bar: the header is board casing, present on every surface -->
    <header class="sticky top-0 z-20 bg-board-frame text-board-flap-ink shadow-lg shadow-black/20">
      <div class="mx-auto flex max-w-5xl items-center gap-3 px-4 py-3">
        <RouterLink to="/" class="flex items-center gap-2 text-lg font-extrabold">
          <img
            src="/brand/output-smallpngtools.png"
            alt=""
            class="h-8 w-8 rounded-[4px] object-cover"
          />
          <span class="uppercase tracking-wide">Nimiq <span class="text-accent-ink">Mini Apps</span></span>
        </RouterLink>
        <nav class="ml-auto hidden gap-1 md:flex">
          <RouterLink
            v-for="item in desktopNavItems" :key="item.to" :to="item.to"
            class="rounded-[4px] px-3 py-1.5 text-sm font-bold uppercase tracking-wide transition-colors duration-200 hover:bg-board-flap-hover"
            :class="isActive(item.to) ? 'bg-board-flap-hover text-accent-ink' : 'text-board-flap-muted'"
          >{{ t(item.key) }}</RouterLink>
        </nav>
        <WalletLoginButton class="ml-auto md:ml-0" />
        <button @click="toggleTheme" :aria-label="isDark ? t('theme.light') : t('theme.dark')"
          class="grid h-9 w-9 cursor-pointer place-items-center rounded-[4px] border border-board-hairline bg-board-flap-hover text-board-flap-muted transition-colors duration-200 hover:border-lamp-live/50 hover:text-accent-ink md:ml-0">
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
        <div v-if="!insideNimiqPay" class="board relative p-6 md:p-10">
          <div class="relative z-10 flex flex-col gap-5 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 class="text-xl font-extrabold uppercase tracking-wide text-board-flap-ink md:text-2xl">{{ t('footer.title') }}</h2>
              <p class="mt-1 max-w-md text-board-flap-muted">
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

    <!-- bottom nav (mobile): a fixed row of board plates -->
    <nav class="fixed inset-x-0 bottom-0 z-20 bg-board-frame pb-safe shadow-[0_-0.5rem_1.5rem_rgba(0,0,0,0.3)] md:hidden">
      <div class="grid grid-cols-4">
        <RouterLink
          v-for="item in navItems" :key="item.to" :to="item.to"
          class="flex flex-col items-center gap-1 border-t-2 py-2.5 text-[11px] font-bold uppercase tracking-wide transition-colors duration-200"
          :class="isActive(item.to) ? 'border-lamp-live text-accent-ink' : 'border-transparent text-board-flap-muted'"
        >
          <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path :d="item.icon" />
          </svg>
          {{ t(item.key) }}
        </RouterLink>
      </div>
    </nav>
  </div>
</template>
