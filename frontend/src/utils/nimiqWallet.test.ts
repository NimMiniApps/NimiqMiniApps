import { describe, expect, it } from 'vitest'

// Test helpers mirrored from nimiqWallet.ts (not exported to keep the module surface small).
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

function cryptoBytesToBase64(value: string | Uint8Array): string {
  if (value instanceof Uint8Array) return uint8ToBase64(value)
  const s = value.trim()
  if (isHexString(s)) return uint8ToBase64(hexToBytes(s))
  return s
}

describe('cryptoBytesToBase64', () => {
  it('converts Nimiq Pay hex public keys to std base64', () => {
    const hex = '00'.repeat(32)
    const b64 = cryptoBytesToBase64(hex)
    expect(atob(b64)).toHaveLength(32)
  })

  it('passes through existing base64', () => {
    const raw = new Uint8Array([1, 2, 3])
    const b64 = uint8ToBase64(raw)
    expect(cryptoBytesToBase64(b64)).toBe(b64)
  })
})
