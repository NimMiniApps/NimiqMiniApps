package main

import (
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
