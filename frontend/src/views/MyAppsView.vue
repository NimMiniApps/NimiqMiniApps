<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { useI18n } from '../composables/useI18n'
import {
  getMyApps,
  getAppStats,
  addAppOwner,
  removeAppOwner,
  listAppCredentials,
  createAppCredential,
  revokeAppCredential,
  type App,
  type AppStats,
  type AppCredential,
} from '../api'
import AppCard from '../components/AppCard.vue'
import EmptyState from '../components/EmptyState.vue'
import StatsSparkline from '../components/StatsSparkline.vue'
import AddressIdenticon from '../components/AddressIdenticon.vue'

const { walletAddress, checking } = useWalletAuth()
const { t } = useI18n()

const apps = ref<(App & { has_pending_revision: boolean })[]>([])
type StatsState =
  | { status: 'loading' }
  | { status: 'error' }
  | { status: 'ready'; data: AppStats }
const statsBySlug = reactive<Record<string, StatsState>>({})
const loading = ref(true)
const error = ref('')

type ManageTab = 'owners' | 'keys'
const manageSlug = ref('')
const manageTab = ref<ManageTab>('owners')

const newOwnerInput = reactive<Record<string, string>>({})
const ownerError = reactive<Record<string, string>>({})
const ownerBusy = reactive<Record<string, boolean>>({})

const credentialsBySlug = reactive<Record<string, AppCredential[]>>({})
const credLabelInput = reactive<Record<string, string>>({})
const credError = reactive<Record<string, string>>({})
const credBusy = reactive<Record<string, boolean>>({})
const issuedKeyBySlug = reactive<Record<string, string>>({})
const copiedSlug = ref('')
const credLoaded = reactive<Record<string, boolean>>({})

const listedStatuses = new Set(['approved', 'verified', 'experimental'])

function truncateAddress(address: string): string {
  return address.length > 16 ? `${address.slice(0, 7)}…${address.slice(-5)}` : address
}

function last7DaysSum(daily: AppStats['daily'], metric: 'opens' | 'views'): number {
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - 6)
  const cutoffStr = cutoff.toISOString().slice(0, 10)
  return daily
    .filter((d) => d.date >= cutoffStr)
    .reduce((sum, d) => sum + (metric === 'opens' ? d.opens : d.views), 0)
}

async function loadStats(slugs: string[]) {
  for (const slug of slugs) {
    statsBySlug[slug] = { status: 'loading' }
  }
  await Promise.all(slugs.map(async (slug) => {
    try {
      statsBySlug[slug] = { status: 'ready', data: await getAppStats(slug) }
    } catch {
      statsBySlug[slug] = { status: 'error' }
    }
  }))
}

async function load() {
  if (!walletAddress.value) {
    loading.value = false
    return
  }
  loading.value = true
  error.value = ''
  try {
    apps.value = await getMyApps()
    for (const app of apps.value) {
      statsBySlug[app.slug] = { status: 'loading' }
    }
    void loadStats(apps.value.map((a) => a.slug))
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('myApps.errorBody')
  } finally {
    loading.value = false
  }
}

async function ensureCredentials(slug: string, status: string) {
  if (!listedStatuses.has(status) || credLoaded[slug]) return
  await loadCredentials(slug)
  credLoaded[slug] = true
}

async function openManage(slug: string, tab: ManageTab, status: string) {
  if (manageSlug.value === slug && manageTab.value === tab) {
    manageSlug.value = ''
    return
  }
  manageSlug.value = slug
  manageTab.value = tab
  if (tab === 'keys') await ensureCredentials(slug, status)
}

async function handleAddOwner(slug: string) {
  const wallet = (newOwnerInput[slug] || '').trim()
  if (!wallet) return
  ownerBusy[slug] = true
  ownerError[slug] = ''
  try {
    await addAppOwner(slug, wallet)
    newOwnerInput[slug] = ''
    await load()
  } catch (err) {
    ownerError[slug] = err instanceof Error ? err.message : 'Failed to add owner'
  } finally {
    ownerBusy[slug] = false
  }
}

async function handleRemoveOwner(slug: string, wallet: string) {
  ownerBusy[slug] = true
  ownerError[slug] = ''
  try {
    await removeAppOwner(slug, wallet)
    await load()
  } catch (err) {
    ownerError[slug] = err instanceof Error ? err.message : 'Failed to remove owner'
  } finally {
    ownerBusy[slug] = false
  }
}

