<script setup lang="ts">
import {
  CATALOG_ISSUES_URL,
  CATALOG_REPO_URL,
  COMPETITION_CYCLE1_WINNERS_URL,
  COMPETITION_SKOOL_URL,
  COMPETITION_URL,
} from '../utils/catalogLinks'

const designResources = [
  {
    title: 'Nimiq UI Kit',
    description:
      'Design tokens, CSS components, live @nimiq/vue-components catalog, identicons, utilities, and icons — the authoritative reference for matching Nimiq Pay.',
    href: 'https://nimiqtoolbox.github.io/nimiq-ui-kit/',
    github: 'https://github.com/NimiqToolbox/nimiq-ui-kit',
  },
]

const docs = [
  {
    title: 'Mini Apps Overview',
    description: 'What mini apps are and how they run inside Nimiq Pay.',
    href: 'https://nimiq.dev/mini-apps/',
  },
  {
    title: 'Build Your First Mini App',
    description: 'Step-by-step tutorial for your first Nimiq Pay mini app.',
    href: 'https://nimiq.dev/mini-apps/mini-app-tutorial',
  },
  {
    title: 'Dual-Chain Mini App',
    description: 'Use NIM and EVM tokens (USDT, USDC) in one app.',
    href: 'https://nimiq.dev/mini-apps/dual-chain-mini-app-tutorial',
  },
  {
    title: 'EVM Tokens Guide',
    description: 'Working with ERC-20 tokens across supported chains.',
    href: 'https://nimiq.dev/mini-apps/evm-tokens',
  },
  {
    title: 'Load a Local Mini App',
    description: 'Test your app on a phone over your LAN during development.',
    href: 'https://nimiq.dev/mini-apps/development/load-local-mini-app',
  },
]

const examples = [
  {
    title: 'evm-mini-wallet',
    description: 'Full EVM wallet mini app — token sends, balances, multi-chain.',
    href: 'https://github.com/Albermonte/evm-mini-wallet',
  },
  {
    title: 'nimiq-mini-app-demo',
    description: 'Demo of all Nimiq and Ethereum provider methods.',
    href: 'https://github.com/Eligioo/nimiq-mini-app-demo',
  },
]
</script>

