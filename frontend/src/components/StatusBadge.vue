<script setup lang="ts">
defineProps<{ status: string }>()

const lamp: Record<string, string> = {
  verified: 'board-lamp-live',
  approved: 'board-lamp-live',
  experimental: 'board-lamp-info',
  submitted: 'board-lamp-pending',
  rejected: 'board-lamp-cancelled',
}

const titles: Record<string, string> = {
  verified: 'Reviewed and verified by the directory moderators',
  approved: 'Approved and listed in the public directory',
  experimental: 'Early or experimental — may change or break',
}
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-[3px] border border-board-hairline bg-board-flap px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wide text-board-flap-ink"
    :title="titles[status]"
  >
    <span class="board-lamp" :class="lamp[status] || 'board-lamp-pending'" aria-hidden="true"></span>
    {{ status }}
  </span>
</template>
