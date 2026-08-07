package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLaunchOrigins(t *testing.T) {
	website := "https://www.example.com/path"
	got := launchOrigins("app.example.com", &website)
	want := []string{"https://app.example.com", "https://www.example.com"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestValidateDeclaredScopes(t *testing.T) {
	a := App{
		Slug: "scope-app", Name: "Scope", Domain: "scope.example.com", Category: "Games",
		DeveloperSlug: "dev", DeveloperName: "Dev", Tagline: "t", Description: "d",
		Status: "approved", DeclaredScopes: []string{"Friends:Read", "friends:read", "bad"},
	}
	err := validateApp(&a)
	if err == nil {
		t.Fatal("expected validation error")
	}
	a.DeclaredScopes = []string{"friends:read", "achievements:read"}
	if err := validateApp(&a); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(a.DeclaredScopes) != 2 {
		t.Fatalf("scopes=%v", a.DeclaredScopes)
	}
}

func TestGetAppClientAuthAndStatuses(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	slug := "registry-client-" + time.Now().Format("150405.000000")
	website := "https://alt.example.com/"
	_, err := pool.Exec(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status, website_url, declared_scopes)
		VALUES ($1, 'Registry Client', 'registry.example.com', 'Utilities', 'reg-dev', 'Reg Dev', 'tag', 'desc', 'submitted', $2, '{friends:read}')`,
		slug, website)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE slug=$1`, slug) })

	s := &server{pool: pool, nimconnectServiceToken: "registry-token"}

	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+slug+"/client", nil)
	rec := httptest.NewRecorder()
	s.registryServiceAuth(s.getAppClient)(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no auth status %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/apps/"+slug+"/client", nil)
	req.Header.Set("Authorization", "Bearer registry-token")
	req.SetPathValue("slug", slug)
	rec = httptest.NewRecorder()
	s.registryServiceAuth(s.getAppClient)(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("submitted status %d body %s", rec.Code, rec.Body.String())
	}

	_, err = pool.Exec(ctx, `UPDATE apps SET status='approved' WHERE slug=$1`, slug)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/apps/"+slug+"/client", nil)
	req.Header.Set("Authorization", "Bearer registry-token")
	req.SetPathValue("slug", slug)
	rec = httptest.NewRecorder()
	s.registryServiceAuth(s.getAppClient)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("approved status %d body %s", rec.Code, rec.Body.String())
	}
	var client AppClientRecord
	if err := json.NewDecoder(rec.Body).Decode(&client); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if client.AppID == "" || client.Slug != slug || client.Verified || client.Status != "approved" {
		t.Fatalf("client=%+v", client)
	}
	if len(client.DeclaredScopes) != 1 || client.DeclaredScopes[0] != "friends:read" {
		t.Fatalf("scopes=%v", client.DeclaredScopes)
	}
	if len(client.LaunchOrigins) != 2 {
		t.Fatalf("origins=%v", client.LaunchOrigins)
	}

	_, err = pool.Exec(ctx, `UPDATE apps SET status='experimental', declared_scopes='{}' WHERE slug=$1`, slug)
	if err != nil {
		t.Fatalf("experimental: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/apps/"+slug+"/client", nil)
	req.Header.Set("Authorization", "Bearer registry-token")
	req.SetPathValue("slug", slug)
	rec = httptest.NewRecorder()
	s.registryServiceAuth(s.getAppClient)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("experimental status %d", rec.Code)
	}
	if err := json.NewDecoder(rec.Body).Decode(&client); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if client.Status != "experimental" || client.Verified || len(client.DeclaredScopes) != 0 {
		t.Fatalf("experimental client=%+v", client)
	}

	_, err = pool.Exec(ctx, `UPDATE apps SET status='rejected' WHERE slug=$1`, slug)
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/apps/"+slug+"/client", nil)
	req.Header.Set("Authorization", "Bearer registry-token")
	req.SetPathValue("slug", slug)
	rec = httptest.NewRecorder()
	s.registryServiceAuth(s.getAppClient)(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("rejected status %d", rec.Code)
	}
}

func TestListAppClientChanges(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	slug := "registry-changed-" + time.Now().Format("150405.000000")
	var appID string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status)
		VALUES ($1, 'Changed', 'changed.example.com', 'Utilities', 'chg-dev', 'Chg Dev', 'tag', 'desc', 'approved')
		RETURNING id::text`, slug).Scan(&appID)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE slug=$1`, slug) })

	since := time.Now().UTC().Add(-time.Second).Format(time.RFC3339Nano)
	s := &server{pool: pool, nimconnectServiceToken: "registry-token"}
	s.recordClientChange(ctx, appID, clientChangeOwnerTransfer)
	s.recordClientChange(ctx, appID, clientChangeScopesChanged)
	s.recordClientChange(ctx, appID, clientChangeStatusChanged)

	req := httptest.NewRequest(http.MethodGet, "/api/apps/changed", nil)
	req.Header.Set("Authorization", "Bearer registry-token")
	rec := httptest.NewRecorder()
	s.registryServiceAuth(s.listAppClientChanges)(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing since status %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/apps/changed?since="+since, nil)
	req.Header.Set("Authorization", "Bearer registry-token")
	rec = httptest.NewRecorder()
	s.registryServiceAuth(s.listAppClientChanges)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Changes []AppClientChange `json:"changes"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var found *AppClientChange
	for i := range body.Changes {
		if body.Changes[i].AppID == appID {
			found = &body.Changes[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("app not in changes: %+v", body.Changes)
	}
	if found.Slug != slug {
		t.Fatalf("slug=%s", found.Slug)
	}
	for _, reason := range []string{clientChangeOwnerTransfer, clientChangeScopesChanged, clientChangeStatusChanged} {
		ok := false
		for _, r := range found.Reasons {
			if r == reason {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("missing reason %s in %v", reason, found.Reasons)
		}
	}
}
