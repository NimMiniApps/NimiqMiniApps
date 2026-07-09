<script setup lang="ts">
import type { SocialLink } from '../api'
import LinkIconButton from './LinkIconButton.vue'
import { socialLabel } from '../utils/socials'

defineProps<{ items: SocialLink[] }>()
</script>

<template>
  <LinkIconButton
    v-for="item in items"
    :key="`${item.platform}-${item.url}`"
    :href="item.url"
    :platform="item.platform"
    :label="socialLabel(item.platform)"
  />
</template>
