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
    class="inline-flex items-center rounded-[3px] border border-board-hairline bg-board-flap-hover font-bold uppercase tracking-wide text-emerald-300 transition-colors hover:border-emerald-400/50"
    :class="compact ? 'px-2 py-0.5 text-[11px]' : 'px-2.5 py-1 text-xs'"
    :title="label"
  >
    {{ label }}
  </RouterLink>
</template>
