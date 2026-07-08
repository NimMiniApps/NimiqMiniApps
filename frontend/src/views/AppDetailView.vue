<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getApp, type App } from '../api'
import StatusBadge from '../components/StatusBadge.vue'

const route = useRoute()
const app = ref<App | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    app.value = await getApp(route.params.slug as string)
  } catch (e) {
    error.value = (e as Error).message
  }
})
</script>

<template>
  <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-200">{{ error }}</p>
  <div v-else-if="app" class="space-y-6">
    <img v-if="app.banner_url" :src="app.banner_url" :alt="app.name" class="h-40 w-full rounded-2xl object-cover md:h-56" />

    <div class="flex items-start gap-4">
      <img v-if="app.icon_url" :src="app.icon_url" :alt="app.name" class="h-16 w-16 rounded-2xl object-cover" />
      <div v-else class="grid h-16 w-16 shrink-0 place-items-center rounded-2xl bg-nq-gold text-3xl font-extrabold text-nq-blue-darker">
        {{ app.name[0] }}
      </div>
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-2xl font-extrabold">{{ app.name }}</h1>
          <StatusBadge :status="app.status" />
        </div>
        <p class="text-white/70">{{ app.tagline }}</p>
        <RouterLink :to="`/developers/${app.developer_slug}`" class="text-sm text-nq-gold hover:underline">
          by {{ app.developer_name }}
        </RouterLink>
      </div>
    </div>

    <div class="flex flex-col gap-2 sm:flex-row">
      <a :href="app.open_url" target="_blank" rel="noopener"
        class="rounded-xl bg-nq-gold px-6 py-3 text-center font-bold text-nq-blue-darker hover:bg-nq-gold-dark">
        Open in Nimiq Pay
      </a>
      <a v-if="app.website_url" :href="app.website_url" target="_blank" rel="noopener"
        class="rounded-xl border border-white/20 px-6 py-3 text-center font-semibold hover:bg-white/10">Website</a>
      <a v-if="app.github_url" :href="app.github_url" target="_blank" rel="noopener"
        class="rounded-xl border border-white/20 px-6 py-3 text-center font-semibold hover:bg-white/10">GitHub</a>
    </div>

    <div class="flex flex-wrap items-center gap-1.5 text-sm">
      <span class="rounded-full bg-nq-gold/15 px-2.5 py-1 font-semibold text-nq-gold">{{ app.category }}</span>
      <span v-for="asset in app.assets" :key="asset" class="rounded-full bg-white/10 px-2.5 py-1 font-semibold">{{ asset }}</span>
      <span v-for="tag in app.tags" :key="tag" class="rounded-full bg-white/5 px-2.5 py-1 text-white/60">#{{ tag }}</span>
    </div>

    <section class="rounded-2xl border border-white/10 bg-nq-card p-5">
      <h2 class="mb-2 font-bold">About</h2>
      <p class="whitespace-pre-line text-white/80">{{ app.description }}</p>
      <p class="mt-4 text-xs text-white/40">Domain: {{ app.domain }}</p>
    </section>

    <section v-if="app.screenshots.length" class="space-y-2">
      <h2 class="font-bold">Screenshots</h2>
      <div class="flex gap-3 overflow-x-auto pb-2">
        <img v-for="(shot, i) in app.screenshots" :key="i" :src="shot" class="h-64 rounded-xl object-cover" />
      </div>
    </section>
  </div>
  <p v-else class="py-10 text-center text-white/50">Loading…</p>
</template>