async function loadCredentials(slug: string) {
  credBusy[slug] = true
  credError[slug] = ''
  try {
    const res = await listAppCredentials(slug)
    credentialsBySlug[slug] = res.credentials
  } catch (err) {
    credError[slug] = err instanceof Error ? err.message : 'Failed to load credentials'
  } finally {
    credBusy[slug] = false
  }
}

async function handleIssueCredential(slug: string) {
  credBusy[slug] = true
  credError[slug] = ''
  issuedKeyBySlug[slug] = ''
  try {
    const created = await createAppCredential(slug, (credLabelInput[slug] || '').trim())
    issuedKeyBySlug[slug] = created.key
    credLabelInput[slug] = ''
    credLoaded[slug] = true
    await loadCredentials(slug)
  } catch (err) {
    credError[slug] = err instanceof Error ? err.message : 'Failed to issue credential'
  } finally {
    credBusy[slug] = false
  }
}

async function handleRevokeCredential(slug: string, id: number) {
  credBusy[slug] = true
  credError[slug] = ''
  try {
    await revokeAppCredential(slug, id)
    await loadCredentials(slug)
  } catch (err) {
    credError[slug] = err instanceof Error ? err.message : 'Failed to revoke credential'
  } finally {
    credBusy[slug] = false
  }
}

async function copyIssuedKey(slug: string) {
  const key = issuedKeyBySlug[slug]
  if (!key) return
  try {
    await navigator.clipboard.writeText(key)
    copiedSlug.value = slug
    window.setTimeout(() => {
      if (copiedSlug.value === slug) copiedSlug.value = ''
    }, 2000)
  } catch {
    credError[slug] = 'Could not copy to clipboard'
  }
}

function formatUsed(cred: AppCredential): string {
  if (cred.revoked_at) return t('myApps.credentials.revoked')
  if (!cred.last_used_at) return t('myApps.credentials.neverUsed')
  return `${t('myApps.credentials.lastUsed')}: ${new Date(cred.last_used_at).toLocaleString()}`
}

function activeCredentialCount(slug: string): number {
  return (credentialsBySlug[slug] || []).filter((c) => !c.revoked_at).length
}

watch([checking, walletAddress], () => {
  if (!checking.value) void load()
}, { immediate: true })
</script>

