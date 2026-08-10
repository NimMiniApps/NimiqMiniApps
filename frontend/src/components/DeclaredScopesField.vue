<script setup lang="ts">
import { computed } from 'vue'
import { DECLARED_SCOPES } from '../api'

const props = defineProps<{
  modelValue: string
  help?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const knownSet = new Set<string>(DECLARED_SCOPES.map((s) => s.value))

const selected = computed(() =>
  props.modelValue.split(',').map((token) => token.trim()).filter(Boolean),
)

const extraScopes = computed(() =>
  selected.value.filter((scope) => !knownSet.has(scope)),
)

const summary = computed(() => {
  if (!selected.value.length) return 'None selected'
  if (selected.value.length === 1) return selected.value[0]
  return `${selected.value.length} scopes selected`
})

function toggle(scope: string) {
  const next = new Set(selected.value)
  if (next.has(scope)) {
    next.delete(scope)
  } else {
    next.add(scope)
  }
  emit('update:modelValue', Array.from(next).join(', '))
}
</script>

<template>
  <details class="group text-sm sm:col-span-2">
    <summary class="flex cursor-pointer list-none items-center justify-between gap-2 rounded-lg border border-line bg-surface-2 px-3 py-2 font-semibold text-muted outline-none marker:content-none [&::-webkit-details-marker]:hidden">
      <span>
        Declared scopes
        <span class="ml-2 font-normal text-ink">{{ summary }}</span>
      </span>
      <span class="text-muted transition-transform group-open:rotate-180" aria-hidden="true">⌄</span>
    </summary>
    <div class="mt-2 space-y-2 rounded-lg border border-line bg-surface-2/50 p-3">
      <label
        v-for="scope in DECLARED_SCOPES"
        :key="scope.value"
        class="flex cursor-pointer items-start gap-2 rounded-md px-1 py-1 hover:bg-surface-2"
      >
        <input
          type="checkbox"
          class="mt-0.5 h-4 w-4 shrink-0 accent-[#1F74FF]"
          :checked="selected.includes(scope.value)"
          @change="toggle(scope.value)"
        />
        <span>
          <span class="block font-semibold text-ink">{{ scope.label }}</span>
          <span class="block font-mono text-[11px] text-muted">{{ scope.value }}</span>
          <span class="mt-0.5 block text-xs leading-snug text-muted">{{ scope.description }}</span>
        </span>
      </label>

      <div v-if="extraScopes.length" class="space-y-1 border-t border-line pt-2">
        <p class="text-xs text-muted">Already on this app (kept unless unchecked):</p>
        <label
          v-for="scope in extraScopes"
          :key="scope"
          class="flex cursor-pointer items-center gap-2 rounded-md px-1 py-1 hover:bg-surface-2"
        >
          <input
            type="checkbox"
            class="h-4 w-4 accent-[#1F74FF]"
            :checked="true"
            @change="toggle(scope)"
          />
          <span class="font-mono text-xs text-ink">{{ scope }}</span>
        </label>
      </div>
    </div>
    <span class="mt-1 block text-xs leading-snug text-muted">
      {{ help || 'Scopes this app may request via NimConnect. Leave all unchecked for none.' }}
    </span>
  </details>
</template>
