<script setup lang="ts">
import { ref, watch } from 'vue'
import type { MediaItem } from '../api'
import {
  compileMediaRows,
  mediaItemsToRows,
  mediaTypeLabel,
  newMediaRow,
  normalizeMediaItem,
  youtubeThumbnail,
  type MediaRow,
} from '../utils/media'

const model = defineModel<MediaItem[]>({ default: () => [] })

const rows = ref<MediaRow[]>([])
const rowErrors = ref<Record<string, string>>({})
const previewErrors = ref<Record<string, boolean>>({})
let syncingFromModel = false

function syncRowsFromModel(items: MediaItem[]) {
  syncingFromModel = true
  rows.value = mediaItemsToRows(items)
  rowErrors.value = {}
  previewErrors.value = {}
  syncingFromModel = false
}

watch(
  () => model.value,
  (items) => {
    if (syncingFromModel) return
    const compiled = compileMediaRows(rows.value)
    const same = compiled.length === items.length
      && compiled.every((item, i) => item.type === items[i]?.type && item.url === items[i]?.url)
    if (!same) syncRowsFromModel(items)
  },
  { immediate: true, deep: true },
)

function emitModel() {
  if (syncingFromModel) return
  try {
    model.value = compileMediaRows(rows.value)
  } catch {
    // keep partial model until rows are valid
  }
}

function addRow() {
  rows.value.push(newMediaRow())
}

function removeRow(id: string) {
  rows.value = rows.value.filter((row) => row.id !== id)
  delete rowErrors.value[id]
  delete previewErrors.value[id]
  emitModel()
}

function normalizeRow(row: MediaRow) {
  delete rowErrors.value[row.id]
  delete previewErrors.value[row.id]

  const value = row.url.trim()
  if (!value) {
    row.type = null
    emitModel()
    return
  }

  try {
    const item = normalizeMediaItem(value)
    row.url = item.url
    row.type = item.type
    emitModel()
  } catch (e) {
    row.type = null
    rowErrors.value[row.id] = (e as Error).message
  }
}

function onInputBlur(row: MediaRow) {
  normalizeRow(row)
}

function onPreviewError(rowId: string) {
  previewErrors.value[rowId] = true
}

function previewUrl(row: MediaRow): string | null {
  if (!row.url.trim() || rowErrors.value[row.id]) return null
  if (row.type === 'youtube') return youtubeThumbnail(row.url)
  if (row.type === 'image') return row.url
  try {
    const item = normalizeMediaItem(row.url)
    return item.type === 'youtube' ? youtubeThumbnail(item.url) : item.url
  } catch {
    return null
  }
}

function validate(): MediaItem[] {
  for (const row of rows.value) {
    if (row.url.trim()) normalizeRow(row)
  }
  if (Object.keys(rowErrors.value).length > 0) {
    throw new Error('Fix the media links highlighted below')
  }
  return compileMediaRows(rows.value)
}

defineExpose({ validate })
</script>

<template>
  <div class="space-y-2">
    <p v-if="!rows.length" class="text-xs text-muted">
      Optional — add screenshots or a YouTube demo. Use direct image links (https://…) hosted anywhere public.
    </p>

    <div
      v-for="row in rows"
      :key="row.id"
      class="space-y-2 rounded-xl border border-line bg-surface-2/50 p-3"
    >
      <div class="flex gap-3">
        <div
          class="flex h-16 w-24 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-line bg-surface text-[10px] font-semibold uppercase tracking-wide text-muted"
          :class="row.type === 'youtube' ? 'bg-black' : ''"
        >
          <img
            v-if="previewUrl(row) && !previewErrors[row.id]"
            :src="previewUrl(row)!"
            alt=""
            class="h-full w-full object-cover"
            @error="onPreviewError(row.id)"
          />
          <span v-else-if="row.type === 'youtube'" class="px-1 text-center text-white/80">Video</span>
          <span v-else class="px-1 text-center">Preview</span>
        </div>

        <div class="min-w-0 flex-1 space-y-1">
          <label class="block text-sm">
            <span class="sr-only">Image or YouTube URL</span>
            <input
              v-model="row.url"
              type="url"
              placeholder="https://… screenshot or YouTube link"
              class="w-full rounded-lg border border-line bg-surface px-3 py-2 outline-none transition-colors duration-200 placeholder:text-muted/60 focus:border-accent"
              :class="rowErrors[row.id] ? 'border-red-400/70' : ''"
              @blur="onInputBlur(row)"
            />
          </label>
          <p v-if="row.type && !rowErrors[row.id]" class="text-xs text-muted">
            {{ mediaTypeLabel(row.type) }}
          </p>
          <p v-if="previewErrors[row.id] && row.type === 'image'" class="text-xs text-amber-700 dark:text-amber-200">
            Preview failed — check the URL is a direct link to an image file.
          </p>
        </div>

        <button
          type="button"
          class="cursor-pointer shrink-0 self-start rounded-lg border border-line px-3 py-2 text-xs font-semibold text-muted transition-colors duration-200 hover:border-red-400/50 hover:text-red-600 dark:hover:text-red-300"
          aria-label="Remove media"
          @click="removeRow(row.id)"
        >
          Remove
        </button>
      </div>
      <p v-if="rowErrors[row.id]" class="text-xs text-red-600 dark:text-red-300">{{ rowErrors[row.id] }}</p>
    </div>

    <button
      type="button"
      class="cursor-pointer rounded-lg border border-dashed border-line px-3 py-2 text-sm font-semibold text-accent-ink transition-colors duration-200 hover:border-accent/50 hover:bg-surface-2"
      @click="addRow"
    >
      + Add screenshot or video
    </button>
  </div>
</template>
