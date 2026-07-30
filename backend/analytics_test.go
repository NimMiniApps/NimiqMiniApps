package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestRecordAnalyticsEventRequiresHashSecret(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	t.Setenv("ANALYTICS_HASH_SECRET", "")

	s := &server{pool: pool}
	sessionID := newTestIdentifier(t)

	_, err := s.recordAnalyticsEvent(ctx, analyticsEventInput{
		EventType: analyticsCatalogVisit,
		VisitorID: newTestIdentifier(t),
		SessionID: sessionID,
	})
	if err == nil {
		t.Fatal("want error when ANALYTICS_HASH_SECRET is unset")
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE session_hash = $1`,
		sessionID).Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Fatal("expected no rows inserted when the hash secret is unset")
	}
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

func TestTrackAnalyticsEventAcceptsStrictPublicEvents(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	secret := "test-secret-track-strict"
	t.Setenv("ANALYTICS_HASH_SECRET", secret)
	s := &server{pool: pool, analyticsHashSecret: secret, analyticsLimiter: newRateLimiter(120, time.Minute)}

	t.Run("catalog_visit", func(t *testing.T) {
		visitorID := newTestIdentifier(t)
		sessionID := newTestIdentifier(t)
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM analytics_events WHERE session_hash = $1`,
				hashAnalyticsIdentifier(secret, "session", sessionID))
		})
		body, _ := json.Marshal(map[string]string{
			"event": "catalog_visit", "visitor_id": visitorID, "session_id": sessionID,
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.1:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status %d, want 204: %s", rec.Code, rec.Body.String())
		}
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE event_type='catalog_visit' AND session_hash=$1`,
			hashAnalyticsIdentifier(secret, "session", sessionID)).Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 1 {
			t.Fatalf("count = %d, want 1", count)
		}
	})

	t.Run("app_view", func(t *testing.T) {
		slug := "analytics-track-appview-" + time.Now().Format("150405.000000")
		var appID string
		err := pool.QueryRow(ctx, `
			INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
			VALUES ($1, 'Track App View', 'trackappview.example.com', 'Utilities', 'analytics-dev', 'Analytics Dev', 'tag', 'desc', 'approved')
			RETURNING id`, slug).Scan(&appID)
		if err != nil {
			t.Fatalf("insert app: %v", err)
		}
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appID)
		})

		visitorID := newTestIdentifier(t)
		sessionID := newTestIdentifier(t)
		body, _ := json.Marshal(map[string]string{
			"event": "app_view", "visitor_id": visitorID, "session_id": sessionID, "app_slug": slug,
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.2:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status %d, want 204: %s", rec.Code, rec.Body.String())
		}
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics_events WHERE event_type='app_view' AND app_id=$1`, appID).Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 1 {
			t.Fatalf("count = %d, want 1", count)
		}
	})
}

