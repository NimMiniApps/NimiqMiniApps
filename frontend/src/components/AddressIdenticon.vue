<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import Identicons from '@nimiq/identicons'
import identiconsSvgUrl from '@nimiq/identicons/dist/identicons.min.svg?url'

Identicons.svgPath = identiconsSvgUrl

const props = withDefaults(defineProps<{ address?: string; imgClass?: string }>(), {
  address: '',
  imgClass: 'h-10 w-10',
})

const imageUrl = ref(Identicons.placeholderToDataUrl('#d7deeb', 1))

async function render() {
  if (!props.address) {
    imageUrl.value = Identicons.placeholderToDataUrl('#d7deeb', 1)
    return
  }
  try {
    imageUrl.value = await Identicons.toDataUrl(props.address)
  } catch {
    imageUrl.value = Identicons.placeholderToDataUrl('#d7deeb', 1)
  }
}

onMounted(render)
watch(() => props.address, render)
</script>

<template>
  <img :class="[imgClass, 'rounded-full']" :src="imageUrl" alt="" />
</template>
