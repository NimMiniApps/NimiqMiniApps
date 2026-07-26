<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  text: string
  delayStep?: number
}>(), {
  delayStep: 0.018,
})

const words = computed(() => {
  let i = 0
  return props.text.split(' ').map((word) => [...word].map((char) => {
    const style = { animationDelay: `${i * props.delayStep}s` }
    i += 1
    return { char, style }
  }))
})
</script>

<template>
  <span aria-hidden="true">
    <template v-for="(word, w) in words" :key="w">
      <span v-if="w > 0" class="inline-block" style="width: 0.3em"></span>
      <span class="inline-block whitespace-nowrap">
        <span v-for="(c, i) in word" :key="i" class="board-flap-char" :style="c.style">{{ c.char }}</span>
      </span>
    </template>
  </span>
  <span class="sr-only">{{ text }}</span>
</template>
