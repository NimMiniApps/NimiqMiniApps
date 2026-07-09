<script setup lang="ts">
import { ref, watch } from 'vue'
import { submitAppReview, type AppReview } from '../api'

const props = defineProps<{ slug: string; existing: AppReview | null }>()
const emit = defineEmits<{ saved: [AppReview] }>()

const rating = ref(props.existing?.rating ?? 0)
const body = ref(props.existing?.body ?? '')
const submitting = ref(false)
const error = ref('')

watch(
  () => props.existing,
  (value) => {
    rating.value = value?.rating ?? 0
    body.value = value?.body ?? ''
  },
)

async function submit() {
  if (rating.value < 1) {
    error.value = 'Pick a star rating'
    return
  }
  submitting.value = true
  error.value = ''
  try {
    const review = await submitAppReview(props.slug, rating.value, body.value.trim())
    emit('saved', review)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to submit review'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form class="flex flex-col gap-3 rounded-xl border border-line bg-surface p-4" @submit.prevent="submit">
    <div class="flex gap-1">
      <button
        v-for="n in 5" :key="n" type="button"
        class="text-2xl leading-none"
        :class="n <= rating ? 'text-accent-ink' : 'text-muted'"
        @click="rating = n"
      >★</button>
    </div>
    <textarea
      v-model="body"
      rows="3"
      maxlength="1000"
      placeholder="Share your experience with this app (optional)"
      class="rounded-lg border border-line bg-surface-2 p-2 text-sm"
    />
    <div class="flex items-center justify-between">
      <p v-if="error" class="text-xs text-red-500">{{ error }}</p>
      <button
        type="submit"
        class="ml-auto rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
        :disabled="submitting"
      >{{ submitting ? 'Saving…' : existing ? 'Update review' : 'Post review' }}</button>
    </div>
  </form>
</template>
