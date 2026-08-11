package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// signedAuthVerifyBody builds a fresh challenge/response pair and returns the
// JSON body for a POST /api/auth/verify request signed by priv/pub for address.
func signedAuthVerifyBody(t *testing.T, s *server, pub ed25519.PublicKey, priv ed25519.PrivateKey, address string, extra map[string]string) []byte {
	t.Helper()
	challengeBody, _ := json.Marshal(map[string]string{"wallet_address": address})
	challengeReq := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/challenge", bytes.NewReader(challengeBody))
	challengeRec := httptest.NewRecorder()
	s.authChallenge(challengeRec, challengeReq)
	if challengeRec.Code != http.StatusOK {
		t.Fatalf("challenge status %d: %s", challengeRec.Code, challengeRec.Body.String())
	}
	var challengeResp struct {
		Nonce   string `json:"nonce"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(challengeRec.Body).Decode(&challengeResp); err != nil {
		t.Fatalf("decode challenge: %v", err)
	}

	prefix := "\x16Nimiq Signed Message:\n"
	payload := prefix + strconv.Itoa(len(challengeResp.Message)) + challengeResp.Message
	hash := sha256.Sum256([]byte(payload))
	sig := ed25519.Sign(priv, hash[:])

	fields := map[string]string{
		"wallet_address": address,
		"nonce":          challengeResp.Nonce,
		"signature":      base64.StdEncoding.EncodeToString(sig),
		"public_key":     base64.StdEncoding.EncodeToString(pub),
	}
	for k, v := range extra {
		fields[k] = v
	}
	body, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal verify body: %v", err)
	}
	return body
}

func TestAuthVerifyRecordsWalletLoginOnlyAfterSuccess(t *testing.T) {
	pool := testPool(t)
	secret := "test-secret-authverify-success"
	t.Setenv("ANALYTICS_HASH_SECRET", secret)
	walletSecret := "wallet-auth-secret"
	s := &server{pool: pool, walletAuthSecret: walletSecret, nonces: newNonceStore()}

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	address := userFriendlyAddressFromPublicKey(pub)

	t.Run("successful verification records wallet_login", func(t *testing.T) {
		visitorID := newTestIdentifier(t)
		sessionID := newTestIdentifier(t)
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM analytics_events WHERE visitor_hash = $1`,
				hashAnalyticsIdentifier(secret, "visitor", visitorID))
		})

		body := signedAuthVerifyBody(t, s, pub, priv, address, map[string]string{
			"analytics_visitor_id": visitorID,
			"analytics_session_id": sessionID,
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/verify", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		s.authVerify(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("verify status %d: %s", rec.Code, rec.Body.String())
		}

		var count int
		if err := pool.QueryRow(context.Background(), `
			SELECT count(*) FROM analytics_events
			WHERE event_type='wallet_login' AND wallet_hash=$1 AND session_hash=$2`,
			hashAnalyticsIdentifier(secret, "wallet", address),
			hashAnalyticsIdentifier(secret, "session", sessionID)).Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 1 {
			t.Fatalf("count = %d, want 1 wallet_login analytics event", count)
		}
	})

	t.Run("failed verification records no login", func(t *testing.T) {
		badVisitorID := newTestIdentifier(t)
		badSessionID := newTestIdentifier(t)

		// A fresh, valid challenge, but signed with the wrong key so verification fails.
		wrongPub, wrongPriv, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatal(err)
		}
		_ = wrongPub
		body := signedAuthVerifyBody(t, s, pub, wrongPriv, address, map[string]string{
			"analytics_visitor_id": badVisitorID,
			"analytics_session_id": badSessionID,
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/verify", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		s.authVerify(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("verify status %d, want 401: %s", rec.Code, rec.Body.String())
		}

		var count int
		if err := pool.QueryRow(context.Background(), `
			SELECT count(*) FROM analytics_events WHERE session_hash=$1`,
			hashAnalyticsIdentifier(secret, "session", badSessionID)).Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 0 {
			t.Fatal("failed verification must not record a wallet_login event")
		}
	})

	t.Run("claimed address mismatch records no login", func(t *testing.T) {
		otherPub, otherPriv, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatal(err)
		}
		otherAddress := userFriendlyAddressFromPublicKey(otherPub)
		badSessionID := newTestIdentifier(t)

		// Claim `address` but sign with a different key/address's challenge message content mismatch.
		challengeBody, _ := json.Marshal(map[string]string{"wallet_address": address})
		challengeReq := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/challenge", bytes.NewReader(challengeBody))
		challengeRec := httptest.NewRecorder()
		s.authChallenge(challengeRec, challengeReq)
		var challengeResp struct {
			Nonce   string `json:"nonce"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(challengeRec.Body).Decode(&challengeResp); err != nil {
			t.Fatalf("decode challenge: %v", err)
		}
		prefix := "\x16Nimiq Signed Message:\n"
		payload := prefix + strconv.Itoa(len(challengeResp.Message)) + challengeResp.Message
		hash := sha256.Sum256([]byte(payload))
		sig := ed25519.Sign(otherPriv, hash[:])

		verifyBody, _ := json.Marshal(map[string]string{
			"wallet_address":       address,
			"nonce":                challengeResp.Nonce,
			"signature":            base64.StdEncoding.EncodeToString(sig),
			"public_key":           base64.StdEncoding.EncodeToString(otherPub),
			"analytics_visitor_id": newTestIdentifier(t),
			"analytics_session_id": badSessionID,
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/verify", bytes.NewReader(verifyBody))
		rec := httptest.NewRecorder()
		s.authVerify(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("verify status %d, want 401: %s", rec.Code, rec.Body.String())
		}
		_ = otherAddress

		var count int
		if err := pool.QueryRow(context.Background(), `
			SELECT count(*) FROM analytics_events WHERE session_hash=$1`,
			hashAnalyticsIdentifier(secret, "session", badSessionID)).Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 0 {
			t.Fatal("mismatched claimed address must not record a wallet_login event")
		}
	})
}

func TestAuthVerifyStillSucceedsWhenAnalyticsUnavailable(t *testing.T) {
	pool := testPool(t)
	t.Setenv("ANALYTICS_HASH_SECRET", "")
	walletSecret := "wallet-auth-secret-no-analytics"
	s := &server{pool: pool, walletAuthSecret: walletSecret, nonces: newNonceStore()}

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	address := userFriendlyAddressFromPublicKey(pub)

	body := signedAuthVerifyBody(t, s, pub, priv, address, map[string]string{
		"analytics_visitor_id": newTestIdentifier(t),
		"analytics_session_id": newTestIdentifier(t),
	})
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/verify", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.authVerify(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify status %d, want 200 even when analytics is unavailable: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		WalletAddress string `json:"wallet_address"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode verify response: %v", err)
	}
	if resp.WalletAddress != address {
		t.Fatalf("wallet_address = %q, want %q", resp.WalletAddress, address)
	}
}

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
	// Change an encoded payload byte, not the final Base64URL character. The
	// final character of a 32-byte HMAC has unused bits and can decode unchanged.
	tampered := "x" + value[1:]
	if _, err := verifyWalletCookie(secret, tampered); err == nil {
		t.Fatal("expected tampered cookie to fail verification")
	}
}

