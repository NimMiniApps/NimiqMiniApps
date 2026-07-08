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
        <img v-else-if="item.type === 'image'"
          :src="item.url"
          :alt="`Screenshot ${i + 1}`"
          loading="lazy"
          class="h-64 w-full rounded-2xl border border-line object-cover shadow-sm"
        />
      </template>
    </div>
  </section>
</template>
