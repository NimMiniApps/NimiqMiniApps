import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./WalletLoginButton.vue', import.meta.url), 'utf8')

describe('WalletLoginButton admin menu', () => {
  it('shows analytics directly below moderation for admins', () => {
    const adminIndex = source.indexOf('to="/admin"')
    const analyticsIndex = source.indexOf('to="/admin/analytics"')
    const logoutIndex = source.indexOf('@click="handleLogout"')

    expect(adminIndex).toBeGreaterThan(-1)
    expect(analyticsIndex).toBeGreaterThan(adminIndex)
    expect(logoutIndex).toBeGreaterThan(analyticsIndex)
    expect(source).toMatch(
      /<RouterLink\s+v-if="isAdmin"\s+to="\/admin\/analytics"[\s\S]*?@click="closeMenu"[\s\S]*?>\s*Analytics\s*<\/RouterLink>/,
    )
  })
})
