<script setup lang="ts">
defineProps<{ stage: string }>()

const styles: Record<string, string> = {
  released: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-500/20 dark:text-emerald-300',
  beta: 'bg-amber-100 text-amber-800 dark:bg-amber-500/20 dark:text-amber-300',
  alpha: 'bg-orange-100 text-orange-800 dark:bg-orange-500/20 dark:text-orange-300',
  concept: 'bg-slate-200 text-slate-700 dark:bg-slate-500/20 dark:text-slate-300',
}

const labels: Record<string, string> = {
  released: 'Released',
  beta: 'Beta',
  alpha: 'Alpha',
  concept: 'Concept',
}
</script>

<template>
  <span class="rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide"
    :class="styles[stage] || styles.released">{{ labels[stage] || stage }}</span>
</template>
