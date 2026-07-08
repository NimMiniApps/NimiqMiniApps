<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listCategories, type Category } from '../api'

const categories = ref<Category[]>([])
const error = ref('')

onMounted(async () => {
  try {
    categories.value = await listCategories()
  } catch (e) {
    error.value = (e as Error).message
  }
})
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-2xl font-extrabold">Categories</h1>
    <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-200">{{ error }}</p>
    <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3">
      <RouterLink v-for="c in categories" :key="c.name" :to="{ path: '/apps', query: { category: c.name } }"
        class="flex items-center justify-between rounded-2xl border border-white/10 bg-nq-card p-5 hover:border-nq-gold/50">
        <span class="font-bold">{{ c.name }}</span>
        <span class="rounded-full bg-nq-gold/15 px-2.5 py-0.5 text-sm font-semibold text-nq-gold">{{ c.count }}</span>
      </RouterLink>
    </div>
  </div>
</template>
