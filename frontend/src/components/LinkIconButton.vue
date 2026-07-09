<script setup lang="ts">
import { computed } from 'vue'
import { linkIcon } from '../utils/linkIcons'

const props = defineProps<{
  href: string
  platform: string
  label: string
}>()

const icon = computed(() => linkIcon(props.platform))
</script>

<template>
  <a
    :href="href"
    target="_blank"
    rel="noopener"
    :aria-label="label"
    :title="label"
    class="grid h-10 w-10 shrink-0 cursor-pointer place-items-center rounded-xl border border-line bg-surface-2/60 text-muted transition-colors duration-200 hover:border-accent/50 hover:bg-surface-2 hover:text-accent-ink"
  >
    <svg
      :viewBox="icon.viewBox"
      class="h-5 w-5"
      :fill="icon.fill ? 'currentColor' : 'none'"
      :stroke="icon.fill ? 'none' : 'currentColor'"
      stroke-width="1.75"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
    >
      <path v-for="(path, index) in icon.paths" :key="index" :d="path" />
    </svg>
  </a>
</template>
