<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { MediaItem } from '../api'
import { clampMediaIndex, moveMediaIndex, orderMediaItems } from '../utils/mediaGallery'
import { youtubeEmbedUrl } from '../utils/youtube'

const props = defineProps<{ items: MediaItem[]; title?: string }>()

const selectedIndex = ref(0)
const imageDialog = ref<HTMLDialogElement>()
const orderedItems = computed(() => orderMediaItems(props.items))
const selectedItem = computed(() => orderedItems.value[selectedIndex.value])

watch(
  () => props.items,
  (items) => {
    selectedIndex.value = clampMediaIndex(selectedIndex.value, items.length)
  },
  { deep: true },
)

function selectMedia(index: number) {
  selectedIndex.value = clampMediaIndex(index, orderedItems.value.length)
}

function moveSelection(direction: 'previous' | 'next') {
  selectedIndex.value = moveMediaIndex(selectedIndex.value, orderedItems.value.length, direction)
}

function openImageDialog() {
  if (selectedItem.value?.type === 'image' && !imageDialog.value?.open) {
    imageDialog.value?.showModal()
  }
}

function closeImageDialog() {
  imageDialog.value?.close()
}
</script>

<template>
  <section v-if="orderedItems.length" class="space-y-3">
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
        <button
          v-else-if="selectedItem?.type === 'image'"
          type="button"
          class="h-full w-full cursor-zoom-in focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-accent"
          :aria-label="`View screenshot ${selectedIndex + 1} full screen`"
          @click="openImageDialog"
        >
          <img
            :src="selectedItem.url"
            :alt="`Screenshot ${selectedIndex + 1}`"
            class="h-full w-full object-contain"
          />
        </button>
        <div v-else class="flex h-full items-center justify-center text-sm font-medium text-white/70">
          Video preview unavailable
        </div>
      </div>

    </div>

    <div class="flex items-center justify-center gap-2">
      <button
        v-if="orderedItems.length > 1"
        type="button"
        class="grid h-10 w-10 shrink-0 place-items-center rounded-full border border-line bg-surface text-ink shadow-sm transition hover:bg-surface-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
        aria-label="Show previous media"
        @click="moveSelection('previous')"
      >
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-current" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="m15 18-6-6 6-6" />
        </svg>
      </button>

      <div class="flex max-w-[calc(100vw-8rem)] snap-x snap-mandatory gap-3 overflow-x-auto pb-1" role="group" aria-label="Media thumbnails">
        <button
          v-for="(item, index) in orderedItems"
          :key="`${item.type}-${item.url}-${index}`"
          type="button"
          class="relative h-16 w-24 shrink-0 snap-start overflow-hidden rounded-xl border border-line bg-slate-950 p-1 text-left transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          :class="index === selectedIndex ? 'ring-2 ring-inset ring-accent' : 'hover:border-ink/40'"
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
            class="h-full w-full rounded-lg object-cover"
          />
          <span v-else class="flex h-full w-full flex-col items-center justify-center gap-1 rounded-lg bg-slate-900 px-2 text-center text-xs font-semibold text-white">
            <svg viewBox="0 0 24 24" class="h-5 w-5 fill-current" aria-hidden="true">
              <path d="M8 5v14l11-7z" />
            </svg>
            Video
          </span>
        </button>
      </div>

      <button
        v-if="orderedItems.length > 1"
        type="button"
        class="grid h-10 w-10 shrink-0 place-items-center rounded-full border border-line bg-surface text-ink shadow-sm transition hover:bg-surface-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
        aria-label="Show next media"
        @click="moveSelection('next')"
      >
        <svg viewBox="0 0 24 24" class="h-5 w-5 fill-none stroke-current" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="m9 18 6-6-6-6" />
        </svg>
      </button>
    </div>

    <dialog
      v-if="selectedItem?.type === 'image'"
      ref="imageDialog"
      class="fixed inset-0 m-auto h-[100dvh] w-[100vw] max-h-none max-w-none items-center justify-center overflow-hidden border-0 bg-transparent p-4 backdrop:bg-slate-950/90 open:flex sm:p-8"
      aria-label="Full-screen screenshot"
      @click.self="closeImageDialog"
    >
      <img
        :src="selectedItem.url"
        :alt="`Screenshot ${selectedIndex + 1} full screen`"
        class="max-h-full max-w-full rounded-lg object-contain shadow-2xl"
      />
      <button
        type="button"
        class="fixed right-4 top-4 grid h-11 w-11 place-items-center rounded-full border border-white/20 bg-slate-950/80 text-white shadow-lg backdrop-blur transition hover:bg-slate-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
        aria-label="Close full-screen image"
        @click="closeImageDialog"
      >
        <svg viewBox="0 0 24 24" class="h-6 w-6 fill-none stroke-current" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M6 6l12 12M18 6 6 18" />
        </svg>
      </button>
    </dialog>
  </section>
</template>
