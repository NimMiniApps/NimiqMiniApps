package main

import (
	"errors"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	slugRe            = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	youtubeURLRe      = regexp.MustCompile(`(?:youtube\.com/(?:watch\?v=|embed/|shorts/)|youtu\.be/)[a-zA-Z0-9_-]{11}`)
	validCategories   = []string{"Games", "Utilities", "Finance", "Maps", "Social", "Experiments"}
	validStatuses     = []string{"submitted", "approved", "verified", "experimental", "rejected"}
	validReleaseStage = []string{"concept", "alpha", "beta", "released"}
	validMediaTypes   = []string{"image", "youtube"}
	validSocials      = []string{"twitter", "discord", "telegram", "bluesky", "instagram", "youtube", "linkedin", "mastodon", "reddit", "tiktok"}
	validAssets       = []string{"NIM", "USDT", "USDC", "BTC", "ETH"}
)

func validateOptionalURL(field string, raw *string) string {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return ""
	}
	u, err := url.Parse(*raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return field + " must be a valid http or https URL"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return field + " must use http or https"
	}
	return ""
}

func validateSubmitterContact(contact string) string {
	if strings.TrimSpace(contact) == "" {
		return "submitter_contact is required (e.g. Telegram @handle or email)"
	}
	if len(contact) > 200 {
		return "submitter_contact must be at most 200 characters"
	}
	return ""
}

func normalizeDomain(domain string) string {
	d := strings.TrimSpace(domain)
	for {
		lower := strings.ToLower(d)
		switch {
		case strings.HasPrefix(lower, "https://"):
			d = d[8:]
		case strings.HasPrefix(lower, "http://"):
			d = d[7:]
		default:
			return strings.TrimSuffix(d, "/")
		}
	}
}

func validateApp(a *App) error {
	var problems []string
	a.Domain = normalizeDomain(a.Domain)
	if !slugRe.MatchString(a.Slug) {
		problems = append(problems, "slug is required and must be lowercase and url-safe (a-z, 0-9, hyphens)")
	}
	for field, val := range map[string]string{
		"name": a.Name, "domain": a.Domain, "category": a.Category, "tagline": a.Tagline,
		"developer_slug": a.DeveloperSlug, "developer_name": a.DeveloperName,
	} {
		if strings.TrimSpace(val) == "" {
			problems = append(problems, field+" is required")
		}
	}
	if strings.Contains(a.Domain, "://") {
		problems = append(problems, "domain must not include a scheme like https://")
	}
	if !slices.Contains(validCategories, a.Category) {
		problems = append(problems, "category must be one of: "+strings.Join(validCategories, ", "))
	}
	if !slices.Contains(validStatuses, a.Status) {
		problems = append(problems, "status must be one of: "+strings.Join(validStatuses, ", "))
	}
	if a.ReleaseStage == "" {
		a.ReleaseStage = "released"
	}
	if !slices.Contains(validReleaseStage, a.ReleaseStage) {
		problems = append(problems, "release_stage must be one of: "+strings.Join(validReleaseStage, ", "))
	}
	if a.FeaturedOrder < 0 {
		problems = append(problems, "featured_order must be 0 or greater")
	}
	for i, item := range a.Media {
		if !slices.Contains(validMediaTypes, item.Type) {
			problems = append(problems, "media item "+strconv.Itoa(i+1)+" type must be image or youtube")
		}
		if strings.TrimSpace(item.URL) == "" {
			problems = append(problems, "media item "+strconv.Itoa(i+1)+" url is required")
		}
		if item.Type == "youtube" && !youtubeURLRe.MatchString(item.URL) {
			problems = append(problems, "media item "+strconv.Itoa(i+1)+" is not a valid YouTube URL")
		}
		if item.Type == "image" {
			if msg := validateOptionalURL("media item "+strconv.Itoa(i+1)+" url", &item.URL); msg != "" {
				problems = append(problems, msg)
			}
		}
	}
	for i, item := range a.Socials {
		platform := strings.ToLower(strings.TrimSpace(item.Platform))
		if !slices.Contains(validSocials, platform) {
			problems = append(problems, "social item "+strconv.Itoa(i+1)+" platform must be one of: "+strings.Join(validSocials, ", "))
		}
		if strings.TrimSpace(item.URL) == "" {
			problems = append(problems, "social item "+strconv.Itoa(i+1)+" url is required")
		} else if msg := validateOptionalURL("social item "+strconv.Itoa(i+1)+" url", &item.URL); msg != "" {
			problems = append(problems, msg)
		}
	}
	for _, field := range []struct {
		name string
		val  *string
	}{
		{"website_url", a.WebsiteURL},
		{"github_url", a.GithubURL},
		{"icon_url", a.IconURL},
		{"banner_url", a.BannerURL},
	} {
		if msg := validateOptionalURL(field.name, field.val); msg != "" {
			problems = append(problems, msg)
		}
	}
	for _, asset := range a.Assets {
		if !slices.Contains(validAssets, asset) {
			problems = append(problems, "asset "+asset+" is not one of: "+strings.Join(validAssets, ", "))
		}
	}
	if a.Description == "" {
		a.Description = a.Tagline // description NOT NULL; fall back to tagline
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}
