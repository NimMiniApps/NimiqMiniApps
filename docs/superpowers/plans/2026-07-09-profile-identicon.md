# User profile: display name + identicon Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a connected wallet set an optional display name, and show a deterministic identicon (Nimiq's official wallet-style avatar) wherever a wallet address currently appears.

**Architecture:** A new `users` table (keyed by `wallet_address`) backs a `GET/PUT /api/profile` pair gated by the existing `walletAuthMiddleware`. The reviews list query gains a `LEFT JOIN` so review authors' names show up automatically. The frontend adds `@nimiq/identicons` (already proven in the sibling `NimFeed` repo) and a ported `AddressIdenticon.vue`, used in the header button, review list, and a new `/profile` page.

**Tech Stack:** Go 1.25 stdlib `net/http` + `pgx/v5` (backend, existing); Vue 3 `<script setup>` + TypeScript + Tailwind v4 + `vue-router` (frontend, existing); `@nimiq/identicons` (new frontend dep).

## Global Constraints

- No new backend framework or storage — same `server` struct, `writeJSON`/`writeError`, `walletAuthMiddleware` conventions as the review feature.
- `display_name` is optional, self-set, no uniqueness check, max 50 chars.
- Identicon is a pure function of the wallet address (`@nimiq/identicons`) — no upload, no backend image storage.
- Match existing Tailwind tokens (`border-line`, `bg-surface`, `bg-surface-2`, `text-accent-ink`, `bg-accent`, `text-muted`).
- Follow the repo's table-driven `*_test.go` stdlib `testing` pattern.

---

### Task 1: `users` table migration

**Files:**
- Create: `backend/migrations/011_users.sql`

**Interfaces:**
- Produces: table `users(wallet_address, display_name, created_at, updated_at)`, consumed by Task 2 and Task 3.

- [ ] **Step 1: Write the migration**

```sql
CREATE TABLE IF NOT EXISTS users (
    wallet_address TEXT PRIMARY KEY,
    display_name TEXT CHECK (display_name IS NULL OR char_length(display_name) <= 50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- [ ] **Step 2: Verify it applies cleanly**

Run: `cd backend && DATABASE_URL=... go run .` (against a local dev Postgres), watch the logs.
Expected: log line `"applied migration" name=011_users.sql`. Stop the process afterward.

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/011_users.sql
git commit -m "Add users table migration for wallet display names"
```

---

### Task 2: Profile handlers (`profile.go`)

**Files:**
- Create: `backend/profile.go`
- Test: `backend/profile_test.go`

**Interfaces:**
- Consumes: `s.pool`, `writeJSON`/`writeError` (`backend/handlers.go`), `walletAuthMiddleware` (`backend/walletauth.go`).
- Produces: `validateDisplayName(name string) string`, `(s *server) getProfile(w, r, address string)`, `(s *server) updateProfile(w, r, address string)` — consumed by Task 4 (main.go routes).

- [ ] **Step 1: Write the failing test**

```go
// backend/profile_test.go
package main

import (
	"strings"
	"testing"
)

func TestValidateDisplayName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty clears name", "", false},
		{"normal name", "Satoshi", false},
		{"max length", strings.Repeat("a", 50), false},
		{"too long", strings.Repeat("a", 51), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDisplayName(tc.input)
			if (err != "") != tc.wantErr {
				t.Fatalf("validateDisplayName(%q) error=%q, wantErr=%v", tc.input, err, tc.wantErr)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./... -run TestValidateDisplayName -v`
Expected: FAIL — `undefined: validateDisplayName`.

- [ ] **Step 3: Write the implementation**

