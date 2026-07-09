# Nimiq wallet login + app ratings/reviews Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let visitors connect a Nimiq wallet (Hub popup on desktop, injected Nimiq Pay signer in-app) and leave a 1-5 star rating with an optional written review on a directory-listed app.

**Architecture:** A stdlib `net/http` backend (no new framework) adds a challenge/verify signature-auth flow producing a stateless HMAC-signed cookie, plus a `app_reviews` table with one editable review per wallet per app. The Vue 3 frontend adds a wallet-connect composable and two small review components wired into the existing app detail page.

**Tech Stack:** Go 1.25 stdlib `net/http` + `pgx/v5` + Postgres (backend, existing); Vue 3 `<script setup>` + TypeScript + Tailwind v4 (frontend, existing); `@nimiq/hub-api` and `@nimiq/mini-app-sdk` (new frontend deps); `golang.org/x/crypto/blake2b` (new backend dep, for Nimiq address derivation).

## Global Constraints

- No new backend web framework or session store — stdlib `net/http` + the existing `authMiddleware`/`writeError`/`writeJSON` conventions in `backend/handlers.go` only.
- One review per `(app_id, wallet_address)` — a second write from the same wallet edits, never duplicates.
- Rating required 1-5; review body optional, 0-1000 chars.
- Wallet session cookie: `HttpOnly; Secure; SameSite=Lax`, 7-day sliding expiry, no server-side revocation.
- Reviews auto-publish; admin can delete via the existing admin-token pattern.
- Match existing Tailwind design tokens (`border-line`, `bg-surface`, `bg-surface-2`, `text-accent-ink`, `bg-accent`, `text-muted`) — no ad hoc colors, no gold/yellow.
- Follow the repo's existing table-driven `*_test.go` stdlib `testing` pattern (see `backend/validate_test.go`, `backend/pagination_test.go`).

---

### Task 1: `app_reviews` migration

**Files:**
- Create: `backend/migrations/009_reviews.sql`

**Interfaces:**
- Produces: table `app_reviews(id, app_id, wallet_address, rating, body, created_at, updated_at)` with `UNIQUE (app_id, wallet_address)`, consumed by Task 4.

- [ ] **Step 1: Write the migration**

```sql
CREATE TABLE IF NOT EXISTS app_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    wallet_address TEXT NOT NULL,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    body TEXT NOT NULL DEFAULT '' CHECK (char_length(body) <= 1000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (app_id, wallet_address)
);

CREATE INDEX IF NOT EXISTS app_reviews_app_id_idx ON app_reviews (app_id);
```

- [ ] **Step 2: Verify it applies cleanly**

Run: `cd backend && go run . &` (with `DATABASE_URL` pointed at a local dev Postgres, e.g. via `docker-compose up -d db`), watch the logs.
Expected: log line `"applied migration" name=009_reviews.sql`, then stop the process (`kill %1` or Ctrl-C).

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/009_reviews.sql
git commit -m "Add app_reviews table migration"
```

---

### Task 2: Nimiq address derivation (`nimiqaddress.go`)

Ports the tested address-derivation logic from the sibling `nimiq-2048` repo (`backend/platform/nimiq/address.go` there) so the verify handler in Task 3 can confirm a signed message's public key actually belongs to the claimed address.

**Files:**
- Create: `backend/nimiqaddress.go`
- Test: `backend/nimiqaddress_test.go`
- Modify: `backend/go.mod` (add `golang.org/x/crypto`)

**Interfaces:**
- Produces: `userFriendlyAddressFromPublicKey(pub ed25519.PublicKey) string`, `normalizeUserFriendlyAddress(s string) string`, `publicKeyMatchesClaimedAddress(pub ed25519.PublicKey, claimed string) bool` — consumed by Task 3's `authVerify`.

- [ ] **Step 1: Add the dependency**

Run: `cd backend && go get golang.org/x/crypto@latest`
Expected: `go.mod` gains a direct `require golang.org/x/crypto vX.Y.Z` line.

- [ ] **Step 2: Write the failing test**

```go
// backend/nimiqaddress_test.go
package main

import (
	"crypto/ed25519"
	"testing"
)

