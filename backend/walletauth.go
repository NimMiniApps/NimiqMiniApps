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
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	nimiq "github.com/NimMiniApps/nimiq-go"
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

func verifyWalletSignature(address, message, signatureB64, publicKeyB64 string) (ed25519.PublicKey, error) {
	addr, err := nimiq.ParseAddress(address)
	if err != nil {
		return nil, errors.New("invalid wallet address")
	}
	sig, err := nimiq.ParseSignature(signatureB64)
	if err != nil {
		return nil, errors.New("invalid signature encoding")
	}
	pub, err := nimiq.ParsePublicKey(publicKeyB64)
	if err != nil {
		return nil, errors.New("invalid public key encoding")
	}
	if err := nimiq.VerifyMessageFrom(addr, pub, []byte(message), sig); err != nil {
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
	if strings.TrimSpace(req.Address) == "" {
		writeError(w, http.StatusBadRequest, "wallet_address is required")
		return
	}
	address, err := parseCanonicalWalletAddress(req.Address)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid wallet address")
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
		Address            string `json:"wallet_address"`
		Nonce              string `json:"nonce"`
		Signature          string `json:"signature"`
		PublicKey          string `json:"public_key"`
		AnalyticsVisitorID string `json:"analytics_visitor_id"`
		AnalyticsSessionID string `json:"analytics_session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	address, err := parseCanonicalWalletAddress(req.Address)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid wallet address")
		return
	}
	entry, ok := s.nonces.get(req.Nonce)
	if !ok || entry.address != address || time.Now().After(entry.expires) {
		writeError(w, http.StatusUnauthorized, "invalid or expired challenge")
		return
	}
	message := buildWalletAuthMessage(address, req.Nonce, entry.expires)
	if _, err := verifyWalletSignature(address, message, req.Signature, req.PublicKey); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err := s.nonces.markUsed(req.Nonce, address); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	setWalletCookie(w, r, s.walletAuthSecret, address)

	// Analytics is best-effort: a successful login must never be undone by a
	// hashing/storage hiccup downstream. Skip when the client omitted IDs.
	if req.AnalyticsVisitorID != "" && req.AnalyticsSessionID != "" {
		if _, err := s.recordAnalyticsEvent(r.Context(), analyticsEventInput{
			EventType:     analyticsWalletLogin,
			VisitorID:     req.AnalyticsVisitorID,
			SessionID:     req.AnalyticsSessionID,
			WalletAddress: address,
		}); err != nil {
			slog.Warn("failed to record wallet_login analytics event", "error", err.Error())
		}
	}

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
