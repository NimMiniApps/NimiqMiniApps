<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  APP_CATEGORIES, APP_RELEASE_STAGES, adminListApps, adminCreateApp, adminUpdateApp, adminDeleteApp, adminSetStatus, adminCheckDomains,
  adminListRevisions, adminApproveRevision, adminRejectRevision, adminSearchUsers,
  type App, type RevisionReviewItem, type AdminUserResult,
} from '../api'
import { useWalletAuth } from '../composables/useWalletAuth'
import StatusBadge from '../components/StatusBadge.vue'
import ReleaseStageBadge from '../components/ReleaseStageBadge.vue'
import { formatMediaLines, parseMediaLines } from '../utils/media'
import { formatSocialLines, parseSocialLines } from '../utils/socials'
import { diffRevision } from '../utils/revisionDiff'

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

const emptyForm = {
  slug: '', name: '', domain: '', category: '', developer_slug: '', developer_name: '',
  developer_wallet_address: null as string | null,
  tagline: '', description: '', long_description: '', tags: '', assets: 'NIM', status: 'submitted',
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
    developerResults.value = developerQuery.value.trim()
      ? await adminSearchUsers(developerQuery.value.trim())
      : []
  }, 250)
}

function pickDeveloper(user: AdminUserResult) {
  form.developer_wallet_address = user.wallet_address
  developerQuery.value = user.display_name ?? user.wallet_address
  developerResults.value = []
}

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
  editingSlug.value = ''
  showForm.value = true
}

function startEdit(app: App) {
  Object.assign(form, {
    slug: app.slug, name: app.name, domain: app.domain, category: app.category,
    developer_slug: app.developer_slug, developer_name: app.developer_name,
    developer_wallet_address: app.developer_wallet_address,
    tagline: app.tagline, description: app.description, long_description: app.long_description || '',
    tags: app.tags.join(', '), assets: app.assets.join(', '),
    status: app.status, release_stage: app.release_stage, featured: app.featured,
    featured_order: app.featured_order ?? 0,
    website_url: app.website_url || '', github_url: app.github_url || '',
    icon_url: app.icon_url || '', banner_url: app.banner_url || '',
    media: formatMediaLines(app.media),
    socials: formatSocialLines(app.socials),
    submitter_contact: app.submitter_contact || '',
  })
  developerQuery.value = app.developer_name
  editingSlug.value = app.slug
  showForm.value = true
}

const csv = (s: string) => s.split(',').map((x) => x.trim()).filter(Boolean)

async function submit() {
  error.value = ''
  const payload = {
    ...form,
    developer_wallet_address: form.developer_wallet_address || null,
    tags: csv(form.tags),
    assets: csv(form.assets),
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

const fields: [keyof typeof emptyForm, string, boolean][] = [
  ['slug', 'Slug (lowercase, url-safe)', true],
  ['name', 'Name', true],
  ['domain', 'Domain (no https://)', true],
  ['developer_slug', 'Developer slug', true],
  ['developer_name', 'Developer name', true],
  ['submitter_contact', 'Submitter contact (private)', false],
  ['tagline', 'Tagline', true],
  ['tags', 'Tags (comma-separated)', false],
  ['assets', 'Assets (NIM, USDT, USDC, BTC, ETH)', false],
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
        <button @click="saveToken" class="rounded-xl bg-nq-blue px-4 py-2.5 font-bold text-white hover:bg-nq-blue-dark">
          Save
        </button>
      </div>
    </details>

    <p v-if="error" class="rounded-xl bg-red-500/20 p-4 text-red-600 dark:text-red-600 dark:text-red-300">{{ error }}</p>
    <p v-if="notice" class="rounded-xl bg-emerald-500/15 p-4 text-emerald-700 dark:text-emerald-700 dark:text-emerald-300">{{ notice }}</p>

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
        <label v-for="[key, label, required] in fields" :key="key" class="text-sm">
          <span class="mb-1 block text-muted">{{ label }}{{ required ? ' *' : '' }}</span>
          <input v-model="(form as any)[key]" :required="required"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 focus:border-accent outline-none" />
        </label>
        <label class="relative text-sm sm:col-span-2">
          <span class="mb-1 block font-semibold text-muted">Owning developer (optional)</span>
          <input v-model="developerQuery" @input="onDeveloperQueryInput"
            placeholder="Search by display name or wallet address — leave blank for an unclaimed/anonymous app"
            class="w-full rounded-lg border border-line bg-surface-2 px-3 py-2 outline-none transition-colors duration-200 focus:border-accent" />
          <ul v-if="developerResults.length" class="absolute z-10 mt-1 w-full rounded-lg border border-line bg-surface shadow-lg">
            <li v-for="user in developerResults" :key="user.wallet_address"
              @click="pickDeveloper(user)"
              class="cursor-pointer px-3 py-2 text-sm hover:bg-surface-2">
              {{ user.display_name ?? 'No display name' }}
              <span class="block font-mono text-xs text-muted">{{ user.wallet_address }}</span>
            </li>
          </ul>
          <span v-if="form.developer_wallet_address" class="mt-1 block text-xs text-muted">
            Linked to <span class="font-mono">{{ form.developer_wallet_address }}</span>
            <button type="button" @click="form.developer_wallet_address = null; developerQuery = ''" class="ml-1 text-accent-ink hover:underline">clear</button>
          </span>
        </label>
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
        <button type="submit" class="rounded-xl bg-nq-blue px-5 py-2 font-bold text-white hover:bg-nq-blue-dark">
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
            <button @click="setStatus(app, 'reject')" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-600 dark:text-red-300 hover:bg-red-500/30">Reject</button>
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
    <div class="space-y-2">
      <div v-for="app in apps" :key="app.id"
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
        </div>
        <div class="flex flex-wrap gap-1.5 text-xs font-semibold">
          <button @click="setStatus(app, 'approve')" class="rounded-lg bg-sky-500/20 px-2.5 py-1.5 text-sky-700 dark:text-sky-300 hover:bg-sky-500/30">Approve</button>
          <button @click="setStatus(app, 'verify')" class="rounded-lg bg-emerald-500/20 px-2.5 py-1.5 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/30">Verify</button>
          <button @click="setStatus(app, 'reject')" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-600 dark:text-red-300 hover:bg-red-500/30">Reject</button>
          <button @click="startEdit(app)" class="rounded-lg bg-surface-2 px-2.5 py-1.5 hover:bg-line">Edit</button>
          <button @click="remove(app)" class="rounded-lg bg-red-500/20 px-2.5 py-1.5 text-red-600 dark:text-red-300 hover:bg-red-500/30">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>
