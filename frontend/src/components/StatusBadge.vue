<script setup lang="ts">
defineProps<{ status: string }>()

const styles: Record<string, string> = {
  verified: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300',
  approved: 'bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-300',
  experimental: 'bg-violet-100 text-violet-700 dark:bg-violet-500/20 dark:text-violet-300',
  submitted: 'bg-surface-2 text-muted',
  rejected: 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300',
}
</script>

<template>
  <span class="rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide"
    :class="styles[status] || styles.submitted">{{ status }}</span>
</template>
