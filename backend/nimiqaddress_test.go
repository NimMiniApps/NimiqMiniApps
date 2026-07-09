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
