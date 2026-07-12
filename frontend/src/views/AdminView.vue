<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  APP_CATEGORIES, APP_RELEASE_STAGES, adminListApps, adminCreateApp, adminUpdateApp, adminDeleteApp, adminSetStatus, adminCheckDomains,
  adminListRevisions, adminApproveRevision, adminRejectRevision, adminSearchUsers, adminAddAppOwner, adminRemoveAppOwner,
  type App, type RevisionReviewItem, type AdminUserResult,
} from '../api'
import { useWalletAuth } from '../composables/useWalletAuth'
import StatusBadge from '../components/StatusBadge.vue'
import ReleaseStageBadge from '../components/ReleaseStageBadge.vue'
import TokenMultiSelect from '../components/TokenMultiSelect.vue'
import { formatMediaLines, parseMediaLines } from '../utils/media'
import { formatSocialLines, parseSocialLines } from '../utils/socials'
import { diffRevision } from '../utils/revisionDiff'
import { normalizeDomain } from '../utils/domain'

const route = useRoute()
const router = useRouter()
const { isAdmin: walletIsAdmin } = useWalletAuth()
const token = ref(localStorage.getItem('admin_token') || '')
const apps = ref<App[]>([])
const pendingRevisions = ref<RevisionReviewItem[]>([])
const error = ref('')
const notice = ref('')
const reordering = ref(false)
const checkingDomains = ref(false)
const rejectTarget = ref<App | null>(null)
const rejectNote = ref('')
type AppSortKey = 'name' | 'total_opens' | 'total_views'
const sortKey = ref<AppSortKey>('name')
const sortAsc = ref(true)

function toggleSort(key: AppSortKey) {
  if (sortKey.value === key) {
    sortAsc.value = !sortAsc.value
  } else {
    sortKey.value = key
    sortAsc.value = key === 'name'
  }
}

const sortedApps = computed(() => {
  const list = [...apps.value]
  const dir = sortAsc.value ? 1 : -1
  list.sort((a, b) => {
    if (sortKey.value === 'name') {
      return dir * a.name.localeCompare(b.name)
    }
    const av = sortKey.value === 'total_opens' ? (a.total_opens ?? 0) : (a.total_views ?? 0)
    const bv = sortKey.value === 'total_opens' ? (b.total_opens ?? 0) : (b.total_views ?? 0)
    if (av !== bv) return dir * (av - bv)
    return a.name.localeCompare(b.name)
  })
  return list
})

const emptyForm = {
  slug: '', name: '', domain: '', category: '', developer_slug: '', developer_name: '',
  tagline: '', description: '', long_description: '', tags: '', assets: 'NIM', reward_assets: '', status: 'submitted',
  release_stage: 'released', featured: false, featured_order: 0,
  website_url: '', github_url: '', icon_url: '', banner_url: '', media: '', socials: '',
  submitter_contact: '',
}
const form = reactive({ ...emptyForm })
const editingSlug = ref('') // '' = create mode
const showForm = ref(false)
const developerQuery = ref('')
const developerResults = ref<AdminUserResult[]>([])
let developerSearchTimer: ReturnType<typeof setTimeout> | undefined

function onDeveloperQueryInput() {
  clearTimeout(developerSearchTimer)
  developerSearchTimer = setTimeout(async () => {
    const q = developerQuery.value.trim()
    if (!q) {
      developerResults.value = []
      return
    }
    try {
      developerResults.value = await adminSearchUsers(q)
    } catch {
      developerResults.value = []
    }
  }, 250)
}

const currentOwners = ref<string[]>([])
const ownerBusy = ref(false)

async function addOwnerFromPicker(user: AdminUserResult) {
  if (!user.display_name?.trim()) {
    error.value = 'This user must set a display name on their profile before they can own an app.'
    return
  }
  error.value = ''
  ownerBusy.value = true
  try {
    await adminAddAppOwner(editingSlug.value, user.wallet_address)
    if (!currentOwners.value.includes(user.wallet_address)) currentOwners.value.push(user.wallet_address)
    if (!form.developer_name) form.developer_name = user.display_name
    if (!form.developer_slug) form.developer_slug = slugify(user.display_name)
    developerQuery.value = ''
    developerResults.value = []
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    ownerBusy.value = false
  }
}

