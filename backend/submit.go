package main

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// ponytail: in-memory per-IP rate limit, resets on restart; move to postgres/redis if we ever run >1 replica
var (
	submitMu  sync.Mutex
	submitLog = map[string][]time.Time{}
)

const (
	submitLimit  = 5
	submitWindow = time.Hour
)

func allowSubmit(ip string, now time.Time) bool {
	submitMu.Lock()
	defer submitMu.Unlock()
	recent := submitLog[ip][:0]
	for _, t := range submitLog[ip] {
		if now.Sub(t) < submitWindow {
			recent = append(recent, t)
		}
	}
	if len(recent) >= submitLimit {
		submitLog[ip] = recent
		return false
	}
	submitLog[ip] = append(recent, now)
	return true
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
	if !allowSubmit(clientIP(r), time.Now()) {
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
