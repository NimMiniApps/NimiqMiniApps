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

	"github.com/jackc/pgx/v5"
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

func cookieSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func setWalletCookie(w http.ResponseWriter, r *http.Request, secret, address string) {
	expires := time.Now().Add(walletSessionTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     walletCookieName,
		Value:    signWalletCookie(secret, address, expires),
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   cookieSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearWalletCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     walletCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   cookieSecure(r),
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
		setWalletCookie(w, r, secret, address)
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

func decodeCryptoBytes(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty value")
	}
	if isHexCryptoString(s) {
		h := strings.TrimPrefix(strings.ToLower(s), "0x")
		return hex.DecodeString(h)
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return nil, errors.New("invalid encoding")
}

func isHexCryptoString(s string) bool {
	h := strings.TrimPrefix(strings.TrimSpace(s), "0x")
	if len(h) == 0 || len(h)%2 != 0 {
		return false
	}
	for _, c := range h {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

func verifyWalletSignature(message, signatureB64, publicKeyB64 string) (ed25519.PublicKey, error) {
	sig, err := decodeCryptoBytes(signatureB64)
	if err != nil {
		return nil, errors.New("invalid signature encoding")
	}
	pubBytes, err := decodeCryptoBytes(publicKeyB64)
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
	setWalletCookie(w, r, s.walletAuthSecret, address)
	writeJSON(w, http.StatusOK, map[string]string{"wallet_address": address})
}

func (s *server) authMe(w http.ResponseWriter, r *http.Request, address string) {
	var displayName *string
	err := s.pool.QueryRow(r.Context(),
		`SELECT display_name FROM users WHERE wallet_address=$1`, address,
	).Scan(&displayName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"wallet_address": address,
		"display_name":   displayName,
		"is_admin":       isAdminWallet(s.adminWallets, address),
	})
}

func (s *server) authLogout(w http.ResponseWriter, r *http.Request) {
	clearWalletCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}
