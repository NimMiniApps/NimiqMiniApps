package main

import (
	"net"
	"net/http"
	"sync"
	"time"
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

func (s *server) submitApp(w http.ResponseWriter, r *http.Request) {
	if !allowSubmit(clientIP(r), time.Now()) {
		writeError(w, http.StatusTooManyRequests, "too many submissions, try again later")
		return
	}
	s.decodeAndInsert(w, r, func(a *App) {
		a.Status = "submitted"
		a.Featured = false
	}, true)
}
