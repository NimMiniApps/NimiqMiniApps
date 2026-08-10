package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

var socialCrawlerPattern = regexp.MustCompile(`(?i)(facebookexternalhit|facebot|twitterbot|linkedinbot|slackbot|discordbot|whatsapp|telegrambot|applebot|pinterestbot|googlebot|bingbot|embedly|quora link preview|vkshare|redditbot)`)

func isSocialCrawler(ua string) bool {
	return socialCrawlerPattern.MatchString(ua)
}

func (s *server) ogAppHTML(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	a, err := s.fetchApp(r, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isPublicStatus(a.Status) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}

	base := siteURL()
	pageURL := base + "/apps/" + slug
	title := htmlEscape(a.Name + " · Nimiq Mini Apps")
	description := htmlEscape(a.Tagline)
	if description == "" {
		description = htmlEscape(a.Description)
	}
	image := base + "/og-default.svg"
	if a.BannerURL != nil && *a.BannerURL != "" {
		image = *a.BannerURL
	} else if icon := effectiveIconURL(a.IconURL, a.DiscoveredIconURL); icon != "" {
		image = icon
	}
	image = htmlEscape(image)
	body := htmlEscape(a.Description)
	if body == "" {
		body = description
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover"/>
<title>%s</title>
<meta name="description" content="%s"/>
<meta property="og:title" content="%s"/>
<meta property="og:description" content="%s"/>
<meta property="og:url" content="%s"/>
<meta property="og:type" content="website"/>
<meta property="og:image" content="%s"/>
<meta name="twitter:card" content="summary_large_image"/>
<meta name="twitter:title" content="%s"/>
<meta name="twitter:description" content="%s"/>
<meta name="twitter:image" content="%s"/>
<link rel="canonical" href="%s"/>
</head>
<body>
<h1>%s</h1>
<p>%s</p>
<p><a href="%s">Open %s in Nimiq Mini Apps</a></p>
</body>
</html>`,
		title, description, title, description, htmlEscape(pageURL), image,
		title, description, image, htmlEscape(pageURL),
		title, body, htmlEscape(pageURL), htmlEscape(a.Name))
}

func (s *server) ogAppsListHTML(w http.ResponseWriter, r *http.Request) {
	base := siteURL()
	rows, err := s.pool.Query(r.Context(),
		`SELECT slug, name, tagline FROM apps WHERE `+publicStatuses+` ORDER BY updated_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	title := "All Mini Apps · Nimiq Mini Apps"
	description := "Browse every community-curated mini app for the Nimiq Pay wallet — games, tools, maps and more."

	var links strings.Builder
	for rows.Next() {
		var slug, name, tagline string
		if err := rows.Scan(&slug, &name, &tagline); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		links.WriteString(fmt.Sprintf(`<li><a href="%s/apps/%s">%s</a> &mdash; %s</li>`,
			base, htmlEscape(slug), htmlEscape(name), htmlEscape(tagline)))
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover"/>
<title>%s</title>
<meta name="description" content="%s"/>
<meta property="og:title" content="%s"/>
<meta property="og:description" content="%s"/>
<meta property="og:type" content="website"/>
<link rel="canonical" href="%s/apps"/>
</head>
<body>
<h1>%s</h1>
<ul>%s</ul>
</body>
</html>`,
		title, description, title, description, base, title, links.String())
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return r.Replace(s)
}
