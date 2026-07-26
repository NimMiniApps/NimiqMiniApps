<script setup lang="ts">
defineProps<{ stage: string }>()

const lamp: Record<string, string> = {
  released: 'board-lamp-live',
  beta: 'board-lamp-pending',
  alpha: 'board-lamp-pending',
  concept: 'board-lamp-info',
}

const labels: Record<string, string> = {
  released: 'Released',
  beta: 'Beta',
  alpha: 'Alpha',
  concept: 'Concept',
}
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-[3px] border border-board-hairline bg-board-flap px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-board-flap-ink"
  >
    <span class="board-lamp" :class="lamp[stage] || 'board-lamp-live'" aria-hidden="true"></span>
    {{ labels[stage] || stage }}
  </span>
</template>
