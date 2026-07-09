<script setup lang="ts">
import { computed } from 'vue'
import type { App } from '../api'
import { useIsMobileDevice } from '../utils/device'
import StatusBadge from '../components/StatusBadge.vue'
import ReleaseStageBadge from './ReleaseStageBadge.vue'
import DomainStatus from './DomainStatus.vue'
import AppIcon from './AppIcon.vue'
import HostedByBadge from './HostedByBadge.vue'

const isMobile = useIsMobileDevice()

const props = defineProps<{ app: App }>()

const categoryThemes: Record<string, { accent: string; soft: string; ink: string }> = {
  games: { accent: '#1f74ff', soft: 'rgba(31, 116, 255, 0.13)', ink: '#1557c7' },
  utilities: { accent: '#14b8a6', soft: 'rgba(20, 184, 166, 0.14)', ink: '#0f766e' },
  finance: { accent: '#22c55e', soft: 'rgba(34, 197, 94, 0.14)', ink: '#15803d' },
  maps: { accent: '#f59e0b', soft: 'rgba(245, 158, 11, 0.16)', ink: '#b45309' },
  social: { accent: '#f43f5e', soft: 'rgba(244, 63, 94, 0.14)', ink: '#be123c' },
  experiments: { accent: '#a855f7', soft: 'rgba(168, 85, 247, 0.15)', ink: '#7e22ce' },
}

const appIdentityThemes = [
  { accent: '#1f74ff', soft: 'rgba(31, 116, 255, 0.13)' },
  { accent: '#14b8a6', soft: 'rgba(20, 184, 166, 0.14)' },
  { accent: '#f59e0b', soft: 'rgba(245, 158, 11, 0.16)' },
  { accent: '#f43f5e', soft: 'rgba(244, 63, 94, 0.14)' },
  { accent: '#a855f7', soft: 'rgba(168, 85, 247, 0.15)' },
  { accent: '#22c55e', soft: 'rgba(34, 197, 94, 0.14)' },
]

const fallbackTheme = { accent: '#64748b', soft: 'rgba(100, 116, 139, 0.14)', ink: '#475569' }
const categoryTheme = computed(() => {
  const key = props.app.category.toLowerCase()
  return categoryThemes[key] ?? fallbackTheme
})

const identityTheme = computed(() => {
  const source = props.app.slug || props.app.name
  const themeIndex = [...source].reduce((sum, char) => sum + char.charCodeAt(0), 0) % appIdentityThemes.length
  return appIdentityThemes[themeIndex]
})

const previewTags = computed(() => props.app.tags.slice(0, 3))
const extraTagCount = computed(() => Math.max(0, props.app.tags.length - previewTags.value.length))
</script>

<template>
  <div
    class="relative flex flex-col gap-3 overflow-hidden rounded-2xl border border-line bg-surface p-4 shadow-sm shadow-slate-950/5 transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md dark:shadow-black/20"
    :style="{ borderColor: `${identityTheme.accent}55`, background: `linear-gradient(135deg, ${identityTheme.soft}, transparent 44%), var(--nq-surface)` }"
  >
    <div class="absolute inset-x-0 top-0 h-1.5" :style="{ backgroundColor: identityTheme.accent }" aria-hidden="true"></div>
    <div class="flex items-start gap-3">
      <AppIcon :app="app" />
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <h3 class="truncate font-bold">{{ app.name }}</h3>
          <ReleaseStageBadge v-if="app.release_stage !== 'released'" :stage="app.release_stage" />
          <StatusBadge :status="app.status" />
          <HostedByBadge :domain="app.domain" />
          <DomainStatus
            v-if="app.domain_reachable != null"
            :reachable="app.domain_reachable"
            show-online
            compact
          />
        </div>
        <p class="text-sm text-muted line-clamp-2">{{ app.tagline }}</p>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-1.5 text-xs">
      <span
        class="rounded-full px-2 py-0.5 font-semibold ring-1"
        :style="{ backgroundColor: categoryTheme.soft, color: categoryTheme.ink, borderColor: categoryTheme.accent }"
      >{{ app.category }}</span>
      <span v-for="asset in app.assets" :key="asset"
        class="rounded-full bg-surface-2 px-2 py-0.5 font-semibold text-ink">
        <RouterLink :to="`/apps?asset=${encodeURIComponent(asset)}`"
          class="transition-colors hover:text-accent-ink">{{ asset }}</RouterLink>
      </span>
      <span v-for="tag in previewTags" :key="tag" class="rounded-full bg-surface-2 px-2 py-0.5 text-muted">
        <RouterLink :to="`/apps?tag=${encodeURIComponent(tag)}`"
          class="transition-colors hover:text-accent-ink">#{{ tag }}</RouterLink>
      </span>
      <span v-if="extraTagCount" class="rounded-full px-2 py-0.5 text-muted/80">
        +{{ extraTagCount }} more
      </span>
    </div>

    <div class="mt-auto flex gap-2">
      <a v-if="isMobile" :href="app.open_url" target="_blank" rel="noopener"
        class="flex-1 cursor-pointer rounded-xl bg-nq-blue px-3 py-2 text-center text-sm font-bold text-white shadow-sm shadow-blue-700/20 transition duration-200 hover:bg-nq-blue-dark">
        Open in Nimiq Pay
      </a>
      <RouterLink :to="`/apps/${app.slug}`"
        class="flex-1 cursor-pointer rounded-xl border border-line px-3 py-2 text-center text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
        Details
      </RouterLink>
    </div>
  </div>
</template>