<template>
  <div class="space-y-5">
    <div>
      <h1 class="text-2xl font-extrabold">My apps</h1>
      <p class="mt-1 text-sm text-muted">Apps linked to your wallet — edit listings or check review status.</p>
    </div>

    <p v-if="checking || loading" class="text-sm text-muted">{{ t('common.loading') }}</p>

    <EmptyState
      v-else-if="!walletAddress"
      title="My apps"
      :description="t('myApps.connectWallet')"
    />

    <EmptyState
      v-else-if="error"
      :title="t('myApps.errorTitle')"
      :description="t('myApps.errorBody')"
      variant="error"
    >
      <template #actions>
        <button
          type="button"
          class="cursor-pointer board-plate board-plate-ghost px-5 py-2.5 text-sm"
          @click="load"
        >
          {{ t('common.retry') }}
        </button>
      </template>
    </EmptyState>

    <EmptyState
      v-else-if="apps.length === 0"
      :title="t('myApps.emptyTitle')"
      :description="t('myApps.emptyBody')"
    >
      <template #actions>
        <RouterLink
          to="/submit"
          class="cursor-pointer board-plate board-plate-primary px-5 py-2.5 text-sm"
        >
          {{ t('nav.submit') }}
        </RouterLink>
        <RouterLink
          to="/admin"
          class="cursor-pointer board-plate board-plate-ghost px-5 py-2.5 text-sm"
        >
          {{ t('nav.admin') }}
        </RouterLink>
      </template>
    </EmptyState>

    <div v-else class="grid items-stretch gap-4 sm:grid-cols-2">
      <div v-for="app in apps" :key="app.id" class="flex min-h-0 flex-col gap-2">
        <AppCard
          class="h-full min-h-0 flex-1"
          :app="app"
          owned
          :pending-update="app.has_pending_revision"
          show-manage-actions
        />

        <div class="board shrink-0 space-y-3 p-3">
          <div v-if="statsBySlug[app.slug]?.status === 'loading'" class="grid min-h-[5.5rem] grid-cols-2 gap-3">
            <div v-for="n in 2" :key="n" class="space-y-1">
              <div class="h-3 w-12 animate-pulse rounded bg-board-hairline/60" />
              <div class="h-6 w-10 animate-pulse rounded bg-board-hairline/60" />
              <div class="h-3 w-20 animate-pulse rounded bg-board-hairline/60" />
              <div class="mt-1 h-8 w-full animate-pulse rounded bg-board-hairline/40" />
            </div>
          </div>
          <div
            v-else-if="statsBySlug[app.slug]?.status === 'ready'"
            class="grid min-h-[5.5rem] grid-cols-2 gap-3"
          >
            <div class="flex flex-col">
              <p class="font-mono text-[10px] font-bold uppercase tracking-wide text-board-flap-muted">{{ t('myApps.stats.opens') }}</p>
              <p class="text-lg font-bold text-board-flap-ink">{{ statsBySlug[app.slug].data.totals.opens.toLocaleString() }}</p>
              <p class="text-xs text-board-flap-muted">
                {{ t('myApps.stats.last7Days') }}: {{ last7DaysSum(statsBySlug[app.slug].data.daily, 'opens').toLocaleString() }}
              </p>
              <div class="mt-auto pt-1">
                <StatsSparkline :daily="statsBySlug[app.slug].data.daily" metric="opens" class="w-full" />
              </div>
            </div>
            <div class="flex flex-col">
              <p class="font-mono text-[10px] font-bold uppercase tracking-wide text-board-flap-muted">{{ t('myApps.stats.views') }}</p>
              <p class="text-lg font-bold text-board-flap-ink">{{ statsBySlug[app.slug].data.totals.views.toLocaleString() }}</p>
              <p class="text-xs text-board-flap-muted">
                {{ t('myApps.stats.last7Days') }}: {{ last7DaysSum(statsBySlug[app.slug].data.daily, 'views').toLocaleString() }}
              </p>
              <div class="mt-auto pt-1">
                <StatsSparkline :daily="statsBySlug[app.slug].data.daily" metric="views" class="w-full" />
              </div>
            </div>
          </div>
          <p v-else class="min-h-[5.5rem] text-xs text-board-flap-muted">Stats unavailable</p>

          <div class="flex gap-1 border-t border-board-hairline pt-2">
            <button
              type="button"
              class="flex-1 rounded-[3px] px-2 py-1.5 text-[11px] font-bold uppercase tracking-wide transition-colors"
              :class="manageSlug === app.slug && manageTab === 'owners'
                ? 'bg-board-flap text-board-flap-ink'
                : 'text-board-flap-muted hover:bg-board-flap-hover hover:text-board-flap-ink'"
              @click="openManage(app.slug, 'owners', app.status)"
            >
              {{ t('myApps.manage.owners') }}
              <span class="ml-1 font-mono text-board-flap-muted">({{ app.owner_wallet_addresses.length }})</span>
            </button>
            <button
              type="button"
              class="flex-1 rounded-[3px] px-2 py-1.5 text-[11px] font-bold uppercase tracking-wide transition-colors"
              :class="manageSlug === app.slug && manageTab === 'keys'
                ? 'bg-board-flap text-board-flap-ink'
                : 'text-board-flap-muted hover:bg-board-flap-hover hover:text-board-flap-ink'"
              @click="openManage(app.slug, 'keys', app.status)"
            >
              {{ t('myApps.manage.keys') }}
              <span v-if="credLoaded[app.slug]" class="ml-1 font-mono text-board-flap-muted">({{ activeCredentialCount(app.slug) }})</span>
            </button>
          </div>

          <div v-if="manageSlug === app.slug" class="space-y-3 border-t border-board-hairline pt-3">
            <template v-if="manageTab === 'owners'">
              <ul class="space-y-1.5">
                <li
                  v-for="wallet in app.owner_wallet_addresses"
                  :key="wallet"
                  class="flex items-center gap-2 rounded-[3px] bg-board-flap px-2 py-1.5"
                >
                  <AddressIdenticon :address="wallet" img-class="h-6 w-6 rounded-[3px]" />
                  <span class="min-w-0 flex-1 truncate font-mono text-xs text-board-flap-ink" :title="wallet">
                    {{ truncateAddress(wallet) }}
                  </span>
                  <button
                    type="button"
                    :disabled="ownerBusy[app.slug] || app.owner_wallet_addresses.length <= 1"
                    class="board-plate board-plate-ghost shrink-0 px-2 py-1 text-[10px] disabled:cursor-default disabled:opacity-40"
                    @click="handleRemoveOwner(app.slug, wallet)"
                  >
                    {{ t('myApps.manage.remove') }}
                  </button>
                </li>
              </ul>
              <div class="flex gap-2">
                <input
                  v-model="newOwnerInput[app.slug]"
                  :placeholder="t('myApps.manage.ownerPlaceholder')"
                  class="min-w-0 flex-1 rounded-[3px] border border-board-hairline bg-board-flap px-2 py-1.5 font-mono text-xs text-board-flap-ink outline-none placeholder:text-board-flap-muted focus:border-lamp-live"
                />
                <button
                  type="button"
                  :disabled="ownerBusy[app.slug]"
                  class="board-plate board-plate-primary shrink-0 px-3 py-1.5 text-[11px] disabled:opacity-50"
                  @click="handleAddOwner(app.slug)"
                >
                  {{ t('myApps.manage.add') }}
                </button>
              </div>
              <p v-if="ownerError[app.slug]" class="text-xs text-lamp-cancelled">{{ ownerError[app.slug] }}</p>
            </template>

            <template v-else>
              <p class="text-xs text-board-flap-muted">
                <RouterLink
                  :to="t('myApps.credentials.helpLink')"
                  class="font-semibold text-accent-ink hover:underline"
                >{{ t('myApps.credentials.help') }}</RouterLink>
              </p>
              <p v-if="!listedStatuses.has(app.status)" class="text-xs text-board-flap-muted">
                {{ t('myApps.credentials.notListed') }}
              </p>
              <template v-else>
                <p v-if="credBusy[app.slug] && !credentialsBySlug[app.slug]" class="text-xs text-board-flap-muted">
                  {{ t('common.loading') }}
                </p>
                <ul v-else-if="(credentialsBySlug[app.slug] || []).length" class="space-y-1.5">
                  <li
                    v-for="cred in credentialsBySlug[app.slug]"
                    :key="cred.id"
                    class="flex items-start justify-between gap-2 rounded-[3px] bg-board-flap px-2 py-1.5"
                  >
                    <div class="min-w-0">
                      <p class="truncate font-mono text-xs text-board-flap-ink">{{ cred.key_prefix }}…</p>
                      <p class="text-[11px] text-board-flap-muted">
                        <span v-if="cred.label">{{ cred.label }} · </span>{{ formatUsed(cred) }}
                      </p>
                    </div>
                    <button
                      v-if="!cred.revoked_at"
                      type="button"
                      :disabled="credBusy[app.slug]"
                      class="board-plate board-plate-ghost shrink-0 px-2 py-1 text-[10px] disabled:opacity-40"
                      @click="handleRevokeCredential(app.slug, cred.id)"
                    >
                      {{ t('myApps.credentials.revoke') }}
                    </button>
                  </li>
                </ul>
                <p v-else class="text-xs text-board-flap-muted">{{ t('myApps.credentials.empty') }}</p>

                <div
                  v-if="issuedKeyBySlug[app.slug]"
                  class="space-y-2 rounded-[3px] border border-lamp-live/40 bg-board-flap px-2.5 py-2"
                >
                  <p class="text-[11px] font-bold uppercase tracking-wide text-accent-ink">
                    {{ t('myApps.credentials.createdOnce') }}
                  </p>
                  <p class="break-all font-mono text-[11px] text-board-flap-ink">{{ issuedKeyBySlug[app.slug] }}</p>
                  <button
                    type="button"
                    class="board-plate board-plate-primary px-3 py-1.5 text-[11px]"
                    @click="copyIssuedKey(app.slug)"
                  >
                    {{ copiedSlug === app.slug ? t('myApps.credentials.copied') : t('myApps.credentials.copyKey') }}
                  </button>
                </div>

                <div class="flex gap-2">
                  <input
                    v-model="credLabelInput[app.slug]"
                    :placeholder="t('myApps.credentials.labelPlaceholder')"
                    class="min-w-0 flex-1 rounded-[3px] border border-board-hairline bg-board-flap px-2 py-1.5 text-xs text-board-flap-ink outline-none placeholder:text-board-flap-muted focus:border-lamp-live"
                  />
                  <button
                    type="button"
                    :disabled="credBusy[app.slug]"
                    class="board-plate board-plate-primary shrink-0 px-3 py-1.5 text-[11px] disabled:opacity-50"
                    @click="handleIssueCredential(app.slug)"
                  >
                    {{ t('myApps.credentials.issue') }}
                  </button>
                </div>
              </template>
              <p v-if="credError[app.slug]" class="text-xs text-lamp-cancelled">{{ credError[app.slug] }}</p>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
