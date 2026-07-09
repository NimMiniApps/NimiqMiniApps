<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '../composables/useI18n'

const props = defineProps<{
  reachable: boolean | null | undefined
  variant?: 'badge' | 'banner'
  showOnline?: boolean
  compact?: boolean
}>()

const { t } = useI18n()

const showBadge = computed(() => {
  if (props.reachable === false) return true
  if (props.reachable === true && props.showOnline) return true
  return false
})

const isOnline = computed(() => props.reachable === true)
</script>

<template>
  <p
    v-if="variant === 'banner' && reachable === false"
    class="rounded-xl border border-line bg-surface-2/80 px-4 py-2.5 text-sm text-muted"
  >
    <span class="mr-1.5 inline-block h-1.5 w-1.5 translate-y-[-1px] rounded-full bg-amber-500/80" aria-hidden="true" />
    {{ t('appDetail.offlineBanner') }}
  </p>

  <span
    v-else-if="showBadge"
    class="inline-flex shrink-0 items-center gap-1 text-xs text-muted"
    :title="isOnline ? t('appDetail.onlineHint') : t('appDetail.offlineHint')"
  >
    <span
      class="rounded-full"
      :class="compact ? 'h-2 w-2' : 'h-1.5 w-1.5 shrink-0'"
      :style="{ backgroundColor: isOnline ? 'rgb(16 185 129 / 0.75)' : 'rgb(245 158 11 / 0.85)' }"
      :aria-label="isOnline ? t('appDetail.online') : t('appDetail.offline')"
      role="img"
    />
    <span v-if="!compact" :class="isOnline ? 'text-muted' : 'text-amber-800/90 dark:text-amber-200/90'">
      {{ isOnline ? t('appDetail.online') : t('appDetail.offline') }}
    </span>
  </span>
</template>
