<script setup lang="ts">
import { ref, watch } from 'vue'
import type { SocialLink } from '../api'
import {
  SOCIAL_PLATFORMS,
  compileSocialRows,
  detectPlatformFromUrl,
  newSocialRow,
  normalizeSocialUrl,
  socialLabel,
  socialLinksToRows,
  socialPlaceholder,
  type SocialLinkRow,
  type SocialPlatform,
} from '../utils/socials'

const model = defineModel<SocialLink[]>({ default: () => [] })

const rows = ref<SocialLinkRow[]>([])
const rowErrors = ref<Record<string, string>>({})
let syncingFromModel = false

function syncRowsFromModel(links: SocialLink[]) {
  syncingFromModel = true
  rows.value = socialLinksToRows(links)
  rowErrors.value = {}
  syncingFromModel = false
}

watch(
  () => model.value,
  (links) => {
    if (syncingFromModel) return
    const compiled = compileSocialRows(rows.value)
    const same = compiled.length === links.length
      && compiled.every((item, i) => item.platform === links[i]?.platform && item.url === links[i]?.url)
    if (!same) syncRowsFromModel(links)
  },
  { immediate: true, deep: true },
)

function emitModel() {
  if (syncingFromModel) return
  model.value = compileSocialRows(rows.value)
}

function addRow() {
  const used = new Set(rows.value.map((row) => row.platform).filter(Boolean))
  const next = SOCIAL_PLATFORMS.find((p) => !used.has(p)) ?? 'twitter'
  rows.value.push(newSocialRow({ platform: next, url: '' }))
}

function removeRow(id: string) {
  rows.value = rows.value.filter((row) => row.id !== id)
  delete rowErrors.value[id]
  emitModel()
}

function onPlatformChange(row: SocialLinkRow) {
  delete rowErrors.value[row.id]
  if (row.value.trim()) normalizeRow(row)
  emitModel()
}

function normalizeRow(row: SocialLinkRow) {
  delete rowErrors.value[row.id]
  const value = row.value.trim()
  if (!value) {
    emitModel()
    return
  }

  if (!row.platform) {
    const detected = detectPlatformFromUrl(value)
    if (detected) row.platform = detected
  }

  if (!row.platform) {
    rowErrors.value[row.id] = 'Choose a platform first'
    return
  }

  try {
    const detected = detectPlatformFromUrl(value)
    if (detected && !row.platform) row.platform = detected
    row.value = normalizeSocialUrl(row.platform as SocialPlatform, value)
    emitModel()
  } catch (e) {
    rowErrors.value[row.id] = (e as Error).message
  }
}

function onInputBlur(row: SocialLinkRow) {
  normalizeRow(row)
}

function validate(): SocialLink[] {
  for (const row of rows.value) {
    if (row.value.trim()) normalizeRow(row)
  }
  if (Object.keys(rowErrors.value).length > 0) {
    throw new Error('Fix the social links highlighted below')
  }
  return compileSocialRows(rows.value)
}

defineExpose({ validate })
</script>

<template>
  <div class="space-y-2">
    <p v-if="!rows.length" class="text-xs text-muted">
      Optional — add links to your community profiles so users can follow updates.
    </p>

    <div
      v-for="row in rows"
      :key="row.id"
      class="space-y-1 rounded-xl border border-line bg-surface-2/50 p-3"
    >
      <div class="flex flex-col gap-2 sm:flex-row sm:items-start">
        <label class="block shrink-0 text-sm sm:w-36">
          <span class="sr-only">Platform</span>
          <select
            v-model="row.platform"
            class="w-full cursor-pointer rounded-lg border border-line bg-surface px-3 py-2 outline-none transition-colors duration-200 focus:border-accent"
            @change="onPlatformChange(row)"
          >
            <option value="" disabled>Platform</option>
            <option v-for="platform in SOCIAL_PLATFORMS" :key="platform" :value="platform">
              {{ socialLabel(platform) }}
            </option>
          </select>
        </label>
        <label class="block min-w-0 flex-1 text-sm">
          <span class="sr-only">Profile URL or handle</span>
          <input
            v-model="row.value"
            type="text"
            :placeholder="row.platform ? socialPlaceholder(row.platform as SocialPlatform) : 'Paste a link or @handle'"
            class="w-full rounded-lg border border-line bg-surface px-3 py-2 outline-none transition-colors duration-200 placeholder:text-muted/60 focus:border-accent"
            :class="rowErrors[row.id] ? 'border-red-400/70' : ''"
            @blur="onInputBlur(row)"
          />
        </label>
        <button
          type="button"
          class="cursor-pointer self-start rounded-lg border border-line px-3 py-2 text-xs font-semibold text-muted transition-colors duration-200 hover:border-red-400/50 hover:text-red-600 dark:hover:text-red-300 sm:mt-0"
          aria-label="Remove social link"
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
      + Add social link
    </button>
  </div>
</template>