func TestAuthChallengeRejectsInvalidAddress(t *testing.T) {
	s := &server{walletAuthSecret: "secret", nonces: newNonceStore()}
	body, _ := json.Marshal(map[string]string{"wallet_address": "NQ05 NOT-A-VALID-ADDRESS!!!!"})
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/challenge", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	s.authChallenge(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400; body %s", rec.Code, rec.Body.String())
	}
}

func TestAuthVerifyAcceptsAddressSpacingVariant(t *testing.T) {
	s := &server{walletAuthSecret: "secret", nonces: newNonceStore()}
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	canonical := userFriendlyAddressFromPublicKey(pub)
	compact := normalizeUserFriendlyAddress(canonical)

	challengeBody, _ := json.Marshal(map[string]string{"wallet_address": compact})
	challengeReq := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/challenge", bytes.NewReader(challengeBody))
	challengeRec := httptest.NewRecorder()
	s.authChallenge(challengeRec, challengeReq)
	if challengeRec.Code != http.StatusOK {
		t.Fatalf("challenge status %d: %s", challengeRec.Code, challengeRec.Body.String())
	}
	var challengeResp struct {
		Nonce   string `json:"nonce"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(challengeRec.Body).Decode(&challengeResp); err != nil {
		t.Fatalf("decode challenge: %v", err)
	}
	if !strings.Contains(challengeResp.Message, "address="+canonical) {
		t.Fatalf("challenge message should use canonical address; got %q", challengeResp.Message)
	}

	prefix := "\x16Nimiq Signed Message:\n"
	payload := prefix + strconv.Itoa(len(challengeResp.Message)) + challengeResp.Message
	hash := sha256.Sum256([]byte(payload))
	sig := ed25519.Sign(priv, hash[:])
	verifyBody, _ := json.Marshal(map[string]string{
		"wallet_address": canonical, // different spacing than challenge request
		"nonce":          challengeResp.Nonce,
		"signature":      base64.StdEncoding.EncodeToString(sig),
		"public_key":     base64.StdEncoding.EncodeToString(pub),
	})
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/verify", bytes.NewReader(verifyBody))
	rec := httptest.NewRecorder()
	s.authVerify(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify status %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		WalletAddress string `json:"wallet_address"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.WalletAddress != canonical {
		t.Fatalf("wallet_address = %q, want canonical %q", resp.WalletAddress, canonical)
	}
}

func TestVerifyWalletSignatureRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	address := userFriendlyAddressFromPublicKey(pub)
	message := "test message"
	prefix := "\x16Nimiq Signed Message:\n"
	payload := prefix + strconv.Itoa(len(message)) + message
	hash := sha256.Sum256([]byte(payload))
	sig := ed25519.Sign(priv, hash[:])
	sigB64 := base64.StdEncoding.EncodeToString(sig)
	pubB64 := base64.StdEncoding.EncodeToString(pub)

	gotPub, err := verifyWalletSignature(address, message, sigB64, pubB64)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotPub.Equal(pub) {
		t.Fatal("public key mismatch")
	}
}

func TestVerifyWalletSignatureAcceptsHex(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	address := userFriendlyAddressFromPublicKey(pub)
	message := "test message"
	prefix := "\x16Nimiq Signed Message:\n"
	payload := prefix + strconv.Itoa(len(message)) + message
	hash := sha256.Sum256([]byte(payload))
	sig := ed25519.Sign(priv, hash[:])

	gotPub, err := verifyWalletSignature(address, message, hex.EncodeToString(sig), hex.EncodeToString(pub))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotPub.Equal(pub) {
		t.Fatal("public key mismatch")
	}
}

func TestDecodeCryptoBytesHex(t *testing.T) {
	pub := make([]byte, ed25519.PublicKeySize)
	for i := range pub {
		pub[i] = byte(i)
	}
	got, err := decodeCryptoBytes(hex.EncodeToString(pub))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != ed25519.PublicKeySize {
		t.Fatalf("got len %d, want %d", len(got), ed25519.PublicKeySize)
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
