import { ref } from 'vue'
import { authChallenge, authVerify, authMe, authLogout } from '../api'
import { chooseWalletAddress, signLoginChallenge } from '../utils/nimiqWallet'

const walletAddress = ref<string | null>(null)
const displayName = ref<string | null>(null)
const isAdmin = ref(false)
const checking = ref(true)
const loggingIn = ref(false)
const error = ref('')

let checked = false

async function applySession() {
  try {
    const me = await authMe()
    walletAddress.value = me.wallet_address
    displayName.value = me.display_name
    isAdmin.value = me.is_admin
  } catch {
    walletAddress.value = null
    displayName.value = null
    isAdmin.value = false
  } finally {
    checking.value = false
  }
}

export function useWalletAuth() {
  if (!checked) {
    checked = true
    applySession()
  }

  async function login() {
    loggingIn.value = true
    error.value = ''
    try {
      const address = await chooseWalletAddress()
      const challenge = await authChallenge(address)
      const signed = await signLoginChallenge(challenge.message, address)
      await authVerify({
        wallet_address: address,
        nonce: challenge.nonce,
        signature: signed.signature,
        public_key: signed.publicKey,
      })
      await applySession()
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to connect wallet'
      error.value = message === 'Failed to open popup'
        ? 'Your browser blocked the wallet popup. Allow popups for this site and try again.'
        : message
      throw err
    } finally {
      loggingIn.value = false
    }
  }

  async function logout() {
    await authLogout()
    walletAddress.value = null
    displayName.value = null
    isAdmin.value = false
  }

  async function refreshSession() {
    checking.value = true
    await applySession()
  }

  return { walletAddress, displayName, isAdmin, checking, loggingIn, error, login, logout, refreshSession }
}