<template>
  <div class="space-y-8">
    <section class="rounded-3xl border border-line bg-surface p-6 shadow-sm md:p-10">
      <p class="text-sm font-bold uppercase tracking-widest text-accent-ink">For builders</p>
      <h1 class="mt-2 text-3xl font-extrabold">Build a mini app</h1>
      <p class="mt-3 max-w-2xl text-muted md:text-lg">
        Mini apps are web apps that run inside Nimiq Pay. Use the Nimiq provider for NIM payments
        and staking, and <code class="rounded bg-surface-2 px-1.5 py-0.5 text-sm">window.ethereum</code>
        for EVM chains and tokens — no app store, no installs.
      </p>
      <div class="mt-6 flex flex-wrap gap-3">
        <a href="https://nimiq.dev/mini-apps/mini-app-tutorial" target="_blank" rel="noopener"
          class="cursor-pointer board-plate board-plate-primary px-6 py-3 text-sm">
          Start the tutorial
        </a>
        <RouterLink to="/submit"
          class="cursor-pointer board-plate board-plate-ghost px-6 py-3 text-sm">
          Submit your app
        </RouterLink>
      </div>
    </section>

    <section class="relative overflow-hidden rounded-3xl border border-amber-500/35 bg-gradient-to-br from-amber-500/15 via-surface to-surface p-6 shadow-sm md:p-8">
      <p class="text-sm font-bold uppercase tracking-widest text-amber-800 dark:text-amber-200">Mini Apps Competition</p>
      <h2 class="mt-2 text-2xl font-extrabold md:text-3xl">Cycle 1 is in the books</h2>
      <p class="mt-3 max-w-2xl text-muted">
        62 Mini Apps submitted, scored by the Nimiq Community Council. Results are final.
        Winners keep their work under MIT; prizes are arranged directly with the podium.
      </p>

      <ol class="mt-5 space-y-3 text-sm">
        <li class="flex flex-wrap items-center gap-x-2 gap-y-1">
          <span class="inline-flex items-center gap-1 rounded-[3px] border border-amber-500/70 bg-gradient-to-b from-amber-300 to-amber-500 px-2 py-0.5 text-[11px] font-extrabold uppercase tracking-wide text-amber-950">
            <svg class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M4.5 1.5h7v1.2h1.3A1.7 1.7 0 0 1 14.5 4.4c0 2.2-1.5 3.6-3.4 4.1-.4 1-.9 1.6-1.6 2v1.5h2.2v1.5H4.3V11.9h2.2V10.4c-.7-.4-1.2-1-1.6-2C2.9 8 1.5 6.6 1.5 4.4A1.7 1.7 0 0 1 3.2 2.7h1.3V1.5Zm1.5 1.2v1.2H10V2.7H6Zm-2.8 2.9c0 1.3.8 2.2 2 2.6V5.6H3.2Zm9.6 0H9.8v2.6c1.2-.4 2-1.3 2-2.6Z" /></svg>
            1st place
          </span>
          <RouterLink to="/apps/nimiq-space" class="font-bold text-accent-ink hover:underline">Nimiq Space</RouterLink>
          <span class="text-muted">— open multiplayer world with live rooms and a persistent collaborative floor</span>
        </li>
        <li class="flex flex-wrap items-center gap-x-2 gap-y-1">
          <span class="inline-flex items-center gap-1 rounded-[3px] border border-slate-400/80 bg-gradient-to-b from-slate-100 to-slate-300 px-2 py-0.5 text-[11px] font-extrabold uppercase tracking-wide text-slate-900">
            <svg class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M4.5 1.5h7v1.2h1.3A1.7 1.7 0 0 1 14.5 4.4c0 2.2-1.5 3.6-3.4 4.1-.4 1-.9 1.6-1.6 2v1.5h2.2v1.5H4.3V11.9h2.2V10.4c-.7-.4-1.2-1-1.6-2C2.9 8 1.5 6.6 1.5 4.4A1.7 1.7 0 0 1 3.2 2.7h1.3V1.5Zm1.5 1.2v1.2H10V2.7H6Zm-2.8 2.9c0 1.3.8 2.2 2 2.6V5.6H3.2Zm9.6 0H9.8v2.6c1.2-.4 2-1.3 2-2.6Z" /></svg>
            2nd place
          </span>
          <RouterLink to="/apps/nimjump" class="font-bold text-accent-ink hover:underline">NimJump</RouterLink>
          <span class="text-muted">— arcade run with server-side replay instead of client-trusted scores</span>
        </li>
        <li class="flex flex-wrap items-center gap-x-2 gap-y-1">
          <span class="inline-flex items-center gap-1 rounded-[3px] border border-orange-700/60 bg-gradient-to-b from-orange-300 to-orange-600 px-2 py-0.5 text-[11px] font-extrabold uppercase tracking-wide text-orange-950">
            <svg class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true"><path d="M4.5 1.5h7v1.2h1.3A1.7 1.7 0 0 1 14.5 4.4c0 2.2-1.5 3.6-3.4 4.1-.4 1-.9 1.6-1.6 2v1.5h2.2v1.5H4.3V11.9h2.2V10.4c-.7-.4-1.2-1-1.6-2C2.9 8 1.5 6.6 1.5 4.4A1.7 1.7 0 0 1 3.2 2.7h1.3V1.5Zm1.5 1.2v1.2H10V2.7H6Zm-2.8 2.9c0 1.3.8 2.2 2 2.6V5.6H3.2Zm9.6 0H9.8v2.6c1.2-.4 2-1.3 2-2.6Z" /></svg>
            3rd place
          </span>
          <RouterLink to="/apps/nimquest" class="font-bold text-accent-ink hover:underline">NimQuest</RouterLink>
          <span class="text-muted">— learn-and-prove quizzes with wallet-signed completion</span>
        </li>
      </ol>

      <p class="mt-4 text-sm text-muted">
        Full write-up on the
        <a :href="COMPETITION_CYCLE1_WINNERS_URL" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">Nimiq blog</a>.
        Catalog apps show Cycle scores (0 until official totals are published).
      </p>

      <div class="mt-8 border-t border-amber-500/25 pt-6">
        <h3 class="text-lg font-extrabold">Cycle 2 opens August 24</h3>
        <p class="mt-2 max-w-2xl text-muted">
          Same <strong class="font-semibold text-ink">$17,000 USDT</strong> prize pool
          ($10k / $5k / $2k). Four Sip &amp; Ship calls. Submissions close
          <strong class="font-semibold text-ink">September 18 at 23:59 UTC</strong>.
          Non-winners can refine and resubmit; prize-winning apps cannot be resubmitted as the same entry.
        </p>
        <div class="mt-6 flex flex-wrap gap-3">
          <a :href="COMPETITION_URL" target="_blank" rel="noopener"
            class="cursor-pointer board-plate board-plate-primary px-6 py-3 text-sm">
            Register &amp; rules
          </a>
          <a :href="COMPETITION_SKOOL_URL" target="_blank" rel="noopener"
            class="cursor-pointer board-plate board-plate-ghost px-6 py-3 text-sm">
            Join Skool community
          </a>
          <RouterLink to="/submit"
            class="cursor-pointer rounded-xl border border-amber-500/40 bg-amber-500/10 px-6 py-3 font-bold text-amber-900 transition-colors duration-200 hover:bg-amber-500/20 dark:text-amber-100">
            Submit to directory
          </RouterLink>
        </div>
      </div>
    </section>

    <section class="space-y-3">
      <h2 class="text-xl font-extrabold">Design system</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <div
          v-for="resource in designResources"
          :key="resource.href"
          class="rounded-2xl border border-line bg-surface p-4 shadow-sm sm:col-span-2"
        >
          <a
            :href="resource.href"
            target="_blank"
            rel="noopener"
            class="block transition duration-200 hover:-translate-y-0.5"
          >
            <p class="font-bold">{{ resource.title }}</p>
            <p class="mt-1 text-sm text-muted">{{ resource.description }}</p>
          </a>
          <p class="mt-3 text-xs">
            <a
              :href="resource.href"
              target="_blank"
              rel="noopener"
              class="font-semibold text-accent-ink hover:underline"
            >Browse the kit</a>
            <span class="mx-1 text-muted" aria-hidden="true">·</span>
            <a
              :href="resource.github"
              target="_blank"
              rel="noopener"
              class="font-semibold text-accent-ink hover:underline"
            >GitHub</a>
          </p>
        </div>
      </div>
    </section>

    <section id="api-credentials" class="scroll-mt-24 space-y-3 rounded-3xl border border-line bg-surface p-6 shadow-sm md:p-8">
      <p class="text-sm font-bold uppercase tracking-widest text-accent-ink">Catalog integrations</p>
      <h2 class="text-2xl font-extrabold">API keys &amp; NimConnect scopes</h2>
      <p class="max-w-2xl text-muted">
        Once your app is listed, you can issue your own API credential from the catalog — no need to ask a human to mint a key.
        Ecosystem services (for example NimConnect awards) verify that key against this registry and trust your app’s
        <strong class="font-semibold text-ink">declared scopes</strong>.
      </p>
      <ol class="mt-4 list-decimal space-y-3 pl-5 text-sm text-muted">
        <li>
          <span class="font-semibold text-ink">Declare scopes</span> when you
          <RouterLink to="/submit" class="font-semibold text-accent-ink hover:underline">submit</RouterLink>
          or request an update — tick only what the app actually needs
          (<code class="rounded bg-surface-2 px-1 py-0.5 text-xs">friends:read</code>,
          <code class="rounded bg-surface-2 px-1 py-0.5 text-xs">achievements:write</code>,
          <code class="rounded bg-surface-2 px-1 py-0.5 text-xs">events:write</code>, …).
          Changes on a live listing go through review.
        </li>
        <li>
          After the app is <strong class="font-semibold text-ink">approved</strong> (or verified / experimental), open
          <RouterLink to="/my-apps" class="font-semibold text-accent-ink hover:underline">My apps</RouterLink>
          → <strong class="font-semibold text-ink">API keys</strong> → <strong class="font-semibold text-ink">Issue key</strong>.
          Copy the secret immediately — it is shown once.
        </li>
        <li>
          Send that key to the consumer service as they document (usually a bearer / app key header).
          They call the registry verify endpoint; revoked or unlisted apps stop verifying.
        </li>
      </ol>
      <ul class="mt-4 space-y-1 text-sm text-muted">
        <li class="flex gap-2"><span class="text-accent-ink" aria-hidden="true">→</span> Rotate by issuing a new key, then revoking the old one (no downtime).</li>
        <li class="flex gap-2"><span class="text-accent-ink" aria-hidden="true">→</span> Never commit the plaintext key; store it in your app’s server env.</li>
        <li class="flex gap-2"><span class="text-accent-ink" aria-hidden="true">→</span> Empty declared scopes means the app cannot request NimConnect permissions.</li>
      </ul>
      <div class="mt-6 flex flex-wrap gap-3">
        <RouterLink to="/my-apps" class="cursor-pointer board-plate board-plate-primary px-5 py-2.5 text-sm">
          Open My apps
        </RouterLink>
        <RouterLink to="/submit" class="cursor-pointer board-plate board-plate-ghost px-5 py-2.5 text-sm">
          Submit or update listing
        </RouterLink>
      </div>
    </section>

    <section class="space-y-3">
      <h2 class="text-xl font-extrabold">Documentation</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <a
          v-for="doc in docs"
          :key="doc.href"
          :href="doc.href"
          target="_blank"
          rel="noopener"
          class="rounded-2xl border border-line bg-surface p-4 shadow-sm transition duration-200 hover:-translate-y-0.5 hover:border-accent/40"
        >
          <p class="font-bold">{{ doc.title }}</p>
          <p class="mt-1 text-sm text-muted">{{ doc.description }}</p>
        </a>
      </div>
    </section>

    <section class="space-y-3">
      <h2 class="text-xl font-extrabold">Reference implementations</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <a
          v-for="repo in examples"
          :key="repo.href"
          :href="repo.href"
          target="_blank"
          rel="noopener"
          class="rounded-2xl border border-line bg-surface p-4 shadow-sm transition duration-200 hover:-translate-y-0.5 hover:border-accent/40"
        >
          <p class="font-bold">{{ repo.title }}</p>
          <p class="mt-1 text-sm text-muted">{{ repo.description }}</p>
        </a>
      </div>
    </section>

    <section class="rounded-2xl border border-line bg-surface p-5 shadow-sm">
      <p class="font-semibold text-ink">This directory</p>
      <p class="mt-1 text-sm text-muted">
        Nimiq Mini Apps is open source. Questions about listing your app, bugs, or feature ideas?
        <a :href="CATALOG_ISSUES_URL" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">Open a GitHub issue</a>
        or browse the
        <a :href="CATALOG_REPO_URL" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">repository</a>.
      </p>
    </section>

    <section class="rounded-2xl border border-line bg-surface-2 p-5 text-sm text-muted">
      <p class="font-semibold text-ink">Quick tips</p>
      <ul class="mt-2 list-inside list-disc space-y-1">
        <li>Use the <a href="https://nimiqtoolbox.github.io/nimiq-ui-kit/" target="_blank" rel="noopener" class="font-semibold text-accent-ink hover:underline">Nimiq UI Kit</a> for tokens, components, and icons that match Nimiq Pay.</li>
        <li>Design mobile-first — mini apps run in a phone WebView.</li>
        <li>Never access private keys; all wallet actions go through Nimiq Pay dialogs.</li>
        <li>USDT and USDC use 6 decimals, not 18.</li>
        <li>Long-press Nimiq Pay settings for 10s to switch to testnet during development.</li>
        <li>Need awards or plaza events?
          <RouterLink to="/build#api-credentials" class="font-semibold text-accent-ink hover:underline">Issue a catalog API key</RouterLink>
          after your app is listed.</li>
      </ul>
    </section>
  </div>
</template>
