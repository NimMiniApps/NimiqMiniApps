<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '../composables/useI18n'

const props = defineProps<{
  title: string
  url?: string
}>()

const { t } = useI18n()
const copied = ref(false)
const sharing = ref(false)

const shareUrl = computed(() => props.url ?? (typeof window !== 'undefined' ? window.location.href : ''))
const useNativeShare = typeof navigator !== 'undefined' && typeof navigator.share === 'function'

async function share() {
  if (useNativeShare) {
    sharing.value = true
    try {
      await navigator.share({ title: props.title, url: shareUrl.value })
    } catch (e) {
      if ((e as Error).name !== 'AbortError') {
        await copyUrl()
      }
    } finally {
      sharing.value = false
    }
    return
  }
  await copyUrl()
}

async function copyUrl() {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    /* clipboard may be unavailable */
  }
}
</script>

<template>
  <button
    type="button"
    :disabled="sharing"
    :aria-label="copied ? t('common.copied') : useNativeShare ? t('common.share') : t('common.copyLink')"
    :title="copied ? t('common.copied') : useNativeShare ? t('common.share') : t('common.copyLink')"
    class="grid h-10 w-10 shrink-0 cursor-pointer place-items-center rounded-xl border border-line bg-surface text-muted transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink disabled:opacity-60"
    @click="share"
  >
    <svg v-if="copied" viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M20 6L9 17l-5-5" />
    </svg>
    <svg v-else viewBox="0 0 24 24" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <path d="M18 8a3 3 0 100-6 3 3 0 000 6zM6 15a3 3 0 100-6 3 3 0 000 6zM18 22a3 3 0 100-6 3 3 0 000 6zM8.59 13.51l6.83 3.98M15.41 6.51L8.59 10.49" />
    </svg>
  </button>
</template>
