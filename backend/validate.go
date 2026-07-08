package main

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

var (
	slugRe        = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	validStatuses = []string{"submitted", "approved", "verified", "experimental", "rejected"}
	validAssets   = []string{"NIM", "USDT", "BTC", "ETH"}
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
	if !slices.Contains(validStatuses, a.Status) {
		problems = append(problems, "status must be one of: "+strings.Join(validStatuses, ", "))
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
