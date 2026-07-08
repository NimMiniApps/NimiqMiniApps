<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import StoreBadges from './components/StoreBadges.vue'

const route = useRoute()

const navItems = [
  { to: '/', label: 'Home', icon: 'M3 12l9-9 9 9M5 10v10h5v-6h4v6h5V10' },
  { to: '/apps', label: 'Apps', icon: 'M4 4h7v7H4zM13 4h7v7h-7zM4 13h7v7H4zM13 13h7v7h-7z' },
  { to: '/submit', label: 'Submit', icon: 'M12 5v14M5 12h14' },
]
const isActive = (to: string) =>
  to === '/' ? route.path === '/' : route.path.startsWith(to)

const isDark = ref(document.documentElement.classList.contains('dark'))
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.theme = isDark.value ? 'dark' : 'light'
}
</script>

<template>
  <div class="flex min-h-screen flex-col pb-16 md:pb-0">
    <!-- top bar -->
    <header class="sticky top-0 z-20 border-b border-line bg-surface/90 shadow-sm shadow-slate-950/5 backdrop-blur dark:shadow-black/20">
      <div class="mx-auto flex max-w-5xl items-center gap-3 px-4 py-3">
        <RouterLink to="/" class="flex items-center gap-2 text-lg font-extrabold">
          <span class="grid h-8 w-8 place-items-center rounded-lg bg-nq-blue font-extrabold text-white shadow-sm shadow-blue-700/25">N</span>
          <span>Nimiq <span class="text-accent-ink">Mini Apps</span></span>
        </RouterLink>
        <nav class="ml-auto hidden gap-1 md:flex">
          <RouterLink
            v-for="item in navItems" :key="item.to" :to="item.to"
            class="rounded-lg px-3 py-1.5 text-sm font-semibold transition-colors duration-200 hover:bg-surface-2"
            :class="isActive(item.to) ? 'bg-surface-2 text-accent-ink' : 'text-muted'"
          >{{ item.label }}</RouterLink>
        </nav>
        <button @click="toggleTheme" :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
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
              <h2 class="text-xl font-extrabold md:text-2xl">Don't have Nimiq Pay yet?</h2>
              <p class="mt-1 max-w-md text-white/75">
                Get the free self-custodial wallet for NIM and BTC Lightning — and open every mini app here with one tap.
              </p>
            </div>
            <StoreBadges />
          </div>
        </div>
        <p class="mt-4 text-center text-xs text-muted">
          Community-curated directory for <a href="https://www.nimiq.com/nimiq-pay/" target="_blank" rel="noopener" class="text-accent-ink hover:underline">Nimiq Pay</a> mini apps.
        </p>
      </div>
    </footer>

    <!-- bottom nav (mobile) -->
    <nav class="fixed inset-x-0 bottom-0 z-20 border-t border-line bg-surface/95 backdrop-blur md:hidden">
      <div class="grid grid-cols-3">
        <RouterLink
          v-for="item in navItems" :key="item.to" :to="item.to"
          class="flex flex-col items-center gap-1 py-2.5 text-[11px] font-semibold transition-colors duration-200"
          :class="isActive(item.to) ? 'bg-surface-2 text-accent-ink' : 'text-muted'"
        >
          <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path :d="item.icon" />
          </svg>
          {{ item.label }}
        </RouterLink>
      </div>
    </nav>
  </div>
</template>
