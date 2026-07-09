<script setup lang="ts">
import { deleteOwnAppReview, type AppReview } from '../api'
import AddressIdenticon from './AddressIdenticon.vue'

const props = defineProps<{ slug: string; reviews: AppReview[]; walletAddress: string | null }>()
const emit = defineEmits<{ deleted: [] }>()

function truncate(address: string): string {
  return address.length > 12 ? address.slice(0, 6) + '…' + address.slice(-4) : address
}

async function remove() {
  await deleteOwnAppReview(props.slug)
  emit('deleted')
}
</script>

<template>
  <ul class="flex flex-col gap-3">
    <li v-for="review in reviews" :key="review.id" class="rounded-xl border border-line bg-surface p-4">
      <div class="flex items-center justify-between">
        <span class="text-accent-ink">{{ '★'.repeat(review.rating) }}{{ '☆'.repeat(5 - review.rating) }}</span>
        <div class="flex items-center gap-1.5">
          <span class="font-mono text-xs text-muted">{{ review.display_name || truncate(review.wallet_address) }}</span>
          <AddressIdenticon :address="review.wallet_address" img-class="h-5 w-5" />
        </div>
      </div>
      <p v-if="review.body" class="mt-2 text-sm">{{ review.body }}</p>
      <button
        v-if="walletAddress === review.wallet_address"
        class="mt-2 text-xs text-muted hover:text-red-500"
        @click="remove"
      >Delete</button>
    </li>
  </ul>
</template>
