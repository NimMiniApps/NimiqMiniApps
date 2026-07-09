<script setup lang="ts">
import { computed } from 'vue'
import { APP_ASSETS } from '../api'

const props = defineProps<{
  modelValue: string
  label: string
  help?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const selected = computed(() =>
  props.modelValue.split(',').map((token) => token.trim()).filter(Boolean),
)

const summary = computed(() => selected.value.length ? selected.value.join(', ') : 'None')

function toggle(token: string) {
  const next = new Set(selected.value)
  if (next.has(token)) {
    next.delete(token)
  } else {
    next.add(token)
  }
  emit('update:modelValue', Array.from(next).join(', '))
}
</script>

<template>
  <div class="text-sm">
    <span class="mb-1 block font-semibold text-muted">{{ label }}</span>
    <details class="group relative">
      <summary
        class="token-select-trigger flex min-h-10 cursor-pointer list-none items-center justify-between gap-2 rounded-lg border px-3 py-2 outline-none transition-colors duration-200"
      >
        <span class="truncate font-medium text-ink">{{ summary }}</span>
        <span class="text-muted transition-transform group-open:rotate-180" aria-hidden="true">⌄</span>
      </summary>
      <div class="token-select-menu absolute z-20 mt-1 w-full rounded-lg border p-2 shadow-lg">
        <label
          v-for="asset in APP_ASSETS"
          :key="asset"
          class="token-select-option flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5"
        >
          <input
            type="checkbox"
            class="h-4 w-4 accent-[#1F74FF]"
            :checked="selected.includes(asset)"
            @change="toggle(asset)"
          />
          <span class="font-semibold text-ink">{{ asset }}</span>
        </label>
      </div>
    </details>
    <span v-if="help" class="mt-1 block text-xs leading-snug text-muted">{{ help }}</span>
  </div>
</template>

<style scoped>
.token-select-trigger {
  background: var(--nq-surface-2);
  border-color: var(--nq-line);
  color: var(--nq-ink);
}

.token-select-trigger:hover,
.token-select-trigger:focus {
  border-color: color-mix(in srgb, var(--nq-accent) 55%, transparent);
}

.token-select-menu {
  background: var(--nq-surface);
  border-color: var(--nq-line);
  color: var(--nq-ink);
  box-shadow: 0 0.75rem 2rem rgba(0, 0, 0, 0.16);
}

.token-select-option {
  color: var(--nq-ink);
}

.token-select-option:hover {
  background: var(--nq-surface-2);
}
</style>