async function removeOwnerFromApp(wallet: string) {
  ownerBusy.value = true
  error.value = ''
  try {
    await adminRemoveAppOwner(editingSlug.value, wallet)
    currentOwners.value = currentOwners.value.filter((w) => w !== wallet)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    ownerBusy.value = false
  }
}

const slugify = (s: string) =>
  s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '')

function saveToken() {
  localStorage.setItem('admin_token', token.value)
  notice.value = 'Token saved.'
  load()
}

async function load() {
  error.value = ''
  try {
    const [listed, revisions] = await Promise.all([
      adminListApps(),
      adminListRevisions().catch(() => [] as RevisionReviewItem[]),
    ])
    apps.value = listed
    pendingRevisions.value = revisions
    openEditFromQuery()
  } catch (e) {
    error.value = (e as Error).message
  }
}

function openEditFromQuery() {
  const slug = route.query.edit
  if (typeof slug !== 'string' || !slug) return
  const app = apps.value.find((entry) => entry.slug === slug)
  if (!app) return
  startEdit(app)
  router.replace({ query: { ...route.query, edit: undefined } })
}

function startCreate() {
  Object.assign(form, emptyForm)
  developerQuery.value = ''
  developerResults.value = []
  currentOwners.value = []
  editingSlug.value = ''
  showForm.value = true
}

function startEdit(app: App) {
  Object.assign(form, {
    slug: app.slug, name: app.name, domain: app.domain, category: app.category,
    developer_slug: app.developer_slug, developer_name: app.developer_name,
    tagline: app.tagline, description: app.description, long_description: app.long_description || '',
    tags: app.tags.join(', '), assets: app.assets.join(', '), reward_assets: app.reward_assets.join(', '),
    status: app.status, release_stage: app.release_stage, featured: app.featured,
    featured_order: app.featured_order ?? 0,
    website_url: app.website_url || '', github_url: app.github_url || '',
    icon_url: app.icon_url || '', banner_url: app.banner_url || '',
    media: formatMediaLines(app.media),
    socials: formatSocialLines(app.socials),
    submitter_contact: app.submitter_contact || '',
  })
  currentOwners.value = [...app.owner_wallet_addresses]
  developerQuery.value = ''
  developerResults.value = []
  editingSlug.value = app.slug
  showForm.value = true
}

const csv = (s: string) => s.split(',').map((x) => x.trim()).filter(Boolean)

