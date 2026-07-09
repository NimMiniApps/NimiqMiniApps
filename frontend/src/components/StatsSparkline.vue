<script setup lang="ts">
import { computed } from 'vue'
import { buildSparklinePoints } from '../utils/sparkline'

const props = defineProps<{
  daily: { date: string; opens: number; views: number }[]
  metric: 'opens' | 'views'
  width?: number
  height?: number
}>()

const chartW = computed(() => props.width ?? 120)
const chartH = computed(() => props.height ?? 32)

const points = computed(() => {
  const w = chartW.value
  const h = chartH.value
  const series = props.daily.map((d) => ({
    date: d.date,
    value: props.metric === 'opens' ? d.opens : d.views,
  }))
  if (series.length === 0) {
    return buildSparklinePoints(
      [{ date: '', value: 0 }, { date: '', value: 0 }],
      w,
      h,
    )
  }
  return buildSparklinePoints(series, w, h)
})
</script>

<template>
  <svg
    :width="chartW"
    :height="chartH"
    class="block w-full max-w-full text-accent/70"
    aria-hidden="true"
    :viewBox="`0 0 ${chartW} ${chartH}`"
    preserveAspectRatio="none"
  >
    <polyline
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
      stroke-linejoin="round"
      stroke-linecap="round"
      vector-effect="non-scaling-stroke"
      :points="points"
    />
  </svg>
</template>
