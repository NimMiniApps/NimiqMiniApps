<script setup lang="ts">
defineProps<{ status: string }>()

const styles: Record<string, string> = {
  verified: 'bg-emerald-500/20 text-emerald-300',
  approved: 'bg-sky-500/20 text-sky-300',
  experimental: 'bg-amber-500/20 text-amber-300',
  submitted: 'bg-white/10 text-white/70',
  rejected: 'bg-red-500/20 text-red-300',
}
</script>

<template>
  <span class="rounded-full px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide"
    :class="styles[status] || styles.submitted">{{ status }}</span>
</template>
