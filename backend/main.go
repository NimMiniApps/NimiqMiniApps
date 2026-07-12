package main

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name=$1)`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sql, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, name); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		slog.Info("applied migration", "name", name)
	}
	return nil
}

func corsMiddleware(allowed []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (slices.Contains(allowed, "*") || slices.Contains(allowed, origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
	})
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dbURL := env("DATABASE_URL", "postgres://nimiq:nimiq@localhost:5432/nimiq_miniapps?sslmode=disable")
	adminToken := env("ADMIN_TOKEN", "")
	adminWallets := parseAdminWallets(env("ADMIN_WALLET_ADDRESSES", ""))
	walletAuthSecret := env("WALLET_AUTH_SECRET", "")
	addr := env("HTTP_ADDR", ":8080")
	corsOrigins := strings.Split(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}
	if adminToken == "" && len(adminWallets) == 0 {
		slog.Warn("ADMIN_TOKEN and ADMIN_WALLET_ADDRESSES are empty; admin endpoints are disabled")
	} else if len(adminWallets) > 0 {
		slog.Info("admin wallet allowlist configured", "count", len(adminWallets))
	}
	if walletAuthSecret == "" {
		slog.Warn("WALLET_AUTH_SECRET is empty; wallet login endpoints are disabled")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 30; i++ { // ponytail: dumb retry loop covers postgres starting slower than us in compose
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			err = pool.Ping(ctx)
		}
		if err == nil {
			break
		}
		slog.Info("waiting for database", "attempt", i+1, "error", err.Error())
		time.Sleep(time.Second)
	}
	if err != nil {
		slog.Error("database unreachable", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	if err := migrate(ctx, pool); err != nil {
		slog.Error("migration failed", "error", err.Error())
		os.Exit(1)
	}

	s := &server{
		pool:             pool,
		nonces:           newNonceStore(),
		walletAuthSecret: walletAuthSecret,
		adminToken:       adminToken,
		adminWallets:     adminWallets,
		reviewLimiter:    newRateLimiter(5, time.Hour),
		statsLimiter:     newRateLimiter(20, time.Minute),
	}
	s.startDomainHealthWorker(ctx)
	s.startIconDiscoveryBackfill(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /openapi.json", s.openAPIJSON)
	mux.HandleFunc("GET /openapi.yaml", s.openAPIYAML)
	mux.HandleFunc("GET /robots.txt", s.robotsTxt)
	mux.HandleFunc("GET /sitemap.xml", s.sitemapXML)
	mux.HandleFunc("GET /og/apps/{slug}", s.ogAppHTML)
	mux.HandleFunc("GET /api/apps", s.listApps)
	mux.HandleFunc("GET /api/apps/{slug}/status", s.getSubmissionStatus)
	mux.HandleFunc("GET /api/apps/{slug}/related", s.getRelatedApps)
	mux.HandleFunc("GET /api/apps/{slug}", s.getApp)
	mux.HandleFunc("GET /api/categories", s.listCategories)
	mux.HandleFunc("GET /api/collections", s.listCollections)
	mux.HandleFunc("GET /api/developers", s.listDevelopers)
	mux.HandleFunc("GET /api/developers/{slug}", s.getDeveloper)
	mux.HandleFunc("POST /api/apps/submit", walletAuthMiddleware(walletAuthSecret, s.submitApp))
	mux.HandleFunc("POST /api/apps/{slug}/request-update", walletAuthMiddleware(walletAuthSecret, s.requestAppUpdate))
	mux.HandleFunc("POST /api/auth/challenge", s.authChallenge)
	mux.HandleFunc("POST /api/auth/verify", s.authVerify)
	mux.HandleFunc("GET /api/auth/me", walletAuthMiddleware(walletAuthSecret, s.authMe))
	mux.HandleFunc("GET /api/profile", walletAuthMiddleware(walletAuthSecret, s.getProfile))
	mux.HandleFunc("PUT /api/profile", walletAuthMiddleware(walletAuthSecret, s.updateProfile))
	mux.HandleFunc("POST /api/auth/logout", s.authLogout)
	mux.HandleFunc("POST /api/apps/{slug}/track", s.trackApp)
	mux.HandleFunc("GET /api/apps/{slug}/stats", s.ownerOrAdminAuth(s.appStats))
	mux.HandleFunc("GET /api/apps/{slug}/reviews", s.listReviews)
	mux.HandleFunc("POST /api/apps/{slug}/reviews", walletAuthMiddleware(walletAuthSecret, s.upsertReview))
	mux.HandleFunc("DELETE /api/apps/{slug}/reviews", walletAuthMiddleware(walletAuthSecret, s.deleteOwnReview))
	mux.HandleFunc("GET /api/my/apps", walletAuthMiddleware(walletAuthSecret, s.myApps))
	mux.HandleFunc("GET /api/my/favorites", walletAuthMiddleware(walletAuthSecret, s.myFavorites))
	mux.HandleFunc("POST /api/apps/{slug}/favorite", walletAuthMiddleware(walletAuthSecret, s.addFavorite))
	mux.HandleFunc("DELETE /api/apps/{slug}/favorite", walletAuthMiddleware(walletAuthSecret, s.removeFavorite))
	mux.HandleFunc("POST /api/apps/{slug}/owners", walletAuthMiddleware(walletAuthSecret, s.addAppOwnerSelf))
	mux.HandleFunc("DELETE /api/apps/{slug}/owners/{wallet}", walletAuthMiddleware(walletAuthSecret, s.removeAppOwnerSelf))
	mux.HandleFunc("GET /api/admin/stats", s.adminAuth(s.adminStats))
	mux.HandleFunc("GET /api/admin/revisions", s.adminAuth(s.adminListRevisions))
	mux.HandleFunc("POST /api/admin/revisions/{id}/approve", s.adminAuth(s.approveRevision))
	mux.HandleFunc("POST /api/admin/revisions/{id}/reject", s.adminAuth(s.rejectRevision))
	mux.HandleFunc("GET /api/admin/apps", s.adminAuth(s.adminListApps))
	mux.HandleFunc("GET /api/admin/users", s.adminAuth(s.adminSearchUsers))
	mux.HandleFunc("POST /api/admin/check-domains", s.adminAuth(s.adminCheckDomains))
	mux.HandleFunc("POST /api/admin/apps", s.adminAuth(s.createApp))
	mux.HandleFunc("PUT /api/admin/apps/{slug}", s.adminAuth(s.updateApp))
	mux.HandleFunc("PATCH /api/admin/apps/{slug}", s.adminAuth(s.updateApp))
	mux.HandleFunc("POST /api/admin/apps/{slug}/owners", s.adminAuth(s.adminAddAppOwner))
	mux.HandleFunc("DELETE /api/admin/apps/{slug}/owners/{wallet}", s.adminAuth(s.adminRemoveAppOwner))
	mux.HandleFunc("DELETE /api/admin/apps/{slug}", s.adminAuth(s.deleteApp))
	mux.HandleFunc("POST /api/admin/apps/{slug}/verify", s.adminAuth(s.setStatus("verified")))
	mux.HandleFunc("POST /api/admin/apps/{slug}/approve", s.adminAuth(s.setStatus("approved")))
	mux.HandleFunc("POST /api/admin/apps/{slug}/reject", s.adminAuth(s.rejectApp))
	mux.HandleFunc("DELETE /api/admin/apps/{slug}/reviews/{id}", s.adminAuth(s.adminDeleteReview))

	srv := &http.Server{
		Addr:    addr,
		Handler: logMiddleware(corsMiddleware(corsOrigins, mux)),
	}

	go func() {
		slog.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err.Error())
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