func TestPublicKeyMatchesClaimedAddress(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	addr := userFriendlyAddressFromPublicKey(pub)
	if !publicKeyMatchesClaimedAddress(pub, addr) {
		t.Fatal("expected match for derived address")
	}
	if !publicKeyMatchesClaimedAddress(pub, normalizeUserFriendlyAddress(addr)) {
		t.Fatal("expected match for normalized form")
	}
	wrongPub, _, _ := ed25519.GenerateKey(nil)
	if publicKeyMatchesClaimedAddress(wrongPub, addr) {
		t.Fatal("expected mismatch for wrong key")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd backend && go test ./... -run TestPublicKeyMatchesClaimedAddress -v`
Expected: FAIL — `undefined: userFriendlyAddressFromPublicKey` (function not defined yet).

- [ ] **Step 4: Write the implementation**

```go
// backend/nimiqaddress.go
package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base32"
	"errors"
	"strconv"
	"strings"

	"golang.org/x/crypto/blake2b"
)

var errInvalidIBANChar = errors.New("invalid IBAN character")

// ibanEncoding is Nimiq's base32 alphabet for user-friendly addresses (no I/O).
var ibanEncoding = base32.NewEncoding("0123456789ABCDEFGHJKLMNPQRSTUVXY")

// publicKeyToAddressBytes derives the 20-byte on-chain address from an Ed25519 public key
// (first 20 bytes of Blake2b-256(pubkey), matching Nimiq Core).
func publicKeyToAddressBytes(pub ed25519.PublicKey) [20]byte {
	h := blake2b.Sum256(pub)
	var addr [20]byte
	copy(addr[:], h[:20])
	return addr
}

// userFriendlyAddressFromPublicKey returns the canonical user-friendly Nimiq address
// (with spaces), e.g. "NQ12 3456 ...", for the given Ed25519 public key.
func userFriendlyAddressFromPublicKey(pub ed25519.PublicKey) string {
	addr := publicKeyToAddressBytes(pub)
	return addressToUserFriendly(&addr)
}

// normalizeUserFriendlyAddress strips spaces and uppercases for comparisons.
func normalizeUserFriendlyAddress(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	return strings.ToUpper(s)
}

// publicKeyMatchesClaimedAddress reports whether the Ed25519 public key corresponds to
// the claimed Nimiq user-friendly address (spacing and case ignored).
func publicKeyMatchesClaimedAddress(pub ed25519.PublicKey, claimedAddress string) bool {
	if len(pub) != ed25519.PublicKeySize {
		return false
	}
	expected := userFriendlyAddressFromPublicKey(pub)
	return normalizeUserFriendlyAddress(expected) == normalizeUserFriendlyAddress(claimedAddress)
}

// addressToUserFriendly encodes a 20-byte address to the NQ… user-friendly form (with spaces).
func addressToUserFriendly(addr *[20]byte) string {
	var noSpaces [36]byte
	copy(noSpaces[0:4], "NQ00")
	ibanEncoding.Encode(noSpaces[4:], addr[:])

	check, _ := calcIBANAddressCheck(&noSpaces)
	check = 98 - check

	var b strings.Builder
	b.WriteString("NQ")
	b.Write([]byte{
		0x30 + (uint8(check%100) / 10),
		0x30 + uint8(check%10),
	})
	for i := 4; i < 36; i += 4 {
		b.WriteByte(' ')
		b.Write(noSpaces[i : i+4])
	}
	return b.String()
}

func calcIBANAddressCheck(userFriendly *[36]byte) (uint8, error) {
	var sumBuffer bytes.Buffer

	nextChars := func(slice []byte) error {
		for _, char := range slice {
			switch {
			case char > 0x60 && char <= 0x7A:
				char -= 0x20
				fallthrough
			case char > 0x40 && char <= 0x5A:
				num := char - 0x37
				sumBuffer.WriteString(strconv.FormatUint(uint64(num), 10))
			case char >= 0x30 && char <= 0x39:
				sumBuffer.WriteByte(char)
			default:
				return errInvalidIBANChar
			}
		}
		return nil
	}

	if err := nextChars(userFriendly[4:]); err != nil {
		return 0, err
	}
	if err := nextChars(userFriendly[0:4]); err != nil {
		return 0, err
	}

	sum := sumBuffer.Bytes()
	var tmpBuffer bytes.Buffer
	blockCount := (len(sum) + 5) / 6

	for i := 0; true; i++ {
		offset := i * 6
		var stop int
		if len(sum) <= offset+6 {
			stop = len(sum)
		} else {
			stop = offset + 6
		}
		block := sum[offset:stop]
		tmpBuffer.Write(block)
		tmp := tmpBuffer.String()
		tmpNum, _ := strconv.ParseUint(tmp, 10, 64)
		tmpNum %= 97
		if (i + 1) < blockCount {
			tmpBuffer.Reset()
			tmpBuffer.WriteString(strconv.FormatUint(tmpNum, 10))
		} else {
			return uint8(tmpNum), nil
		}
	}
	panic("unreachable")
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd backend && go test ./... -run TestPublicKeyMatchesClaimedAddress -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/nimiqaddress.go backend/nimiqaddress_test.go backend/go.mod backend/go.sum
git commit -m "Add Nimiq address derivation for signature verification"
```

---

### Task 3: Wallet auth — nonce store, signed cookie, challenge/verify/me/logout (`walletauth.go`)

**Files:**
- Create: `backend/walletauth.go`
- Test: `backend/walletauth_test.go`
- Modify: `backend/handlers.go:98-100` (add fields to `server` struct)

**Interfaces:**
- Consumes: `writeJSON`, `writeError` (`backend/handlers.go`), `publicKeyMatchesClaimedAddress` (Task 2).
- Produces: `(s *server) authChallenge`, `(s *server) authVerify`, `(s *server) authMe`, `(s *server) authLogout` (all `func(w http.ResponseWriter, r *http.Request)` except `authMe`/`authVerify`'s inner handlers which take an extra `address string`), `walletAuthMiddleware(secret string, next func(w http.ResponseWriter, r *http.Request, address string)) http.HandlerFunc` — consumed by Task 4 (reviews) and Task 5 (main.go wiring). `server` struct gains `nonces *nonceStore` and `walletAuthSecret string`.

- [ ] **Step 1: Add fields to the `server` struct**

In `backend/handlers.go`, change:

```go
type server struct {
	pool *pgxpool.Pool
}
```

to:

```go
type server struct {
	pool             *pgxpool.Pool
	nonces           *nonceStore
	walletAuthSecret string
	reviewLimiter    *rateLimiter
}
```

(`rateLimiter` is defined in Task 4; this struct literal won't compile until that task lands — that's expected, both tasks are part of the same backend build and Task 4 follows immediately.)

- [ ] **Step 2: Write the failing tests**

```go
// backend/walletauth_test.go
package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"testing"
	"time"
)

func TestWalletCookieRoundTrip(t *testing.T) {
	secret := "test-secret"
	address := "NQ07 0000 0000 0000 0000 0000 0000 0000 0000"
	value := signWalletCookie(secret, address, time.Now().Add(time.Hour))
	got, err := verifyWalletCookie(secret, value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != address {
		t.Fatalf("got %q, want %q", got, address)
	}
}

func TestWalletCookieExpired(t *testing.T) {
	secret := "test-secret"
	value := signWalletCookie(secret, "NQ07", time.Now().Add(-time.Hour))
	if _, err := verifyWalletCookie(secret, value); err == nil {
		t.Fatal("expected expired cookie to fail verification")
	}
}

func TestWalletCookieTampered(t *testing.T) {
	secret := "test-secret"
	value := signWalletCookie(secret, "NQ07", time.Now().Add(time.Hour))
	tampered := value[:len(value)-1] + "x"
	if _, err := verifyWalletCookie(secret, tampered); err == nil {
		t.Fatal("expected tampered cookie to fail verification")
	}
}

func TestVerifyWalletSignatureRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	message := "test message"
	prefix := "\x16Nimiq Signed Message:\n"
	payload := prefix + strconv.Itoa(len(message)) + message
	hash := sha256.Sum256([]byte(payload))
	sig := ed25519.Sign(priv, hash[:])
	sigB64 := base64.StdEncoding.EncodeToString(sig)
	pubB64 := base64.StdEncoding.EncodeToString(pub)

	gotPub, err := verifyWalletSignature(message, sigB64, pubB64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotPub.Equal(pub) {
		t.Fatal("public key mismatch")
	}
}

func TestNonceStoreReplay(t *testing.T) {
	ns := newNonceStore()
	nonce, _, err := ns.create("NQ07")
	if err != nil {
		t.Fatal(err)
	}
	if err := ns.markUsed(nonce, "NQ07"); err != nil {
		t.Fatalf("first use should succeed: %v", err)
	}
	if err := ns.markUsed(nonce, "NQ07"); err == nil {
		t.Fatal("expected replay to fail")
	}
}

func TestNonceStoreAddressMismatch(t *testing.T) {
	ns := newNonceStore()
	nonce, _, err := ns.create("NQ07")
	if err != nil {
		t.Fatal(err)
	}
	if err := ns.markUsed(nonce, "NQ08"); err == nil {
		t.Fatal("expected address mismatch to fail")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd backend && go test ./... -run 'TestWalletCookie|TestVerifyWalletSignature|TestNonceStore' -v`
Expected: FAIL — undefined symbols (`signWalletCookie`, `verifyWalletCookie`, `verifyWalletSignature`, `newNonceStore`).

- [ ] **Step 4: Write the implementation**

```go
// backend/walletauth.go
package main

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	walletCookieName = "wallet_session"
	walletSessionTTL = 7 * 24 * time.Hour
	authChallengeTTL = 5 * time.Minute
	authDomain       = "nimiqminiapps.com"
	authPurpose      = "login to leave an app rating and review"
)

// --- nonce store (in-memory; single backend instance — move to a table if it ever runs >1 replica) ---

type nonceEntry struct {
	address string
	expires time.Time
	used    bool
}

type nonceStore struct {
	mu      sync.Mutex
	entries map[string]nonceEntry
}

func newNonceStore() *nonceStore {
	return &nonceStore{entries: map[string]nonceEntry{}}
}

func (n *nonceStore) create(address string) (nonce string, expires time.Time, err error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, err
	}
	nonce = hex.EncodeToString(buf)
	expires = time.Now().Add(authChallengeTTL)

	n.mu.Lock()
	defer n.mu.Unlock()
	for k, e := range n.entries {
		if time.Now().After(e.expires) {
			delete(n.entries, k)
		}
	}
	n.entries[nonce] = nonceEntry{address: address, expires: expires}
	return nonce, expires, nil
}

func (n *nonceStore) get(nonce string) (nonceEntry, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	e, ok := n.entries[nonce]
	return e, ok
}

func (n *nonceStore) markUsed(nonce, address string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	e, ok := n.entries[nonce]
	if !ok {
		return errors.New("unknown or expired nonce")
	}
	if e.used {
		return errors.New("nonce already used")
	}
	if time.Now().After(e.expires) {
		delete(n.entries, nonce)
		return errors.New("nonce expired")
	}
	if e.address != address {
		return errors.New("nonce does not match address")
	}
	e.used = true
	n.entries[nonce] = e
	return nil
}

// --- signed session cookie: base64(address|expiry) + "." + base64(hmac-sha256) ---

func signWalletCookie(secret, address string, expires time.Time) string {
	payload := address + "|" + strconv.FormatInt(expires.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." +
		base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func verifyWalletCookie(secret, value string) (address string, err error) {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return "", errors.New("malformed session cookie")
	}
	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("malformed session cookie")
	}
	sigRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("malformed session cookie")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadRaw)
	if !hmac.Equal(mac.Sum(nil), sigRaw) {
		return "", errors.New("invalid session signature")
	}
	payload := string(payloadRaw)
	idx := strings.LastIndex(payload, "|")
	if idx < 0 {
		return "", errors.New("malformed session payload")
	}
	address = payload[:idx]
	expUnix, err := strconv.ParseInt(payload[idx+1:], 10, 64)
	if err != nil {
		return "", errors.New("malformed session expiry")
	}
	if time.Now().Unix() > expUnix {
		return "", errors.New("session expired")
	}
	return address, nil
}

func setWalletCookie(w http.ResponseWriter, secret, address string) {
	expires := time.Now().Add(walletSessionTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     walletCookieName,
		Value:    signWalletCookie(secret, address, expires),
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearWalletCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     walletCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// walletAuthMiddleware validates the signed cookie, refreshes it (sliding expiry),
// then calls next with the authenticated wallet address.
func walletAuthMiddleware(secret string, next func(w http.ResponseWriter, r *http.Request, address string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(walletCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "wallet login required")
			return
		}
		address, err := verifyWalletCookie(secret, cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "wallet login required")
			return
		}
		setWalletCookie(w, secret, address)
		next(w, r, address)
	}
}

// --- signed-message challenge/verify ---

func buildWalletAuthMessage(address, nonce string, expires time.Time) string {
	timestamp := expires.Add(-authChallengeTTL).UTC().Format(time.RFC3339)
	return "Nimiq Mini Apps login challenge:" +
		"\naddress=" + address +
		"\nnonce=" + nonce +
		"\ntimestamp=" + timestamp +
		"\ndomain=" + authDomain +
		"\npurpose=" + authPurpose +
		"\nexpires=" + expires.UTC().Format(time.RFC3339)
}

func verifyWalletSignature(message, signatureB64, publicKeyB64 string) (ed25519.PublicKey, error) {
	sig, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, errors.New("invalid signature encoding")
	}
	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, errors.New("invalid public key encoding")
	}
	if len(pubBytes) != ed25519.PublicKeySize || len(sig) != ed25519.SignatureSize {
		return nil, errors.New("invalid signature or public key size")
	}
	pub := ed25519.PublicKey(pubBytes)
	prefix := "\x16Nimiq Signed Message:\n"
	payload := prefix + strconv.Itoa(len(message)) + message
	hash := sha256.Sum256([]byte(payload))
	if !ed25519.Verify(pub, hash[:], sig) {
		return nil, errors.New("signature verification failed")
	}
	return pub, nil
}

