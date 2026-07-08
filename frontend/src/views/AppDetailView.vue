<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { getApp, type App } from '../api'
import StatusBadge from '../components/StatusBadge.vue'
import ReleaseStageBadge from '../components/ReleaseStageBadge.vue'
import MediaGallery from '../components/MediaGallery.vue'
import { useIsMobileDevice } from '../utils/device'

const route = useRoute()
const isMobile = useIsMobileDevice()
const app = ref<App | null>(null)
const error = ref('')

const aboutText = computed(() => {
  if (!app.value) return ''
  return app.value.long_description || app.value.description
})

onMounted(async () => {
  try {
    app.value = await getApp(route.params.slug as string)
  } catch (e) {
    error.value = (e as Error).message
  }
})
</script>

<template>
  <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>
  <div v-else-if="app" class="space-y-8">
    <img v-if="app.banner_url" :src="app.banner_url" :alt="app.name" class="h-40 w-full rounded-2xl object-cover md:h-56" />

    <div class="flex items-start gap-4">
      <img v-if="app.icon_url" :src="app.icon_url" :alt="app.name" class="h-16 w-16 rounded-2xl object-cover" />
      <div v-else class="grid h-16 w-16 shrink-0 place-items-center rounded-2xl bg-nq-blue text-3xl font-extrabold text-white">
        {{ app.name[0] }}
      </div>
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-extrabold">{{ app.name }}</h1>
          <ReleaseStageBadge v-if="app.release_stage !== 'released'" :stage="app.release_stage" />
          <StatusBadge :status="app.status" />
        </div>
        <p class="text-muted">{{ app.tagline }}</p>
        <RouterLink :to="`/developers/${app.developer_slug}`" class="text-sm text-accent-ink-dark hover:underline dark:text-accent-ink">
          by {{ app.developer_name }}
        </RouterLink>
      </div>
    </div>

    <div class="flex flex-col gap-2 sm:flex-row">
      <a v-if="isMobile" :href="app.open_url" target="_blank" rel="noopener"
        class="cursor-pointer rounded-xl bg-nq-blue px-6 py-3 text-center font-bold text-white transition duration-200 hover:bg-nq-blue-dark">
        Open in Nimiq Pay
      </a>
      <p v-else class="rounded-xl border border-line bg-surface-2 px-6 py-3 text-center text-sm text-muted">
        Open in Nimiq Pay is available on mobile only — browse the details below.
      </p>
      <a v-if="app.website_url" :href="app.website_url" target="_blank" rel="noopener"
        class="cursor-pointer rounded-xl border border-line px-6 py-3 text-center font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">Website</a>
      <a v-if="app.github_url" :href="app.github_url" target="_blank" rel="noopener"
        class="cursor-pointer rounded-xl border border-line px-6 py-3 text-center font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">GitHub</a>
    </div>

    <div class="flex flex-wrap items-center gap-1.5 text-sm">
      <span class="rounded-full bg-accent/10 px-2.5 py-1 font-semibold text-accent-ink">{{ app.category }}</span>
      <span v-for="asset in app.assets" :key="asset" class="rounded-full bg-surface-2 px-2.5 py-1 font-semibold">{{ asset }}</span>
      <span v-for="tag in app.tags" :key="tag" class="rounded-full bg-surface-2 px-2.5 py-1 text-muted">#{{ tag }}</span>
    </div>

    <MediaGallery v-if="app.media?.length" :items="app.media" title="Screenshots &amp; video" />

    <section class="rounded-2xl border border-line bg-surface p-5 shadow-sm md:p-6">
      <h2 class="mb-3 text-lg font-bold">About</h2>
      <p v-if="app.description && app.long_description" class="mb-4 font-semibold text-ink">{{ app.description }}</p>
      <div class="whitespace-pre-line leading-relaxed text-muted">{{ aboutText }}</div>
      <p class="mt-6 text-xs text-muted/70">Domain: {{ app.domain }}</p>
    </section>
  </div>
  <div v-else class="space-y-6" aria-hidden="true">
    <div class="h-40 animate-pulse rounded-2xl border border-line bg-surface md:h-56"></div>
    <div class="h-24 animate-pulse rounded-2xl border border-line bg-surface"></div>
  </div>
</template>
