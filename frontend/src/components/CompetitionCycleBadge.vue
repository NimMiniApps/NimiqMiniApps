<script setup lang="ts">
import { computed } from 'vue'
import type { CompetitionResult } from '../api'
import {
  competitionPlaceLabel,
  competitionResultLabel,
  displayCompetitionResult,
  isPodiumPlace,
} from '../utils/competition'

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

const podium = computed(() => (isPodiumPlace(result.value?.place) ? result.value!.place : null))

const label = computed(() => {
  if (result.value) return competitionResultLabel(result.value)
  return props.cycle ? `Cycle ${props.cycle}` : ''
})

const filterCycle = computed(() => result.value?.cycle ?? props.cycle)

const placeWord = computed(() => {
  if (!podium.value) return ''
  return `${competitionPlaceLabel(podium.value)} place`
})
</script>

<template>
  <RouterLink
    v-if="filterCycle && label"
    :to="`/apps?competition_cycle=${filterCycle}`"
    class="inline-flex items-center gap-1.5 rounded-[3px] border font-bold uppercase tracking-wide transition-colors"
    :class="[
      compact ? 'px-1.5 py-0.5 text-[10px]' : 'px-2.5 py-1 text-xs',
      podium === 1
        ? 'border-amber-500/70 bg-gradient-to-b from-amber-300 to-amber-500 text-amber-950 shadow-[inset_0_1px_0_rgba(255,255,255,0.45)] hover:brightness-105'
        : podium === 2
          ? 'border-slate-400/80 bg-gradient-to-b from-slate-100 to-slate-300 text-slate-900 shadow-[inset_0_1px_0_rgba(255,255,255,0.55)] hover:brightness-105'
          : podium === 3
            ? 'border-orange-700/60 bg-gradient-to-b from-orange-300 to-orange-600 text-orange-950 shadow-[inset_0_1px_0_rgba(255,255,255,0.35)] hover:brightness-105'
            : 'border-board-hairline bg-board-flap text-board-flap-ink hover:border-accent/50 hover:text-accent-ink',
    ]"
    :title="label"
  >
    <svg
      v-if="podium"
      class="shrink-0"
      :class="compact ? 'h-3 w-3' : 'h-3.5 w-3.5'"
      viewBox="0 0 16 16"
      fill="currentColor"
      aria-hidden="true"
    >
      <path d="M4.5 1.5h7v1.2h1.3A1.7 1.7 0 0 1 14.5 4.4c0 2.2-1.5 3.6-3.4 4.1-.4 1-.9 1.6-1.6 2v1.5h2.2v1.5H4.3V11.9h2.2V10.4c-.7-.4-1.2-1-1.6-2C2.9 8 1.5 6.6 1.5 4.4A1.7 1.7 0 0 1 3.2 2.7h1.3V1.5Zm1.5 1.2v1.2H10V2.7H6Zm-2.8 2.9c0 1.3.8 2.2 2 2.6V5.6H3.2Zm9.6 0H9.8v2.6c1.2-.4 2-1.3 2-2.6Z" />
    </svg>
    <span v-else class="board-lamp board-lamp-info" aria-hidden="true"></span>
    <template v-if="podium">
      <span>{{ placeWord }}</span>
      <span v-if="!compact" class="opacity-80">· C{{ result?.cycle }}</span>
    </template>
    <template v-else>
      {{ label }}
    </template>
  </RouterLink>
</template>
