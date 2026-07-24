<script setup lang="ts">
import type { MediaItem } from '../api'
import { youtubeEmbedUrl } from '../utils/youtube'

defineProps<{ items: MediaItem[]; title?: string }>()
</script>

<template>
  <section v-if="items.length" class="space-y-3">
    <h2 class="font-bold">{{ title || 'Media' }}</h2>
    <div class="grid gap-4 sm:grid-cols-2">
      <template v-for="(item, i) in items" :key="`${item.type}-${item.url}-${i}`">
        <div v-if="item.type === 'youtube' && youtubeEmbedUrl(item.url)"
          class="overflow-hidden rounded-2xl border border-line bg-black shadow-sm sm:col-span-2">
          <div class="aspect-video">
            <iframe
              :src="youtubeEmbedUrl(item.url)!"
              :title="`Video ${i + 1}`"
              class="h-full w-full"
              loading="lazy"
              allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
              allowfullscreen
            />
          </div>
        </div>
        <div v-else-if="item.type === 'image'"
          class="aspect-video w-full overflow-hidden rounded-2xl border border-line bg-slate-900 shadow-sm">
          <img
            :src="item.url"
            :alt="`Screenshot ${i + 1}`"
            loading="lazy"
            class="h-full w-full object-contain"
          />
        </div>
      </template>
    </div>
  </section>
</template>
