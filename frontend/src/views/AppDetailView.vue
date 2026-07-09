<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getApp, getRelatedApps, type App } from '../api'
import AppCard from '../components/AppCard.vue'
import AppBreadcrumb from '../components/AppBreadcrumb.vue'
import EmptyState from '../components/EmptyState.vue'
import DomainStatus from '../components/DomainStatus.vue'
import StatusBadge from '../components/StatusBadge.vue'
import ReleaseStageBadge from '../components/ReleaseStageBadge.vue'
import MediaGallery from '../components/MediaGallery.vue'
import OpenInWalletPanel from '../components/OpenInWalletPanel.vue'
import SocialLinks from '../components/SocialLinks.vue'
import LinkIconButton from '../components/LinkIconButton.vue'
import MarkdownContent from '../components/MarkdownContent.vue'
import AppIcon from '../components/AppIcon.vue'
import HostedByBadge from '../components/HostedByBadge.vue'
import { displayIconUrl } from '../utils/appIcon'
import { useIsMobileDevice } from '../utils/device'
import { useAdminAuth } from '../composables/useAdminAuth'
import { useI18n } from '../composables/useI18n'
import { setPageMeta, resetPageMeta } from '../utils/meta'

const route = useRoute()
const isMobile = useIsMobileDevice()
const { isAdmin } = useAdminAuth()
const { t } = useI18n()
const app = ref<App | null>(null)
const related = ref<App[]>([])
const error = ref('')
const loading = ref(true)
const notFound = ref(false)

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
    const [loaded, relatedApps] = await Promise.all([
      getApp(slug),
      getRelatedApps(slug).catch(() => [] as App[]),
    ])
    app.value = loaded
    related.value = relatedApps
  } catch (e) {
    const message = (e as Error).message.toLowerCase()
    notFound.value = message.includes('not found')
    error.value = (e as Error).message
    resetPageMeta()
  } finally {
    loading.value = false
  }
}

watch(app, (value) => {
  if (!value) return
  setPageMeta({
    title: value.name,
    description: value.tagline || value.description,
    image: value.banner_url || displayIconUrl(value) || undefined,
    url: window.location.href,
  })
})

watch(() => route.params.slug, (slug) => {
  if (typeof slug === 'string') loadApp(slug)
})

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
        class="cursor-pointer rounded-xl bg-nq-blue px-5 py-2.5 text-sm font-bold text-white transition duration-200 hover:bg-nq-blue-dark"
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
        class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
        @click="loadApp(route.params.slug as string)"
      >
        {{ t('common.retry') }}
      </button>
    </template>
  </EmptyState>

  <div v-else-if="app" class="space-y-8">
    <AppBreadcrumb :category="app.category" :app-name="app.name" />

    <DomainStatus :reachable="app.domain_reachable" variant="banner" />

    <img v-if="app.banner_url" :src="app.banner_url" :alt="app.name" class="h-40 w-full rounded-2xl object-cover md:h-56" />

    <div class="flex items-start gap-4">
      <AppIcon :app="app" size="md" />
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-extrabold">{{ app.name }}</h1>
          <ReleaseStageBadge v-if="app.release_stage !== 'released'" :stage="app.release_stage" />
          <StatusBadge :status="app.status" />
          <HostedByBadge :domain="app.domain" />
          <DomainStatus v-if="app.domain_reachable === false" :reachable="app.domain_reachable" />
        </div>
        <p class="text-muted">{{ app.tagline }}</p>
        <RouterLink :to="`/apps?developer=${encodeURIComponent(app.developer_slug)}`" class="text-sm text-accent-ink-dark hover:underline dark:text-accent-ink">
          {{ t('common.by') }} {{ app.developer_name }}
        </RouterLink>
      </div>
    </div>

    <div class="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
      <div class="flex flex-wrap items-center gap-2">
        <a v-if="isMobile" :href="app.open_url" target="_blank" rel="noopener"
          class="inline-flex h-10 cursor-pointer items-center rounded-xl bg-nq-blue px-5 text-sm font-bold text-white transition duration-200 hover:bg-nq-blue-dark">
          {{ t('appDetail.openInWallet') }}
        </a>
        <ShareButton :title="app.name" />
        <RouterLink :to="`/apps/${app.slug}/update`"
          class="inline-flex h-10 cursor-pointer items-center rounded-xl border border-line bg-surface px-4 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
          {{ t('appDetail.suggestUpdate') }}
        </RouterLink>
        <LinkIconButton
          v-if="app.website_url"
          :href="app.website_url"
          platform="website"
          :label="t('appDetail.website')"
        />
        <LinkIconButton
          v-if="app.github_url"
          :href="app.github_url"
          platform="github"
          :label="t('appDetail.github')"
        />
        <SocialLinks v-if="app.socials?.length" :items="app.socials" />
        <RouterLink v-if="isAdmin" :to="`/admin?edit=${app.slug}`"
          class="inline-flex h-10 cursor-pointer items-center rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 text-sm font-semibold text-amber-800 transition-colors duration-200 hover:border-amber-500/60 hover:bg-amber-500/15 dark:text-amber-200">
          {{ t('appDetail.edit') }}
        </RouterLink>
      </div>
      <OpenInWalletPanel v-if="!isMobile" :open-url="app.open_url" class="sm:ml-auto" />
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
      <RouterLink v-for="tag in app.tags" :key="tag" :to="`/apps?tag=${encodeURIComponent(tag)}`"
        class="rounded-full bg-surface-2 px-2.5 py-1 text-muted transition-colors hover:bg-accent/10 hover:text-accent-ink">
        #{{ tag }}
      </RouterLink>
    </div>

    <MediaGallery v-if="app.media?.length" :items="app.media" :title="t('appDetail.media')" />

    <section class="rounded-2xl border border-line bg-surface p-5 shadow-sm md:p-6">
      <h2 class="mb-3 text-lg font-bold">{{ t('appDetail.about') }}</h2>
      <p v-if="app.description && app.long_description" class="mb-4 font-semibold text-ink">{{ app.description }}</p>
      <MarkdownContent :source="aboutSource" />
      <p class="mt-6 text-xs text-muted/70">{{ t('appDetail.domain') }}: {{ app.domain }}</p>
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
