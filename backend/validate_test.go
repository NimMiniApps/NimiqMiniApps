package main

import "testing"

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
	}
	for name, mutate := range bad {
		a := valid()
		mutate(&a)
		if err := validateApp(&a); err == nil {
			t.Errorf("%s: expected error, got none", name)
		}
	}
}
