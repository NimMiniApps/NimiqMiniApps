<script setup lang="ts">
import { computed } from 'vue'
import type { AnalyticsDailyPoint } from '../api'
import { buildTrendPoints, type TrendMetric } from '../utils/adminAnalytics'

const props = defineProps<{
  daily: AnalyticsDailyPoint[]
  metric: TrendMetric
  label: string
}>()

const width = 640
const height = 160

const points = computed(() => buildTrendPoints(props.daily, props.metric, width, height))
const total = computed(() => props.daily.reduce((sum, d) => sum + (d[props.metric] ?? 0), 0))
const summary = computed(() => {
  if (props.daily.length === 0) return `No ${props.label} data`
  return `${props.label}: ${total.value} across ${props.daily.length} days`
})
</script>

<template>
  <div class="space-y-2">
    <svg
      class="h-40 w-full text-accent"
      :viewBox="`0 0 ${width} ${height}`"
      role="img"
      :aria-label="summary"
      preserveAspectRatio="none"
    >
      <polyline
        v-if="points"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        :points="points"
      />
    </svg>
    <table class="sr-only">
      <caption>{{ summary }}</caption>
      <thead>
        <tr>
          <th>Day</th>
          <th>{{ label }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="point in daily" :key="point.day">
          <td>{{ point.day }}</td>
          <td>{{ point[metric] }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
