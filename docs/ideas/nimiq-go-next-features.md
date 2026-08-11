# nimiq-go — next features (handoff)

Repo: https://github.com/NimMiniApps/nimiq-go  
Current consumer: Nimiq Mini Apps catalog (`NimiqMiniApps` backend) uses **v0.3.1** for wallet login only.

## Context for the implementing agent

The catalog already consumes:

- `ParseAddress` / `Address.String()` — checksum + canonical spaced form on challenge/verify
- `AddressFromPublicKey`
- `VerifyMessageFrom` — Hub `signMessage` login (sig + address binding)

Catalog auth now uses `ParsePublicKey` / `ParseSignature` (local `decodeCryptoBytes` removed).

Still hand-rolled in the catalog (glue, not chain logic):

- Session cookies, nonce store, challenge message text (app-specific — keep out of SDK)

Do **not** pull RPC, `signer`, cashlinks, mnemonics, or tx builders into the catalog.

---

## Priority features to implement in nimiq-go

### P0 — Hub wire helpers (unblocks cleaner catalog auth)

Add decode helpers for the encodings Hub/Keyguard emit on `signMessage` results.

**API sketch:**

```go
// ParsePublicKey accepts std base64, raw/padded base64url, and optional 0x-hex.
func ParsePublicKey(s string) (ed25519.PublicKey, error)

// ParseSignature accepts the same encodings; length must be 64.
func ParseSignature(s string) ([]byte, error)
```

**Acceptance:**

- Golden vectors from a real Hub `signMessage` response (base64 pubkey + sig)
- Reject wrong lengths and all-zero pubkey (same policy as proofs)
- Docs snippet: wallet login = `ParseAddress` + `ParsePublicKey`/`ParseSignature` + `VerifyMessageFrom`
- Catalog can delete local `decodeCryptoBytes` for auth once this ships

### P1 — Docs: “server-side wallet login” recipe

Short README (or `docs/wallet-login.md`) showing the exact server flow mini-app backends need:

1. Parse & canonicalize address  
2. Issue app-owned challenge string (app-specific; SDK does not define format)  
3. Verify with `VerifyMessageFrom`  
4. Note: spacing/case irrelevant after `ParseAddress`

No challenge/nonce protocol in the SDK.

### P2 — ES256 / passkey message verification (when wallets need it)

`proof.go` already names `AlgorithmES256` for passkey tx proofs but cannot verify. If Nimiq Pay / Hub login moves to WebAuthn messages:

- Define how signed messages work for passkey accounts (mirror core-rs / Hub)
- `VerifyMessage` / `VerifyMessageFrom` (or sibling) for ES256
- Tests against Hub/Keyguard fixtures

Skip until there is a real consumer asking for passkey login.

### P3 — Nice-to-haves (only if a second consumer asks)

| Idea | Notes |
|------|--------|
| `MustParseAddress` already exists | Prefer `ParseAddress` in request handlers |
| Address equality helpers | `[20]byte` compare is enough; optional `Equal` is noise |
| Typed encode helpers (`PublicKey.Base64()`) | Only if multiple apps encode the same way |
| Challenge builders | **Out of scope** — app-specific SIWE-like text |
| Session / cookie helpers | **Out of scope** — not Nimiq |

---

## Explicit non-goals for this handoff

- Expanding catalog to use `rpc`, cashlinks, vesting, staking, or `signer`
- Wrapping more Albatross RPC methods “just in case” (use `Call` escape hatch)
- Changing Hub JS / `@nimiq/hub-api` (frontend stays on Hub)

---

## Suggested first PR

1. `ParsePublicKey` + `ParseSignature` + tests + README login recipe  
2. Tag e.g. `v0.3.1` or `v0.4.0`  
3. Follow-up in `NimiqMiniApps`: bump module, replace `decodeCryptoBytes` in auth path

## Done when

Another Go service can verify a Hub login signature with **only** nimiq-go imports (plus its own challenge string), without copying base64/hex decode or message-prefix hashing.
