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

const props = defineProps<{
  app: App
  owned?: boolean
  pendingUpdate?: boolean
  showManageActions?: boolean
}>()

const categoryThemes: Record<string, { accent: string; soft: string; ink: string }> = {
  games: { accent: '#0582ca', soft: 'rgba(5, 130, 202, 0.1)', ink: '#0582ca' },
  utilities: { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)', ink: '#168f80' },
  finance: { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)', ink: '#168f80' },
  maps: { accent: '#e9b213', soft: 'rgba(233, 178, 19, 0.16)', ink: '#9c7300' },
  social: { accent: '#fa7268', soft: 'rgba(250, 114, 104, 0.13)', ink: '#c44941' },
  experiments: { accent: '#5f4b8b', soft: 'rgba(95, 75, 139, 0.13)', ink: '#5f4b8b' },
}

const appIdentityThemes = [
  { accent: '#0582ca', soft: 'rgba(5, 130, 202, 0.1)' },
  { accent: '#21bca5', soft: 'rgba(33, 188, 165, 0.12)' },
  { accent: '#e9b213', soft: 'rgba(233, 178, 19, 0.16)' },
  { accent: '#fa7268', soft: 'rgba(250, 114, 104, 0.13)' },
  { accent: '#5f4b8b', soft: 'rgba(95, 75, 139, 0.13)' },
  { accent: '#fc8702', soft: 'rgba(252, 135, 2, 0.14)' },
]

const fallbackTheme = { accent: '#1f2348', soft: 'rgba(31, 35, 72, 0.06)', ink: '#1f2348' }
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
    class="nq-card-shadow relative flex flex-col gap-3 overflow-hidden rounded-[10px] border border-line bg-surface p-4 transition-all duration-200 hover:-translate-y-0.5 dark:shadow-black/20"
    :style="{ borderColor: `${identityTheme.accent}55`, background: `linear-gradient(135deg, ${identityTheme.soft}, transparent 44%), var(--nq-surface)` }"
  >
    <div class="absolute inset-x-0 top-0 h-1.5" :style="{ backgroundColor: identityTheme.accent }" aria-hidden="true"></div>
    <div class="flex items-start gap-3">
      <AppIcon :app="app" />
      <div class="min-w-0 flex-1">
        <h3 class="truncate font-bold">{{ app.name }}</h3>
        <div class="mt-1 flex flex-wrap items-center gap-1.5">
          <span
            v-if="owned"
            class="rounded-full bg-emerald-500/15 px-2 py-0.5 text-[11px] font-semibold text-emerald-700 dark:text-emerald-300"
            title="Linked to your wallet"
          >Yours</span>
          <span
            v-if="pendingUpdate"
            class="rounded-full bg-amber-500/15 px-2 py-0.5 text-[11px] font-semibold text-amber-800 dark:text-amber-200"
          >Update pending</span>
          <ReleaseStageBadge v-if="app.release_stage !== 'released'" :stage="app.release_stage" />
          <StatusBadge :status="app.status" />
          <HostedByBadge :domain="app.domain" compact />
          <DomainStatus
            v-if="app.domain_reachable != null"
            :reachable="app.domain_reachable"
            show-online
            compact
          />
        </div>
        <p class="mt-1 text-sm text-muted line-clamp-2">{{ app.tagline }}</p>
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
      <a v-if="isMobile && !showManageActions" :href="app.open_url" target="_blank" rel="noopener"
        class="nq-primary flex-1 cursor-pointer rounded-[500px] px-3 py-2 text-center text-sm font-bold text-white transition duration-200">
        Open in Nimiq Pay
      </a>
      <RouterLink
        v-if="showManageActions"
        :to="`/apps/${app.slug}/update`"
        class="nq-primary flex-1 cursor-pointer rounded-[500px] px-3 py-2 text-center text-sm font-bold text-white transition duration-200"
        :class="{ 'opacity-60 pointer-events-none': pendingUpdate }"
        :aria-disabled="pendingUpdate"
        :title="pendingUpdate ? 'An update is already pending review' : undefined"
      >
        Edit listing
      </RouterLink>
      <RouterLink :to="`/apps/${app.slug}`"
        class="flex-1 cursor-pointer rounded-[500px] border border-line bg-surface px-3 py-2 text-center text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
        {{ showManageActions ? 'View' : 'Details' }}
      </RouterLink>
      <a v-if="isMobile && showManageActions" :href="app.open_url" target="_blank" rel="noopener"
        class="cursor-pointer rounded-[500px] border border-line bg-surface px-3 py-2 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
        Open
      </a>
    </div>
  </div>
</template>
