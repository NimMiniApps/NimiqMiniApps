<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listApps, type App } from '../api'
import AppCard from '../components/AppCard.vue'

const featured = ref<App[]>([])
const newest = ref<App[]>([])
const error = ref('')

onMounted(async () => {
  try {
    ;[featured.value, newest.value] = await Promise.all([
      listApps({ featured: 'true' }),
      listApps({ sort: 'newest' }),
    ])
  } catch (e) {
    error.value = (e as Error).message
  }
})
</script>

<template>
  <div class="space-y-8">
    <section class="rounded-3xl bg-gradient-to-br from-nq-blue to-nq-blue-dark border border-white/10 p-6 md:p-10">
      <h1 class="text-2xl md:text-4xl font-extrabold">
        Discover <span class="text-nq-gold">Nimiq Pay</span> Mini Apps
      </h1>
      <p class="mt-2 max-w-xl text-white/70">
        A community directory of apps you can open straight from your Nimiq Pay wallet.
      </p>
      <div class="mt-5 flex flex-wrap gap-2">
        <RouterLink to="/apps"
          class="rounded-xl bg-nq-gold px-5 py-2.5 font-bold text-nq-blue-darker hover:bg-nq-gold-dark">
          Browse all apps
        </RouterLink>
        <RouterLink to="/submit"
          class="rounded-xl border border-nq-gold/60 px-5 py-2.5 font-bold text-nq-gold hover:bg-nq-gold/10">
          Submit your app
        </RouterLink>
      </div>
    </section>

    <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-200">{{ error }}</p>

    <section v-if="featured.length">
      <h2 class="mb-3 text-lg font-bold">⭐ Featured</h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in featured" :key="app.id" :app="app" />
      </div>
    </section>

    <section v-if="newest.length">
      <h2 class="mb-3 text-lg font-bold">🆕 Newest</h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <AppCard v-for="app in newest.slice(0, 6)" :key="app.id" :app="app" />
      </div>
    </section>
  </div>
</template>
