<script setup lang="ts">
import { computed, ref } from 'vue'
import type { App } from '../api'
import { appIdentityAccent, displayIconUrl } from '../utils/appIcon'

const props = withDefaults(defineProps<{
  app: Pick<App, 'slug' | 'name' | 'icon_url' | 'discovered_icon_url'>
  size?: 'sm' | 'md'
}>(), {
  size: 'sm',
})

const failed = ref(false)
const iconUrl = computed(() => displayIconUrl(props.app))
const showImage = computed(() => !!iconUrl.value && !failed.value)
const accent = computed(() => appIdentityAccent(props.app))

const sizeClass = computed(() => props.size === 'md'
  ? 'h-16 w-16 rounded-2xl text-3xl'
  : 'h-12 w-12 rounded-xl text-xl')

function onError() {
  failed.value = true
}
</script>

<template>
  <img
    v-if="showImage"
    :src="iconUrl!"
    :alt="app.name"
    loading="lazy"
    :class="[sizeClass, 'object-cover']"
    @error="onError"
  />
  <div
    v-else
    class="grid shrink-0 place-items-center font-extrabold text-white shadow-sm"
    :class="sizeClass"
    :style="{ backgroundColor: accent }"
  >
    {{ app.name[0] }}
  </div>
</template>
