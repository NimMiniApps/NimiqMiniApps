package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://nimiq:nimiq@localhost:5432/nimiq_miniapps?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("database unavailable: %v", err)
	}
	if err := migrate(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestTrackEventUnknownEvent(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	var appID string
	err := pool.QueryRow(ctx, `SELECT id FROM apps LIMIT 1`).Scan(&appID)
	if err != nil {
		t.Skip("no apps in database")
	}
	if err := trackEvent(ctx, pool, appID, "click"); err != errUnknownTrackEvent {
		t.Fatalf("got %v, want errUnknownTrackEvent", err)
	}
}

func TestTrackEventUpsert(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	slug := "stats-test-" + time.Now().Format("150405")
	var appID string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'Stats Test', 'stats.example.com', 'Utilities', 'stats-dev', 'Stats Dev', 'tag', 'desc', 'approved')
		RETURNING id`, slug).Scan(&appID)
	if err != nil {
		t.Fatalf("insert app: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appID)
	})

	if err := trackEvent(ctx, pool, appID, "open"); err != nil {
		t.Fatalf("first open: %v", err)
	}
	var opens, views int
	err = pool.QueryRow(ctx, `SELECT opens, views FROM app_stats_daily WHERE app_id=$1 AND day=CURRENT_DATE`, appID).Scan(&opens, &views)
	if err != nil {
		t.Fatalf("query row: %v", err)
	}
	if opens != 1 || views != 0 {
		t.Fatalf("after first open: opens=%d views=%d, want 1/0", opens, views)
	}

	if err := trackEvent(ctx, pool, appID, "open"); err != nil {
		t.Fatalf("second open: %v", err)
	}
	if err := trackEvent(ctx, pool, appID, "view"); err != nil {
		t.Fatalf("view: %v", err)
	}
	err = pool.QueryRow(ctx, `SELECT opens, views FROM app_stats_daily WHERE app_id=$1 AND day=CURRENT_DATE`, appID).Scan(&opens, &views)
	if err != nil {
		t.Fatalf("query row: %v", err)
	}
	if opens != 2 || views != 1 {
		t.Fatalf("after increment: opens=%d views=%d, want 2/1", opens, views)
	}
}

func TestTrackAppUnknownSlugNoOps(t *testing.T) {
	pool := testPool(t)
	s := &server{pool: pool, statsLimiter: newRateLimiter(20, time.Minute)}
	body, _ := json.Marshal(map[string]string{"event": "open"})
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/apps/no-such-slug-xyz/track", bytes.NewReader(body))
	req.SetPathValue("slug", "no-such-slug-xyz")
	rec := httptest.NewRecorder()
	s.trackApp(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d, want 204", rec.Code)
	}
}

func TestTrackAppStatsRateLimit(t *testing.T) {
	s := &server{statsLimiter: newRateLimiter(20, time.Minute)}
	body, _ := json.Marshal(map[string]string{"event": "open"})
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodPost, "http://example.com/api/apps/demo/track", bytes.NewReader(body))
		req.SetPathValue("slug", "demo")
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		if !s.statsLimiter.allow(clientIP(req) + ":demo") {
			t.Fatalf("request %d should be allowed", i+1)
		}
		_ = rec
	}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/apps/demo/track", bytes.NewReader(body))
	req.SetPathValue("slug", "demo")
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()
	s.trackApp(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429", rec.Code)
	}

	// A different ephemeral port from the same client IP must share the same bucket —
	// r.RemoteAddr's port changes per TCP connection, so keying on it verbatim would
	// let every request through as an apparently "new" client.
	reqOtherPort := httptest.NewRequest(http.MethodPost, "http://example.com/api/apps/demo/track", bytes.NewReader(body))
	reqOtherPort.SetPathValue("slug", "demo")
	reqOtherPort.RemoteAddr = "1.2.3.4:5678"
	recOtherPort := httptest.NewRecorder()
	s.trackApp(recOtherPort, reqOtherPort)
	if recOtherPort.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429 (different port, same IP should still be rate-limited)", recOtherPort.Code)
	}
}

func TestOwnerOrAdminAuth(t *testing.T) {
	pool := testPool(t)
	secret := "stats-test-secret"
	adminWallet := "NQ07 0000 0000 0000 0000 0000 0000 0000 0000"
	ownerWallet := "NQ11 ABCD EFGH IJKL MNOP QRST UVWX YZ01 2345"
	otherWallet := "NQ22 0000 0000 0000 0000 0000 0000 0000 0000"
	slug := "stats-auth-" + time.Now().Format("150405")

	ctx := context.Background()
	var appID string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'Auth Test', 'auth.example.com', 'Utilities', 'auth-dev', 'Auth Dev', 'tag', 'desc', 'approved')
		RETURNING id`, slug).Scan(&appID)
	if err != nil {
		t.Fatalf("insert app: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1`, appID)
	})
	_, err = pool.Exec(ctx, `INSERT INTO app_owners (app_slug, wallet_address) VALUES ($1, $2)`, slug, ownerWallet)
	if err != nil {
		t.Fatalf("insert owner: %v", err)
	}

	s := &server{
		pool:             pool,
		walletAuthSecret: secret,
		adminToken:       "admin-token",
		adminWallets:     parseAdminWallets(adminWallet),
	}

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/apps/"+slug+"/stats", nil)
		req.SetPathValue("slug", slug)
		rec := httptest.NewRecorder()
		s.ownerOrAdminAuth(okHandler)(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status %d, want 401", rec.Code)
		}
	})

	t.Run("owner wallet returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/apps/"+slug+"/stats", nil)
		req.SetPathValue("slug", slug)
		req.AddCookie(&http.Cookie{Name: walletCookieName, Value: signWalletCookie(secret, ownerWallet, time.Now().Add(time.Hour))})
		rec := httptest.NewRecorder()
		s.ownerOrAdminAuth(okHandler)(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d, want 200", rec.Code)
		}
	})

	t.Run("admin wallet returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/apps/"+slug+"/stats", nil)
		req.SetPathValue("slug", slug)
		req.AddCookie(&http.Cookie{Name: walletCookieName, Value: signWalletCookie(secret, adminWallet, time.Now().Add(time.Hour))})
		rec := httptest.NewRecorder()
		s.ownerOrAdminAuth(okHandler)(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d, want 200", rec.Code)
		}
	})

	t.Run("admin bearer returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/apps/"+slug+"/stats", nil)
		req.SetPathValue("slug", slug)
		req.Header.Set("Authorization", "Bearer admin-token")
		rec := httptest.NewRecorder()
		s.ownerOrAdminAuth(okHandler)(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d, want 200", rec.Code)
		}
	})

	t.Run("other wallet returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/api/apps/"+slug+"/stats", nil)
		req.SetPathValue("slug", slug)
		req.AddCookie(&http.Cookie{Name: walletCookieName, Value: signWalletCookie(secret, otherWallet, time.Now().Add(time.Hour))})
		rec := httptest.NewRecorder()
		s.ownerOrAdminAuth(okHandler)(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status %d, want 403", rec.Code)
		}
	})
}

