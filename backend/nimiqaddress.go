package main

import (
	"crypto/ed25519"
	"strings"

	nimiq "github.com/NimMiniApps/nimiq-go"
)

// parseCanonicalWalletAddress parses a Nimiq user-friendly address and returns
// the canonical spaced form (Address.String). Spacing and case are ignored.
func parseCanonicalWalletAddress(s string) (string, error) {
	addr, err := nimiq.ParseAddress(s)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

// userFriendlyAddressFromPublicKey returns the canonical user-friendly Nimiq address
// for the given Ed25519 public key.
func userFriendlyAddressFromPublicKey(pub ed25519.PublicKey) string {
	addr, err := nimiq.AddressFromPublicKey(pub)
	if err != nil {
		return ""
	}
	return addr.String()
}

// normalizeUserFriendlyAddress strips spaces and uppercases for comparisons
// (admin allowlists and similar non-crypto equality checks).
func normalizeUserFriendlyAddress(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	return strings.ToUpper(s)
}

// publicKeyMatchesClaimedAddress reports whether the Ed25519 public key corresponds to
// the claimed Nimiq user-friendly address (spacing and case ignored).
func publicKeyMatchesClaimedAddress(pub ed25519.PublicKey, claimedAddress string) bool {
	addr, err := nimiq.AddressFromPublicKey(pub)
	if err != nil {
		return false
	}
	claimed, err := nimiq.ParseAddress(claimedAddress)
	if err != nil {
		return false
	}
	return addr == claimed
}
