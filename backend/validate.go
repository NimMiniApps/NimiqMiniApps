package main

import (
	"errors"
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
	validAssets       = []string{"NIM", "USDT", "BTC", "ETH"}
)

func validateApp(a *App) error {
	var problems []string
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