func TestAdminListAppsStatsTotals(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	slugWith := "stats-admin-with-" + time.Now().Format("150405")
	slugWithout := "stats-admin-without-" + time.Now().Format("150405")

	var appIDWith string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'With Stats', 'with.example.com', 'Utilities', 'dev', 'Dev', 'tag', 'desc', 'approved')
		RETURNING id`, slugWith).Scan(&appIDWith)
	if err != nil {
		t.Fatalf("insert with: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'Without Stats', 'without.example.com', 'Utilities', 'dev', 'Dev', 'tag', 'desc', 'approved')`, slugWithout)
	if err != nil {
		t.Fatalf("insert without: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE slug = ANY($1)`, []string{slugWith, slugWithout})
	})

	_, err = pool.Exec(ctx, `
		INSERT INTO app_stats_daily (app_id, day, opens, views) VALUES ($1, CURRENT_DATE, 5, 10)`, appIDWith)
	if err != nil {
		t.Fatalf("insert stats: %v", err)
	}

	s := &server{pool: pool}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/admin/apps", nil)
	rec := httptest.NewRecorder()
	s.adminListApps(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	var apps []App
	if err := json.NewDecoder(rec.Body).Decode(&apps); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var withStats, withoutStats *App
	for i := range apps {
		switch apps[i].Slug {
		case slugWith:
			withStats = &apps[i]
		case slugWithout:
			withoutStats = &apps[i]
		}
	}
	if withStats == nil {
		t.Fatal("app with stats not found in response")
	}
	if withStats.TotalOpens != 5 || withStats.TotalViews != 10 {
		t.Fatalf("with stats: opens=%d views=%d, want 5/10", withStats.TotalOpens, withStats.TotalViews)
	}
	if withoutStats == nil {
		t.Fatal("app without stats not found in response")
	}
	if withoutStats.TotalOpens != 0 || withoutStats.TotalViews != 0 {
		t.Fatalf("without stats: opens=%d views=%d, want 0/0", withoutStats.TotalOpens, withoutStats.TotalViews)
	}
}
