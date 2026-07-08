<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getDeveloper, type Developer } from '../api'
import AppCard from '../components/AppCard.vue'

const route = useRoute()
const dev = ref<Developer | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    dev.value = await getDeveloper(route.params.slug as string)
  } catch (e) {
    error.value = (e as Error).message
  }
})
</script>

<template>
  <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-200">{{ error }}</p>
  <div v-else-if="dev" class="space-y-5">
    <div class="flex items-center gap-4">
      <div class="grid h-14 w-14 place-items-center rounded-full bg-nq-gold text-2xl font-extrabold text-nq-blue-darker">
        {{ dev.name[0] }}
      </div>
      <div>
        <h1 class="text-2xl font-extrabold">{{ dev.name }}</h1>
        <p class="text-sm text-white/60">{{ dev.apps.length }} app{{ dev.apps.length === 1 ? '' : 's' }}</p>
      </div>
    </div>
    <div class="grid gap-4 sm:grid-cols-2">
      <AppCard v-for="app in dev.apps" :key="app.id" :app="app" />
    </div>
  </div>
  <p v-else class="py-10 text-center text-white/50">Loading…</p>
</template>
