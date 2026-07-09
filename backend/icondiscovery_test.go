package main

import (
	"net/url"
	"sort"
	"testing"
)

func TestParseIconCandidatesPrefersAppleTouchIcon(t *testing.T) {
	base, _ := url.Parse("https://play.example.com/")
	html := `<html><head>
<link rel="icon" href="/favicon.ico">
<link rel="apple-touch-icon" href="/apple-touch-icon.png">
<link rel="icon" type="image/png" sizes="32x32" href="/icon-32.png">
</head></html>`

	candidates := parseIconCandidates(html, base)
	if len(candidates) != 3 {
		t.Fatalf("got %d candidates, want 3", len(candidates))
	}
	sortCandidates := append([]iconCandidate(nil), candidates...)
	sort.Slice(sortCandidates, func(i, j int) bool {
		return sortCandidates[i].priority > sortCandidates[j].priority
	})
	if sortCandidates[0].href != "https://play.example.com/apple-touch-icon.png" {
		t.Fatalf("best candidate = %q", sortCandidates[0].href)
	}
}

func TestParseIconCandidatesIgnoresNonIconLinks(t *testing.T) {
	base, _ := url.Parse("https://example.com/")
	html := `<link rel="stylesheet" href="/style.css">
<link rel="canonical" href="https://example.com/">`

	if got := parseIconCandidates(html, base); len(got) != 0 {
		t.Fatalf("expected no candidates, got %v", got)
	}
}

func TestParseIconSizes(t *testing.T) {
	if got := parseIconSizes("16x16 32x32 192x192"); got != 192 {
		t.Fatalf("parseIconSizes = %d, want 192", got)
	}
}

func TestResolveIconHref(t *testing.T) {
	base, _ := url.Parse("https://app.nimiq.com/path/")
	if got := resolveIconHref(base, "//cdn.nimiq.com/icon.png"); got != "https://cdn.nimiq.com/icon.png" {
		t.Fatalf("resolveIconHref = %q", got)
	}
	if got := resolveIconHref(base, "data:image/png;base64,abc"); got != "" {
		t.Fatalf("expected empty for data URI, got %q", got)
	}
}

func TestNeedsIconDiscovery(t *testing.T) {
	empty := ""
	icon := "https://cdn.example.com/icon.png"
	discovered := "https://cdn.example.com/favicon.ico"

	if needsIconDiscovery(&icon, nil) {
		t.Fatal("should not discover when icon_url is set")
	}
	if !needsIconDiscovery(nil, nil) {
		t.Fatal("should discover when both empty")
	}
	if needsIconDiscovery(nil, &discovered) {
		t.Fatal("should not discover when discovered_icon_url is set")
	}
	if !needsIconDiscovery(&empty, nil) {
		t.Fatal("should discover when icon_url is blank")
	}
}
