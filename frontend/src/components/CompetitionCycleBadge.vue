<script setup lang="ts">
import { computed } from 'vue'
import type { CompetitionResult } from '../api'
import { competitionResultLabel, displayCompetitionResult } from '../utils/competition'

const props = defineProps<{
  cycle?: number | null
  results?: CompetitionResult[]
  compact?: boolean
}>()

const result = computed(() =>
  displayCompetitionResult({
    competition_cycle: props.cycle,
    competition_results: props.results,
  }),
)

const label = computed(() => {
  if (result.value) return competitionResultLabel(result.value)
  return props.cycle ? `Cycle ${props.cycle}` : ''
})

const filterCycle = computed(() => result.value?.cycle ?? props.cycle)
</script>

<template>
  <RouterLink
    v-if="filterCycle && label"
    :to="`/apps?competition_cycle=${filterCycle}`"
    class="inline-flex items-center gap-1.5 rounded-[3px] border border-board-hairline bg-board-flap font-bold uppercase tracking-wide text-board-flap-ink transition-colors hover:border-accent/50 hover:text-accent-ink"
    :class="compact ? 'px-1.5 py-0.5 text-[10px]' : 'px-2.5 py-1 text-xs'"
    :title="label"
  >
    <span
      class="board-lamp"
      :class="result?.place === 1 ? 'board-lamp-ok' : result?.place ? 'board-lamp-info' : 'board-lamp-info'"
      aria-hidden="true"
    ></span>
    {{ label }}
  </RouterLink>
</template>
