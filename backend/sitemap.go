package main

import (
	"net/http"
	"os"
	"strings"
	"time"
)

func siteURL() string {
	if v := strings.TrimRight(os.Getenv("SITE_URL"), "/"); v != "" {
		return v
	}
	return "https://nimiqminiapps.com"
}

func (s *server) robotsTxt(w http.ResponseWriter, r *http.Request) {
	apiURL := strings.TrimRight(os.Getenv("API_PUBLIC_URL"), "/")
	if apiURL == "" {
		apiURL = siteURL()
	}
	body := "User-agent: *\nAllow: /\n\nSitemap: " + apiURL + "/sitemap.xml\n"
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(body))
}

func (s *server) sitemapXML(w http.ResponseWriter, r *http.Request) {
	base := siteURL()
	rows, err := s.pool.Query(r.Context(),
		`SELECT slug, updated_at FROM apps WHERE `+publicStatuses+` ORDER BY updated_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)

	static := []string{"", "/apps", "/build", "/submit"}
	now := time.Now().UTC().Format("2006-01-02")
	for _, path := range static {
		b.WriteString("<url><loc>")
		b.WriteString(xmlEscape(base + path))
		b.WriteString("</loc><lastmod>")
		b.WriteString(now)
		b.WriteString("</lastmod></url>")
	}

	for rows.Next() {
		var slug string
		var updated time.Time
		if err := rows.Scan(&slug, &updated); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		b.WriteString("<url><loc>")
		b.WriteString(xmlEscape(base + "/apps/" + slug))
		b.WriteString("</loc><lastmod>")
		b.WriteString(updated.UTC().Format("2006-01-02"))
		b.WriteString("</lastmod></url>")
	}

	rows2, err := s.pool.Query(r.Context(),
		`SELECT developer_slug, MAX(updated_at) FROM apps WHERE `+publicStatuses+` GROUP BY developer_slug`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var slug string
			var updated time.Time
			if err := rows2.Scan(&slug, &updated); err != nil {
				break
			}
			b.WriteString("<url><loc>")
			b.WriteString(xmlEscape(base + "/apps?developer=" + slug))
			b.WriteString("</loc><lastmod>")
			b.WriteString(updated.UTC().Format("2006-01-02"))
			b.WriteString("</lastmod></url>")
		}
	}

	b.WriteString("</urlset>")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write([]byte(b.String()))
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return r.Replace(s)
}
