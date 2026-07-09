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
