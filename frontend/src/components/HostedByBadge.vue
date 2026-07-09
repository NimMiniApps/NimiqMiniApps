<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '../composables/useI18n'
import { isNimiqMiniAppsHosted } from '../utils/hosting'

const props = withDefaults(defineProps<{
  domain: string
  compact?: boolean
}>(), {
  compact: false,
})

const { t } = useI18n()
const isHosted = computed(() => isNimiqMiniAppsHosted(props.domain))
const label = computed(() => props.compact ? t('common.hostedByShort') : t('common.hostedBy'))
</script>

<template>
  <span
    v-if="isHosted"
    class="inline-flex shrink-0 items-center rounded-full border border-blue-500/20 bg-blue-500/10 px-2 py-0.5 text-xs font-semibold text-blue-700 dark:border-blue-300/20 dark:bg-blue-300/10 dark:text-blue-200"
    :title="compact ? t('common.hostedBy') : undefined"
  >
    {{ label }}
  </span>
</template>
