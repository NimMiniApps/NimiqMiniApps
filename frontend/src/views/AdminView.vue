<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import {
  adminListApps, adminCreateApp, adminUpdateApp, adminDeleteApp, adminSetStatus, type App,
} from '../api'
import StatusBadge from '../components/StatusBadge.vue'

const token = ref(localStorage.getItem('admin_token') || '')
const apps = ref<App[]>([])
const error = ref('')
const notice = ref('')

const emptyForm = {
  slug: '', name: '', domain: '', category: '', developer_slug: '', developer_name: '',
  tagline: '', description: '', tags: '', assets: 'NIM', status: 'submitted', featured: false,
  website_url: '', github_url: '', icon_url: '', banner_url: '',
}
const form = reactive({ ...emptyForm })
const editingSlug = ref('') // '' = create mode
const showForm = ref(false)

function saveToken() {
  localStorage.setItem('admin_token', token.value)
  notice.value = 'Token saved.'
  load()
}

async function load() {
  error.value = ''
  try {
    apps.value = await adminListApps()
  } catch (e) {
    error.value = (e as Error).message
  }
}

function startCreate() {
  Object.assign(form, emptyForm)
  editingSlug.value = ''
  showForm.value = true
}

function startEdit(app: App) {
  Object.assign(form, {
    slug: app.slug, name: app.name, domain: app.domain, category: app.category,
    developer_slug: app.developer_slug, developer_name: app.developer_name,
    tagline: app.tagline, description: app.description,
    tags: app.tags.join(', '), assets: app.assets.join(', '),
    status: app.status, featured: app.featured,
    website_url: app.website_url || '', github_url: app.github_url || '',
    icon_url: app.icon_url || '', banner_url: app.banner_url || '',
  })
  editingSlug.value = app.slug
  showForm.value = true
}

const csv = (s: string) => s.split(',').map((x) => x.trim()).filter(Boolean)