func (s *server) authChallenge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"wallet_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	address := strings.TrimSpace(req.Address)
	if address == "" {
		writeError(w, http.StatusBadRequest, "wallet_address is required")
		return
	}
	nonce, expires, err := s.nonces.create(address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate challenge")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"nonce":   nonce,
		"message": buildWalletAuthMessage(address, nonce, expires),
	})
}

func (s *server) authVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address   string `json:"wallet_address"`
		Nonce     string `json:"nonce"`
		Signature string `json:"signature"`
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	address := strings.TrimSpace(req.Address)
	entry, ok := s.nonces.get(req.Nonce)
	if !ok || entry.address != address || time.Now().After(entry.expires) {
		writeError(w, http.StatusUnauthorized, "invalid or expired challenge")
		return
	}
	message := buildWalletAuthMessage(address, req.Nonce, entry.expires)
	pub, err := verifyWalletSignature(message, req.Signature, req.PublicKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if !publicKeyMatchesClaimedAddress(pub, address) {
		writeError(w, http.StatusUnauthorized, "public key does not match claimed address")
		return
	}
	if err := s.nonces.markUsed(req.Nonce, address); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	setWalletCookie(w, s.walletAuthSecret, address)
	writeJSON(w, http.StatusOK, map[string]string{"wallet_address": address})
}

