<script setup lang="ts">
import { CATALOG_ISSUES_URL, CATALOG_REPO_URL, COMPETITION_SKOOL_URL, COMPETITION_URL } from '../utils/catalogLinks'

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
          class="cursor-pointer rounded-[500px] nq-primary px-6 py-3 font-bold text-white transition duration-200">
          Start the tutorial
        </a>
        <RouterLink to="/submit"
          class="cursor-pointer rounded-xl border border-line bg-surface px-6 py-3 font-bold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
          Submit your app
        </RouterLink>
      </div>
    </section>

    <section class="relative overflow-hidden rounded-3xl border border-amber-500/35 bg-gradient-to-br from-amber-500/15 via-surface to-surface p-6 shadow-sm md:p-8">
      <p class="text-sm font-bold uppercase tracking-widest text-amber-800 dark:text-amber-200">Mini Apps Competition</p>
      <h2 class="mt-2 text-2xl font-extrabold md:text-3xl">Build, compete, win prizes</h2>
      <p class="mt-3 max-w-2xl text-muted">
        The Nimiq Mini Apps Competition runs in cycles with
        <strong class="font-semibold text-ink">$17,000 USDT</strong> in prizes per cycle
        ($10k / $5k / $2k for the top three). Ship a mini app with wallet or payment features,
        get scored on design, functionality, and originality — then list it here when it is ready.
      </p>
      <ul class="mt-4 grid gap-2 text-sm text-muted sm:grid-cols-2">
        <li class="flex gap-2"><span class="text-amber-600 dark:text-amber-300" aria-hidden="true">→</span> Open globally (18+, public GitHub, Nimiq wallet)</li>
        <li class="flex gap-2"><span class="text-amber-600 dark:text-amber-300" aria-hidden="true">→</span> Teams up to 5 — games, tools, social apps welcome</li>
        <li class="flex gap-2"><span class="text-amber-600 dark:text-amber-300" aria-hidden="true">→</span> AI-assisted builds OK (Cursor, Copilot, etc.)</li>
        <li class="flex gap-2"><span class="text-amber-600 dark:text-amber-300" aria-hidden="true">→</span> Skool community for feedback, co-founders, live calls</li>
      </ul>
      <div class="mt-6 flex flex-wrap gap-3">
        <a :href="COMPETITION_URL" target="_blank" rel="noopener"
          class="cursor-pointer rounded-[500px] nq-primary px-6 py-3 font-bold text-white transition duration-200">
          Register &amp; rules
        </a>
        <a :href="COMPETITION_SKOOL_URL" target="_blank" rel="noopener"
          class="cursor-pointer rounded-xl border border-line bg-surface px-6 py-3 font-bold transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink">
          Join Skool community
        </a>
        <RouterLink to="/submit"
          class="cursor-pointer rounded-xl border border-amber-500/40 bg-amber-500/10 px-6 py-3 font-bold text-amber-900 transition-colors duration-200 hover:bg-amber-500/20 dark:text-amber-100">
          Submit to directory
        </RouterLink>
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
      </ul>
    </section>
  </div>
</template>
