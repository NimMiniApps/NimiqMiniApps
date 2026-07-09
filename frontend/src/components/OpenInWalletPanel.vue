<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import QRCode from 'qrcode'
import { useI18n } from '../composables/useI18n'

const props = defineProps<{ openUrl: string }>()
const { t } = useI18n()

const qrDataUrl = ref('')
const copied = ref(false)

async function renderQr() {
  if (!props.openUrl) return
  try {
    qrDataUrl.value = await QRCode.toDataURL(props.openUrl, {
      width: 80,
      margin: 1,
      color: { dark: '#0f172a', light: '#ffffff' },
    })
  } catch {
    qrDataUrl.value = ''
  }
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(props.openUrl)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    /* clipboard may be unavailable */
  }
}

watch(() => props.openUrl, renderQr)
onMounted(renderQr)
</script>

<template>
  <div
    class="inline-flex shrink-0 items-center gap-2.5 rounded-xl border border-line bg-surface-2/60 px-2.5 py-1.5"
    :title="t('openWallet.copyTitle')"
  >
    <img
      v-if="qrDataUrl"
      :src="qrDataUrl"
      alt="QR code to open in Nimiq Pay"
      width="56"
      height="56"
      class="rounded-md border border-line bg-white p-0.5"
      :title="t('openWallet.scanTitle')"
    />
    <p class="max-w-[9rem] text-xs leading-snug text-muted">
      <span class="font-semibold text-ink">{{ t('openWallet.scanTitle') }}</span>
      {{ t('openWallet.scanBody') }}
    </p>
    <button type="button" @click="copyLink"
      class="inline-flex h-10 cursor-pointer items-center rounded-xl border border-line bg-surface px-4 text-sm font-semibold whitespace-nowrap transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink"
      :title="t('openWallet.copyTitle')">
      {{ copied ? t('common.copied') : t('common.copyLink') }}
    </button>
  </div>
</template>