func (s *server) authMe(w http.ResponseWriter, r *http.Request, address string) {
	writeJSON(w, http.StatusOK, map[string]string{"wallet_address": address})
}

func (s *server) authLogout(w http.ResponseWriter, r *http.Request) {
	clearWalletCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend && go build ./... && go test ./... -run 'TestWalletCookie|TestVerifyWalletSignature|TestNonceStore' -v`
Expected: build succeeds (note: `reviewLimiter`/`rateLimiter` reference from Step 1 will not compile until Task 4 adds `rateLimiter` — if running this task in isolation, temporarily comment out the `reviewLimiter *rateLimiter` field, run the tests, then restore it before Task 4). All four/five tests PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/walletauth.go backend/walletauth_test.go backend/handlers.go
git commit -m "Add wallet challenge/verify auth with stateless signed session cookie"
```

---

### Task 4: App reviews CRUD (`reviews.go`)

**Files:**
- Create: `backend/reviews.go`
- Test: `backend/reviews_test.go`

**Interfaces:**
- Consumes: `s.pool` (`backend/handlers.go`), `writeJSON`/`writeError`, `walletAuthMiddleware` (Task 3).
- Produces: `(s *server) listReviews`, `(s *server) upsertReview(w, r, address string)`, `(s *server) deleteOwnReview(w, r, address string)`, `(s *server) adminDeleteReview`, `validateReviewInput(rating int, body string) string`, `rateLimiter` type + `newRateLimiter(limit int, window time.Duration) *rateLimiter` — consumed by Task 5 (main.go routes) and referenced by Task 3's `server` struct.

- [ ] **Step 1: Write the failing tests**

```go
// backend/reviews_test.go
package main

import (
	"strings"
	"testing"
	"time"
)

func TestValidateReviewInput(t *testing.T) {
	cases := []struct {
		name    string
		rating  int
		body    string
		wantErr bool
	}{
		{"valid with body", 5, "Great app!", false},
		{"valid rating only", 3, "", false},
		{"rating too low", 0, "", true},
		{"rating too high", 6, "", true},
		{"body too long", 4, strings.Repeat("a", 1001), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReviewInput(tc.rating, tc.body)
			if (err != "") != tc.wantErr {
				t.Fatalf("validateReviewInput(%d, ...) error=%q, wantErr=%v", tc.rating, err, tc.wantErr)
			}
		})
	}
}

func TestRateLimiterAllow(t *testing.T) {
	rl := newRateLimiter(2, time.Minute)
	if !rl.allow("wallet-a") {
		t.Fatal("expected first request to be allowed")
	}
	if !rl.allow("wallet-a") {
		t.Fatal("expected second request to be allowed")
	}
	if rl.allow("wallet-a") {
		t.Fatal("expected third request to be denied")
	}
	if !rl.allow("wallet-b") {
		t.Fatal("expected a different key to have its own bucket")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend && go test ./... -run 'TestValidateReviewInput|TestRateLimiterAllow' -v`
Expected: FAIL — undefined: `validateReviewInput`, `newRateLimiter`.

- [ ] **Step 3: Write the implementation**

```go
// backend/reviews.go
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Review struct {
	ID            string    `json:"id"`
	AppID         string    `json:"app_id"`
	WalletAddress string    `json:"wallet_address"`
	Rating        int       `json:"rating"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func validateReviewInput(rating int, body string) string {
	if rating < 1 || rating > 5 {
		return "rating must be between 1 and 5"
	}
	if len(body) > 1000 {
		return "body must be at most 1000 characters"
	}
	return ""
}

// rateLimiter is a per-key fixed-window limiter (in-memory; single backend instance).
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{attempts: map[string][]time.Time{}, limit: limit, window: window}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	kept := rl.attempts[key][:0]
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.limit {
		rl.attempts[key] = kept
		return false
	}
	rl.attempts[key] = append(kept, now)
	return true
}

