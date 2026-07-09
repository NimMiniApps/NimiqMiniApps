package main

import "testing"

func TestSlugifyDisplayName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Satoshi", "satoshi"},
		{"two words", "Satoshi Nakamoto", "satoshi-nakamoto"},
		{"extra whitespace and punctuation", "  Multi   Space! ", "multi-space"},
		{"mixed case with numbers", "Team42 Studio", "team42-studio"},
		{"all symbols", "💯💯", ""},
		{"leading/trailing separators collapse", "-Foo-Bar-", "foo-bar"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := slugifyDisplayName(tc.input)
			if got != tc.want {
				t.Fatalf("slugifyDisplayName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
