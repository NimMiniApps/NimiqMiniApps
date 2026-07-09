package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const iconDiscoveryMaxHTML = 512 << 10 // 512 KiB

var linkTagRe = regexp.MustCompile(`(?is)<link\s+([^>]+)>`)
var attrRe = regexp.MustCompile(`(?i)([a-z_:][-a-z0-9_:.]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)

type iconCandidate struct {
	href     string
	priority int
	size     int
}

func discoverIconURL(domain string) (string, error) {
	pageURL := probeDomainURL(domain)
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "NimiqMiniApps-IconDiscovery/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := domainCheckClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, iconDiscoveryMaxHTML))
	if err != nil {
		return "", err
	}

	base, _ := url.Parse(pageURL)
	candidates := parseIconCandidates(string(body), base)
	if len(candidates) == 0 {
		fallback := resolveIconHref(base, "/favicon.ico")
		if fallback == "" {
			return "", fmt.Errorf("no icon found")
		}
		if ok := iconURLReachable(fallback); !ok {
			return "", fmt.Errorf("favicon unreachable")
		}
		return fallback, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority > candidates[j].priority
		}
		return candidates[i].size > candidates[j].size
	})

	for _, c := range candidates {
		if ok := iconURLReachable(c.href); ok {
			return c.href, nil
		}
	}
	return "", fmt.Errorf("no reachable icon found")
}

func parseIconCandidates(html string, base *url.URL) []iconCandidate {
	var out []iconCandidate
	for _, tag := range linkTagRe.FindAllStringSubmatch(html, -1) {
		attrs := parseHTMLAttrs(tag[1])
		rel := strings.ToLower(strings.TrimSpace(attrs["rel"]))
		if rel == "" {
			continue
		}
		href := strings.TrimSpace(attrs["href"])
		if href == "" {
			continue
		}
		abs := resolveIconHref(base, href)
		if abs == "" {
			continue
		}
		if !strings.Contains(rel, "icon") && !strings.Contains(rel, "apple-touch-icon") {
			continue
		}
		priority := 40
		if strings.Contains(rel, "apple-touch-icon") {
			priority = 100
		} else if strings.Contains(rel, "shortcut") {
			priority = 50
		}
		out = append(out, iconCandidate{
			href:     abs,
			priority: priority,
			size:     parseIconSizes(attrs["sizes"]),
		})
	}
	return out
}

func parseHTMLAttrs(raw string) map[string]string {
	attrs := map[string]string{}
	for _, m := range attrRe.FindAllStringSubmatch(raw, -1) {
		val := m[2]
		if val == "" {
			val = m[3]
		}
		if val == "" {
			val = m[4]
		}
		attrs[strings.ToLower(m[1])] = val
	}
	return attrs
}

func parseIconSizes(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	best := 0
	for part := range strings.SplitSeq(raw, " ") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		w, h, ok := strings.Cut(part, "x")
		if !ok {
			continue
		}
		for _, dim := range []string{w, h} {
			if n, err := strconv.Atoi(dim); err == nil && n > best {
				best = n
			}
		}
	}
	return best
}

func resolveIconHref(base *url.URL, href string) string {
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	if abs.Host == "" {
		return ""
	}
	return abs.String()
}

func iconURLReachable(iconURL string) bool {
	req, err := http.NewRequest(http.MethodHead, iconURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "NimiqMiniApps-IconDiscovery/1.0")
	resp, err := domainCheckClient.Do(req)
	if err != nil {
		req, err = http.NewRequest(http.MethodGet, iconURL, nil)
		if err != nil {
			return false
		}
		req.Header.Set("User-Agent", "NimiqMiniApps-IconDiscovery/1.0")
		resp, err = domainCheckClient.Do(req)
		if err != nil {
			return false
		}
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func needsIconDiscovery(iconURL, discoveredIconURL *string) bool {
	if iconURL != nil && strings.TrimSpace(*iconURL) != "" {
		return false
	}
	if discoveredIconURL != nil && strings.TrimSpace(*discoveredIconURL) != "" {
		return false
	}
	return true
}

func effectiveIconURL(iconURL, discoveredIconURL *string) string {
	if iconURL != nil && strings.TrimSpace(*iconURL) != "" {
		return strings.TrimSpace(*iconURL)
	}
	if discoveredIconURL != nil && strings.TrimSpace(*discoveredIconURL) != "" {
		return strings.TrimSpace(*discoveredIconURL)
	}
	return ""
}

func (s *server) startIconDiscoveryBackfill(ctx context.Context) {
	go func() {
		rows, err := s.pool.Query(ctx, `
			SELECT slug, domain FROM apps
			WHERE status != 'rejected'
			AND (icon_url IS NULL OR btrim(icon_url) = '')
			AND (discovered_icon_url IS NULL OR btrim(discovered_icon_url) = '')`)
		if err != nil {
			slog.Error("icon discovery backfill query failed", "error", err.Error())
			return
		}
		defer rows.Close()
		var queued int
		for rows.Next() {
			var slug, domain string
			if err := rows.Scan(&slug, &domain); err != nil {
				slog.Error("icon discovery backfill scan failed", "error", err.Error())
				return
			}
			queued++
			go s.tryDiscoverAppIcon(ctx, slug, domain)
		}
		if queued > 0 {
			slog.Info("icon discovery backfill started", "apps", queued)
		}
	}()
}

func (s *server) tryDiscoverAppIcon(ctx context.Context, slug, domain string) {
	iconURL, err := discoverIconURL(domain)
	if err != nil {
		slog.Info("icon discovery failed", "slug", slug, "domain", domain, "reason", err.Error())
		return
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE apps SET discovered_icon_url=$1, updated_at=updated_at
		 WHERE slug=$2 AND (icon_url IS NULL OR btrim(icon_url) = '')
		 AND (discovered_icon_url IS NULL OR btrim(discovered_icon_url) = '')`,
		iconURL, slug)
	if err != nil {
		slog.Error("icon discovery save failed", "slug", slug, "error", err.Error())
	}
}
