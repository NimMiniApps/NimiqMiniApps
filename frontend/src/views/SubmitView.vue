<script setup lang="ts">
import { reactive, ref } from 'vue'
import { APP_CATEGORIES, APP_RELEASE_STAGES, submitApp, type MediaItem, type SocialLink } from '../api'
import SocialLinksEditor from '../components/SocialLinksEditor.vue'
import MediaEditor from '../components/MediaEditor.vue'
import WalletLoginButton from '../components/WalletLoginButton.vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { CATALOG_ISSUES_URL } from '../utils/catalogLinks'

const { walletAddress, displayName, checking } = useWalletAuth()

const socialEditor = ref<InstanceType<typeof SocialLinksEditor>>()
const mediaEditor = ref<InstanceType<typeof MediaEditor>>()
const socials = ref<SocialLink[]>([])
const media = ref<MediaItem[]>([])

const form = reactive({
  name: '', slug: '', domain: '', category: '',
  tagline: '', description: '', long_description: '', release_stage: 'beta', tags: '', assets: 'NIM',
  icon_url: '', banner_url: '', website_url: '', github_url: '',
  submitter_contact: '',
})
const error = ref('')
const submitted = ref(false)
const submittedSlug = ref('')
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
      tags: csv(form.tags),
      assets: csv(form.assets),
      media: mediaEditor.value?.validate() ?? [],
      socials: socialEditor.value?.validate() ?? [],
      icon_url: form.icon_url || null,
      banner_url: form.banner_url || null,
      website_url: form.website_url || null,
      github_url: form.github_url || null,
    })
    submittedSlug.value = slugify(form.slug)
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
  ['submitter_contact', 'Contact (Telegram, email, etc.)', true, '@yourhandle or you@example.com'],
  ['tagline', 'Tagline', true, 'One sentence about your app'],
  ['tags', 'Tags (comma-separated)', false, 'games, multiplayer'],
  ['assets', 'Assets (NIM, USDT, USDC, BTC, ETH)', false, 'NIM'],
  ['icon_url', 'Icon URL', false, 'https://…'],
  ['banner_url', 'Banner URL', false, 'https://…'],
  ['website_url', 'Website URL', false, 'https://…'],
  ['github_url', 'GitHub URL', false, 'https://github.com/…'],
]
</script>

<template>
  <div class="mx-auto max-w-2xl space-y-5">
    <template v-if="!submitted">
      <div>
        <h1 class="text-2xl font-extrabold">Submit your app</h1>
        <p class="mt-1 text-sm text-muted">
          Get your Nimiq Pay mini app listed in the directory — beta apps welcome. Submissions
          are reviewed before they appear publicly. We need a way to reach you if we have questions.
          You can also
          <a :href="CATALOG_ISSUES_URL" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">open a GitHub issue</a>
          if you need help or want to follow up.
        </p>
      </div>

      <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>

      <div v-if="checking" class="rounded-2xl border border-line bg-surface p-5 text-sm text-muted">
        Checking wallet session…
      </div>
      <div v-else-if="!walletAddress" class="rounded-2xl border border-line bg-surface p-5 text-center">
        <p class="text-sm text-muted">Connect your Nimiq wallet to submit an app. It will be linked to your wallet as the developer of record — admins can reassign it later.</p>
        <WalletLoginButton class="mt-3 inline-block" />
      </div>
      <div v-else-if="!displayName" class="rounded-2xl border border-line bg-surface p-5 text-center">
        <p class="text-sm text-muted">Set a display name on your profile before submitting — it becomes your public developer name.</p>
        <RouterLink to="/profile" class="mt-3 inline-block rounded-xl bg-nq-blue px-5 py-2.5 font-bold text-white">Go to profile</RouterLink>
      </div>
      <template v-else>
        <p class="text-xs text-muted">
          Submitting as <span class="font-mono">{{ walletAddress }}</span> — this app will be linked to your wallet as the developer of record; admins can reassign it later.
        </p>
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
            <span class="mt-1 block text-xs text-muted">Markdown supported: **bold**, lists, [links](https://…), ## headings.</span>
          </label>
          <div class="block text-sm">
            <span class="mb-2 block font-semibold text-muted">Screenshots &amp; video</span>
            <MediaEditor ref="mediaEditor" v-model="media" />
          </div>
          <div class="block text-sm">
            <span class="mb-2 block font-semibold text-muted">Social links</span>
            <SocialLinksEditor ref="socialEditor" v-model="socials" />
          </div>
          <button type="submit" :disabled="submitting"
            class="w-full cursor-pointer rounded-xl bg-nq-blue px-5 py-3 font-bold text-white transition duration-200 hover:bg-nq-blue-dark disabled:cursor-default disabled:opacity-60 sm:w-auto">
            {{ submitting ? 'Submitting…' : 'Submit for review' }}
          </button>
        </form>
      </template>
    </template>

    <div v-else class="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-8 text-center">
      <svg viewBox="0 0 24 24" class="mx-auto h-12 w-12 fill-none stroke-emerald-500" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="10" />
        <path d="M8 12.5l2.5 2.5L16 9.5" />
      </svg>
      <h1 class="mt-3 text-xl font-extrabold">Thanks, {{ form.name }} is submitted!</h1>
      <p class="mt-1 text-muted">It will appear in the directory once it's reviewed and approved.</p>
      <p class="mt-3 text-sm text-muted">
        Track review status at
        <RouterLink :to="`/status/${submittedSlug}`" class="font-semibold text-accent-ink hover:underline">
          /status/{{ submittedSlug }}
        </RouterLink>
        ·
        <a :href="CATALOG_ISSUES_URL" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">GitHub issues</a>
      </p>
      <div class="mt-5 flex flex-wrap justify-center gap-2">
        <RouterLink :to="`/status/${submittedSlug}`"
          class="inline-block cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 font-bold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
          Check status
        </RouterLink>
        <RouterLink to="/apps" class="inline-block cursor-pointer rounded-xl bg-nq-blue px-5 py-2.5 font-bold text-white transition duration-200 hover:bg-nq-blue-dark">
          Browse apps
        </RouterLink>
      </div>
    </div>
  </div>
</template>
