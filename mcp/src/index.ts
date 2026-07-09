#!/usr/bin/env node

import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { z } from 'zod'
import * as api from './api.js'

const categories = z.enum(['Games', 'Utilities', 'Finance', 'Maps', 'Social', 'Experiments'])
const statuses = z.enum(['submitted', 'approved', 'verified', 'experimental', 'rejected'])
const releaseStages = z.enum(['concept', 'alpha', 'beta', 'released'])
const assets = z.enum(['NIM', 'USDT', 'USDC', 'BTC', 'ETH'])
const mediaItem = z.object({
  type: z.enum(['image', 'youtube']),
  url: z.string(),
})
const socialItem = z.object({
  platform: z.enum(['twitter', 'discord', 'telegram', 'bluesky', 'instagram', 'youtube', 'linkedin', 'mastodon', 'reddit', 'tiktok']),
  url: z.string(),
})

const descriptionField = z.string().optional().describe(
  'Plain-text short summary for listings and SEO. Keep to 1–3 sentences; no Markdown.',
)
const longDescriptionField = z.string().optional().describe(
  'Full About text for the app detail page. Markdown supported: **bold**, lists, [links](https://…), ## headings, `code`. Short description stays plain text when both are set.',
)

const appFields = {
  slug: z.string().describe('URL-safe id, lowercase with hyphens'),
  name: z.string(),
  domain: z.string().describe('Mini app host/path — https:// and http:// are stripped automatically if pasted'),
  category: categories,
  developer_slug: z.string().optional().describe(
    'Public catalog developer slug. Always required — set directly, or use admin_add_app_owner afterward to link a wallet (ownership no longer travels through create/update).',
  ),
  developer_name: z.string().optional().describe(
    'Public catalog developer name. Always required — set directly; unaffected by ownership.',
  ),
  tagline: z.string().describe('One-line pitch shown on cards and in search results'),
  description: descriptionField,
  long_description: longDescriptionField,
  tags: z.array(z.string()).optional(),
  assets: z.array(assets).optional(),
  reward_assets: z.array(assets).optional().describe(
    'Assets users can actually receive from this app: daily rewards, leaderboard prizes, payouts, tips, faucets, or similar receive-side flows. Leave empty when the app merely accepts, displays, swaps, or supports a token.',
  ),
  status: statuses.optional(),
  release_stage: releaseStages.optional(),
  featured: z.boolean().optional(),
  website_url: z.string().nullable().optional(),
  github_url: z.string().nullable().optional(),
  icon_url: z.string().nullable().optional(),
  banner_url: z.string().nullable().optional(),
  media: z.array(mediaItem).optional(),
  socials: z.array(socialItem).optional(),
  submitter_contact: z.string().optional().describe(
    'Private submitter contact (Telegram, email). Admin-only in API responses; not used for wallet-based public submit.',
  ),
}

function toolError(error: unknown) {
  const message = error instanceof Error ? error.message : String(error)
  return { content: [{ type: 'text' as const, text: `Error: ${message}` }], isError: true }
}

const server = new McpServer({
  name: 'nimiq-miniapps',
  version: '1.0.0',
})

server.registerTool(
  'miniapps_info',
  {
    description: 'Show configured API URL and whether admin credentials are set',
  },
  async () => api.asToolResult(api.apiInfo()),
)

