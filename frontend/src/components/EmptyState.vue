<script setup lang="ts">
defineProps<{
  title: string
  description: string
  variant?: 'empty' | 'error' | 'notFound'
}>()
</script>

<template>
  <div class="board p-8 text-center">
    <div
      class="mx-auto mb-4 grid h-11 w-11 place-items-center rounded-[3px] border border-board-hairline bg-board-flap-hover"
      :class="{
        'text-board-flap-muted': !variant || variant === 'empty',
        'text-lamp-cancelled': variant === 'error',
        'text-accent-ink': variant === 'notFound',
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
    <h2 class="text-lg font-extrabold uppercase tracking-wide text-board-flap-ink">{{ title }}</h2>
    <p class="mx-auto mt-2 max-w-md text-sm leading-relaxed text-board-flap-muted">{{ description }}</p>
    <div v-if="$slots.actions" class="mt-5 flex flex-wrap items-center justify-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
