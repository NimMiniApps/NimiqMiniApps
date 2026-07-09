<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  APP_CATEGORIES, APP_RELEASE_STAGES, getApp, getSubmissionStatus, requestAppUpdate,
  type MediaItem, type SocialLink,
} from '../api'
import SocialLinksEditor from '../components/SocialLinksEditor.vue'
import MediaEditor from '../components/MediaEditor.vue'
import TokenMultiSelect from '../components/TokenMultiSelect.vue'
import { normalizeDomain } from '../utils/domain'

const route = useRoute()
const socialEditor = ref<InstanceType<typeof SocialLinksEditor>>()
const mediaEditor = ref<InstanceType<typeof MediaEditor>>()
const socials = ref<SocialLink[]>([])
const media = ref<MediaItem[]>([])

const slug = ref('')
const error = ref('')
const loadError = ref('')
const loading = ref(true)
const submitted = ref(false)
const submitting = ref(false)
const updatePending = ref(false)

const form = reactive({
  name: '', domain: '', category: '',
  tagline: '', description: '', long_description: '', release_stage: 'released', tags: '', assets: 'NIM', reward_assets: '',
  icon_url: '', banner_url: '', website_url: '', github_url: '',
  author_note: '',
})

const csv = (s: string) => s.split(',').map((x) => x.trim()).filter(Boolean)

async function load(slugParam: string) {
  loading.value = true
  loadError.value = ''
  error.value = ''
  try {
    const [app, status] = await Promise.all([
      getApp(slugParam),
      getSubmissionStatus(slugParam).catch(() => null),
    ])
    slug.value = app.slug
    updatePending.value = !!status?.update_pending
    Object.assign(form, {
      name: app.name,
      domain: app.domain,
      category: app.category,
      tagline: app.tagline,
      description: app.description,
      long_description: app.long_description || '',
      release_stage: app.release_stage,
      tags: app.tags.join(', '),
      assets: app.assets.join(', '),
      reward_assets: app.reward_assets.join(', '),
      icon_url: app.icon_url || '',
      banner_url: app.banner_url || '',
      website_url: app.website_url || '',
      github_url: app.github_url || '',
      author_note: '',
    })
    socials.value = app.socials ?? []
    media.value = app.media ?? []
  } catch (e) {
    loadError.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!slug.value) return
  error.value = ''
  submitting.value = true
  try {
    await requestAppUpdate(slug.value, {
      name: form.name,
      domain: normalizeDomain(form.domain),
      category: form.category,
      tagline: form.tagline,
      description: form.description,
      long_description: form.long_description,
      release_stage: form.release_stage,
      tags: csv(form.tags),
      assets: csv(form.assets),
      reward_assets: csv(form.reward_assets),
      media: mediaEditor.value?.validate() ?? [],
      socials: socialEditor.value?.validate() ?? [],
      icon_url: form.icon_url || null,
      banner_url: form.banner_url || null,
      website_url: form.website_url || null,
      github_url: form.github_url || null,
      author_note: form.author_note,
    })
    submitted.value = true
    updatePending.value = true
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  const s = route.params.slug
  if (typeof s === 'string' && s) load(s)
})

const fields: [keyof typeof form, string, boolean, string?][] = [
  ['name', 'App name', true],
  ['domain', 'Domain (no https://)', true],
  ['tagline', 'Tagline', true],
  ['tags', 'Tags (comma-separated)', false],
  ['icon_url', 'Icon URL', false],
  ['banner_url', 'Banner URL', false],
  ['website_url', 'Website URL', false],
  ['github_url', 'GitHub URL', false],
]
</script>

