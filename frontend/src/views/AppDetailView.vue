<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getApp, getRelatedApps, getSubmissionStatus, listAppReviews, type App, type AppReviewsResponse } from '../api'
import { trackProductEvent } from '../utils/analytics'
import AppCard from '../components/AppCard.vue'
import AppBreadcrumb from '../components/AppBreadcrumb.vue'
import EmptyState from '../components/EmptyState.vue'
import DomainStatus from '../components/DomainStatus.vue'
import StatusBadge from '../components/StatusBadge.vue'
import ReleaseStageBadge from '../components/ReleaseStageBadge.vue'
import MediaGallery from '../components/MediaGallery.vue'
import OpenInWalletPanel from '../components/OpenInWalletPanel.vue'
import ShareButton from '../components/ShareButton.vue'
import SocialLinks from '../components/SocialLinks.vue'
import LinkIconButton from '../components/LinkIconButton.vue'
import MarkdownContent from '../components/MarkdownContent.vue'
import AppIcon from '../components/AppIcon.vue'
import HostedByBadge from '../components/HostedByBadge.vue'
import RewardBadge from '../components/RewardBadge.vue'
import CompetitionCycleBadge from '../components/CompetitionCycleBadge.vue'
import ReviewForm from '../components/ReviewForm.vue'
import ReviewList from '../components/ReviewList.vue'
import { displayIconUrl } from '../utils/appIcon'
import { useIsMobileDevice } from '../utils/device'
import { useAdminAuth } from '../composables/useAdminAuth'
import { useWalletAuth } from '../composables/useWalletAuth'
import { useI18n } from '../composables/useI18n'
import { setPageMeta, resetPageMeta } from '../utils/meta'
import { walletOwnsApp } from '../utils/wallet'
import { resolveAppOpenUrl } from '../utils/nimiqWallet'
import { nimConnectPublicUrl, resolveNimConnectHandle } from '../utils/nimconnect'

const route = useRoute()
const isMobile = useIsMobileDevice()
const { isAdmin } = useAdminAuth()
const { walletAddress } = useWalletAuth()
const { t } = useI18n()
const app = ref<App | null>(null)
const related = ref<App[]>([])
const reviewsData = ref<AppReviewsResponse>({ items: [], average: 0, count: 0 })
const myReview = computed(() => reviewsData.value.items.find((rv) => rv.wallet_address === walletAddress.value) ?? null)
const error = ref('')
const loading = ref(true)
const notFound = ref(false)
const updatePending = ref(false)
const nimConnectHandle = ref<string | null>(null)

const isOwner = computed(() => walletOwnsApp(walletAddress.value, app.value?.owner_wallet_addresses))
const openUrl = computed(() => (app.value ? resolveAppOpenUrl(app.value) : ''))
const nimConnectUrl = computed(() =>
  nimConnectHandle.value ? nimConnectPublicUrl(nimConnectHandle.value) : null,
)

let nimConnectLookupSeq = 0
watch(
  () => app.value?.owner_wallet_addresses?.[0],
  async (address) => {
    const seq = ++nimConnectLookupSeq
    nimConnectHandle.value = null
    if (!address) return
    const handle = await resolveNimConnectHandle(address)
    if (seq === nimConnectLookupSeq) nimConnectHandle.value = handle
  },
)

const aboutSource = computed(() => {
  if (!app.value) return ''
  return app.value.long_description || app.value.description
})

async function loadApp(slug: string) {
  error.value = ''
  notFound.value = false
  app.value = null
  related.value = []
  loading.value = true
  try {
    const [loaded, relatedApps, reviews] = await Promise.all([
      getApp(slug),
      getRelatedApps(slug).catch(() => [] as App[]),
      listAppReviews(slug).catch(() => ({ items: [], average: 0, count: 0 }) as AppReviewsResponse),
    ])
    app.value = loaded
    trackProductEvent('app_view', slug)
    related.value = relatedApps
    reviewsData.value = reviews
  } catch (e) {
    const message = (e as Error).message.toLowerCase()
    notFound.value = message.includes('not found')
    error.value = (e as Error).message
    resetPageMeta()
  } finally {
    loading.value = false
  }
}

async function refreshReviews() {
  if (!app.value) return
  reviewsData.value = await listAppReviews(app.value.slug).catch(() => reviewsData.value)
}

watch([app, reviewsData], ([value, reviews]) => {
  if (!value) return
  setPageMeta({
    title: value.name,
    description: value.tagline || value.description,
    image: value.banner_url || displayIconUrl(value) || undefined,
    url: window.location.href,
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: value.name,
      description: value.tagline || value.description,
      applicationCategory: value.category,
      operatingSystem: 'Nimiq Pay Wallet',
      image: value.banner_url || displayIconUrl(value) || undefined,
      url: window.location.href,
      author: { '@type': 'Organization', name: value.developer_name },
      ...(reviews.count
        ? {
            aggregateRating: {
              '@type': 'AggregateRating',
              ratingValue: reviews.average,
              reviewCount: reviews.count,
            },
          }
        : {}),
    },
  })
})

