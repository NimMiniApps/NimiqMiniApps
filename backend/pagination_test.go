package main

import (
	"net/url"
	"testing"
)

func TestParsePagination(t *testing.T) {
	t.Run("defaults when paginate", func(t *testing.T) {
		limit, offset, paginate, err := parsePagination(url.Values{"paginate": {"1"}})
		if err != "" || !paginate || limit != defaultPageLimit || offset != 0 {
			t.Fatalf("got limit=%d offset=%d paginate=%v err=%q", limit, offset, paginate, err)
		}
	})

	t.Run("invalid offset", func(t *testing.T) {
		_, _, _, err := parsePagination(url.Values{"offset": {"-1"}})
		if err != "invalid offset" {
			t.Fatalf("expected invalid offset, got %q", err)
		}
	})

	t.Run("caps limit", func(t *testing.T) {
		limit, _, paginate, err := parsePagination(url.Values{"paginate": {"1"}, "limit": {"999"}})
		if err != "" || !paginate || limit != maxPageLimit {
			t.Fatalf("got limit=%d paginate=%v err=%q", limit, paginate, err)
		}
	})
}

func TestProbeDomainURL(t *testing.T) {
	if got := probeDomainURL("https://example.com/path"); got != "https://example.com/path" {
		t.Fatalf("unexpected url: %s", got)
	}
	if got := probeDomainURL("my.app.example"); got != "https://my.app.example" {
		t.Fatalf("unexpected url: %s", got)
	}
}

func TestIsSocialCrawler(t *testing.T) {
	if !isSocialCrawler("Twitterbot/1.0") {
		t.Fatal("expected twitterbot match")
	}
	if isSocialCrawler("Mozilla/5.0 Chrome") {
		t.Fatal("expected no match for normal browser")
	}
}

func TestHtmlEscape(t *testing.T) {
	if got := htmlEscape(`Tom & "Jerry"`); got != "Tom &amp; &quot;Jerry&quot;" {
		t.Fatalf("got %q", got)
	}
}