<template>
  <div class="mx-auto max-w-2xl space-y-5">
    <p v-if="loading" class="text-muted">Loading…</p>
    <p v-else-if="loadError" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ loadError }}</p>

    <template v-else-if="submitted">
      <div class="rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-8 text-center">
        <h1 class="text-xl font-extrabold">Update request submitted</h1>
        <p class="mt-2 text-muted">
          Your proposed changes for <strong>{{ form.name }}</strong> are queued for review.
          The live listing stays unchanged until a moderator approves them.
        </p>
        <div class="mt-5 flex flex-wrap justify-center gap-2">
          <RouterLink :to="`/apps/${slug}`"
            class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 font-bold hover:border-accent/50">
            Back to app
          </RouterLink>
          <RouterLink :to="`/status/${slug}`"
            class="cursor-pointer rounded-[500px] nq-primary px-5 py-2.5 font-bold text-white">
            Check status
          </RouterLink>
        </div>
      </div>
    </template>

    <template v-else>
      <div>
        <RouterLink :to="`/apps/${slug}`" class="text-sm font-semibold text-accent-ink hover:underline">← Back to {{ form.name }}</RouterLink>
        <h1 class="mt-2 text-2xl font-extrabold">Request an update</h1>
        <p class="mt-1 text-muted">
          Propose changes to your listing. Moderators will review your request before anything goes live.
          Slug: <code class="rounded bg-surface-2 px-1.5 py-0.5 text-sm">{{ slug }}</code>
        </p>
      </div>

      <p v-if="updatePending" class="rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-sm text-amber-900 dark:text-amber-100">
        An update is already pending review for this app. Submitting again will fail until it is approved or rejected.
      </p>

      <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>

      <form @submit.prevent="submit" class="space-y-3 rounded-2xl border border-line bg-surface p-5 shadow-sm">
        <div class="grid gap-3 sm:grid-cols-2">
          <label v-for="[key, label, required, help] in fields" :key="key" class="text-sm">
            <span class="mb-1 block font-semibold text-muted">{{ label }}{{ required ? ' *' : '' }}</span>
            <input v-model="form[key]" :required="required"
              class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none focus:border-accent" />
            <span v-if="help" class="mt-1 block text-xs leading-snug text-muted">{{ help }}</span>
          </label>
          <TokenMultiSelect
            v-model="form.assets"
            label="Assets"
            help="Tokens your app uses, accepts, reads, or supports."
          />
          <TokenMultiSelect
            v-model="form.reward_assets"
            label="Reward assets"
            help="Only select tokens users can actually receive from your app, such as daily rewards, leaderboard prizes, payouts, or tips. Leave empty if the app only uses or accepts the token."
          />
          <label class="text-sm">
            <span class="mb-1 block font-semibold text-muted">Category *</span>
            <select v-model="form.category" required class="w-full cursor-pointer rounded-lg border border-line bg-surface-2 px-3 py-2">
              <option v-for="category in APP_CATEGORIES" :key="category" :value="category">{{ category }}</option>
            </select>
          </label>
          <label class="text-sm">
            <span class="mb-1 block font-semibold text-muted">Release stage</span>
            <select v-model="form.release_stage" class="w-full cursor-pointer rounded-lg border border-line bg-surface-2 px-3 py-2">
              <option v-for="stage in APP_RELEASE_STAGES" :key="stage" :value="stage">{{ stage }}</option>
            </select>
          </label>
        </div>
        <label class="block text-sm">
          <span class="mb-1 block font-semibold text-muted">Short description</span>
          <textarea v-model="form.description" rows="3" class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none"></textarea>
        </label>
        <label class="block text-sm">
          <span class="mb-1 block font-semibold text-muted">Full description</span>
          <textarea v-model="form.long_description" rows="6" class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none"></textarea>
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
        <label class="block text-sm">
          <span class="mb-1 block font-semibold text-muted">Note for moderators (optional)</span>
          <textarea v-model="form.author_note" rows="2" placeholder="What changed and why?"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none"></textarea>
        </label>
        <button type="submit" :disabled="submitting || updatePending"
          class="w-full cursor-pointer rounded-[500px] nq-primary px-5 py-3 font-bold text-white disabled:opacity-60 sm:w-auto">
          {{ submitting ? 'Submitting…' : 'Submit update for review' }}
        </button>
      </form>
    </template>
  </div>
</template>
