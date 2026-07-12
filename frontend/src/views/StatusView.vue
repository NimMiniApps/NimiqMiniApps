<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getSubmissionStatus, type SubmissionStatus } from '../api'
import StatusBadge from '../components/StatusBadge.vue'
import EmptyState from '../components/EmptyState.vue'
import { useI18n } from '../composables/useI18n'
import { CATALOG_ISSUES_URL } from '../utils/catalogLinks'

const route = useRoute()
const { t } = useI18n()
const status = ref<SubmissionStatus | null>(null)
const error = ref('')
const notFound = ref(false)
const loading = ref(true)

async function load(slug: string) {
  loading.value = true
  error.value = ''
  notFound.value = false
  status.value = null
  try {
    status.value = await getSubmissionStatus(slug)
  } catch (e) {
    const message = (e as Error).message
    notFound.value = message.toLowerCase().includes('not found')
    error.value = message
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
      <h1 class="text-2xl font-extrabold">{{ t('status.title') }}</h1>
      <p class="mt-1 text-sm text-muted">{{ t('status.subtitle') }}</p>
    </div>

    <EmptyState
      v-if="error && notFound"
      :title="t('status.notFoundTitle')"
      :description="t('status.notFoundBody')"
      variant="notFound"
    >
      <template #actions>
        <RouterLink
          to="/submit"
          class="cursor-pointer rounded-[500px] nq-primary px-5 py-2.5 text-sm font-bold text-white transition duration-200"
        >
          {{ t('status.submitAnother') }}
        </RouterLink>
      </template>
    </EmptyState>

    <EmptyState
      v-else-if="error"
      :title="t('status.errorTitle')"
      :description="t('status.errorBody')"
      variant="error"
    >
      <template #actions>
        <button
          type="button"
          class="cursor-pointer rounded-xl border border-line bg-surface px-5 py-2.5 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
          @click="load(route.params.slug as string)"
        >
          {{ t('common.retry') }}
        </button>
      </template>
    </EmptyState>

    <div v-else-if="loading" class="h-32 animate-pulse rounded-2xl border border-line bg-surface" aria-hidden="true"></div>

    <div v-else-if="status" class="space-y-4 rounded-2xl border border-line bg-surface p-6 shadow-sm">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-xl font-bold">{{ status.name }}</h2>
        <StatusBadge :status="status.raw_status" />
      </div>
      <p class="font-mono text-sm text-muted">/{{ status.slug }}</p>

      <div class="rounded-xl border border-line bg-surface-2 p-4">
        <p class="font-semibold text-ink">
          {{
            status.status === 'pending' ? t('status.pendingTitle')
            : status.status === 'live' ? t('status.liveTitle')
            : status.status === 'rejected' ? t('status.rejectedTitle')
            : status.status
          }}
        </p>
        <p class="mt-1 text-sm text-muted">
          {{
            status.status === 'pending' ? t('status.pendingBody')
            : status.status === 'live' ? t('status.liveBody')
            : status.status === 'rejected' ? t('status.rejectedBody')
            : ''
          }}
        </p>
      </div>

      <div
        v-if="status.rejection_note"
        class="rounded-xl border border-red-500/30 bg-red-500/10 p-4 text-sm"
      >
        <p class="font-semibold text-ink">{{ t('status.rejectionNoteLabel') }}</p>
        <p class="mt-1 whitespace-pre-wrap text-muted">{{ status.rejection_note }}</p>
      </div>

      <div v-if="status.update_pending" class="rounded-xl border border-sky-500/30 bg-sky-500/10 p-4 text-sm">
        <p class="font-semibold text-ink">{{ t('status.updatePendingTitle') }}</p>
        <p class="mt-1 text-muted">{{ t('status.updatePendingBody') }}</p>
      </div>

      <p class="text-xs text-muted">{{ t('status.lastUpdated') }} {{ new Date(status.updated_at).toLocaleString() }}</p>

      <p class="text-sm text-muted">
        {{ t('status.questions') }}
        <a :href="CATALOG_ISSUES_URL" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">{{ t('status.githubIssue') }}</a>
      </p>

      <div class="flex flex-wrap gap-2">
        <RouterLink
          v-if="status.public"
          :to="`/apps/${status.slug}`"
          class="cursor-pointer rounded-[500px] nq-primary px-5 py-2.5 font-bold text-white transition duration-200"
        >
          {{ t('status.viewListing') }}
        </RouterLink>
        <RouterLink to="/submit" class="cursor-pointer rounded-xl border border-line px-5 py-2.5 font-semibold hover:border-accent/50 hover:text-accent-ink">
          {{ t('status.submitAnother') }}
        </RouterLink>
      </div>
    </div>
  </div>
</template>
