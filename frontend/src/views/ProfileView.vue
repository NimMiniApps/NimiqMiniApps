<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getProfile, updateProfile } from '../api'
import { nimConnectPublicUrl, resolveNimConnectHandle } from '../utils/nimconnect'
import { useI18n } from '../composables/useI18n'
import AddressIdenticon from '../components/AddressIdenticon.vue'

const { t } = useI18n()
const { walletAddress, checking, refreshSession } = useWalletAuth()

const displayName = ref('')
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const saved = ref(false)
const nimConnectHandle = ref<string | null>(null)

const nimConnectUrl = computed(() =>
  nimConnectHandle.value ? nimConnectPublicUrl(nimConnectHandle.value) : null,
)

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
  nimConnectHandle.value = await resolveNimConnectHandle(walletAddress.value)
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
    <h1 class="text-xl font-extrabold uppercase tracking-wide text-ink">Profile</h1>

    <p v-if="checking || loading" class="text-sm text-muted">Loading…</p>
    <p v-else-if="!walletAddress" class="text-sm text-muted">Connect your wallet to edit your profile.</p>

    <div v-else class="board space-y-4 p-5">
      <div class="flex items-center gap-3">
        <AddressIdenticon :address="walletAddress" img-class="h-14 w-14 rounded-[4px]" />
        <div class="min-w-0">
          <span class="block break-all font-mono text-sm text-board-flap-muted">{{ walletAddress }}</span>
          <a
            v-if="nimConnectUrl"
            :href="nimConnectUrl"
            target="_blank"
            rel="noopener"
            :aria-label="t('common.nimconnectProfile')"
            :title="t('common.nimconnectProfile')"
            class="mt-0.5 inline-block font-mono text-sm text-accent-ink hover:underline"
          >@{{ nimConnectHandle }}</a>
        </div>
      </div>

      <label class="block space-y-1">
        <span class="font-mono text-xs font-bold uppercase tracking-wide text-board-flap-muted">Display name</span>
        <input
          v-model="displayName"
          type="text"
          maxlength="50"
          placeholder="Not set"
          class="w-full rounded-[4px] border border-board-hairline bg-board-flap p-2 text-sm text-board-flap-ink outline-none transition-colors duration-200 placeholder:text-board-flap-muted focus:border-lamp-live"
        />
        <span class="text-xs text-board-flap-muted">Must be unique. Shown on your reviews instead of your wallet address.</span>
      </label>

      <div class="flex flex-wrap items-center gap-3">
        <button
          class="board-plate board-plate-primary px-4 py-2 text-xs disabled:opacity-50"
          :disabled="saving"
          @click="save"
        >{{ saving ? 'Saving…' : 'Save' }}</button>
        <RouterLink to="/my-apps" class="text-sm font-semibold text-accent-ink hover:underline">My apps</RouterLink>
        <span v-if="saved" class="text-sm text-board-flap-muted">Saved.</span>
        <span v-if="error" class="text-sm text-lamp-cancelled">{{ error }}</span>
      </div>
    </div>
  </div>
</template>
