<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getSubmissionStatus, type SubmissionStatus } from '../api'
import StatusBadge from '../components/StatusBadge.vue'
import { CATALOG_ISSUES_URL } from '../utils/catalogLinks'

const route = useRoute()
const status = ref<SubmissionStatus | null>(null)
const error = ref('')
const loading = ref(true)

const statusMessages: Record<string, { title: string; body: string }> = {
  pending: {
    title: 'Awaiting review',
    body: 'Your submission is in the queue. Moderators will review it before it appears in the public directory.',
  },
  live: {
    title: 'Live in the directory',
    body: 'This app is approved and visible to everyone browsing Nimiq Mini Apps.',
  },
  rejected: {
    title: 'Not listed',
    body: 'This submission was not accepted. You can submit an updated version with a new slug, or open a GitHub issue if you have questions.',
  },
}

async function load(slug: string) {
  loading.value = true
  error.value = ''
  status.value = null
  try {
    status.value = await getSubmissionStatus(slug)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

watch(() => route.params.slug, (slug) => {
  if (typeof slug === 'string') load(slug)
})

onMounted(() => {
  const slug = route.params.slug as string
  if (slug) load(slug)
})
</script>

<template>
  <div class="mx-auto max-w-lg space-y-5">
    <div>
      <h1 class="text-2xl font-extrabold">Submission status</h1>
      <p class="mt-1 text-sm text-muted">Check whether your app has been reviewed.</p>
    </div>

    <p v-if="error" class="rounded-xl bg-red-500/15 p-4 text-red-600 dark:text-red-300">{{ error }}</p>

    <div v-else-if="loading" class="h-32 animate-pulse rounded-2xl border border-line bg-surface" aria-hidden="true"></div>

    <div v-else-if="status" class="space-y-4 rounded-2xl border border-line bg-surface p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-xl font-bold">{{ status.name }}</h2>
        <StatusBadge :status="status.raw_status" />
      </div>
      <p class="font-mono text-sm text-muted">/{{ status.slug }}</p>

      <div class="rounded-xl border border-line bg-surface-2 p-4">
        <p class="font-semibold text-ink">{{ statusMessages[status.status]?.title ?? status.status }}</p>
        <p class="mt-1 text-sm text-muted">{{ statusMessages[status.status]?.body }}</p>
      </div>

      <div v-if="status.update_pending" class="rounded-xl border border-sky-500/30 bg-sky-500/10 p-4 text-sm">
        <p class="font-semibold text-ink">Update pending review</p>
        <p class="mt-1 text-muted">A change request is in the queue. The public listing will update once moderators approve it.</p>
      </div>

      <p class="text-xs text-muted">Last updated {{ new Date(status.updated_at).toLocaleString() }}</p>

      <p class="text-sm text-muted">
        Questions about your submission?
        <a :href="CATALOG_ISSUES_URL" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">Open a GitHub issue</a>
      </p>

      <div class="flex flex-wrap gap-2">
        <RouterLink
          v-if="status.public"
          :to="`/apps/${status.slug}`"
          class="cursor-pointer rounded-[500px] nq-primary px-5 py-2.5 font-bold text-white transition duration-200"
        >
          View listing
        </RouterLink>
        <RouterLink to="/submit" class="cursor-pointer rounded-xl border border-line px-5 py-2.5 font-semibold hover:border-accent/50 hover:text-accent-ink">
          Submit another app
        </RouterLink>
      </div>
    </div>
  </div>
</template>