func (s *server) appIDForSlug(ctx context.Context, slug string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id FROM apps WHERE slug=$1`, slug).Scan(&id)
	return id, err
}

func (s *server) listReviews(w http.ResponseWriter, r *http.Request) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT id, app_id, wallet_address, rating, body, created_at, updated_at
		FROM app_reviews WHERE app_id=$1 ORDER BY created_at DESC`, appID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	reviews := []Review{}
	for rows.Next() {
		var rv Review
		if err := rows.Scan(&rv.ID, &rv.AppID, &rv.WalletAddress, &rv.Rating, &rv.Body, &rv.CreatedAt, &rv.UpdatedAt); err != nil {
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

func (s *server) upsertReview(w http.ResponseWriter, r *http.Request, address string) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if !s.reviewLimiter.allow(address) {
		writeError(w, http.StatusTooManyRequests, "too many reviews, try again later")
		return
	}
	var req struct {
		Rating int    `json:"rating"`
		Body   string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	body := strings.TrimSpace(req.Body)
	if msg := validateReviewInput(req.Rating, body); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	var rv Review
	err = s.pool.QueryRow(r.Context(), `
		INSERT INTO app_reviews (app_id, wallet_address, rating, body)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (app_id, wallet_address)
		DO UPDATE SET rating=$3, body=$4, updated_at=now()
		RETURNING id, app_id, wallet_address, rating, body, created_at, updated_at`,
		appID, address, req.Rating, body).
		Scan(&rv.ID, &rv.AppID, &rv.WalletAddress, &rv.Rating, &rv.Body, &rv.CreatedAt, &rv.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rv)
}

func (s *server) deleteOwnReview(w http.ResponseWriter, r *http.Request, address string) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	tag, err := s.pool.Exec(r.Context(), `DELETE FROM app_reviews WHERE app_id=$1 AND wallet_address=$2`, appID, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) adminDeleteReview(w http.ResponseWriter, r *http.Request) {
	tag, err := s.pool.Exec(r.Context(), `DELETE FROM app_reviews WHERE id=$1`, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd backend && go build ./... && go test ./... -run 'TestValidateReviewInput|TestRateLimiterAllow' -v`
Expected: build succeeds (the `server.reviewLimiter` field from Task 3 now resolves), both tests PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/reviews.go backend/reviews_test.go
git commit -m "Add app review CRUD handlers with rate limiting"
```

---

### Task 5: Wire routes into `main.go`

**Files:**
- Modify: `backend/main.go:111-183`

**Interfaces:**
- Consumes: `s.authChallenge`, `s.authVerify`, `s.authMe`, `s.authLogout`, `walletAuthMiddleware` (Task 3); `s.listReviews`, `s.upsertReview`, `s.deleteOwnReview`, `s.adminDeleteReview`, `newRateLimiter` (Task 4); `newNonceStore` (Task 3).

- [ ] **Step 1: Add the `WALLET_AUTH_SECRET` env var and server construction**

In `backend/main.go`, after the existing `adminToken` block (around line 121-123):

```go
	adminToken := env("ADMIN_TOKEN", "")
	walletAuthSecret := env("WALLET_AUTH_SECRET", "")
	addr := env("HTTP_ADDR", ":8080")
	corsOrigins := strings.Split(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}
	if adminToken == "" {
		slog.Warn("ADMIN_TOKEN is empty; admin endpoints are disabled")
	}
	if walletAuthSecret == "" {
		slog.Warn("WALLET_AUTH_SECRET is empty; wallet login endpoints are disabled")
	}
```

- [ ] **Step 2: Wire the new fields into `server{...}`**

Change:

```go
	s := &server{pool: pool}
```

to:

```go
	s := &server{
		pool:             pool,
		nonces:           newNonceStore(),
		walletAuthSecret: walletAuthSecret,
		reviewLimiter:    newRateLimiter(5, time.Hour),
	}
```

- [ ] **Step 3: Register the routes**

After the existing `mux.HandleFunc("POST /api/apps/{slug}/request-update", ...)` line, add:

```go
	mux.HandleFunc("POST /api/auth/challenge", s.authChallenge)
	mux.HandleFunc("POST /api/auth/verify", s.authVerify)
	mux.HandleFunc("GET /api/auth/me", walletAuthMiddleware(walletAuthSecret, s.authMe))
	mux.HandleFunc("POST /api/auth/logout", s.authLogout)
	mux.HandleFunc("GET /api/apps/{slug}/reviews", s.listReviews)
	mux.HandleFunc("POST /api/apps/{slug}/reviews", walletAuthMiddleware(walletAuthSecret, s.upsertReview))
	mux.HandleFunc("DELETE /api/apps/{slug}/reviews", walletAuthMiddleware(walletAuthSecret, s.deleteOwnReview))
```

After the existing `mux.HandleFunc("POST /api/admin/apps/{slug}/reject", ...)` line, add:

```go
	mux.HandleFunc("DELETE /api/admin/apps/{slug}/reviews/{id}", authMiddleware(adminToken, s.adminDeleteReview))
```

- [ ] **Step 4: Send `Access-Control-Allow-Credentials` for cookie auth**

In `corsMiddleware` (`backend/main.go`), inside the `if origin != "" && ...` block, add one line so cross-origin dev setups (frontend on a different port than the backend) can still send/receive the wallet cookie:

```go
		if origin != "" && (slices.Contains(allowed, "*") || slices.Contains(allowed, origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
```

- [ ] **Step 5: Build and smoke-test**

Run: `cd backend && go build ./... && go vet ./...`
Expected: no errors.

Run: `cd backend && WALLET_AUTH_SECRET=dev-secret ADMIN_TOKEN=dev-admin go run .` (with a local Postgres reachable via `DATABASE_URL`), then in another terminal:

```bash
curl -s -X POST localhost:8080/api/auth/challenge -H 'Content-Type: application/json' \
  -d '{"wallet_address":"NQ07 TEST"}'
```

Expected: JSON body with `nonce` and `message` fields. Stop the server after checking.

- [ ] **Step 6: Commit**

```bash
git add backend/main.go
git commit -m "Wire wallet auth and app review routes into the HTTP server"
```

---

### Task 6: Frontend wallet signing utility (`nimiqWallet.ts`)

**Files:**
- Modify: `frontend/package.json` (add `@nimiq/hub-api`, `@nimiq/mini-app-sdk`)
- Create: `frontend/src/utils/nimiqWallet.ts`

**Interfaces:**
- Produces: `hasInjectedNimiqPayHost(): boolean`, `chooseWalletAddress(): Promise<string>`, `signLoginChallenge(message: string, address: string): Promise<{ signature: string; publicKey: string }>` — consumed by Task 8 (`useWalletAuth`).

- [ ] **Step 1: Install the dependencies**

Run: `cd frontend && npm install @nimiq/hub-api @nimiq/mini-app-sdk`
Expected: both packages added to `frontend/package.json` `dependencies` and `frontend/package-lock.json` updated.

- [ ] **Step 2: Write the utility**

```ts
// frontend/src/utils/nimiqWallet.ts
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
  const w = window as Window & { nimiq?: unknown; nimiqPay?: unknown }
  return Boolean(w.nimiqPay || w.nimiq)
}

function uint8ToBase64(arr: Uint8Array): string {
  let s = ''
  for (let i = 0; i < arr.length; i++) s += String.fromCharCode(arr[i])
  return btoa(s)
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
  const address = String(accounts[0] || '')
  if (!address) throw new Error('No Nimiq Pay wallet account is available')
  return address
}

async function signWithNimiqPay(message: string): Promise<WalletSignature> {
  const provider = await getNimiqPayProvider()
  const result = await provider.sign({ message })
  if (!result.publicKey || !result.signature) {
    throw new Error('Nimiq Pay returned an incomplete signature')
  }
  return {
    signature: uint8ToBase64(result.signature),
    publicKey: uint8ToBase64(result.publicKey),
  }
}

export async function chooseWalletAddress(): Promise<string> {
  return hasInjectedNimiqPayHost() ? chooseNimiqPayAddress() : chooseHubAddress()
}

export async function signLoginChallenge(message: string, address: string): Promise<WalletSignature> {
  return hasInjectedNimiqPayHost() ? signWithNimiqPay(message) : signWithHub(message, address)
}
```

- [ ] **Step 3: Typecheck**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors (if `@nimiq/hub-api` ships no bundled types, add a `frontend/src/types/nimiq-hub-api.d.ts` with `declare module '@nimiq/hub-api'` — check the package's own `package.json` `types` field first via `cat node_modules/@nimiq/hub-api/package.json | grep types` before adding a shim, since most current Nimiq SDKs ship their own `.d.ts`).

- [ ] **Step 4: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/utils/nimiqWallet.ts
git commit -m "Add Nimiq Hub / Nimiq Pay wallet signing utility"
```

---

### Task 7: API client additions (`api.ts`)

**Files:**
- Modify: `frontend/src/api.ts` (append near the existing admin section, after line 242)

**Interfaces:**
- Consumes: `request<T>` (existing, `frontend/src/api.ts:114`).
- Produces: `AppReview`, `AppReviewsResponse` types, `authChallenge`, `authVerify`, `authMe`, `authLogout`, `listAppReviews`, `submitAppReview`, `deleteOwnAppReview` — consumed by Task 8 (`useWalletAuth`) and Task 10 (review components).

- [ ] **Step 1: Append to `frontend/src/api.ts`**

```ts
// --- wallet auth & reviews ---

export interface AppReview {
  id: string
  app_id: string
  wallet_address: string
  rating: number
  body: string
  created_at: string
  updated_at: string
}

export interface AppReviewsResponse {
  items: AppReview[]
  average: number
  count: number
}

export const authChallenge = (wallet_address: string) =>
  request<{ nonce: string; message: string }>('/api/auth/challenge', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ wallet_address }),
  })

export const authVerify = (payload: {
  wallet_address: string
  nonce: string
  signature: string
  public_key: string
}) =>
  request<{ wallet_address: string }>('/api/auth/verify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(payload),
  })

export const authMe = () =>
  request<{ wallet_address: string }>('/api/auth/me', { credentials: 'include' })

export const authLogout = () =>
  request<void>('/api/auth/logout', { method: 'POST', credentials: 'include' })

export const listAppReviews = (slug: string) =>
  request<AppReviewsResponse>(`/api/apps/${encodeURIComponent(slug)}/reviews`)

export const submitAppReview = (slug: string, rating: number, body: string) =>
  request<AppReview>(`/api/apps/${encodeURIComponent(slug)}/reviews`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ rating, body }),
  })