func TestTrackAnalyticsEventRequiresAppForAppEvents(t *testing.T) {
	pool := testPool(t)
	s := &server{pool: pool, analyticsHashSecret: "test-secret-track-requires-app", analyticsLimiter: newRateLimiter(120, time.Minute)}

	body, _ := json.Marshal(map[string]string{
		"event": "app_open", "visitor_id": newTestIdentifier(t), "session_id": newTestIdentifier(t),
	})
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
	req.RemoteAddr = "9.9.9.3:1234"
	rec := httptest.NewRecorder()
	s.trackAnalyticsEvent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestTrackAnalyticsEventRejectsWalletLoginFromPublicPayload(t *testing.T) {
	pool := testPool(t)
	secret := "test-secret-track-wallet"
	s := &server{pool: pool, analyticsHashSecret: secret, analyticsLimiter: newRateLimiter(120, time.Minute)}

	t.Run("wallet_login event type rejected", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"event": "wallet_login", "visitor_id": newTestIdentifier(t), "session_id": newTestIdentifier(t),
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.4:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status %d, want 400", rec.Code)
		}
	})

	t.Run("wallet field on payload rejected", func(t *testing.T) {
		sessionID := newTestIdentifier(t)
		body, _ := json.Marshal(map[string]string{
			"event": "catalog_visit", "visitor_id": newTestIdentifier(t), "session_id": sessionID,
			"wallet_address": "NQ07 0000 0000 0000 0000 0000 0000 0000 0000",
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.5:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status %d, want 400", rec.Code)
		}
		var count int
		if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM analytics_events WHERE session_hash=$1`,
			hashAnalyticsIdentifier(secret, "session", sessionID)).Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 0 {
			t.Fatal("payload carrying a wallet field must not be recorded")
		}
	})
}

func TestTrackAnalyticsEventRejectsUnknownEventAndIdentifier(t *testing.T) {
	pool := testPool(t)
	s := &server{pool: pool, analyticsHashSecret: "test-secret-track-unknown", analyticsLimiter: newRateLimiter(120, time.Minute)}

	t.Run("unknown event", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"event": "bogus_event", "visitor_id": newTestIdentifier(t), "session_id": newTestIdentifier(t),
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.6:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status %d, want 400", rec.Code)
		}
	})

	t.Run("malformed identifier", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"event": "catalog_visit", "visitor_id": "not-a-uuid", "session_id": newTestIdentifier(t),
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.7:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status %d, want 400", rec.Code)
		}
	})

	t.Run("unconfigured analytics returns 503", func(t *testing.T) {
		s := &server{pool: pool, analyticsLimiter: newRateLimiter(120, time.Minute)}
		body, _ := json.Marshal(map[string]string{
			"event": "catalog_visit", "visitor_id": newTestIdentifier(t), "session_id": newTestIdentifier(t),
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.10:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status %d, want 503", rec.Code)
		}
	})
}

func TestTrackAnalyticsEventRateLimitsByClientAndVisitor(t *testing.T) {
	pool := testPool(t)
	secret := "test-secret-track-ratelimit"
	t.Setenv("ANALYTICS_HASH_SECRET", secret)
	s := &server{pool: pool, analyticsHashSecret: secret, analyticsLimiter: newRateLimiter(2, time.Minute)}
	visitorID := newTestIdentifier(t)
	otherVisitor := newTestIdentifier(t)
	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = pool.Exec(ctx, `DELETE FROM analytics_events WHERE visitor_hash = $1 OR visitor_hash = $2`,
			hashAnalyticsIdentifier(secret, "visitor", visitorID), hashAnalyticsIdentifier(secret, "visitor", otherVisitor))
	})

	for i := 0; i < 2; i++ {
		body, _ := json.Marshal(map[string]string{
			"event": "catalog_visit", "visitor_id": visitorID, "session_id": newTestIdentifier(t),
		})
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
		req.RemoteAddr = "9.9.9.9:1234"
		rec := httptest.NewRecorder()
		s.trackAnalyticsEvent(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: status %d, want 204: %s", i+1, rec.Code, rec.Body.String())
		}
	}

	body, _ := json.Marshal(map[string]string{
		"event": "catalog_visit", "visitor_id": visitorID, "session_id": newTestIdentifier(t),
	})
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body))
	req.RemoteAddr = "9.9.9.9:1234"
	rec := httptest.NewRecorder()
	s.trackAnalyticsEvent(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429", rec.Code)
	}

	// A different visitor from the same client IP must not share the bucket.
	body2, _ := json.Marshal(map[string]string{
		"event": "catalog_visit", "visitor_id": otherVisitor, "session_id": newTestIdentifier(t),
	})
	req2 := httptest.NewRequest(http.MethodPost, "http://example.com/api/analytics/events", bytes.NewReader(body2))
	req2.RemoteAddr = "9.9.9.9:1234"
	rec2 := httptest.NewRecorder()
	s.trackAnalyticsEvent(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("different visitor should not be rate limited: status %d, %s", rec2.Code, rec2.Body.String())
	}
}
