package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"
)

// newTestIdentifier returns a random UUID-shaped string suitable for
// analyticsEventInput.VisitorID / SessionID.
func newTestIdentifier(t *testing.T) string {
	t.Helper()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func TestRecordAnalyticsEventHashesIdentifiers(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Setenv("ANALYTICS_HASH_SECRET", "test-secret-hashes")

	s := &server{pool: pool}
	visitorID := newTestIdentifier(t)
	sessionID := newTestIdentifier(t)
	t.Cleanup(func() {
		secret := "test-secret-hashes"
		_, _ = pool.Exec(context.Background(), `DELETE FROM analytics_events WHERE visitor_hash = $1`,
			hashAnalyticsIdentifier(secret, "visitor", visitorID))
	})

	recorded, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
		EventType: analyticsCatalogVisit,
		VisitorID: visitorID,
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("recordAnalyticsEvent: %v", err)
	}
	if !recorded {
		t.Fatal("want recorded=true for first catalog visit")
	}

	var storedVisitorHash, storedSessionHash string
	err = pool.QueryRow(ctx, `
		SELECT visitor_hash, session_hash FROM analytics_events
		WHERE event_type='catalog_visit' AND session_hash = $1`,
		hashAnalyticsIdentifier("test-secret-hashes", "session", sessionID)).
		Scan(&storedVisitorHash, &storedSessionHash)
	if err != nil {
		t.Fatalf("query stored event: %v", err)
	}

	wantVisitorHash := hashAnalyticsIdentifier("test-secret-hashes", "visitor", visitorID)
	wantSessionHash := hashAnalyticsIdentifier("test-secret-hashes", "session", sessionID)
	if storedVisitorHash != wantVisitorHash {
		t.Fatalf("visitor_hash = %q, want %q", storedVisitorHash, wantVisitorHash)
	}
	if storedSessionHash != wantSessionHash {
		t.Fatalf("session_hash = %q, want %q", storedSessionHash, wantSessionHash)
	}
	if storedVisitorHash == visitorID || storedSessionHash == sessionID {
		t.Fatal("raw identifiers must never be stored")
	}
}

func TestRecordAnalyticsEventRejectsRawWalletInput(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Setenv("ANALYTICS_HASH_SECRET", "test-secret-wallet")

	s := &server{pool: pool}
	visitorID := newTestIdentifier(t)
	sessionID := newTestIdentifier(t)

	_, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
		EventType:     analyticsCatalogVisit,
		VisitorID:     visitorID,
		SessionID:     sessionID,
		WalletAddress: "NQ07 0000 0000 0000 0000 0000 0000 0000 0000",
	})
	if err == nil {
		t.Fatal("want error when a client-facing event carries a wallet address")
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE session_hash = $1`,
		hashAnalyticsIdentifier("test-secret-wallet", "session", sessionID)).Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no rows inserted for rejected event, got %d", count)
	}

	var rawCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE wallet_hash = $1`,
		"NQ07 0000 0000 0000 0000 0000 0000 0000 0000").Scan(&rawCount); err != nil {
		t.Fatalf("raw wallet query: %v", err)
	}
	if rawCount != 0 {
		t.Fatal("raw wallet address must never appear in analytics_events")
	}
}

