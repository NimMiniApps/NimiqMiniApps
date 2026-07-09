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
