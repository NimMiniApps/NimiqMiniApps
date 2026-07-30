package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDomainCheckOfflineIntervalShorterThanHealthy(t *testing.T) {
	t.Setenv("DOMAIN_CHECK_INTERVAL", "")
	t.Setenv("DOMAIN_CHECK_OFFLINE_INTERVAL", "")
	if domainCheckOfflineInterval() >= domainCheckHealthyInterval() {
		t.Fatalf("offline interval %v should be shorter than healthy %v",
			domainCheckOfflineInterval(), domainCheckHealthyInterval())
	}
}

func TestDomainCheckTickWithinBounds(t *testing.T) {
	t.Setenv("DOMAIN_CHECK_TICK", "")
	tick := domainCheckTick()
	if tick < time.Minute || tick > 5*time.Minute {
		t.Fatalf("unexpected default tick: %v", tick)
	}
}

func TestDomainCheckCustomIntervals(t *testing.T) {
	t.Setenv("DOMAIN_CHECK_INTERVAL", "2h")
	t.Setenv("DOMAIN_CHECK_OFFLINE_INTERVAL", "10m")
	t.Setenv("DOMAIN_CHECK_TICK", "2m")

	if got := domainCheckHealthyInterval(); got != 2*time.Hour {
		t.Fatalf("healthy interval = %v, want 2h", got)
	}
	if got := domainCheckOfflineInterval(); got != 10*time.Minute {
		t.Fatalf("offline interval = %v, want 10m", got)
	}
	if got := domainCheckTick(); got != 2*time.Minute {
		t.Fatalf("tick = %v, want 2m", got)
	}
}

func TestCheckDomainReachableFallsBackToGetWhenHeadRejected(t *testing.T) {
	var headRequests, getRequests int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			headRequests++
			w.WriteHeader(http.StatusMethodNotAllowed)
		case http.MethodGet:
			getRequests++
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected request method %q", r.Method)
		}
	}))
	defer server.Close()

	previousClient := domainCheckClient
	domainCheckClient = server.Client()
	defer func() {
		domainCheckClient = previousClient
	}()

	reachable, errMessage := checkDomainReachable(strings.TrimPrefix(server.URL, "https://"))
	if !reachable {
		t.Fatalf("expected domain to be reachable, got error %q", errMessage)
	}
	if headRequests != 1 {
		t.Fatalf("HEAD requests = %d, want 1", headRequests)
	}
	if getRequests != 1 {
		t.Fatalf("GET requests = %d, want 1", getRequests)
	}
}
