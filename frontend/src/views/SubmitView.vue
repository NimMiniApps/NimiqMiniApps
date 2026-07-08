<script setup lang="ts">
import { reactive, ref } from 'vue'
import { submitApp } from '../api'

const form = reactive({
  name: '', slug: '', domain: '', category: '', developer_slug: '', developer_name: '',
  tagline: '', description: '', tags: '', assets: 'NIM', website_url: '', github_url: '',
})
const error = ref('')
const submitted = ref(false)
const slugTouched = ref(false)

const slugify = (s: string) =>
  s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')

function onNameInput() {
  if (!slugTouched.value) form.slug = slugify(form.name)
}

const csv = (s: string) => s.split(',').map((x) => x.trim()).filter(Boolean)

async function submit() {
  error.value = ''
  try {
    await submitApp({
      ...form,
      slug: slugify(form.slug),
      developer_slug: slugify(form.developer_slug || form.developer_name),
      tags: csv(form.tags),
      assets: csv(form.assets),
      website_url: form.website_url || null,
      github_url: form.github_url || null,
    })
    submitted.value = true
  } catch (e) {
    error.value = (e as Error).message
  }
}

const fields: [keyof typeof form, string, boolean, string][] = [
  ['name', 'App name', true, 'My Mini App'],
  ['slug', 'Slug', true, 'my-mini-app'],
  ['domain', 'Domain (no https://)', true, 'myapp.example.com'],
  ['category', 'Category', true, 'Games, Utilities, …'],
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
        <p class="mt-1 text-white/70">
          Get your Nimiq Pay mini app listed in the directory. Submissions are reviewed
          before they appear publicly.
        </p>
      </div>

      <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-200">{{ error }}</p>

      <form @submit.prevent="submit" class="space-y-3 rounded-2xl border border-white/10 bg-nq-card p-5">
        <div class="grid gap-3 sm:grid-cols-2">
          <label v-for="[key, label, required, placeholder] in fields" :key="key" class="text-sm">
            <span class="mb-1 block text-white/60">{{ label }}{{ required ? ' *' : '' }}</span>
            <input v-model="form[key]" :required="required" :placeholder="placeholder"
              @input="key === 'name' ? onNameInput() : key === 'slug' ? (slugTouched = true) : null"
              class="w-full rounded-lg border border-white/15 bg-nq-blue-dark px-3 py-2 placeholder:text-white/30 focus:border-nq-gold outline-none" />
          </label>
        </div>
        <label class="block text-sm">
          <span class="mb-1 block text-white/60">Description</span>
          <textarea v-model="form.description" rows="4" placeholder="What does your app do?"
            class="w-full rounded-lg border border-white/15 bg-nq-blue-dark px-3 py-2 placeholder:text-white/30 focus:border-nq-gold outline-none"></textarea>
        </label>
        <button type="submit" class="w-full rounded-xl bg-nq-gold px-5 py-3 font-bold text-nq-blue-darker hover:bg-nq-gold-dark sm:w-auto">
          Submit for review
        </button>
      </form>
    </template>

    <div v-else class="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-8 text-center">
      <p class="text-4xl">🎉</p>
      <h1 class="mt-2 text-xl font-extrabold">Thanks, {{ form.name }} is submitted!</h1>
      <p class="mt-1 text-white/70">It will appear in the directory once it's reviewed and approved.</p>
      <RouterLink to="/apps" class="mt-5 inline-block rounded-xl bg-nq-gold px-5 py-2.5 font-bold text-nq-blue-darker hover:bg-nq-gold-dark">
        Browse apps
      </RouterLink>
    </div>
  </div>
</template>