func TestRecordAnalyticsEventDeduplicatesCatalogVisitPerSession(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Setenv("ANALYTICS_HASH_SECRET", "test-secret-catalog-dedup")

	s := &server{pool: pool}
	sessionID := newTestIdentifier(t)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM analytics_events WHERE session_hash = $1`,
			hashAnalyticsIdentifier("test-secret-catalog-dedup", "session", sessionID))
	})

	first, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
		EventType: analyticsCatalogVisit,
		VisitorID: newTestIdentifier(t),
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("first visit: %v", err)
	}
	if !first {
		t.Fatal("want recorded=true for first catalog visit")
	}

	second, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
		EventType: analyticsCatalogVisit,
		VisitorID: newTestIdentifier(t),
		SessionID: sessionID,
	})
	if err != nil {
		t.Fatalf("second visit: %v", err)
	}
	if second {
		t.Fatal("want recorded=false for duplicate catalog visit in same session")
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE event_type='catalog_visit' AND session_hash=$1`,
		hashAnalyticsIdentifier("test-secret-catalog-dedup", "session", sessionID)).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestRecordAnalyticsEventDeduplicatesAppViewPerAppSession(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Setenv("ANALYTICS_HASH_SECRET", "test-secret-appview-dedup")

	slug := "analytics-appview-" + time.Now().Format("150405.000000")
	var appID string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'Analytics App View', 'appview.example.com', 'Utilities', 'analytics-dev', 'Analytics Dev', 'tag', 'desc', 'approved')
		RETURNING id`, slug).Scan(&appID)
	if err != nil {
		t.Fatalf("insert app: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appID)
	})

	s := &server{pool: pool}
	sessionID := newTestIdentifier(t)

	first, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
		EventType: analyticsAppView,
		VisitorID: newTestIdentifier(t),
		SessionID: sessionID,
		AppID:     appID,
	})
	if err != nil {
		t.Fatalf("first view: %v", err)
	}
	if !first {
		t.Fatal("want recorded=true for first app view")
	}

	second, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
		EventType: analyticsAppView,
		VisitorID: newTestIdentifier(t),
		SessionID: sessionID,
		AppID:     appID,
	})
	if err != nil {
		t.Fatalf("second view: %v", err)
	}
	if second {
		t.Fatal("want recorded=false for duplicate app view in same session")
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE event_type='app_view' AND app_id=$1 AND session_hash=$2`,
		appID, hashAnalyticsIdentifier("test-secret-appview-dedup", "session", sessionID)).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestRecordAnalyticsEventCountsEveryOpenAndCopy(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Setenv("ANALYTICS_HASH_SECRET", "test-secret-open-copy")

	slug := "analytics-opencopy-" + time.Now().Format("150405.000000")
	var appID string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'Analytics Open Copy', 'opencopy.example.com', 'Utilities', 'analytics-dev', 'Analytics Dev', 'tag', 'desc', 'approved')
		RETURNING id`, slug).Scan(&appID)
	if err != nil {
		t.Fatalf("insert app: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appID)
	})

	s := &server{pool: pool}
	sessionID := newTestIdentifier(t)

	for i := 0; i < 3; i++ {
		recorded, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
			EventType: analyticsAppOpen,
			VisitorID: newTestIdentifier(t),
			SessionID: sessionID,
			AppID:     appID,
		})
		if err != nil {
			t.Fatalf("open %d: %v", i, err)
		}
		if !recorded {
			t.Fatalf("open %d: want recorded=true every time", i)
		}
	}
	for i := 0; i < 2; i++ {
		recorded, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
			EventType: analyticsAppLinkCopy,
			VisitorID: newTestIdentifier(t),
			SessionID: sessionID,
			AppID:     appID,
		})
		if err != nil {
			t.Fatalf("copy %d: %v", i, err)
		}
		if !recorded {
			t.Fatalf("copy %d: want recorded=true every time", i)
		}
	}

	var opens, copies int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE event_type='app_open' AND app_id=$1`, appID).Scan(&opens); err != nil {
		t.Fatalf("count opens: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE event_type='app_link_copy' AND app_id=$1`, appID).Scan(&copies); err != nil {
		t.Fatalf("count copies: %v", err)
	}
	if opens != 3 {
		t.Fatalf("opens = %d, want 3", opens)
	}
	if copies != 2 {
		t.Fatalf("copies = %d, want 2", copies)
	}
}

func TestRecordAnalyticsEventUpdatesLegacyAppStats(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Setenv("ANALYTICS_HASH_SECRET", "test-secret-legacy-stats")

	slug := "analytics-legacy-" + time.Now().Format("150405.000000")
	var appID string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'Analytics Legacy', 'legacy.example.com', 'Utilities', 'analytics-dev', 'Analytics Dev', 'tag', 'desc', 'approved')
		RETURNING id`, slug).Scan(&appID)
	if err != nil {
		t.Fatalf("insert app: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appID)
	})

	s := &server{pool: pool}
	viewSession := newTestIdentifier(t)

	// Same session views the app twice; the second is deduplicated and must not
	// double-count legacy views.
	for i := 0; i < 2; i++ {
		if _, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
			EventType: analyticsAppView,
			VisitorID: newTestIdentifier(t),
			SessionID: viewSession,
			AppID:     appID,
		}); err != nil {
			t.Fatalf("view %d: %v", i, err)
		}
	}

	// Opens are never deduplicated, so both must increment legacy opens.
	for i := 0; i < 2; i++ {
		if _, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
			EventType: analyticsAppOpen,
			VisitorID: newTestIdentifier(t),
			SessionID: newTestIdentifier(t),
			AppID:     appID,
		}); err != nil {
			t.Fatalf("open %d: %v", i, err)
		}
	}

	var opens, views int
	err = pool.QueryRow(ctx, `SELECT opens, views FROM app_stats_daily WHERE app_id=$1 AND day=CURRENT_DATE`, appID).
		Scan(&opens, &views)
	if err != nil {
		t.Fatalf("query app_stats_daily: %v", err)
	}
	if views != 1 {
		t.Fatalf("views = %d, want 1 (deduplicated)", views)
	}
	if opens != 2 {
		t.Fatalf("opens = %d, want 2 (not deduplicated)", opens)
	}
}
