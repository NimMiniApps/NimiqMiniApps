<script setup lang="ts">
defineProps<{
  title: string
  description: string
  variant?: 'empty' | 'error' | 'notFound'
}>()
</script>

<template>
  <div
    class="rounded-2xl border border-line bg-surface p-8 text-center shadow-sm"
    :class="variant === 'error' ? 'border-red-500/20' : ''"
  >
    <div
      class="mx-auto mb-4 grid h-12 w-12 place-items-center rounded-2xl"
      :class="{
        'bg-surface-2 text-muted': !variant || variant === 'empty',
        'bg-red-500/10 text-red-600 dark:text-red-300': variant === 'error',
        'bg-accent/10 text-accent-ink': variant === 'notFound',
      }"
      aria-hidden="true"
    >
      <svg v-if="variant === 'error'" viewBox="0 0 24 24" class="h-6 w-6" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
        <circle cx="12" cy="12" r="9" />
        <path d="M12 8v4M12 16h.01" />
      </svg>
      <svg v-else-if="variant === 'notFound'" viewBox="0 0 24 24" class="h-6 w-6" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="11" cy="11" r="7" />
        <path d="M20 20l-3.5-3.5M8 11h6" />
      </svg>
      <svg v-else viewBox="0 0 24 24" class="h-6 w-6" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 4h7v7H4zM13 4h7v7h-7zM4 13h7v7H4zM13 13h7v7h-7z" />
      </svg>
    </div>
    <h2 class="text-lg font-extrabold">{{ title }}</h2>
    <p class="mx-auto mt-2 max-w-md text-sm leading-relaxed text-muted">{{ description }}</p>
    <div v-if="$slots.actions" class="mt-5 flex flex-wrap items-center justify-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
