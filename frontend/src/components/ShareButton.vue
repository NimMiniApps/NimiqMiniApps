<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from '../composables/useI18n'
import { isMobileDevice } from '../utils/device'

const props = defineProps<{
  title: string
  url?: string
}>()

const { t } = useI18n()
const copied = ref(false)
const sharing = ref(false)

const shareUrl = computed(() => props.url ?? (typeof window !== 'undefined' ? window.location.href : ''))
const canNativeShare = typeof navigator !== 'undefined' && typeof navigator.share === 'function'
const useNativeShare = computed(() => isMobileDevice() && canNativeShare)

async function share() {
  if (useNativeShare.value) {
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
    class="inline-flex h-10 shrink-0 cursor-pointer items-center rounded-xl border border-line bg-surface px-4 text-sm font-semibold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink disabled:opacity-60"
    @click="share"
  >
    {{ copied ? t('common.copied') : useNativeShare ? t('common.share') : t('common.copyLink') }}
  </button>
</template>