watch(() => route.params.slug, (slug) => {
  if (typeof slug === 'string') loadApp(slug)
})

watch([isOwner, () => app.value?.slug], async ([owned, slug]) => {
  if (!owned || !slug) {
    updatePending.value = false
    return
  }
  const status = await getSubmissionStatus(slug).catch(() => null)
  updatePending.value = !!status?.update_pending
}, { immediate: true })

onMounted(() => {
  const slug = route.params.slug as string
  if (slug) loadApp(slug)
})

onUnmounted(resetPageMeta)
</script>

<template>
  <EmptyState
    v-if="error && notFound"
    :title="t('appDetail.notFoundTitle')"
    :description="t('appDetail.notFoundBody')"
    variant="notFound"
  >
    <template #actions>
      <RouterLink
        to="/apps"
        class="cursor-pointer board-plate board-plate-primary px-5 py-2.5 text-sm"
      >
        {{ t('common.browseAll') }}
      </RouterLink>
    </template>
  </EmptyState>

  <EmptyState
    v-else-if="error"
    :title="t('appDetail.errorTitle')"
    :description="t('appDetail.errorBody')"
    variant="error"
  >
    <template #actions>
      <button
        type="button"
        class="cursor-pointer board-plate board-plate-ghost px-5 py-2.5 text-sm"
        @click="loadApp(route.params.slug as string)"
      >
        {{ t('common.retry') }}
      </button>
    </template>
  </EmptyState>

  <div v-else-if="app" class="space-y-8">
    <AppBreadcrumb :category="app.category" :app-name="app.name" />

    <DomainStatus :reachable="app.domain_reachable" variant="banner" />

    <div
      v-if="app.banner_url"
      class="flex h-[clamp(12rem,35vw,28rem)] items-center justify-center overflow-hidden rounded-2xl border border-line bg-surface-2 p-3 sm:p-4"
    >
      <img :src="app.banner_url" :alt="app.name" class="h-full w-full object-contain" />
    </div>

    <div class="flex items-start gap-4">
      <AppIcon :app="app" size="md" />
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-extrabold">{{ app.name }}</h1>
          <ReleaseStageBadge v-if="app.release_stage !== 'released'" :stage="app.release_stage" />
          <StatusBadge :status="app.status" />
          <CompetitionCycleBadge
            v-if="app.competition_cycle || app.competition_results?.length"
            :cycle="app.competition_cycle"
            :results="app.competition_results"
          />
          <RewardBadge :assets="app.reward_assets" />
          <HostedByBadge :domain="app.domain" />
          <DomainStatus v-if="app.domain_reachable === false" :reachable="app.domain_reachable" />
        </div>
        <p class="text-muted">{{ app.tagline }}</p>
        <ul
          v-if="app.competition_results?.length > 1"
          class="mt-2 flex flex-wrap gap-2 text-xs font-semibold uppercase tracking-wide text-muted"
        >
          <li
            v-for="row in app.competition_results"
            :key="row.cycle"
            class="rounded-[3px] border px-2 py-0.5"
            :class="row.place === 1
              ? 'border-amber-500/70 bg-gradient-to-b from-amber-300 to-amber-500 text-amber-950'
              : row.place === 2
                ? 'border-slate-400/80 bg-gradient-to-b from-slate-100 to-slate-300 text-slate-900'
                : row.place === 3
                  ? 'border-orange-700/60 bg-gradient-to-b from-orange-300 to-orange-600 text-orange-950'
                  : 'border-board-hairline bg-board-flap text-board-flap-ink'"
          >
            <template v-if="row.place === 1 || row.place === 2 || row.place === 3">
              {{ row.place === 1 ? '1st' : row.place === 2 ? '2nd' : '3rd' }} place · Cycle {{ row.cycle }}
            </template>
            <template v-else>
              Cycle {{ row.cycle }} · {{ row.score }}
            </template>
          </li>
        </ul>
        <div class="flex flex-wrap items-baseline gap-x-2 gap-y-0.5 text-sm">
          <RouterLink :to="`/apps?developer=${encodeURIComponent(app.developer_slug)}`" class="text-accent-ink-dark hover:underline dark:text-accent-ink">
            {{ t('common.by') }} {{ app.developer_name }}
          </RouterLink>
          <a
            v-if="nimConnectUrl && nimConnectHandle"
            :href="nimConnectUrl"
            target="_blank"
            rel="noopener"
            :aria-label="t('common.nimconnectProfile')"
            :title="t('common.nimconnectProfile')"
            class="text-muted hover:text-accent-ink-dark hover:underline dark:hover:text-accent-ink"
          >@{{ nimConnectHandle }}</a>
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
      <div class="flex flex-wrap items-center gap-2">
        <a v-if="isMobile" :href="openUrl" target="_blank" rel="noopener"
          class="inline-flex h-10 cursor-pointer items-center board-plate board-plate-primary px-5 text-sm"
          @click="trackProductEvent('app_open', app.slug)">
          {{ t('appDetail.openInWallet') }}
        </a>
        <ShareButton :title="app.name" :app-slug="app.slug" />
        <LinkIconButton
          v-if="app.github_url"
          :href="app.github_url"
          platform="github"
          :label="t('appDetail.github')"
        />
        <SocialLinks v-if="app.socials?.length" :items="app.socials" />
        <RouterLink
          v-if="isOwner"
          :to="`/apps/${app.slug}/update`"
          class="inline-flex h-10 cursor-pointer items-center rounded-xl border border-emerald-500/40 bg-emerald-500/10 px-4 text-sm font-semibold text-emerald-800 transition-colors duration-200 hover:border-emerald-500/60 hover:bg-emerald-500/15 dark:text-emerald-200"
          :class="{ 'pointer-events-none opacity-60': updatePending }"
          :aria-disabled="updatePending"
          :title="updatePending ? t('appDetail.updatePendingHint') : undefined"
        >
          {{ t('appDetail.editListing') }}
        </RouterLink>
        <RouterLink v-if="isAdmin" :to="`/admin?edit=${app.slug}`"
          class="inline-flex h-10 cursor-pointer items-center rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 text-sm font-semibold text-amber-800 transition-colors duration-200 hover:border-amber-500/60 hover:bg-amber-500/15 dark:text-amber-200">
          {{ t('appDetail.edit') }}
        </RouterLink>
      </div>
      <OpenInWalletPanel v-if="!isMobile" :open-url="app.open_url" :slug="app.slug" :domain="app.domain" class="sm:ml-auto" />
    </div>

    <div class="flex flex-wrap items-center gap-1.5 text-sm">
      <RouterLink :to="`/apps?category=${encodeURIComponent(app.category)}`"
        class="rounded-full bg-accent/10 px-2.5 py-1 font-semibold text-accent-ink transition-colors hover:bg-accent/20">
        {{ app.category }}
      </RouterLink>
      <RouterLink v-for="asset in app.assets" :key="asset" :to="`/apps?asset=${encodeURIComponent(asset)}`"
        class="rounded-full bg-surface-2 px-2.5 py-1 font-semibold transition-colors hover:bg-accent/10 hover:text-accent-ink">
        {{ asset }}
      </RouterLink>
      <RewardBadge :assets="app.reward_assets" />
      <RouterLink v-for="tag in app.tags" :key="tag" :to="`/apps?tag=${encodeURIComponent(tag)}`"
        class="rounded-full bg-surface-2 px-2.5 py-1 text-muted transition-colors hover:bg-accent/10 hover:text-accent-ink">
        #{{ tag }}
      </RouterLink>
    </div>

    <div v-if="app.declared_scopes?.length" class="space-y-1.5">
      <h2 class="text-sm font-bold">{{ t('appDetail.declaredScopes') }}</h2>
      <p class="text-xs text-muted">{{ t('appDetail.declaredScopesHint') }}</p>
      <div class="flex flex-wrap gap-1.5 text-sm">
        <span
          v-for="scope in app.declared_scopes"
          :key="scope"
          class="rounded-md bg-surface-2 px-2.5 py-1 font-mono text-xs font-semibold"
        >{{ scope }}</span>
      </div>
    </div>

    <MediaGallery v-if="app.media?.length" :items="app.media" :title="t('appDetail.media')" />

    <section class="rounded-2xl border border-line bg-surface p-5 shadow-sm md:p-6">
      <h2 class="mb-3 text-lg font-bold">{{ t('appDetail.about') }}</h2>
      <p v-if="app.description && app.long_description" class="mb-4 font-semibold text-ink">{{ app.description }}</p>
      <MarkdownContent :source="aboutSource" />
      <p class="mt-6 text-xs text-muted/70">{{ t('appDetail.domain') }}: {{ app.domain }}</p>
    </section>

    <section class="space-y-4">
      <div class="flex items-baseline justify-between">
        <h2 class="text-lg font-bold">Reviews</h2>
        <span v-if="reviewsData.count" class="text-sm text-muted">
          {{ reviewsData.average.toFixed(1) }} ★ · {{ reviewsData.count }} review{{ reviewsData.count === 1 ? '' : 's' }}
        </span>
      </div>
      <ReviewForm
        v-if="walletAddress"
        :slug="app.slug"
        :existing="myReview"
        @saved="refreshReviews"
      />
      <p v-else class="text-sm text-muted">Connect your wallet to leave a review.</p>
      <ReviewList
        v-if="reviewsData.items.length"
        :slug="app.slug"
        :reviews="reviewsData.items"
        :wallet-address="walletAddress"
        @deleted="refreshReviews"
      />
    </section>

    <section v-if="related.length" class="space-y-4">
      <h2 class="text-xl font-extrabold">{{ t('appDetail.related') }}</h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="item in related" :key="item.id" :app="item" />
      </div>
    </section>
  </div>

  <div v-else-if="loading" class="space-y-6" aria-hidden="true">
    <div class="h-5 w-48 animate-pulse rounded-lg bg-surface-2"></div>
    <div class="h-40 animate-pulse rounded-2xl border border-line bg-surface md:h-56"></div>
    <div class="h-24 animate-pulse rounded-2xl border border-line bg-surface"></div>
  </div>
</template>
