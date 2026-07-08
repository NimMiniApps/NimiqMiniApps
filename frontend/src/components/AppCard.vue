<script setup lang="ts">
import type { App } from '../api'
import StatusBadge from './StatusBadge.vue'

defineProps<{ app: App }>()
</script>

<template>
  <div class="flex flex-col gap-3 rounded-2xl border border-white/10 bg-nq-card p-4">
    <div class="flex items-start gap-3">
      <!-- icon (placeholder = first letter) -->
      <img v-if="app.icon_url" :src="app.icon_url" :alt="app.name" class="h-12 w-12 rounded-xl object-cover" />
      <div v-else class="grid h-12 w-12 shrink-0 place-items-center rounded-xl bg-nq-gold/90 text-xl font-extrabold text-nq-blue-darker">
        {{ app.name[0] }}
      </div>
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <h3 class="truncate font-bold">{{ app.name }}</h3>
          <StatusBadge :status="app.status" />
        </div>
        <p class="text-sm text-white/70 line-clamp-2">{{ app.tagline }}</p>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-1.5 text-xs">
      <span class="rounded-full bg-nq-gold/15 px-2 py-0.5 font-semibold text-nq-gold">{{ app.category }}</span>
      <span v-for="asset in app.assets" :key="asset" class="rounded-full bg-white/10 px-2 py-0.5 font-semibold">{{ asset }}</span>
      <span v-for="tag in app.tags" :key="tag" class="rounded-full bg-white/5 px-2 py-0.5 text-white/60">#{{ tag }}</span>
    </div>

    <div class="mt-auto flex gap-2">
      <a :href="app.open_url" target="_blank" rel="noopener"
        class="flex-1 rounded-xl bg-nq-gold px-3 py-2 text-center text-sm font-bold text-nq-blue-darker hover:bg-nq-gold-dark">
        Open in Nimiq Pay
      </a>
      <RouterLink :to="`/apps/${app.slug}`"
        class="rounded-xl border border-white/20 px-3 py-2 text-sm font-semibold text-white/90 hover:bg-white/10">
        Details
      </RouterLink>
    </div>
  </div>
</template>