server.registerTool(
  'health_check',
  { description: 'Check API and database connectivity' },
  async () => {
    try {
      return api.asToolResult(await api.healthCheck())
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'list_apps',
  {
    description: 'List public catalog apps with optional filters',
    inputSchema: {
      q: z.string().optional().describe('Search name, tagline, description, tags, assets, or developer'),
      category: categories.optional(),
      tag: z.string().optional().describe('Filter by exact tag match'),
      asset: z.string().optional().describe('Filter by asset (NIM, USDT, USDC, BTC, ETH)'),
      rewards: z.boolean().optional().describe('Only return apps with one or more reward assets'),
      collection: z.enum(['new-week', 'popular', 'rewards', 'games', 'usdt']).optional(),
      status: statuses.optional().describe('Defaults to approved, verified, and experimental'),
      featured: z.boolean().optional(),
      sort: z.enum(['featured', 'newest', 'name']).optional(),
    },
  },
  async (args) => {
    try {
      return api.asToolResult(await api.listApps(args))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'get_app',
  {
    description: 'Get one app by slug from the public catalog. long_description is Markdown source; rendered on the website detail page.',
    inputSchema: { slug: z.string() },
  },
  async ({ slug }) => {
    try {
      return api.asToolResult(await api.getApp(slug))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'list_categories',
  { description: 'List app categories with public app counts' },
  async () => {
    try {
      return api.asToolResult(await api.listCategories())
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'get_developer',
  {
    description: 'Get a developer and their public apps',
    inputSchema: { slug: z.string().describe('Developer slug') },
  },
  async ({ slug }) => {
    try {
      return api.asToolResult(await api.getDeveloper(slug))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'list_developers',
  { description: 'List all developers with public app counts' },
  async () => {
    try {
      return api.asToolResult(await api.listDevelopers())
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'get_related_apps',
  {
    description: 'Get up to 4 related public apps (same developer or category)',
    inputSchema: { slug: z.string().describe('App slug') },
  },
  async ({ slug }) => {
    try {
      return api.asToolResult(await api.getRelatedApps(slug))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'admin_list_apps',
  { description: 'List all apps including submitted and rejected (requires admin token)' },
  async () => {
    try {
      return api.asToolResult(await api.adminListApps())
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'admin_search_users',
  {
    description:
      'Search users by display name or wallet address prefix (requires admin token). Use before admin_add_app_owner.',
    inputSchema: {
      q: z.string().describe('Prefix match on display_name or wallet_address'),
    },
  },
  async ({ q }) => {
    try {
      return api.asToolResult(await api.adminSearchUsers(q))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'admin_add_app_owner',
  {
    description: 'Link a wallet as a co-owner of an app, granting My apps / request-update access (requires admin token). No effect if already an owner.',
    inputSchema: {
      slug: z.string().describe('App slug'),
      wallet_address: z.string().describe('Wallet to add — must have logged in and set a display name'),
    },
  },
  async ({ slug, wallet_address }) => {
    try {
      return api.asToolResult(await api.adminAddAppOwner(slug, wallet_address))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'admin_remove_app_owner',
  {
    description: 'Unlink a wallet from an app\'s ownership (requires admin token). Unlike the self-service endpoint, this can remove the last owner, fully unclaiming the app.',
    inputSchema: {
      slug: z.string().describe('App slug'),
      wallet_address: z.string().describe('Wallet to remove'),
    },
  },
  async ({ slug, wallet_address }) => {
    try {
      return api.asToolResult(await api.adminRemoveAppOwner(slug, wallet_address))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'admin_create_app',
  {
    description:
      'Create a new app (requires admin token). Always set developer_name and developer_slug directly. Use admin_add_app_owner afterward to link one or more wallets.',
    inputSchema: appFields,
  },
  async (fields) => {
    try {
      return api.asToolResult(await api.adminCreateApp(fields))
    } catch (error) {
      return toolError(error)
    }
  },
)

const { slug: _slugSchema, ...mutableAppFields } = appFields

server.registerTool(
  'admin_update_app',
  {
    description:
      'Update an app by slug; merges with the current record (requires admin token). Ownership is managed separately — use admin_add_app_owner / admin_remove_app_owner.',
    inputSchema: {
      slug: z.string().describe('Slug of the app to update'),
      ...Object.fromEntries(
        Object.entries(mutableAppFields).map(([key, schema]) => [key, schema.optional()]),
      ),
    },
  },
  async ({ slug, ...fields }) => {
    try {
      const patch = Object.fromEntries(Object.entries(fields).filter(([, v]) => v !== undefined))
      return api.asToolResult(await api.adminUpdateApp(slug, patch))
    } catch (error) {
      return toolError(error)
    }
  },
)

server.registerTool(
  'admin_delete_app',
  {
    description: 'Permanently delete an app (requires admin token)',
    inputSchema: { slug: z.string() },
  },
  async ({ slug }) => {
    try {
      return api.asToolResult(await api.adminDeleteApp(slug))
    } catch (error) {
      return toolError(error)
    }
  },
)

for (const action of ['approve', 'verify', 'reject'] as const) {
  server.registerTool(
    `admin_${action}_app`,
    {
      description: `${action.charAt(0).toUpperCase() + action.slice(1)} an app (requires admin token)`,
      inputSchema: { slug: z.string() },
    },
    async ({ slug }) => {
      try {
        return api.asToolResult(await api.adminSetStatus(slug, action))
      } catch (error) {
        return toolError(error)
      }
    },
  )
}

async function main() {
  const transport = new StdioServerTransport()
  await server.connect(transport)
}

main().catch((error) => {
  console.error('Fatal error:', error)
  process.exit(1)
})
