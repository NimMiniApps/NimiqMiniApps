<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { MediaItem } from '../api'
import { clampMediaIndex, moveMediaIndex } from '../utils/mediaGallery'
import { youtubeEmbedUrl } from '../utils/youtube'

const props = defineProps<{ items: MediaItem[]; title?: string }>()

const selectedIndex = ref(0)
const selectedItem = computed(() => props.items[selectedIndex.value])

watch(
  () => props.items,
  (items) => {
    selectedIndex.value = clampMediaIndex(selectedIndex.value, items.length)
  },
  { deep: true },
)

function selectMedia(index: number) {
  selectedIndex.value = clampMediaIndex(index, props.items.length)
}

function moveSelection(direction: 'previous' | 'next') {
  selectedIndex.value = moveMediaIndex(selectedIndex.value, props.items.length, direction)
}
</script>

<template>
  <section v-if="props.items.length" class="space-y-3">
    <h2 class="font-bold">{{ props.title || 'Media' }}</h2>

    <div class="relative overflow-hidden rounded-2xl border border-line bg-slate-950 shadow-sm">
      <div class="aspect-video">
        <iframe
          v-if="selectedItem?.type === 'youtube' && youtubeEmbedUrl(selectedItem.url)"
          :src="youtubeEmbedUrl(selectedItem.url)!"
          :title="`Video ${selectedIndex + 1}`"
          class="h-full w-full"
          loading="lazy"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
        />
        <img
          v-else-if="selectedItem?.type === 'image'"
          :src="selectedItem.url"
          :alt="`Screenshot ${selectedIndex + 1}`"
          class="h-full w-full object-contain"
        />
        <div v-else class="flex h-full items-center justify-center text-sm font-medium text-white/70">
          Video preview unavailable
        </div>
      </div>

      <template v-if="props.items.length > 1">
        <button
          type="button"
          class="absolute left-3 top-1/2 grid h-10 w-10 -translate-y-1/2 place-items-center rounded-full bg-slate-950/80 text-white shadow-sm ring-1 ring-white/20 transition hover:bg-slate-950 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          aria-label="Show previous media"
          @click="moveSelection('previous')"
        >
          <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-current" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="m15 18-6-6 6-6" />
          </svg>
        </button>
        <button
          type="button"
          class="absolute right-3 top-1/2 grid h-10 w-10 -translate-y-1/2 place-items-center rounded-full bg-slate-950/80 text-white shadow-sm ring-1 ring-white/20 transition hover:bg-slate-950 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          aria-label="Show next media"
          @click="moveSelection('next')"
        >
          <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-current" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="m9 18 6-6-6-6" />
          </svg>
        </button>
      </template>
    </div>

    <div class="flex snap-x snap-mandatory gap-3 overflow-x-auto pb-1" role="group" aria-label="Media thumbnails">
      <button
        v-for="(item, index) in props.items"
        :key="`${item.type}-${item.url}-${index}`"
        type="button"
        class="relative h-16 w-24 shrink-0 snap-start overflow-hidden rounded-xl border border-line bg-slate-950 text-left transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
        :class="index === selectedIndex ? 'ring-2 ring-accent ring-offset-2 ring-offset-base' : 'hover:border-ink/40'"
        :aria-label="item.type === 'youtube' ? `Show video ${index + 1}` : `Show screenshot ${index + 1}`"
        :aria-current="index === selectedIndex ? 'true' : undefined"
        :aria-pressed="index === selectedIndex"
        @click="selectMedia(index)"
      >
        <img
          v-if="item.type === 'image'"
          :src="item.url"
          :alt="`Screenshot ${index + 1} preview`"
          loading="lazy"
          class="h-full w-full object-cover"
        />
        <span v-else class="flex h-full w-full flex-col items-center justify-center gap-1 bg-slate-900 px-2 text-center text-xs font-semibold text-white">
          <svg viewBox="0 0 24 24" class="h-5 w-5 fill-current" aria-hidden="true">
            <path d="M8 5v14l11-7z" />
          </svg>
          Video
        </span>
      </button>
    </div>
  </section>
</template>
