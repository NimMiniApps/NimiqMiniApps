<script setup lang="ts">
import { computed } from 'vue'
import type { App } from '../api'
import { trackAppEvent } from '../api'
import { useIsMobileDevice } from '../utils/device'
import { useWalletAuth } from '../composables/useWalletAuth'
import { useFavorites } from '../composables/useFavorites'
import { resolveAppOpenUrl } from '../utils/nimiqWallet'
import { categoryTheme } from '../utils/appIcon'
import StatusBadge from '../components/StatusBadge.vue'
import ReleaseStageBadge from './ReleaseStageBadge.vue'
import DomainStatus from './DomainStatus.vue'
import AppIcon from './AppIcon.vue'
import HostedByBadge from './HostedByBadge.vue'
import RewardBadge from './RewardBadge.vue'
import FlapText from './FlapText.vue'

const isMobile = useIsMobileDevice()
const { walletAddress } = useWalletAuth()
const { isFavorite, toggleFavorite } = useFavorites()

const props = defineProps<{
  app: App
  owned?: boolean
  pendingUpdate?: boolean
  showManageActions?: boolean
}>()

const appCategoryTheme = computed(() => categoryTheme(props.app.category))

const openUrl = computed(() => resolveAppOpenUrl(props.app))

const previewTags = computed(() => props.app.tags.slice(0, 3))
const extraTagCount = computed(() => Math.max(0, props.app.tags.length - previewTags.value.length))

function trackOpen() {
  trackAppEvent(props.app.slug, 'open')
}

function onFavoriteClick() {
  toggleFavorite(props.app.slug)
}
</script>

<template>
  <div class="board relative flex h-full min-h-0 flex-col gap-3 p-4">
    <button
      v-if="walletAddress"
      type="button"
      class="absolute right-3 top-3 z-10 rounded-[3px] p-1 text-lg leading-none transition-colors duration-150"
      :class="isFavorite(app.slug) ? 'text-rose-400' : 'text-board-flap-muted hover:text-rose-300'"
      :aria-pressed="isFavorite(app.slug)"
      :aria-label="isFavorite(app.slug) ? 'Remove from favorites' : 'Add to favorites'"
      :title="isFavorite(app.slug) ? 'Remove from favorites' : 'Add to favorites'"
      @click.stop.prevent="onFavoriteClick"
    >
      {{ isFavorite(app.slug) ? '♥' : '♡' }}
    </button>

    <div class="flex items-start gap-3">
      <AppIcon :app="app" />
      <div class="min-w-0 flex-1 pr-6">
        <h3 class="truncate font-bold text-board-flap-ink">
          <FlapText :text="app.name" />
        </h3>
        <div class="mt-1 flex min-h-[2.75rem] flex-wrap items-center gap-1.5">
          <span
            class="rounded-[3px] px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-white"
            :style="{ backgroundColor: appCategoryTheme.accent }"
          >{{ app.category }}</span>
          <span
            v-if="owned"
            class="rounded-[3px] border border-board-hairline bg-board-flap-hover px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-board-flap-ink"
            title="Linked to your wallet"
          >Yours</span>
          <span
            v-if="pendingUpdate"
            class="rounded-[3px] border border-board-hairline bg-board-flap-hover px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-board-flap-ink"
          >Update pending</span>
          <ReleaseStageBadge v-if="app.release_stage !== 'released'" :stage="app.release_stage" />
          <StatusBadge :status="app.status" />
          <RewardBadge :assets="app.reward_assets" compact />
          <HostedByBadge :domain="app.domain" compact />
          <DomainStatus
            v-if="app.domain_reachable != null"
            :reachable="app.domain_reachable"
            show-online
            compact
          />
          <span
            v-if="app.review_count > 0"
            class="rounded-[3px] border border-board-hairline bg-board-flap-hover px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-board-flap-ink"
          >{{ app.avg_rating.toFixed(1) }} ★ ({{ app.review_count }})</span>
        </div>
        <p class="mt-1 text-sm text-board-flap-muted line-clamp-2">{{ app.tagline }}</p>
      </div>
    </div>

    <div class="flex min-h-[1.625rem] flex-wrap items-center gap-1.5 font-mono text-[10px]">
      <span v-for="asset in app.assets" :key="asset"
        class="rounded-[3px] bg-board-flap-hover px-1.5 py-0.5 font-bold text-board-flap-ink">
        <RouterLink :to="`/apps?asset=${encodeURIComponent(asset)}`"
          class="transition-colors hover:text-accent-ink">{{ asset }}</RouterLink>
      </span>
      <span v-for="tag in previewTags" :key="tag" class="rounded-[3px] bg-board-flap-hover px-1.5 py-0.5 text-board-flap-muted">
        <RouterLink :to="`/apps?tag=${encodeURIComponent(tag)}`"
          class="transition-colors hover:text-accent-ink">#{{ tag }}</RouterLink>
      </span>
      <span v-if="extraTagCount" class="rounded-[3px] px-1.5 py-0.5 text-board-flap-muted/80">
        +{{ extraTagCount }} more
      </span>
    </div>

    <div class="mt-auto flex gap-2">
      <a v-if="isMobile && !showManageActions" :href="openUrl" target="_blank" rel="noopener"
        class="board-plate board-plate-primary min-w-0 flex-1 cursor-pointer px-3 py-2 text-center text-xs"
        @click="trackOpen">
        Open in Nimiq Pay
      </a>
      <RouterLink
        v-if="showManageActions"
        :to="`/apps/${app.slug}/update`"
        class="board-plate board-plate-primary min-w-0 flex-1 cursor-pointer px-3 py-2 text-center text-xs"
        :class="{ 'opacity-60 pointer-events-none': pendingUpdate }"
        :aria-disabled="pendingUpdate"
        :title="pendingUpdate ? 'An update is already pending review' : undefined"
      >
        Edit listing
      </RouterLink>
      <RouterLink :to="`/apps/${app.slug}`"
        class="board-plate board-plate-ghost min-w-0 flex-1 cursor-pointer px-3 py-2 text-center text-xs">
        {{ showManageActions ? 'View' : 'Details' }}
      </RouterLink>
      <a v-if="isMobile && showManageActions" :href="openUrl" target="_blank" rel="noopener"
        class="board-plate board-plate-ghost shrink-0 cursor-pointer px-3 py-2 text-center text-xs"
        @click="trackOpen">
        Open
      </a>
    </div>
  </div>
</template>
