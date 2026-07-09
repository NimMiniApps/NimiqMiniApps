package main

import (
	"strings"
	"testing"
)

func TestValidateApp(t *testing.T) {
	valid := func() App {
		return App{
			Slug: "my-app", Name: "My App", Domain: "my.app", Category: "Games",
			DeveloperSlug: "dev", DeveloperName: "Dev", Tagline: "Fun",
			Status: "submitted", Assets: []string{"NIM"},
		}
	}

	a := valid()
	if err := validateApp(&a); err != nil {
		t.Fatalf("valid app rejected: %v", err)
	}
	if a.Description != "Fun" {
		t.Errorf("description should fall back to tagline, got %q", a.Description)
	}

	bad := map[string]func(*App){
		"uppercase slug":   func(a *App) { a.Slug = "MyApp" },
		"empty slug":       func(a *App) { a.Slug = "" },
		"scheme in domain": func(a *App) { a.Domain = "https://my.app" },
		"empty name":       func(a *App) { a.Name = "" },
		"bad category":     func(a *App) { a.Category = "Whatever" },
		"bad status":        func(a *App) { a.Status = "published" },
		"bad release stage": func(a *App) { a.ReleaseStage = "preview" },
		"bad asset":         func(a *App) { a.Assets = []string{"DOGE"} },
		"bad website url":   func(a *App) { u := "not-a-url"; a.WebsiteURL = &u },
		"bad icon url":      func(a *App) { u := "ftp://bad"; a.IconURL = &u },
		"bad social":        func(a *App) { a.Socials = []SocialLink{{Platform: "twitter", URL: "not-a-url"}} },
		"bad social platform": func(a *App) { a.Socials = []SocialLink{{Platform: "myspace", URL: "https://example.com"}} },
	}
	for name, mutate := range bad {
		a := valid()
		mutate(&a)
		if err := validateApp(&a); err == nil {
			t.Errorf("%s: expected error, got none", name)
		}
	}
}

func TestValidateSubmitterContact(t *testing.T) {
	if msg := validateSubmitterContact(""); msg == "" {
		t.Fatal("expected error for empty contact")
	}
	if msg := validateSubmitterContact("  "); msg == "" {
		t.Fatal("expected error for whitespace contact")
	}
	if msg := validateSubmitterContact("@dev"); msg != "" {
		t.Fatalf("valid contact rejected: %q", msg)
	}
	if msg := validateSubmitterContact(strings.Repeat("a", 201)); msg == "" {
		t.Fatal("expected error for long contact")
	}
}