async function submit() {
  error.value = ''
  const payload = {
    ...form,
    tags: csv(form.tags),
    assets: csv(form.assets),
    website_url: form.website_url || null,
    github_url: form.github_url || null,
    icon_url: form.icon_url || null,
    banner_url: form.banner_url || null,
  }
  try {
    if (editingSlug.value) {
      await adminUpdateApp(editingSlug.value, payload)
      notice.value = `Updated ${form.name}.`
    } else {
      await adminCreateApp(payload)
      notice.value = `Created ${form.name}.`
    }
    showForm.value = false
    load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

async function remove(app: App) {
  if (!confirm(`Delete ${app.name}? This cannot be undone.`)) return
  try {
    await adminDeleteApp(app.slug)
    notice.value = `Deleted ${app.name}.`
    load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

async function setStatus(app: App, action: 'verify' | 'approve' | 'reject') {
  try {
    await adminSetStatus(app.slug, action)
    load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

onMounted(load)

const fields: [keyof typeof emptyForm, string, boolean][] = [
  ['slug', 'Slug (lowercase, url-safe)', true],
  ['name', 'Name', true],
  ['domain', 'Domain (no https://)', true],
  ['category', 'Category', true],
  ['developer_slug', 'Developer slug', true],
  ['developer_name', 'Developer name', true],
  ['tagline', 'Tagline', true],
  ['tags', 'Tags (comma-separated)', false],
  ['assets', 'Assets (NIM, USDT, BTC, ETH)', false],
  ['website_url', 'Website URL', false],
  ['github_url', 'GitHub URL', false],
  ['icon_url', 'Icon URL', false],
  ['banner_url', 'Banner URL', false],
]
</script>

<template>
  <div class="space-y-5">
    <h1 class="text-2xl font-extrabold">Admin</h1>

    <!-- token -->
    <div class="flex gap-2">
      <input v-model="token" type="password" placeholder="Admin token"
        class="flex-1 rounded-xl border border-white/15 bg-nq-card px-4 py-2.5 placeholder:text-white/40 focus:border-nq-gold outline-none" />
      <button @click="saveToken" class="rounded-xl bg-nq-gold px-4 py-2.5 font-bold text-nq-blue-darker hover:bg-nq-gold-dark">
        Save
      </button>
    </div>

    <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-200">{{ error }}</p>
    <p v-if="notice" class="rounded-xl bg-emerald-500/15 p-4 text-emerald-200">{{ notice }}</p>

    <button v-if="!showForm" @click="startCreate"
      class="rounded-xl border border-nq-gold/60 px-4 py-2.5 font-bold text-nq-gold hover:bg-nq-gold/10">
      + New app
    </button>

    <!-- create / edit form -->
    <form v-if="showForm" @submit.prevent="submit" class="space-y-3 rounded-2xl border border-white/10 bg-nq-card p-5">
      <h2 class="font-bold">{{ editingSlug ? `Edit ${editingSlug}` : 'New app' }}</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <label v-for="[key, label, required] in fields" :key="key" class="text-sm">
          <span class="mb-1 block text-white/60">{{ label }}{{ required ? ' *' : '' }}</span>
          <input v-model="(form as any)[key]" :required="required"
            class="w-full rounded-lg border border-white/15 bg-nq-blue-dark px-3 py-2 focus:border-nq-gold outline-none" />
        </label>
        <label class="text-sm">
          <span class="mb-1 block text-white/60">Status</span>
          <select v-model="form.status" class="w-full rounded-lg border border-white/15 bg-nq-blue-dark px-3 py-2">
            <option v-for="s in ['submitted', 'approved', 'verified', 'experimental', 'rejected']" :key="s" :value="s">{{ s }}</option>
          </select>
        </label>
        <label class="flex items-end gap-2 pb-2 text-sm">
          <input v-model="form.featured" type="checkbox" class="h-4 w-4 accent-[#e9b213]" />
          Featured
        </label>
      </div>
      <label class="block text-sm">
        <span class="mb-1 block text-white/60">Description</span>
        <textarea v-model="form.description" rows="3"
          class="w-full rounded-lg border border-white/15 bg-nq-blue-dark px-3 py-2 focus:border-nq-gold outline-none"></textarea>
      </label>
      <div class="flex gap-2">
        <button type="submit" class="rounded-xl bg-nq-gold px-5 py-2 font-bold text-nq-blue-darker hover:bg-nq-gold-dark">
          {{ editingSlug ? 'Save changes' : 'Create app' }}
        </button>
        <button type="button" @click="showForm = false" class="rounded-xl border border-white/20 px-5 py-2 font-semibold hover:bg-white/10">
          Cancel
        </button>
      </div>
    </form>

    <!-- app list -->
    <div class="space-y-2">
      <div v-for="app in apps" :key="app.id"
        class="flex flex-col gap-3 rounded-2xl border border-white/10 bg-nq-card p-4 sm:flex-row sm:items-center">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="font-bold">{{ app.name }}</span>
            <StatusBadge :status="app.status" />
            <span v-if="app.featured" class="text-nq-gold" title="Featured">★</span>
          </div>
          <p class="truncate text-sm text-white/60">{{ app.slug }} · {{ app.domain }}</p>
        </div>
        <div class="flex flex-wrap gap-1.5 text-xs font-semibold">
          <button @click="setStatus(app, 'approve')" class="rounded-lg bg-sky-500/20 px-2.5 py-1.5 text-sky-300 hover:bg-sky-500/30">Approve</button>
          <button @click="setStatus(app, 'verify')" class="rounded-lg bg-emerald-500/20 px-2.5 py-1.5 text-emerald-300 hover:bg-emerald-500/30">Verify</button>
          <button @click="setStatus(app, 'reject')" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-300 hover:bg-red-500/30">Reject</button>
          <button @click="startEdit(app)" class="rounded-lg bg-white/10 px-2.5 py-1.5 hover:bg-white/20">Edit</button>
          <button @click="remove(app)" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-300 hover:bg-red-500/30">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>
