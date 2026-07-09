package main

import "testing"

func TestParseAdminWallets(t *testing.T) {
	raw := "NQ07 0000 0000 0000 0000 0000 0000 0000 0000, nq11 abcd efgh ijkl mnop qrst uvwx yz01 2345"
	got := parseAdminWallets(raw)
	if len(got) != 2 {
		t.Fatalf("got %d wallets, want 2", len(got))
	}
	if !isAdminWallet(got, "NQ07 0000 0000 0000 0000 0000 0000 0000 0000") {
		t.Fatal("expected first wallet in allowlist")
	}
	if !isAdminWallet(got, "NQ11 ABCD EFGH IJKL MNOP QRST UVWX YZ01 2345") {
		t.Fatal("expected second wallet in allowlist (case/spacing ignored)")
	}
	if isAdminWallet(got, "NQ99 0000 0000 0000 0000 0000 0000 0000 0000") {
		t.Fatal("unexpected wallet should not match")
	}
}

func TestIsAdminWalletEmptyAllowlist(t *testing.T) {
	if isAdminWallet(nil, "NQ07 0000 0000 0000 0000 0000 0000 0000 0000") {
		t.Fatal("empty allowlist should deny everyone")
	}
}
