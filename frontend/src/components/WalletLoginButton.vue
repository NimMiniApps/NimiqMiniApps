<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { useAdminAuth } from '../composables/useAdminAuth'
import AddressIdenticon from './AddressIdenticon.vue'

const { walletAddress, displayName, checking, loggingIn, error, login, logout } = useWalletAuth()
const { isAdmin, pendingCount } = useAdminAuth()

const menuOpen = ref(false)
const menuRef = ref<HTMLElement | null>(null)

function truncate(address: string): string {
  return address.length > 12 ? address.slice(0, 6) + '…' + address.slice(-4) : address
}

const headerLabel = computed(() => displayName.value || (walletAddress.value ? truncate(walletAddress.value) : ''))

function toggleMenu() {
  menuOpen.value = !menuOpen.value
}

function closeMenu() {
  menuOpen.value = false
}

async function handleLogout() {
  closeMenu()
  await logout()
}

function onDocumentClick(e: MouseEvent) {
  if (!menuOpen.value || !menuRef.value) return
  if (!menuRef.value.contains(e.target as Node)) closeMenu()
}

onMounted(() => document.addEventListener('click', onDocumentClick))
onUnmounted(() => document.removeEventListener('click', onDocumentClick))
</script>

<template>
  <div class="relative shrink-0">
    <span v-if="checking" class="text-sm text-muted">…</span>

    <div v-else-if="walletAddress" ref="menuRef" class="relative">
      <button
        type="button"
        class="flex items-center gap-1 rounded-lg border border-line bg-surface-2 py-1 pl-1 pr-1.5 transition-colors duration-200 hover:border-accent/50"
        :aria-expanded="menuOpen"
        aria-haspopup="menu"
        @click.stop="toggleMenu"
      >
        <AddressIdenticon :address="walletAddress" img-class="h-7 w-7" />
        <span
          class="hidden max-w-[8rem] truncate text-xs text-accent-ink sm:inline"
          :class="displayName ? 'font-semibold' : 'font-mono'"
        >{{ headerLabel }}</span>
        <svg
          viewBox="0 0 24 24"
          class="h-4 w-4 text-muted transition-transform duration-200"
          :class="menuOpen ? 'rotate-180' : ''"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="M6 9l6 6 6-6" />
        </svg>
      </button>

      <div
        v-if="menuOpen"
        role="menu"
        class="absolute right-0 top-full z-30 mt-2 min-w-[11rem] overflow-hidden rounded-xl border border-line bg-surface shadow-lg shadow-slate-950/10 dark:shadow-black/30"
      >
        <p class="border-b border-line px-3 py-2 text-[11px] leading-snug text-muted sm:hidden">
          <span v-if="displayName" class="block font-semibold text-accent-ink">{{ displayName }}</span>
          <span :class="displayName ? 'mt-0.5 block font-mono' : 'font-mono'">{{ truncate(walletAddress) }}</span>
        </p>
        <RouterLink
          to="/profile"
          role="menuitem"
          class="block px-4 py-3 text-sm font-semibold text-ink transition-colors hover:bg-surface-2"
          @click="closeMenu"
        >
          Profile
        </RouterLink>
        <RouterLink
          v-if="isAdmin"
          to="/admin"
          role="menuitem"
          class="flex items-center justify-between px-4 py-3 text-sm font-semibold text-ink transition-colors hover:bg-surface-2"
          @click="closeMenu"
        >
          Admin
          <span
            v-if="pendingCount > 0"
            class="ml-2 inline-flex min-w-[1.25rem] items-center justify-center rounded-full bg-amber-500/20 px-1.5 py-0.5 text-[10px] font-bold text-amber-800 dark:text-amber-200"
          >
            {{ pendingCount }}
          </span>
        </RouterLink>
        <button
          type="button"
          role="menuitem"
          class="block w-full px-4 py-3 text-left text-sm font-semibold text-muted transition-colors hover:bg-surface-2 hover:text-red-500"
          @click="handleLogout"
        >
          Log out
        </button>
      </div>
    </div>

    <button
      v-else
      class="rounded-lg border border-line bg-surface-2 px-2.5 py-1.5 text-xs font-semibold text-accent-ink transition-colors duration-200 hover:border-accent/50 disabled:opacity-50 sm:px-3 sm:text-sm"
      :disabled="loggingIn"
      @click="login"
    >
      {{ loggingIn ? 'Connecting…' : 'Connect Wallet' }}
    </button>

    <p v-if="error" class="absolute right-0 top-full z-30 mt-1 max-w-[12rem] text-right text-[10px] leading-snug text-red-500 sm:max-w-xs sm:text-xs">
      {{ error }}
    </p>
  </div>
</template>
