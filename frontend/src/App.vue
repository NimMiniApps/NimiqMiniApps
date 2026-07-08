<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()

const navItems = [
  { to: '/', label: 'Home', icon: 'M3 12l9-9 9 9M5 10v10h5v-6h4v6h5V10' },
  { to: '/apps', label: 'Apps', icon: 'M4 4h7v7H4zM13 4h7v7h-7zM4 13h7v7H4zM13 13h7v7h-7z' },
  { to: '/categories', label: 'Categories', icon: 'M4 6h16M4 12h16M4 18h10' },
  { to: '/admin', label: 'Admin', icon: 'M12 15a3 3 0 100-6 3 3 0 000 6zM19 12a7 7 0 11-14 0 7 7 0 0114 0z' },
]
const isActive = (to: string) =>
  to === '/' ? route.path === '/' : route.path.startsWith(to)
</script>

<template>
  <div class="min-h-screen pb-20 md:pb-0">
    <!-- top bar -->
    <header class="sticky top-0 z-20 bg-nq-blue/90 backdrop-blur border-b border-white/10">
      <div class="mx-auto max-w-5xl flex items-center gap-3 px-4 py-3">
        <RouterLink to="/" class="flex items-center gap-2 font-bold text-lg">
          <span class="grid h-8 w-8 place-items-center rounded-lg bg-nq-gold text-nq-blue-darker font-extrabold">N</span>
          <span>Nimiq <span class="text-nq-gold">Mini Apps</span></span>
        </RouterLink>
        <nav class="ml-auto hidden md:flex gap-1">
          <RouterLink
            v-for="item in navItems" :key="item.to" :to="item.to"
            class="rounded-lg px-3 py-1.5 text-sm font-medium hover:bg-white/10"
            :class="isActive(item.to) ? 'text-nq-gold' : 'text-white/80'"
          >{{ item.label }}</RouterLink>
        </nav>
      </div>
    </header>

    <main class="mx-auto max-w-5xl px-4 py-6">
      <RouterView />
    </main>

    <!-- bottom nav (mobile) -->
    <nav class="fixed bottom-0 inset-x-0 z-20 border-t border-white/10 bg-nq-blue/95 backdrop-blur md:hidden">
      <div class="grid grid-cols-4">
        <RouterLink
          v-for="item in navItems" :key="item.to" :to="item.to"
          class="flex flex-col items-center gap-1 py-2.5 text-[11px] font-medium"
          :class="isActive(item.to) ? 'text-nq-gold' : 'text-white/60'"
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
