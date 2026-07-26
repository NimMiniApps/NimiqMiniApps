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
    class="inline-flex shrink-0 items-center rounded-[3px] border border-board-hairline bg-board-flap-hover px-2 py-0.5 text-xs font-semibold uppercase tracking-wide text-accent-ink"
    :title="compact ? t('common.hostedBy') : undefined"
  >
    {{ label }}
  </span>
</template>
