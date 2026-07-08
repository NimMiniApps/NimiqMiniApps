package main

import (
	"context"
	"crypto/subtle"
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
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			writeError(w, http.StatusUnauthorized, "invalid or missing admin token")
			return
		}
		next(w, r)
	}
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
	addr := env("HTTP_ADDR", ":8080")
	corsOrigins := strings.Split(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}
	if adminToken == "" {
		slog.Warn("ADMIN_TOKEN is empty; admin endpoints are disabled")
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

	s := &server{pool: pool}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/apps", s.listApps)
	mux.HandleFunc("GET /api/apps/{slug}", s.getApp)
	mux.HandleFunc("GET /api/categories", s.listCategories)
	mux.HandleFunc("GET /api/developers/{slug}", s.getDeveloper)
	mux.HandleFunc("POST /api/apps/submit", s.submitApp)
	mux.HandleFunc("GET /api/admin/apps", authMiddleware(adminToken, s.adminListApps))
	mux.HandleFunc("POST /api/admin/apps", authMiddleware(adminToken, s.createApp))
	mux.HandleFunc("PUT /api/admin/apps/{slug}", authMiddleware(adminToken, s.updateApp))
	mux.HandleFunc("PATCH /api/admin/apps/{slug}", authMiddleware(adminToken, s.updateApp))
	mux.HandleFunc("DELETE /api/admin/apps/{slug}", authMiddleware(adminToken, s.deleteApp))
	mux.HandleFunc("POST /api/admin/apps/{slug}/verify", authMiddleware(adminToken, s.setStatus("verified")))
	mux.HandleFunc("POST /api/admin/apps/{slug}/approve", authMiddleware(adminToken, s.setStatus("approved")))
	mux.HandleFunc("POST /api/admin/apps/{slug}/reject", authMiddleware(adminToken, s.setStatus("rejected")))

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
