<script setup lang="ts">
import { computed } from 'vue'
import { rewardLabel } from '../utils/rewards'

const props = defineProps<{
  assets: string[]
  compact?: boolean
}>()

const label = computed(() => rewardLabel(props.assets))
</script>

<template>
  <RouterLink
    v-if="label"
    to="/apps?collection=rewards"
    class="inline-flex items-center rounded-full bg-emerald-500/15 font-bold text-emerald-800 ring-1 ring-emerald-500/25 transition-colors hover:bg-emerald-500/20 dark:text-emerald-200"
    :class="compact ? 'px-2 py-0.5 text-[11px]' : 'px-2.5 py-1 text-xs'"
    :title="label"
  >
    {{ label }}
  </RouterLink>
</template>
