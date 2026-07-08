<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { listApps, listCategories, type App, type Category } from '../api'
import AppCard from '../components/AppCard.vue'

const route = useRoute()
const apps = ref<App[]>([])
const categories = ref<Category[]>([])
const q = ref('')
const category = ref((route.query.category as string) || '')
const sort = ref('featured')
const error = ref('')
const loading = ref(true)

async function load() {
  loading.value = true
  error.value = ''
  try {
    apps.value = await listApps({ q: q.value, category: category.value, sort: sort.value })
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

let timer: ReturnType<typeof setTimeout>
watch(q, () => { clearTimeout(timer); timer = setTimeout(load, 250) })
watch([category, sort], load)

onMounted(async () => {
  load()
  try { categories.value = await listCategories() } catch { /* filter dropdown just stays empty */ }
})
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-2xl font-extrabold">All Apps</h1>

    <div class="flex flex-col gap-2 sm:flex-row">
      <input v-model="q" type="search" placeholder="Search apps…"
        class="flex-1 rounded-xl border border-white/15 bg-nq-card px-4 py-2.5 outline-none placeholder:text-white/40 focus:border-nq-gold" />
      <div class="flex gap-2">
        <select v-model="category" class="flex-1 rounded-xl border border-white/15 bg-nq-card px-3 py-2.5">
          <option value="">All categories</option>
          <option v-for="c in categories" :key="c.name" :value="c.name">{{ c.name }} ({{ c.count }})</option>
        </select>
        <select v-model="sort" class="rounded-xl border border-white/15 bg-nq-card px-3 py-2.5">
          <option value="featured">Featured</option>
          <option value="newest">Newest</option>
          <option value="name">Name</option>
        </select>
      </div>
    </div>

    <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-200">{{ error }}</p>
    <p v-else-if="loading" class="py-10 text-center text-white/50">Loading…</p>
    <p v-else-if="!apps.length" class="py-10 text-center text-white/50">No apps found.</p>

    <div class="grid gap-4 sm:grid-cols-2">
      <AppCard v-for="app in apps" :key="app.id" :app="app" />
    </div>
  </div>
</template>