async function submit() {
  error.value = ''
  const payload = {
    ...form,
    domain: normalizeDomain(form.domain),
    tags: csv(form.tags),
    assets: csv(form.assets),
    reward_assets: csv(form.reward_assets),
    media: parseMediaLines(form.media),
    socials: parseSocialLines(form.socials),
    website_url: form.website_url || null,
    github_url: form.github_url || null,
    icon_url: form.icon_url || null,
    banner_url: form.banner_url || null,
  }
  try {
    if (editingSlug.value) {
      await adminUpdateApp(editingSlug.value, payload)
      notice.value = `Updated ${form.name}.`
    } else {
      await adminCreateApp(payload)
      notice.value = `Created ${form.name}.`
    }
    showForm.value = false
    load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

async function remove(app: App) {
  if (!confirm(`Delete ${app.name}? This cannot be undone.`)) return
  try {
    await adminDeleteApp(app.slug)
    notice.value = `Deleted ${app.name}.`
    load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

async function setStatus(app: App, action: 'verify' | 'approve' | 'reject') {
  try {
    await adminSetStatus(app.slug, action)
    load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

function startReject(app: App) {
  rejectTarget.value = app
  rejectNote.value = ''
}

function cancelReject() {
  rejectTarget.value = null
  rejectNote.value = ''
}

async function confirmReject() {
  if (!rejectTarget.value) return
  const app = rejectTarget.value
  try {
    await adminSetStatus(app.slug, 'reject', rejectNote.value.trim())
    notice.value = `Rejected ${app.name}.`
    cancelReject()
    load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

const featuredApps = computed(() =>
  apps.value
    .filter((app) => app.featured)
    .sort((a, b) => {
      const ao = a.featured_order > 0 ? a.featured_order : Number.MAX_SAFE_INTEGER
      const bo = b.featured_order > 0 ? b.featured_order : Number.MAX_SAFE_INTEGER
      if (ao !== bo) return ao - bo
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    }),
)

const pendingApps = computed(() =>
  apps.value.filter((app) => app.status === 'submitted'),
)

const unreachableApps = computed(() =>
  apps.value.filter((app) => app.domain_reachable === false),
)

function formatCheckedAt(iso: string | null) {
  if (!iso) return 'never'
  return new Date(iso).toLocaleString()
}

async function recheckDomains() {
  checkingDomains.value = true
  error.value = ''
  try {
    await adminCheckDomains()
    notice.value = 'Domain health check complete.'
    await load()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    checkingDomains.value = false
  }
}

async function approveRevision(item: RevisionReviewItem) {
  try {
    await adminApproveRevision(item.revision.id)
    notice.value = `Approved update for ${item.revision.name}.`
    await load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

async function rejectRevision(item: RevisionReviewItem) {
  try {
    await adminRejectRevision(item.revision.id)
    notice.value = `Rejected update for ${item.revision.name}.`
    await load()
  } catch (e) {
    error.value = (e as Error).message
  }
}

function revisionChanges(item: RevisionReviewItem) {
  return diffRevision(item.current, item.revision)
}

async function moveFeatured(app: App, direction: -1 | 1) {
  const ordered = featuredApps.value
  const i = ordered.findIndex((entry) => entry.id === app.id)
  const j = i + direction
  if (i < 0 || j < 0 || j >= ordered.length) return

  reordering.value = true
  error.value = ''
  try {
    const a = ordered[i]
    const b = ordered[j]
    const aOrder = a.featured_order > 0 ? a.featured_order : (i + 1) * 10
    const bOrder = b.featured_order > 0 ? b.featured_order : (j + 1) * 10
    await Promise.all([
      adminUpdateApp(a.slug, { ...a, featured_order: bOrder }),
      adminUpdateApp(b.slug, { ...b, featured_order: aOrder }),
    ])
    notice.value = `Moved ${app.name} ${direction < 0 ? 'up' : 'down'} in featured order.`
    await load()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    reordering.value = false
  }
}

onMounted(load)

const fields: [keyof typeof emptyForm, string, boolean, string?][] = [
  ['slug', 'Slug (lowercase, url-safe)', true],
  ['name', 'Name', true],
  ['domain', 'Domain (no https://)', true],
  ['submitter_contact', 'Submitter contact (private)', false],
  ['tagline', 'Tagline', true],
  ['tags', 'Tags (comma-separated)', false],
  ['website_url', 'Website URL', false],
  ['github_url', 'GitHub URL', false],
  ['icon_url', 'Icon URL', false],
  ['banner_url', 'Banner URL', false],
]
</script>

<template>
  <div class="space-y-5">
    <h1 class="text-2xl font-extrabold">Admin</h1>

    <p v-if="walletIsAdmin" class="text-sm text-muted">
      Signed in with an admin wallet. Catalog actions use your wallet session.
    </p>

    <!-- bearer token fallback (MCP, scripts, break-glass) -->
    <details v-if="!walletIsAdmin" class="rounded-xl border border-line bg-surface p-4">
      <summary class="cursor-pointer text-sm font-semibold text-muted">Sign in with admin token</summary>
      <div class="mt-3 flex gap-2">
        <input v-model="token" type="password" placeholder="Admin token"
          class="flex-1 rounded-xl border border-line bg-surface-2 px-4 py-2.5 placeholder:text-muted/60 focus:border-accent outline-none" />
        <button @click="saveToken" class="rounded-[500px] nq-primary px-4 py-2.5 font-bold text-white">
          Save
        </button>
      </div>
    </details>

    <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-600 dark:text-red-600 dark:text-red-300">{{ error }}</p>
    <p v-if="notice" class="rounded-xl bg-emerald-500/15 p-4 text-emerald-700 dark:text-emerald-700 dark:text-emerald-300">{{ notice }}</p>

    <div v-if="rejectTarget" class="space-y-3 rounded-2xl border border-red-500/30 bg-red-500/10 p-5">
      <div>
        <h2 class="font-bold">Reject {{ rejectTarget.name }}</h2>
        <p class="text-sm text-muted">Optional note shown to the submitter on the status page.</p>
      </div>
      <textarea
        v-model="rejectNote"
        rows="3"
        maxlength="2000"
        placeholder="e.g. Domain does not load a mini app, or listing duplicates an existing app."
        class="w-full rounded-xl border border-line bg-surface px-4 py-3 text-sm placeholder:text-muted/60 focus:border-accent outline-none"
      />
      <div class="flex flex-wrap gap-2">
        <button type="button" class="rounded-lg bg-red-500/20 px-4 py-2 text-sm font-semibold text-red-600 dark:text-red-300 hover:bg-red-500/30" @click="confirmReject">
          Confirm reject
        </button>
        <button type="button" class="rounded-xl border border-line px-4 py-2 text-sm font-semibold hover:bg-surface-2" @click="cancelReject">
          Cancel
        </button>
      </div>
    </div>

    <button v-if="!showForm" @click="startCreate"
      class="rounded-xl border border-accent/50 px-4 py-2.5 font-bold text-accent-ink hover:bg-accent/10">
      + New app
    </button>

    <button v-if="!showForm" type="button" @click="recheckDomains" :disabled="checkingDomains"
      class="rounded-xl border border-line px-4 py-2.5 text-sm font-semibold hover:bg-surface-2 disabled:opacity-50">
      {{ checkingDomains ? 'Checking domains…' : 'Recheck all domains' }}
    </button>

    <!-- create / edit form -->
    <form v-if="showForm" @submit.prevent="submit" class="space-y-3 rounded-2xl border border-line bg-surface p-5">
      <h2 class="font-bold">{{ editingSlug ? `Edit ${editingSlug}` : 'New app' }}</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <label v-for="[key, label, required, help] in fields" :key="key" class="text-sm">
          <span class="mb-1 block text-muted">{{ label }}{{ required ? ' *' : '' }}</span>
          <input v-model="(form as any)[key]" :required="required"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
          <span v-if="help" class="mt-1 block text-xs leading-snug text-muted">{{ help }}</span>
        </label>
        <TokenMultiSelect
          v-model="form.assets"
          label="Assets"
          help="Tokens the app uses, accepts, reads, or supports."
        />
        <TokenMultiSelect
          v-model="form.reward_assets"
          label="Reward assets"
          help="Moderator-reviewed claim: only select tokens users can actually receive from the app, such as daily rewards, leaderboard prizes, payouts, or tips. Leave empty if the app only uses or accepts the token."
        />

        <div class="space-y-3 rounded-xl border border-line bg-surface-2/50 p-4 sm:col-span-2">
          <div>
            <h3 class="text-sm font-bold">Developer</h3>
            <p class="mt-0.5 text-xs text-muted">
              Link a <strong>wallet owner</strong> to grant My apps access. Catalog name and slug are taken from their profile automatically.
              For legacy listings without a wallet, enter a public name and slug manually below.
            </p>
          </div>

          <div v-if="editingSlug" class="space-y-2">
            <span class="mb-1 block text-sm font-semibold text-muted">Owner wallets</span>
            <ul v-if="currentOwners.length" class="space-y-1">
              <li v-for="wallet in currentOwners" :key="wallet" class="flex items-center justify-between gap-2 rounded-lg bg-surface px-2 py-1.5 text-sm">
                <span class="truncate font-mono text-xs">{{ wallet }}</span>
                <button type="button" :disabled="ownerBusy"
                  class="shrink-0 text-xs font-semibold text-red-600 hover:underline disabled:cursor-default disabled:opacity-40 dark:text-red-400"
                  @click="removeOwnerFromApp(wallet)">
                  Remove
                </button>
              </li>
            </ul>
            <p v-else class="text-xs text-muted">Unclaimed — only admins can edit this listing until a wallet is added.</p>
          </div>
          <label class="relative block text-sm">
            <span class="mb-1 block font-semibold text-muted">Add owner wallet</span>
            <input v-model="developerQuery" @input="onDeveloperQueryInput"
              :disabled="!editingSlug"
              placeholder="Search by display name or wallet address — pick a result to add"
              class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 focus:border-accent disabled:opacity-50" />
            <p v-if="!editingSlug" class="mt-1 text-xs text-muted">Save this app first, then add owner wallets.</p>
            <ul v-if="developerResults.length" class="absolute z-10 mt-1 max-h-48 w-full overflow-y-auto rounded-lg border border-line bg-surface shadow-lg">
              <li v-for="user in developerResults" :key="user.wallet_address"
                @click="addOwnerFromPicker(user)"
                class="cursor-pointer px-3 py-2 text-sm hover:bg-surface-2"
                :class="{ 'opacity-50': !user.display_name?.trim() }">
                {{ user.display_name ?? 'No display name' }}
                <span class="block font-mono text-xs text-muted">{{ user.wallet_address }}</span>
              </li>
            </ul>
            <p v-else-if="developerQuery.trim()" class="mt-1 text-xs text-muted">
              No matching wallets — the user must log in at least once before you can add them.
            </p>
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="text-sm">
              <span class="mb-1 block text-muted">Public developer name *</span>
              <input v-model="form.developer_name" required
                placeholder="Shown on listings"
                class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
            </label>
            <label class="text-sm">
              <span class="mb-1 block text-muted">Public developer slug *</span>
              <input v-model="form.developer_slug" required
                placeholder="Used in developer URLs"
                class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
            </label>
          </div>
        </div>
        <label class="text-sm">
          <span class="mb-1 block text-muted">Category *</span>
          <select v-model="form.category" required class="w-full cursor-pointer rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none">
            <option value="" disabled>Select a category</option>
            <option v-for="category in APP_CATEGORIES" :key="category" :value="category">{{ category }}</option>
          </select>
        </label>
        <label class="text-sm">
          <span class="mb-1 block text-muted">Status</span>
          <select v-model="form.status" class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2">
            <option v-for="s in ['submitted', 'approved', 'verified', 'experimental', 'rejected']" :key="s" :value="s">{{ s }}</option>
          </select>
        </label>
        <label class="text-sm">
          <span class="mb-1 block text-muted">Release stage</span>
          <select v-model="form.release_stage" class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2">
            <option v-for="stage in APP_RELEASE_STAGES" :key="stage" :value="stage">{{ stage }}</option>
          </select>
        </label>
        <label class="flex items-end gap-2 pb-2 text-sm">
          <input v-model="form.featured" type="checkbox" class="h-4 w-4 accent-[#1F74FF]" />
          Featured
        </label>
        <label v-if="form.featured" class="text-sm">
          <span class="mb-1 block text-muted">Featured order</span>
          <input v-model.number="form.featured_order" type="number" min="0" step="1"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
          <span class="mt-1 block text-xs text-muted">Lower numbers appear first. 0 = auto (by date).</span>
        </label>
      </div>
      <label class="block text-sm">
        <span class="mb-1 block text-muted">Short description</span>
        <textarea v-model="form.description" rows="2"
          class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none"></textarea>
      </label>
      <label class="block text-sm">
        <span class="mb-1 block text-muted">Full description</span>
        <textarea v-model="form.long_description" rows="5"
          class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none"></textarea>
        <span class="mt-1 block text-xs text-muted">Markdown supported in full description.</span>
      </label>
      <label class="block text-sm">
        <span class="mb-1 block text-muted">Screenshots &amp; video (one URL per line)</span>
        <textarea v-model="form.media" rows="4"
          class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none"></textarea>
      </label>
      <label class="block text-sm">
        <span class="mb-1 block text-muted">Social links (platform URL per line)</span>
        <textarea v-model="form.socials" rows="3" placeholder="twitter https://x.com/myapp&#10;discord: https://discord.gg/abc"
          class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none"></textarea>
        <span class="mt-1 block text-xs text-muted">twitter, discord, telegram, bluesky, instagram, youtube, linkedin, mastodon, reddit, tiktok</span>
      </label>
      <div class="flex gap-2">
        <button type="submit" class="rounded-[500px] nq-primary px-5 py-2 font-bold text-white">
          {{ editingSlug ? 'Save changes' : 'Create app' }}
        </button>
        <button type="button" @click="showForm = false" class="rounded-xl border border-line px-5 py-2 font-semibold hover:bg-surface-2">
          Cancel
        </button>
      </div>
    </form>

    <section v-if="!showForm && pendingApps.length" class="space-y-3 rounded-2xl border border-amber-500/30 bg-amber-500/10 p-5">
      <div>
        <h2 class="font-bold">Pending review ({{ pendingApps.length }})</h2>
        <p class="text-sm text-muted">New submissions waiting for approval.</p>
      </div>
      <div class="space-y-2">
        <div v-for="app in pendingApps" :key="app.id"
          class="flex flex-col gap-3 rounded-xl border border-line bg-surface px-3 py-3 sm:flex-row sm:items-center">
          <div class="min-w-0 flex-1">
            <p class="font-semibold">{{ app.name }}</p>
            <p class="truncate text-xs text-muted">{{ app.slug }} · {{ app.domain }}</p>
            <p v-if="app.submitter_contact" class="mt-1 text-xs text-muted">
              Contact: <span class="font-medium text-ink">{{ app.submitter_contact }}</span>
            </p>
          </div>
          <div class="flex flex-wrap gap-1.5 text-xs font-semibold">
            <button @click="setStatus(app, 'approve')" class="rounded-lg bg-sky-500/20 px-2.5 py-1.5 text-sky-700 dark:text-sky-300 hover:bg-sky-500/30">Approve</button>
            <button @click="startReject(app)" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-600 dark:text-red-300 hover:bg-red-500/30">Reject</button>
            <button @click="startEdit(app)" class="rounded-lg bg-surface-2 px-2.5 py-1.5 hover:bg-line">Edit</button>
          </div>
        </div>
      </div>
    </section>

    <section v-if="!showForm && pendingRevisions.length" class="space-y-3 rounded-2xl border border-sky-500/30 bg-sky-500/10 p-5">
      <div>
        <h2 class="font-bold">Pending updates ({{ pendingRevisions.length }})</h2>
        <p class="text-sm text-muted">Author-requested changes waiting for approval. Live listings stay unchanged until you approve.</p>
      </div>
      <div v-for="item in pendingRevisions" :key="item.revision.id" class="space-y-3 rounded-xl border border-line bg-surface p-4">
        <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <p class="font-semibold">{{ item.revision.name }} <span class="font-normal text-muted">({{ item.revision.app_slug }})</span></p>
            <p v-if="item.revision.author_note" class="mt-1 text-sm text-muted">Author note: {{ item.revision.author_note }}</p>
            <p class="text-xs text-muted">Requested {{ new Date(item.revision.created_at).toLocaleString() }}</p>
          </div>
          <div class="flex flex-wrap gap-1.5 text-xs font-semibold">
            <button @click="approveRevision(item)" class="rounded-lg bg-emerald-500/20 px-2.5 py-1.5 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/30">Approve</button>
            <button @click="rejectRevision(item)" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-600 dark:text-red-300 hover:bg-red-500/30">Reject</button>
            <RouterLink :to="`/apps/${item.revision.app_slug}`" class="rounded-lg bg-surface-2 px-2.5 py-1.5 hover:bg-line">View live</RouterLink>
          </div>
        </div>
        <div v-if="revisionChanges(item).length" class="overflow-x-auto rounded-lg border border-line bg-surface-2 text-xs">
          <table class="w-full min-w-[28rem]">
            <thead>
              <tr class="border-b border-line text-left text-muted">
                <th class="px-3 py-2 font-semibold">Field</th>
                <th class="px-3 py-2 font-semibold">Current</th>
                <th class="px-3 py-2 font-semibold">Proposed</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="change in revisionChanges(item)" :key="change.field" class="border-b border-line/60 align-top last:border-0">
                <td class="px-3 py-2 font-semibold">{{ change.label }}</td>
                <td class="px-3 py-2 text-muted whitespace-pre-wrap">{{ change.before }}</td>
                <td class="px-3 py-2 whitespace-pre-wrap">{{ change.after }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="text-sm text-muted">No field changes detected (duplicate submission?).</p>
      </div>
    </section>

    <section v-if="!showForm && unreachableApps.length" class="space-y-3 rounded-2xl border border-red-500/30 bg-red-500/10 p-5">
      <div>
        <h2 class="font-bold">Unreachable domains ({{ unreachableApps.length }})</h2>
        <p class="text-sm text-muted">HTTPS probe failed for these domains. Users may not be able to open the app.</p>
      </div>
      <div class="space-y-2">
        <div v-for="app in unreachableApps" :key="app.id"
          class="flex flex-col gap-2 rounded-xl border border-line bg-surface px-3 py-3 sm:flex-row sm:items-center">
          <div class="min-w-0 flex-1">
            <p class="font-semibold">{{ app.name }}</p>
            <p class="truncate text-xs text-muted">{{ app.domain }} · checked {{ formatCheckedAt(app.domain_checked_at) }}</p>
          </div>
          <button @click="startEdit(app)" class="rounded-lg bg-surface-2 px-2.5 py-1.5 text-xs font-semibold hover:bg-line">Edit</button>
        </div>
      </div>
    </section>

    <section v-if="!showForm && featuredApps.length" class="space-y-3 rounded-2xl border border-line bg-surface p-5">
      <div>
        <h2 class="font-bold">Featured order</h2>
        <p class="text-sm text-muted">This is the order shown on the home page. Use the arrows or set a number when editing an app.</p>
      </div>
      <div class="space-y-2">
        <div v-for="(app, index) in featuredApps" :key="app.id"
          class="flex items-center gap-3 rounded-xl border border-line bg-surface-2 px-3 py-2">
          <span class="w-6 text-center text-sm font-bold text-muted">{{ index + 1 }}</span>
          <div class="min-w-0 flex-1">
            <p class="truncate font-semibold">{{ app.name }}</p>
            <p class="text-xs text-muted">
              order {{ app.featured_order > 0 ? app.featured_order : 'auto' }}
            </p>
          </div>
          <div class="flex gap-1">
            <button type="button" :disabled="index === 0 || reordering"
              @click="moveFeatured(app, -1)"
              class="rounded-lg border border-line px-2 py-1 text-xs font-bold hover:bg-surface disabled:opacity-40">
              ↑
            </button>
            <button type="button" :disabled="index === featuredApps.length - 1 || reordering"
              @click="moveFeatured(app, 1)"
              class="rounded-lg border border-line px-2 py-1 text-xs font-bold hover:bg-surface disabled:opacity-40">
              ↓
            </button>
            <button type="button" @click="startEdit(app)"
              class="rounded-lg bg-surface px-2 py-1 text-xs font-semibold hover:bg-line">
              Edit
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- app list -->
    <div v-if="!showForm && apps.length" class="flex flex-wrap items-center gap-2 text-xs font-semibold text-muted">
      <span>Sort:</span>
      <button type="button" class="rounded-lg px-2 py-1 hover:bg-surface-2"
        :class="sortKey === 'name' ? 'text-accent-ink' : ''" @click="toggleSort('name')">
        Name {{ sortKey === 'name' ? (sortAsc ? '↑' : '↓') : '' }}
      </button>
      <button type="button" class="rounded-lg px-2 py-1 hover:bg-surface-2"
        :class="sortKey === 'total_opens' ? 'text-accent-ink' : ''" @click="toggleSort('total_opens')">
        Opens {{ sortKey === 'total_opens' ? (sortAsc ? '↑' : '↓') : '' }}
      </button>
      <button type="button" class="rounded-lg px-2 py-1 hover:bg-surface-2"
        :class="sortKey === 'total_views' ? 'text-accent-ink' : ''" @click="toggleSort('total_views')">
        Views {{ sortKey === 'total_views' ? (sortAsc ? '↑' : '↓') : '' }}
      </button>
    </div>
    <div class="space-y-2">
      <div v-for="app in sortedApps" :key="app.id"
        class="flex flex-col gap-3 rounded-2xl border border-line bg-surface p-4 sm:flex-row sm:items-center">
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="font-bold">{{ app.name }}</span>
            <ReleaseStageBadge v-if="app.release_stage !== 'released'" :stage="app.release_stage" />
            <StatusBadge :status="app.status" />
            <span v-if="app.domain_reachable === false" class="rounded-full bg-red-500/20 px-2 py-0.5 text-xs font-bold text-red-600 dark:text-red-300" title="Domain unreachable">offline</span>
            <span v-else-if="app.domain_reachable === true" class="rounded-full bg-emerald-500/15 px-2 py-0.5 text-xs font-bold text-emerald-700 dark:text-emerald-300" title="Domain reachable">online</span>
            <span v-if="app.featured" class="text-accent-ink" title="Featured">★</span>
          </div>
          <p class="truncate text-sm text-muted">{{ app.slug }} · {{ app.domain }}</p>
          <p class="mt-1 text-xs text-muted">
            Opens: {{ (app.total_opens ?? 0).toLocaleString() }} · Views: {{ (app.total_views ?? 0).toLocaleString() }}
          </p>
        </div>
        <div class="flex flex-wrap gap-1.5 text-xs font-semibold">
          <button @click="setStatus(app, 'approve')" class="rounded-lg bg-sky-500/20 px-2.5 py-1.5 text-sky-700 dark:text-sky-300 hover:bg-sky-500/30">Approve</button>
          <button @click="setStatus(app, 'verify')" class="rounded-lg bg-emerald-500/20 px-2.5 py-1.5 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/30">Verify</button>
          <button @click="startReject(app)" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-600 dark:text-red-300 hover:bg-red-500/30">Reject</button>
          <button @click="startEdit(app)" class="rounded-lg bg-surface-2 px-2.5 py-1.5 hover:bg-line">Edit</button>
          <button @click="remove(app)" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-600 dark:text-red-300 hover:bg-red-500/30">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>