export const deleteOwnAppReview = (slug: string) =>
  request<void>(`/api/apps/${encodeURIComponent(slug)}/reviews`, {
    method: 'DELETE',
    credentials: 'include',
  })
```

- [ ] **Step 2: Typecheck**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api.ts
git commit -m "Add wallet auth and app review API client functions"
```

---

### Task 8: `useWalletAuth` composable

**Files:**
- Create: `frontend/src/composables/useWalletAuth.ts`

**Interfaces:**
- Consumes: `authChallenge`, `authVerify`, `authMe`, `authLogout` (Task 7); `chooseWalletAddress`, `signLoginChallenge` (Task 6).
- Produces: `useWalletAuth(): { walletAddress: Ref<string | null>; checking: Ref<boolean>; loggingIn: Ref<boolean>; error: Ref<string>; login(): Promise<void>; logout(): Promise<void> }` — consumed by Task 9 (`WalletLoginButton.vue`) and Task 10 (review components, to know the current wallet for "is this my review").

- [ ] **Step 1: Write the composable**

```ts
// frontend/src/composables/useWalletAuth.ts
import { ref } from 'vue'
import { authChallenge, authVerify, authMe, authLogout } from '../api'
import { chooseWalletAddress, signLoginChallenge } from '../utils/nimiqWallet'

const walletAddress = ref<string | null>(null)
const checking = ref(true)
const loggingIn = ref(false)
const error = ref('')

let checked = false

async function checkSession() {
  try {
    const me = await authMe()
    walletAddress.value = me.wallet_address
  } catch {
    walletAddress.value = null
  } finally {
    checking.value = false
  }
}

export function useWalletAuth() {
  if (!checked) {
    checked = true
    checkSession()
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
      walletAddress.value = address
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to connect wallet'
      throw err
    } finally {
      loggingIn.value = false
    }
  }

  async function logout() {
    await authLogout()
    walletAddress.value = null
  }

  return { walletAddress, checking, loggingIn, error, login, logout }
}
```