```go
// backend/profile.go
package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Profile struct {
	WalletAddress string  `json:"wallet_address"`
	DisplayName   *string `json:"display_name"`
}

func validateDisplayName(name string) string {
	if len(name) > 50 {
		return "display_name must be at most 50 characters"
	}
	return ""
}

func (s *server) getProfile(w http.ResponseWriter, r *http.Request, address string) {
	var displayName *string
	err := s.pool.QueryRow(r.Context(),
		`SELECT display_name FROM users WHERE wallet_address=$1`, address,
	).Scan(&displayName)
	if err != nil && err.Error() != "no rows in result set" {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, Profile{WalletAddress: address, DisplayName: displayName})
}

func (s *server) updateProfile(w http.ResponseWriter, r *http.Request, address string) {
	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	name := strings.TrimSpace(req.DisplayName)
	if msg := validateDisplayName(name); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	var namePtr *string
	if name != "" {
		namePtr = &name
	}
	_, err := s.pool.Exec(r.Context(), `
		INSERT INTO users (wallet_address, display_name, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (wallet_address) DO UPDATE SET display_name=$2, updated_at=now()`,
		address, namePtr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, Profile{WalletAddress: address, DisplayName: namePtr})
}
```

Note: `pgx` returns `pgx.ErrNoRows` (not the literal string `"no rows in result set"`) — use that directly instead of string comparison:

```go
	"github.com/jackc/pgx/v5"
```

and

```go
	err := s.pool.QueryRow(r.Context(),
		`SELECT display_name FROM users WHERE wallet_address=$1`, address,
	).Scan(&displayName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
```

(add `"errors"` and `"github.com/jackc/pgx/v5"` to the import block; when `pgx.ErrNoRows` is returned, `displayName` stays `nil`, which is the correct "no profile row yet" response.)

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go build ./... && go test ./... -run TestValidateDisplayName -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/profile.go backend/profile_test.go
git commit -m "Add wallet profile get/update handlers"
```

---

### Task 3: Reviews list gains `display_name` (LEFT JOIN)

**Files:**
- Modify: `backend/reviews.go:12-20` (`Review` struct), `backend/reviews.go:69-105` (`listReviews` query + scan)

**Interfaces:**
- Consumes: `users` table (Task 1).
- Produces: `Review.DisplayName *string` — consumed by Task 6 (frontend `AppReview` type).

- [ ] **Step 1: Add `DisplayName` to the `Review` struct**

```go
type Review struct {
	ID            string    `json:"id"`
	AppID         string    `json:"app_id"`
	WalletAddress string    `json:"wallet_address"`
	DisplayName   *string   `json:"display_name"`
	Rating        int       `json:"rating"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Update the `listReviews` query and scan**

```go
func (s *server) listReviews(w http.ResponseWriter, r *http.Request) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT ar.id, ar.app_id, ar.wallet_address, u.display_name, ar.rating, ar.body, ar.created_at, ar.updated_at
		FROM app_reviews ar
		LEFT JOIN users u ON u.wallet_address = ar.wallet_address
		WHERE ar.app_id=$1 ORDER BY ar.created_at DESC`, appID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	reviews := []Review{}
	for rows.Next() {
		var rv Review
		if err := rows.Scan(&rv.ID, &rv.AppID, &rv.WalletAddress, &rv.DisplayName, &rv.Rating, &rv.Body, &rv.CreatedAt, &rv.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		reviews = append(reviews, rv)
	}
	var average float64
	if len(reviews) > 0 {
		var sum int
		for _, rv := range reviews {
			sum += rv.Rating
		}
		average = float64(sum) / float64(len(reviews))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":   reviews,
		"average": average,
		"count":   len(reviews),
	})
}
```

(`upsertReview`'s `RETURNING` scan doesn't need `display_name` — it returns the row it just wrote to `app_reviews`, unrelated to `users`; leave it as-is.)

- [ ] **Step 3: Build and run full backend suite**

Run: `cd backend && go build ./... && go vet ./... && go test ./...`
Expected: builds clean, all tests pass (no existing test asserts on the exact `Review` field list, so this is additive-safe).

- [ ] **Step 4: Commit**

```bash
git add backend/reviews.go
git commit -m "Include reviewer display_name in the app reviews list"
```

---

### Task 4: Wire `/api/profile` routes into `main.go`

**Files:**
- Modify: `backend/main.go` (route registration block added in the reviews feature)

**Interfaces:**
- Consumes: `s.getProfile`, `s.updateProfile` (Task 2), `walletAuthMiddleware` (existing).

- [ ] **Step 1: Register the routes**

In `backend/main.go`, next to the existing `GET /api/auth/me` line, add:

```go
	mux.HandleFunc("GET /api/profile", walletAuthMiddleware(walletAuthSecret, s.getProfile))
	mux.HandleFunc("PUT /api/profile", walletAuthMiddleware(walletAuthSecret, s.updateProfile))
```

- [ ] **Step 2: Build and smoke-test**

Run: `cd backend && go build ./... && go vet ./...`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add backend/main.go
git commit -m "Wire profile routes into the HTTP server"
```

---

### Task 5: Frontend identicon component

**Files:**
- Modify: `frontend/package.json` (add `@nimiq/identicons`)
- Create: `frontend/src/components/AddressIdenticon.vue`

**Interfaces:**
- Produces: `<AddressIdenticon :address :img-class />` — consumed by Task 7 (`WalletLoginButton.vue`), Task 8 (`ReviewList.vue`), Task 9 (`ProfileView.vue`).

- [ ] **Step 1: Install the dependency**

Run: `cd frontend && npm install @nimiq/identicons`
Expected: added to `frontend/package.json` `dependencies` (pin the same major as the sibling `NimFeed` repo uses, `^1.6.2`, since that's the version already proven against this identicon style) and `frontend/package-lock.json` updated.

- [ ] **Step 2: Port the component**

```vue
<!-- frontend/src/components/AddressIdenticon.vue -->
<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import Identicons from '@nimiq/identicons'
import identiconsSvgUrl from '@nimiq/identicons/dist/identicons.min.svg?url'

Identicons.svgPath = identiconsSvgUrl

const props = withDefaults(defineProps<{ address?: string; imgClass?: string }>(), {
  address: '',
  imgClass: 'h-10 w-10',
})

const imageUrl = ref(Identicons.placeholderToDataUrl('#d7deeb', 1))

async function render() {
  if (!props.address) {
    imageUrl.value = Identicons.placeholderToDataUrl('#d7deeb', 1)
    return
  }
  try {
    imageUrl.value = await Identicons.toDataUrl(props.address)
  } catch {
    imageUrl.value = Identicons.placeholderToDataUrl('#d7deeb', 1)
  }
}

onMounted(render)
watch(() => props.address, render)
</script>

<template>
  <img :class="[imgClass, 'rounded-full']" :src="imageUrl" alt="" />
</template>
```

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors (if `@nimiq/identicons` ships no `.d.ts`, add `frontend/src/types/nimiq-identicons.d.ts` with a minimal `declare module '@nimiq/identicons'` — check first with `cat node_modules/@nimiq/identicons/package.json | grep types`).

- [ ] **Step 4: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/components/AddressIdenticon.vue
git commit -m "Add AddressIdenticon component using @nimiq/identicons"
```

---

### Task 6: API client additions (`api.ts`)

**Files:**
- Modify: `frontend/src/api.ts` (the `AppReview` interface and the wallet-auth-and-reviews section added in the reviews feature)

**Interfaces:**
- Produces: `AppReview.display_name`, `getProfile`, `updateProfile` — consumed by Task 8 (`ReviewList.vue`) and Task 9 (`ProfileView.vue`).

- [ ] **Step 1: Add `display_name` to `AppReview`**

```ts
export interface AppReview {
  id: string
  app_id: string
  wallet_address: string
  display_name: string | null
  rating: number
  body: string
  created_at: string
  updated_at: string
}
```

- [ ] **Step 2: Append profile functions**

After the existing `deleteOwnAppReview` export, add:

```ts
export interface Profile {
  wallet_address: string
  display_name: string | null
}

export const getProfile = () =>
  request<Profile>('/api/profile', { credentials: 'include' })

export const updateProfile = (display_name: string) =>
  request<Profile>('/api/profile', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ display_name }),
  })
```

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api.ts
git commit -m "Add profile API client functions and review display_name field"
```

---

### Task 7: Identicon in the header wallet button

**Files:**
- Modify: `frontend/src/components/WalletLoginButton.vue`

**Interfaces:**
- Consumes: `AddressIdenticon` (Task 5).

- [ ] **Step 1: Update the component**

```vue
<script setup lang="ts">
import { useWalletAuth } from '../composables/useWalletAuth'
import AddressIdenticon from './AddressIdenticon.vue'

const { walletAddress, checking, loggingIn, error, login, logout } = useWalletAuth()

function truncate(address: string): string {
  return address.length > 12 ? address.slice(0, 6) + '…' + address.slice(-4) : address
}
</script>

<template>
  <div class="flex items-center gap-2">
    <span v-if="checking" class="text-sm text-muted">…</span>
    <template v-else-if="walletAddress">
      <RouterLink to="/profile" class="flex items-center gap-1.5">
        <AddressIdenticon :address="walletAddress" img-class="h-6 w-6" />
        <span class="font-mono text-sm text-accent-ink">{{ truncate(walletAddress) }}</span>
      </RouterLink>
      <button class="text-sm text-muted hover:text-accent-ink" @click="logout">Log out</button>
    </template>
    <button
      v-else
      class="rounded-lg border border-line bg-surface-2 px-3 py-1.5 text-sm font-semibold text-accent-ink transition-colors duration-200 hover:border-accent/50 disabled:opacity-50"
      :disabled="loggingIn"
      @click="login"
    >
      {{ loggingIn ? 'Connecting…' : 'Connect Wallet' }}
    </button>
    <p v-if="error" class="text-xs text-red-500">{{ error }}</p>
  </div>
</template>
```

(`RouterLink` is a Vue Router global component, already used elsewhere in this project's templates — e.g. `frontend/src/App.vue` — so no import is needed.)

- [ ] **Step 2: Manually verify in the browser**

Run: `cd frontend && npm run dev`, connect a wallet, confirm the identicon renders next to the truncated address in the header and links to `/profile` (the route won't exist until Task 9 — a 404 at this point is expected and fine).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/WalletLoginButton.vue
git commit -m "Show identicon and profile link in the header wallet button"
```

---

### Task 8: Identicon + display name in the reviews list

**Files:**
- Modify: `frontend/src/components/ReviewList.vue`

**Interfaces:**
- Consumes: `AddressIdenticon` (Task 5), `AppReview.display_name` (Task 6).

- [ ] **Step 1: Update the component**

```vue
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
```

- [ ] **Step 2: Typecheck and manually verify**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors.

Run: `cd frontend && npm run dev`, open an app with an existing review, confirm the identicon and (fallback truncated address, since no name is set yet) render.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ReviewList.vue
git commit -m "Show identicon and display name in the reviews list"
```

---

### Task 9: `/profile` page

**Files:**
- Create: `frontend/src/views/ProfileView.vue`
- Modify: `frontend/src/main.ts:9-20` (add the route)

**Interfaces:**
- Consumes: `useWalletAuth` (existing), `AddressIdenticon` (Task 5), `getProfile`, `updateProfile` (Task 6).

- [ ] **Step 1: Write the view**

```vue
<!-- frontend/src/views/ProfileView.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useWalletAuth } from '../composables/useWalletAuth'
import { getProfile, updateProfile } from '../api'
import AddressIdenticon from '../components/AddressIdenticon.vue'

const { walletAddress, checking } = useWalletAuth()

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
    saved.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to save profile'
  } finally {
    saving.value = false
  }
}

onMounted(load)
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
      </label>

      <div class="flex items-center gap-3">
        <button
          class="rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
          :disabled="saving"
          @click="save"
        >{{ saving ? 'Saving…' : 'Save' }}</button>
        <span v-if="saved" class="text-sm text-muted">Saved.</span>
        <span v-if="error" class="text-sm text-red-500">{{ error }}</span>
      </div>
    </div>
  </div>
</template>
```

- [ ] **Step 2: Register the route**

In `frontend/src/main.ts`, add to the `routes` array (next to `/admin`):

```ts
    { path: '/profile', component: () => import('./views/ProfileView.vue'), meta: { title: 'Profile' } },
```

- [ ] **Step 3: Typecheck and manually verify**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors.

Run: `cd frontend && npm run dev`, visit `/profile` while logged out (see "Connect your wallet…"), connect a wallet, visit `/profile` again, set a display name, save, refresh, confirm it persisted.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/ProfileView.vue frontend/src/main.ts
git commit -m "Add /profile page for editing wallet display name"
```

---

## Explicitly out of scope (per spec)

- Uploaded/custom avatars.
- Unique display names / collision handling.
- Developer-account linking, app ownership, or permissions (separate follow-up spec, to build on this `users` table and `/profile` page).
