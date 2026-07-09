<script setup lang="ts">
import { useI18n } from '../composables/useI18n'

defineProps<{
  category: string
  appName: string
}>()

const { t } = useI18n()
</script>

<template>
  <nav aria-label="Breadcrumb" class="text-sm text-muted">
    <ol class="flex flex-wrap items-center gap-1.5">
      <li>
        <RouterLink to="/apps" class="font-semibold transition-colors hover:text-accent-ink">
          {{ t('nav.apps') }}
        </RouterLink>
      </li>
      <li aria-hidden="true" class="text-muted/60">→</li>
      <li>
        <RouterLink
          :to="`/apps?category=${encodeURIComponent(category)}`"
          class="font-semibold transition-colors hover:text-accent-ink"
        >
          {{ category }}
        </RouterLink>
      </li>
      <li aria-hidden="true" class="text-muted/60">→</li>
      <li class="font-semibold text-ink" aria-current="page">{{ appName }}</li>
    </ol>
  </nav>
</template>