- [ ] **Step 2: Typecheck**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/composables/useWalletAuth.ts
git commit -m "Add useWalletAuth composable for wallet login state"
```

---

### Task 9: `WalletLoginButton.vue` and header wiring

**Files:**
- Create: `frontend/src/components/WalletLoginButton.vue`
- Modify: `frontend/src/App.vue:1-11` (import + destructure), `frontend/src/App.vue:87-88` (mount before the theme toggle)

**Interfaces:**
- Consumes: `useWalletAuth` (Task 8).

- [ ] **Step 1: Write the component**

```vue
<!-- frontend/src/components/WalletLoginButton.vue -->
<script setup lang="ts">
import { useWalletAuth } from '../composables/useWalletAuth'

const { walletAddress, checking, loggingIn, error, login, logout } = useWalletAuth()

function truncate(address: string): string {
  return address.length > 12 ? address.slice(0, 6) + '…' + address.slice(-4) : address
}
</script>

<template>
  <div class="flex items-center gap-2">
    <span v-if="checking" class="text-sm text-muted">…</span>
    <template v-else-if="walletAddress">
      <span class="font-mono text-sm text-accent-ink">{{ truncate(walletAddress) }}</span>
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

- [ ] **Step 2: Mount it in the header**

In `frontend/src/App.vue`, change the imports (line 1-7):

```vue
<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import StoreBadges from './components/StoreBadges.vue'
import WalletLoginButton from './components/WalletLoginButton.vue'
import { useAdminAuth } from './composables/useAdminAuth'
import { useI18n } from './composables/useI18n'
import { CATALOG_ISSUES_URL } from './utils/catalogLinks'
```

Then, right before the theme-toggle `<button>` (currently line 88), add:

```vue
        <WalletLoginButton class="ml-auto md:ml-0" />
        <button @click="toggleTheme" :aria-label="isDark ? t('theme.light') : t('theme.dark')"
          class="grid h-9 w-9 cursor-pointer place-items-center rounded-lg border border-line bg-surface-2 text-muted transition-colors duration-200 hover:border-accent/50 hover:text-accent-ink md:ml-0">
```

