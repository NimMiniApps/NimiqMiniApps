<script setup lang="ts">
import { computed } from 'vue'
import { competitionCycleLabel } from '../utils/competition'

const props = defineProps<{
  cycle: number
  compact?: boolean
}>()

const label = computed(() => competitionCycleLabel(props.cycle))
</script>

<template>
  <RouterLink
    :to="`/apps?competition_cycle=${cycle}`"
    class="inline-flex items-center gap-1.5 rounded-[3px] border border-board-hairline bg-board-flap font-bold uppercase tracking-wide text-board-flap-ink transition-colors hover:border-accent/50 hover:text-accent-ink"
    :class="compact ? 'px-1.5 py-0.5 text-[10px]' : 'px-2.5 py-1 text-xs'"
    :title="label"
  >
    <span class="board-lamp board-lamp-info" aria-hidden="true"></span>
    {{ label }}
  </RouterLink>
</template>
