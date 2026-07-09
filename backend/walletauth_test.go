package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

func TestVerifyWalletSignatureAcceptsHex(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	message := "test message"
	prefix := "\x16Nimiq Signed Message:\n"
	payload := prefix + strconv.Itoa(len(message)) + message
	hash := sha256.Sum256([]byte(payload))
	sig := ed25519.Sign(priv, hash[:])

	gotPub, err := verifyWalletSignature(message, hex.EncodeToString(sig), hex.EncodeToString(pub))
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
