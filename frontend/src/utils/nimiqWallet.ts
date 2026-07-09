import HubApi from '@nimiq/hub-api'
import { init } from '@nimiq/mini-app-sdk'

const HUB_URL = import.meta.env.VITE_NIMIQ_HUB_URL || 'https://hub.nimiq.com'
const APP_NAME = 'Nimiq Mini Apps'

let hubApi: HubApi | null = null
function getHubApi(): HubApi {
  if (!hubApi) hubApi = new HubApi(HUB_URL)
  return hubApi
}

export function hasInjectedNimiqPayHost(): boolean {
  return Boolean(window.nimiqPay || window.nimiq)
}

function uint8ToBase64(arr: Uint8Array): string {
  let s = ''
  for (let i = 0; i < arr.length; i++) s += String.fromCharCode(arr[i])
  return btoa(s)
}

function hexToBytes(hex: string): Uint8Array {
  const h = hex.replace(/^0x/i, '').trim()
  const bytes = new Uint8Array(h.length / 2)
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(h.slice(i * 2, i * 2 + 2), 16)
  }
  return bytes
}

function isHexString(value: string): boolean {
  const h = value.replace(/^0x/i, '').trim()
  return h.length > 0 && h.length % 2 === 0 && /^[0-9a-fA-F]+$/.test(h)
}

/** Nimiq Pay returns hex strings; Hub returns Uint8Array — normalize to std base64 for the API. */
function cryptoBytesToBase64(value: string | Uint8Array): string {
  if (value instanceof Uint8Array) return uint8ToBase64(value)
  const s = value.trim()
  if (isHexString(s)) return uint8ToBase64(hexToBytes(s))
  return s
}

function isSdkError(res: unknown): res is { error: { type: string; message: string } } {
  return typeof res === 'object' && res !== null && 'error' in res
}

export interface WalletSignature {
  signature: string
  publicKey: string
}

async function chooseHubAddress(): Promise<string> {
  const result = await getHubApi().chooseAddress({ appName: APP_NAME })
  return result.address
}

async function signWithHub(message: string, signer: string): Promise<WalletSignature> {
  const result = await getHubApi().signMessage({ appName: APP_NAME, message, signer })
  return {
    signature: uint8ToBase64(result.signature),
    publicKey: uint8ToBase64(result.signerPublicKey),
  }
}

let nimiqPayProviderPromise: ReturnType<typeof init> | null = null
function getNimiqPayProvider() {
  if (!nimiqPayProviderPromise) nimiqPayProviderPromise = init({ timeout: 3000 })
  return nimiqPayProviderPromise
}

async function chooseNimiqPayAddress(): Promise<string> {
  const provider = await getNimiqPayProvider()
  const accounts = await provider.listAccounts()
  if (isSdkError(accounts)) {
    throw new Error(accounts.error.message)
  }
  const address = String(accounts[0] || '')
  if (!address) throw new Error('No Nimiq Pay wallet account is available')
  return address
}

async function signWithNimiqPay(message: string): Promise<WalletSignature> {
  const provider = await getNimiqPayProvider()
  const result = await provider.sign({ message })
  if (isSdkError(result)) {
    throw new Error(result.error.message)
  }
  return {
    signature: cryptoBytesToBase64(result.signature),
    publicKey: cryptoBytesToBase64(result.publicKey),
  }
}

export async function chooseWalletAddress(): Promise<string> {
  return hasInjectedNimiqPayHost() ? chooseNimiqPayAddress() : chooseHubAddress()
}

export async function signLoginChallenge(message: string, address: string): Promise<WalletSignature> {
  return hasInjectedNimiqPayHost() ? signWithNimiqPay(message) : signWithHub(message, address)
}