(note the theme button's own `class` loses its `ml-auto` since `WalletLoginButton` now claims that role — keep `md:ml-0` on both so they still sit together at the end of the header on desktop).

- [ ] **Step 3: Manually verify in the browser**

Run: `cd frontend && npm run dev`, open the app, confirm a "Connect Wallet" button appears in the header and clicking it opens the Nimiq Hub popup (a real wallet isn't required to confirm the popup opens — cancel it to finish the check).

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/WalletLoginButton.vue frontend/src/App.vue
git commit -m "Add wallet login button to the site header"
```

---

### Task 10: Review form + list, wired into the app detail page

**Files:**
- Create: `frontend/src/components/ReviewForm.vue`
- Create: `frontend/src/components/ReviewList.vue`
- Modify: `frontend/src/views/AppDetailView.vue:1-21` (imports), `:23-29` (state), `:39-60` (load reviews alongside the app), `:191-196` (mount the new section)

**Interfaces:**
- Consumes: `AppReview`, `listAppReviews`, `submitAppReview`, `deleteOwnAppReview` (Task 7); `useWalletAuth` (Task 8).

- [ ] **Step 1: Write `ReviewForm.vue`**

```vue
<!-- frontend/src/components/ReviewForm.vue -->
<script setup lang="ts">
import { ref, watch } from 'vue'
import { submitAppReview, type AppReview } from '../api'

const props = defineProps<{ slug: string; existing: AppReview | null }>()
const emit = defineEmits<{ saved: [AppReview] }>()

const rating = ref(props.existing?.rating ?? 0)
const body = ref(props.existing?.body ?? '')
const submitting = ref(false)
const error = ref('')

watch(
  () => props.existing,
  (value) => {
    rating.value = value?.rating ?? 0
    body.value = value?.body ?? ''
  },
)

async function submit() {
  if (rating.value < 1) {
    error.value = 'Pick a star rating'
    return
  }
  submitting.value = true
  error.value = ''
  try {
    const review = await submitAppReview(props.slug, rating.value, body.value.trim())
    emit('saved', review)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to submit review'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form class="flex flex-col gap-3 rounded-xl border border-line bg-surface p-4" @submit.prevent="submit">
    <div class="flex gap-1">
      <button
        v-for="n in 5" :key="n" type="button"
        class="text-2xl leading-none"
        :class="n <= rating ? 'text-accent-ink' : 'text-muted'"
        @click="rating = n"
      >★</button>
    </div>
    <textarea
      v-model="body"
      rows="3"
      maxlength="1000"
      placeholder="Share your experience with this app (optional)"
      class="rounded-lg border border-line bg-surface-2 p-2 text-sm"
    />
    <div class="flex items-center justify-between">
      <p v-if="error" class="text-xs text-red-500">{{ error }}</p>
      <button
        type="submit"
        class="ml-auto rounded-lg bg-accent px-3 py-1.5 text-sm font-semibold text-white disabled:opacity-50"
        :disabled="submitting"
      >{{ submitting ? 'Saving…' : existing ? 'Update review' : 'Post review' }}</button>
    </div>
  </form>
</template>
```

- [ ] **Step 2: Write `ReviewList.vue`**

```vue
<!-- frontend/src/components/ReviewList.vue -->
<script setup lang="ts">
import { deleteOwnAppReview, type AppReview } from '../api'

const props = defineProps<{ slug: string; reviews: AppReview[]; walletAddress: string | null }>()
const emit = defineEmits<{ deleted: [] }>()

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
        <span class="font-mono text-xs text-muted">{{ review.wallet_address.slice(0, 9) }}…</span>
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

- [ ] **Step 3: Wire into `AppDetailView.vue` — imports**

Add to the existing import block (after the `useAdminAuth` import on line 18):

```ts
import ReviewForm from '../components/ReviewForm.vue'
import ReviewList from '../components/ReviewList.vue'
import { listAppReviews, type AppReview, type AppReviewsResponse } from '../api'
import { useWalletAuth } from '../composables/useWalletAuth'
```

- [ ] **Step 4: Wire into `AppDetailView.vue` — state**

After the existing `const { isAdmin } = useAdminAuth()` line (line 25), add:

```ts
const { walletAddress } = useWalletAuth()
const reviewsData = ref<AppReviewsResponse>({ items: [], average: 0, count: 0 })
const myReview = computed(() => reviewsData.value.items.find((rv) => rv.wallet_address === walletAddress.value) ?? null)
```

- [ ] **Step 5: Wire into `AppDetailView.vue` — load reviews alongside the app**

Change `loadApp` (lines 39-60) to also fetch reviews:

```ts
async function loadApp(slug: string) {
  error.value = ''
  notFound.value = false
  app.value = null
  related.value = []
  loading.value = true
  try {
    const [loaded, relatedApps, reviews] = await Promise.all([
      getApp(slug),
      getRelatedApps(slug).catch(() => [] as App[]),
      listAppReviews(slug).catch(() => ({ items: [], average: 0, count: 0 }) as AppReviewsResponse),
    ])
    app.value = loaded
    related.value = relatedApps
    reviewsData.value = reviews
  } catch (e) {
    const message = (e as Error).message.toLowerCase()
    notFound.value = message.includes('not found')
    error.value = (e as Error).message
    resetPageMeta()
  } finally {
    loading.value = false
  }
}

async function refreshReviews() {
  if (!app.value) return
  reviewsData.value = await listAppReviews(app.value.slug).catch(() => reviewsData.value)
}
```

- [ ] **Step 6: Wire into `AppDetailView.vue` — template section**

In the `<template>`, right after the About `</section>` (currently line 196) and before the related-apps `<section>` (currently line 198), add:

```vue
    <section class="space-y-4">
      <div class="flex items-baseline justify-between">
        <h2 class="text-lg font-bold">Reviews</h2>
        <span v-if="reviewsData.count" class="text-sm text-muted">
          {{ reviewsData.average.toFixed(1) }} ★ · {{ reviewsData.count }} review{{ reviewsData.count === 1 ? '' : 's' }}
        </span>
      </div>
      <ReviewForm
        v-if="walletAddress"
        :slug="app.slug"
        :existing="myReview"
        @saved="refreshReviews"
      />
      <p v-else class="text-sm text-muted">Connect your wallet to leave a review.</p>
      <ReviewList
        v-if="reviewsData.items.length"
        :slug="app.slug"
        :reviews="reviewsData.items"
        :wallet-address="walletAddress"
        @deleted="refreshReviews"
      />
    </section>
```

- [ ] **Step 7: Typecheck and manually verify**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: no new type errors.

Run: `cd frontend && npm run dev`, open an app detail page, confirm the "Reviews" section renders with "Connect your wallet to leave a review." when logged out.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/components/ReviewForm.vue frontend/src/components/ReviewList.vue frontend/src/views/AppDetailView.vue
git commit -m "Add ratings and reviews UI to the app detail page"
```

---

### Task 11: Deployment config — new env var

**Files:**
- Modify: `docker-compose.yml` and `docker-stack.yml` (wherever `ADMIN_TOKEN` is currently set for the backend service)
- Modify: `docs/DEV.md` (env var reference, if one exists there)

**Interfaces:**
- None — operational wiring only.

- [ ] **Step 1: Add `WALLET_AUTH_SECRET` next to the existing `ADMIN_TOKEN`**

In both `docker-compose.yml` and `docker-stack.yml`, find the backend service's `environment:` block containing `ADMIN_TOKEN` and add a sibling entry, e.g.:

```yaml
      - WALLET_AUTH_SECRET=${WALLET_AUTH_SECRET}
```

(match whatever templating style — `${VAR}` vs a literal — the file already uses for `ADMIN_TOKEN`.)

- [ ] **Step 2: Document it**

If `docs/DEV.md` lists backend env vars (check for an `ADMIN_TOKEN` mention), add a line for `WALLET_AUTH_SECRET` describing it as a random secret used to sign wallet login session cookies (any long random string; rotating it logs everyone out).

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml docker-stack.yml docs/DEV.md
git commit -m "Add WALLET_AUTH_SECRET to deployment config"
```

---

## Explicitly out of scope (per spec)

- Comment threads / replies.
- Profanity/spam ML filtering.
- Instant session revocation / logout-everywhere.
- Cached/denormalized average rating column.
- Nimiq Hub iframe-first signing optimization (popup-only is simpler and sufficient here).
- Any changes to `backend/domaincheck.go` / `backend/icondiscovery.go` SSRF finding (tracked separately, unrelated to this feature).
