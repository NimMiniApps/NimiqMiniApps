package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func domainCheckEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("DOMAIN_CHECK_ENABLED")))
	return v == "" || v == "1" || v == "true" || v == "yes"
}

// domainCheckHealthyInterval is how long to wait before re-checking a reachable domain.
func domainCheckHealthyInterval() time.Duration {
	if v := os.Getenv("DOMAIN_CHECK_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 1 * time.Hour
}

// domainCheckOfflineInterval is how long to wait before re-checking an unreachable domain.
func domainCheckOfflineInterval() time.Duration {
	if v := os.Getenv("DOMAIN_CHECK_OFFLINE_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 15 * time.Minute
}

// domainCheckTick is how often the worker wakes to look for due checks.
func domainCheckTick() time.Duration {
	if v := os.Getenv("DOMAIN_CHECK_TICK"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	tick := domainCheckOfflineInterval() / 3
	if tick < time.Minute {
		return time.Minute
	}
	if tick > 5*time.Minute {
		return 5 * time.Minute
	}
	return tick
}

func domainCheckTimeout() time.Duration {
	if v := os.Getenv("DOMAIN_CHECK_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 10 * time.Second
}

var domainCheckClient = &http.Client{
	Timeout: domainCheckTimeout(),
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

func probeDomainURL(domain string) string {
	d := strings.TrimSpace(domain)
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimSuffix(d, "/")
	return "https://" + d
}

func checkDomainReachable(domain string) (bool, string) {
	url := probeDomainURL(domain)
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return false, err.Error()
	}
	req.Header.Set("User-Agent", "NimiqMiniApps-DomainCheck/1.0")

	resp, err := domainCheckClient.Do(req)
	if err != nil {
		// Some hosts reject HEAD; try GET once.
		req, err = http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return false, err.Error()
		}
		req.Header.Set("User-Agent", "NimiqMiniApps-DomainCheck/1.0")
		resp, err = domainCheckClient.Do(req)
		if err != nil {
			return false, err.Error()
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, ""
	}
	return false, "HTTP " + strconv.Itoa(resp.StatusCode)
}

const domainCheckDueSQL = `
SELECT slug, domain, icon_url, discovered_icon_url FROM apps
WHERE status != 'rejected'
AND (
	domain_checked_at IS NULL
	OR (domain_reachable = false AND domain_checked_at < $1)
	OR (domain_reachable IS DISTINCT FROM false AND domain_checked_at < $2)
)`

func (s *server) runDomainHealthCheck(ctx context.Context, checkAll bool) error {
	now := time.Now()
	var rows pgx.Rows
	var err error

	if checkAll {
		rows, err = s.pool.Query(ctx, `SELECT slug, domain, icon_url, discovered_icon_url FROM apps WHERE status != 'rejected'`)
	} else {
		offlineCutoff := now.Add(-domainCheckOfflineInterval())
		healthyCutoff := now.Add(-domainCheckHealthyInterval())
		rows, err = s.pool.Query(ctx, domainCheckDueSQL, offlineCutoff, healthyCutoff)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	checked := 0
	for rows.Next() {
		var slug, domain string
		var iconURL, discoveredIconURL *string
		if err := rows.Scan(&slug, &domain, &iconURL, &discoveredIconURL); err != nil {
			return err
		}
		reachable, errMsg := checkDomainReachable(domain)
		var reachableVal *bool
		if reachable {
			v := true
			reachableVal = &v
		} else {
			v := false
			reachableVal = &v
		}
		_, err := s.pool.Exec(ctx,
			`UPDATE apps SET domain_reachable=$1, domain_checked_at=$2, updated_at=updated_at WHERE slug=$3`,
			reachableVal, now, slug)
		if err != nil {
			return err
		}
		checked++
		if errMsg != "" {
			slog.Info("domain unreachable", "slug", slug, "domain", domain, "reason", errMsg)
		}
		if reachable && needsIconDiscovery(iconURL, discoveredIconURL) {
			s.tryDiscoverAppIcon(ctx, slug, domain)
		}
	}
	if checkAll {
		slog.Info("domain health check complete", "checked", checked, "mode", "all")
	} else if checked > 0 {
		slog.Info("domain health check complete", "checked", checked, "mode", "due")
	}
	return nil
}

func (s *server) startDomainHealthWorker(ctx context.Context) {
	if !domainCheckEnabled() {
		slog.Info("domain health checks disabled")
		return
	}
	healthy := domainCheckHealthyInterval()
	offline := domainCheckOfflineInterval()
	tick := domainCheckTick()
	slog.Info("domain health worker started",
		"healthy_interval", healthy.String(),
		"offline_interval", offline.String(),
		"tick", tick.String(),
	)

	go func() {
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if err := s.runDomainHealthCheck(ctx, false); err != nil {
					slog.Error("domain health check failed", "error", err.Error())
				}
				timer.Reset(tick)
			}
		}
	}()
}

func (s *server) adminCheckDomains(w http.ResponseWriter, r *http.Request) {
	if err := s.runDomainHealthCheck(r.Context(), true); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
