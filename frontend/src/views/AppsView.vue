<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listApps, listCategories, type App, type Category } from '../api'
import AppCard from '../components/AppCard.vue'

const route = useRoute()
const router = useRouter()
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

function setCategory(name: string) {
  category.value = name
}

let timer: ReturnType<typeof setTimeout>
watch(q, () => { clearTimeout(timer); timer = setTimeout(load, 250) })
watch([category, sort], load)
watch(category, (name) => {
  const queryCategory = (route.query.category as string) || ''
  if (queryCategory !== name) {
    router.replace({ query: { ...route.query, category: name || undefined } })
  }
})
watch(() => route.query.category, (value) => {
  const next = (value as string) || ''
  if (category.value !== next) category.value = next
})

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
        class="flex-1 rounded-xl border border-line bg-surface px-4 py-2.5 outline-none transition-colors duration-200 placeholder:text-muted focus:border-accent" />
      <div class="flex gap-2">
        <select v-model="category" class="flex-1 cursor-pointer rounded-xl border border-line bg-surface px-3 py-2.5">
          <option value="">All categories</option>
          <option v-for="c in categories" :key="c.name" :value="c.name">{{ c.name }} ({{ c.count }})</option>
        </select>
        <select v-model="sort" class="cursor-pointer rounded-xl border border-line bg-surface px-3 py-2.5">
          <option value="featured">Featured</option>
          <option value="newest">Newest</option>
          <option value="name">Name</option>
        </select>
      </div>
    </div>

    <div v-if="categories.length" class="flex flex-wrap gap-2">
      <button type="button" @click="setCategory('')"
        class="cursor-pointer rounded-full border px-3 py-1.5 text-xs font-bold transition duration-200"
        :class="category === '' ? 'border-accent bg-accent/10 text-accent-ink' : 'border-line bg-surface-2 text-muted hover:border-accent/50'">
        All
      </button>
      <button v-for="c in categories" :key="c.name" type="button" @click="setCategory(c.name)"
        class="cursor-pointer rounded-full border px-3 py-1.5 text-xs font-bold transition duration-200"
        :class="category === c.name ? 'border-accent bg-accent/10 text-accent-ink' : 'border-line bg-surface-2 text-muted hover:border-accent/50'">
        {{ c.name }} <span class="opacity-70">{{ c.count }}</span>
      </button>
    </div>

    <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>

    <div v-else-if="loading" class="grid gap-4 sm:grid-cols-2" aria-hidden="true">
      <div v-for="i in 4" :key="i" class="h-40 animate-pulse rounded-2xl border border-line bg-surface"></div>
    </div>

    <p v-else-if="!apps.length" class="py-10 text-center text-muted">No apps found.</p>

    <div class="grid gap-4 sm:grid-cols-2">
      <AppCard v-for="app in apps" :key="app.id" :app="app" />
    </div>
  </div>
</template>
