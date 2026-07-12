package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	submitLimit  = 5
	submitWindow = time.Hour
)

func (s *server) allowSubmit(ctx context.Context, ip string, now time.Time) (bool, error) {
	cutoff := now.Add(-submitWindow)
	var count int
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM submit_rate_limits WHERE ip=$1 AND created_at > $2`, ip, cutoff).
		Scan(&count); err != nil {
		return false, err
	}
	if count >= submitLimit {
		return false, nil
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO submit_rate_limits (ip, created_at) VALUES ($1, $2)`, ip, now); err != nil {
		return false, err
	}
	// Best-effort cleanup; keeps the table small across replicas.
	_, _ = s.pool.Exec(ctx, `DELETE FROM submit_rate_limits WHERE created_at < $1`, cutoff)
	return true, nil
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" { // set by nginx; spoofable only when hitting the backend directly
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *server) submitApp(w http.ResponseWriter, r *http.Request, address string) {
	ok, err := s.allowSubmit(r.Context(), clientIP(r), time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusTooManyRequests, "too many submissions, try again later")
		return
	}

	var displayName *string
	if err := s.pool.QueryRow(r.Context(),
		`SELECT display_name FROM users WHERE wallet_address=$1`, address).
		Scan(&displayName); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if displayName == nil || strings.TrimSpace(*displayName) == "" {
		writeError(w, http.StatusBadRequest, "set a display name on your profile before submitting an app")
		return
	}

	devSlug, err := s.resolveDeveloperSlug(r.Context(), address, *displayName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.decodeAndInsert(w, r, func(a *App) {
		a.Status = "submitted"
		a.Featured = false
		a.DeveloperSlug = devSlug
		a.DeveloperName = *displayName
	}, true, &address)
}
