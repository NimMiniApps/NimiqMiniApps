<script setup lang="ts">
import { reactive, ref } from 'vue'
import { APP_CATEGORIES, APP_RELEASE_STAGES, submitApp } from '../api'
import { parseMediaLines } from '../utils/media'

const form = reactive({
  name: '', slug: '', domain: '', category: '', developer_slug: '', developer_name: '',
  tagline: '', description: '', long_description: '', release_stage: 'beta', tags: '', assets: 'NIM',
  website_url: '', github_url: '', media: '',
})
const error = ref('')
const submitted = ref(false)
const submitting = ref(false)
const slugTouched = ref(false)

const slugify = (s: string) =>
  s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')

function onNameInput() {
  if (!slugTouched.value) form.slug = slugify(form.name)
}

const csv = (s: string) => s.split(',').map((x) => x.trim()).filter(Boolean)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    await submitApp({
      ...form,
      slug: slugify(form.slug),
      developer_slug: slugify(form.developer_slug || form.developer_name),
      tags: csv(form.tags),
      assets: csv(form.assets),
      media: parseMediaLines(form.media),
      website_url: form.website_url || null,
      github_url: form.github_url || null,
    })
    submitted.value = true
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    submitting.value = false
  }
}

const fields: [keyof typeof form, string, boolean, string][] = [
  ['name', 'App name', true, 'My Mini App'],
  ['slug', 'Slug', true, 'my-mini-app'],
  ['domain', 'Domain (no https://)', true, 'myapp.example.com'],
  ['developer_name', 'Developer name', true, 'Your name or team'],
  ['tagline', 'Tagline', true, 'One sentence about your app'],
  ['tags', 'Tags (comma-separated)', false, 'games, multiplayer'],
  ['assets', 'Assets (NIM, USDT, BTC, ETH)', false, 'NIM'],
  ['website_url', 'Website URL', false, 'https://…'],
  ['github_url', 'GitHub URL', false, 'https://github.com/…'],
]
</script>

<template>
  <div class="mx-auto max-w-2xl space-y-5">
    <template v-if="!submitted">
      <div>
        <h1 class="text-2xl font-extrabold">Submit your app</h1>
        <p class="mt-1 text-muted">
          Get your Nimiq Pay mini app listed in the directory — beta apps welcome. Submissions
          are reviewed before they appear publicly.
        </p>
      </div>

      <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>

      <form @submit.prevent="submit" class="space-y-3 rounded-2xl border border-line bg-surface p-5 shadow-sm">
        <div class="grid gap-3 sm:grid-cols-2">
          <label v-for="[key, label, required, placeholder] in fields" :key="key" class="text-sm">
            <span class="mb-1 block font-semibold text-muted">{{ label }}{{ required ? ' *' : '' }}</span>
            <input v-model="form[key]" :required="required" :placeholder="placeholder"
              @input="key === 'name' ? onNameInput() : key === 'slug' ? (slugTouched = true) : null"
              class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 placeholder:text-muted/60 focus:border-accent" />
          </label>
          <label class="text-sm">
            <span class="mb-1 block font-semibold text-muted">Category *</span>
            <select v-model="form.category" required
              class="w-full cursor-pointer rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 focus:border-accent">
              <option value="" disabled>Select a category</option>
              <option v-for="category in APP_CATEGORIES" :key="category" :value="category">{{ category }}</option>
            </select>
          </label>
          <label class="text-sm">
            <span class="mb-1 block font-semibold text-muted">Release stage</span>
            <select v-model="form.release_stage"
              class="w-full cursor-pointer rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 focus:border-accent">
              <option v-for="stage in APP_RELEASE_STAGES" :key="stage" :value="stage">{{ stage }}</option>
            </select>
          </label>
        </div>
        <label class="block text-sm">
          <span class="mb-1 block font-semibold text-muted">Short description</span>
          <textarea v-model="form.description" rows="3" placeholder="Brief summary shown in listings"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 placeholder:text-muted/60 focus:border-accent"></textarea>
        </label>
        <label class="block text-sm">
          <span class="mb-1 block font-semibold text-muted">Full description</span>
          <textarea v-model="form.long_description" rows="6" placeholder="Features, gameplay, how it works — shown on the detail page"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 placeholder:text-muted/60 focus:border-accent"></textarea>
        </label>
        <label class="block text-sm">
          <span class="mb-1 block font-semibold text-muted">Screenshots &amp; video</span>
          <textarea v-model="form.media" rows="4" placeholder="One URL per line — image links or YouTube URLs"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 placeholder:text-muted/60 focus:border-accent"></textarea>
          <span class="mt-1 block text-xs text-muted">Paste screenshot image URLs or YouTube links, one per line.</span>
        </label>
        <button type="submit" :disabled="submitting"
          class="w-full cursor-pointer rounded-xl bg-nq-blue px-5 py-3 font-bold text-white transition duration-200 hover:bg-nq-blue-dark disabled:cursor-default disabled:opacity-60 sm:w-auto">
          {{ submitting ? 'Submitting…' : 'Submit for review' }}
        </button>
      </form>
    </template>

    <div v-else class="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-8 text-center">
      <svg viewBox="0 0 24 24" class="mx-auto h-12 w-12 fill-none stroke-emerald-500" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="10" />
        <path d="M8 12.5l2.5 2.5L16 9.5" />
      </svg>
      <h1 class="mt-3 text-xl font-extrabold">Thanks, {{ form.name }} is submitted!</h1>
      <p class="mt-1 text-muted">It will appear in the directory once it's reviewed and approved.</p>
      <RouterLink to="/apps" class="mt-5 inline-block cursor-pointer rounded-xl bg-nq-blue px-5 py-2.5 font-bold text-white transition duration-200 hover:bg-nq-blue-dark">
        Browse apps
      </RouterLink>
    </div>
  </div>
</template>
