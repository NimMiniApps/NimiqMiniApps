package main

import (
	"strings"
	"testing"
)

func TestValidateDisplayName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty clears name", "", false},
		{"normal name", "Satoshi", false},
		{"max length", strings.Repeat("a", 50), false},
		{"too long", strings.Repeat("a", 51), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDisplayName(tc.input)
			if (err != "") != tc.wantErr {
				t.Fatalf("validateDisplayName(%q) error=%q, wantErr=%v", tc.input, err, tc.wantErr)
			}
		})
	}
}
