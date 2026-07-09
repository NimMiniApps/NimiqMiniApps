<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getProfile, updateProfile } from '../api'
import AddressIdenticon from '../components/AddressIdenticon.vue'

const { walletAddress, checking, refreshSession } = useWalletAuth()

const displayName = ref('')
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const saved = ref(false)

async function load() {
  if (!walletAddress.value) {
    loading.value = false
    return
  }
  try {
    const profile = await getProfile()
    displayName.value = profile.display_name ?? ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load profile'
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    await updateProfile(displayName.value.trim())
    await refreshSession()
    saved.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to save profile'
  } finally {
    saving.value = false
  }
}

watch([checking, walletAddress], () => {
  if (!checking.value) void load()
}, { immediate: true })
</script>

<template>
  <div class="mx-auto max-w-md space-y-6">
    <h1 class="text-xl font-extrabold">Profile</h1>

    <p v-if="checking || loading" class="text-sm text-muted">Loading…</p>
    <p v-else-if="!walletAddress" class="text-sm text-muted">Connect your wallet to edit your profile.</p>

    <div v-else class="space-y-4">
      <div class="flex items-center gap-3">
        <AddressIdenticon :address="walletAddress" img-class="h-14 w-14" />
        <span class="font-mono text-sm text-muted">{{ walletAddress }}</span>
      </div>

      <label class="block space-y-1">
        <span class="text-sm font-semibold">Display name</span>
        <input
          v-model="displayName"
          type="text"
          maxlength="50"
          placeholder="Not set"
          class="w-full rounded-lg border border-line bg-surface-2 p-2 text-sm"
        />
        <span class="text-xs text-muted">Must be unique. Shown on your reviews instead of your wallet address.</span>
      </label>

      <div class="flex flex-wrap items-center gap-3">
        <button
          class="rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
          :disabled="saving"
          @click="save"
        >{{ saving ? 'Saving…' : 'Save' }}</button>
        <RouterLink to="/my-apps" class="text-sm font-semibold text-accent-ink hover:underline">My apps</RouterLink>
        <span v-if="saved" class="text-sm text-muted">Saved.</span>
        <span v-if="error" class="text-sm text-red-500">{{ error }}</span>
      </div>
    </div>
  </div>
</template>
