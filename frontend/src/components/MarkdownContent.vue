<script setup lang="ts">
import { computed } from 'vue'
import { renderMarkdown } from '../utils/markdown'

const props = defineProps<{ source: string }>()

const html = computed(() => renderMarkdown(props.source))
</script>

<template>
  <div v-if="html" class="markdown leading-relaxed text-muted" v-html="html" />
</template>

<style scoped>
.markdown :deep(p + p) {
  margin-top: 0.75rem;
}

.markdown :deep(h2),
.markdown :deep(h3),
.markdown :deep(h4) {
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
  font-weight: 700;
  color: var(--nq-ink, inherit);
}

.markdown :deep(h2) {
  font-size: 1.125rem;
}

.markdown :deep(h3) {
  font-size: 1rem;
}

.markdown :deep(ul),
.markdown :deep(ol) {
  margin-top: 0.5rem;
  padding-left: 1.25rem;
}

.markdown :deep(ul) {
  list-style-type: disc;
}

.markdown :deep(ol) {
  list-style-type: decimal;
}

.markdown :deep(li + li) {
  margin-top: 0.25rem;
}

.markdown :deep(a) {
  font-weight: 600;
  color: var(--nq-accent-ink, #1557c7);
  text-decoration: underline;
  text-underline-offset: 2px;
}

.markdown :deep(code) {
  border-radius: 0.25rem;
  background: var(--nq-surface-2, rgba(0, 0, 0, 0.06));
  padding: 0.1rem 0.35rem;
  font-size: 0.875em;
}

.markdown :deep(pre) {
  margin-top: 0.75rem;
  overflow-x: auto;
  border-radius: 0.5rem;
  border: 1px solid var(--nq-line, rgba(0, 0, 0, 0.1));
  background: var(--nq-surface-2, rgba(0, 0, 0, 0.04));
  padding: 0.75rem 1rem;
}

.markdown :deep(pre code) {
  background: transparent;
  padding: 0;
}

.markdown :deep(blockquote) {
  margin-top: 0.75rem;
  border-left: 3px solid var(--nq-accent, #0582ca);
  padding-left: 0.75rem;
  color: var(--nq-muted, inherit);
}
</style>
