package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGenerateAppCredentialFormat(t *testing.T) {
	plaintext, keyPrefix, hash, err := generateAppCredential("my-app")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.HasPrefix(plaintext, "nma_my-app_") {
		t.Fatalf("plaintext=%s", plaintext)
	}
	if !strings.HasPrefix(keyPrefix, "nma_my-app_") {
		t.Fatalf("keyPrefix=%s", keyPrefix)
	}
	derived, ok := credentialKeyPrefix(plaintext)
	if !ok || derived != keyPrefix {
		t.Fatalf("derived=%s ok=%v want %s", derived, ok, keyPrefix)
	}
	want := sha256.Sum256([]byte(plaintext))
	if !bytes.Equal(hash, want[:]) {
		t.Fatalf("hash mismatch")
	}
	if _, _, _, err := generateAppCredential("Bad_Slug"); err == nil {
		t.Fatal("expected invalid slug error")
	}
}

func TestCredentialKeyPrefixIgnoresSecretUnderscores(t *testing.T) {
	key := "nma_my-app_abc_def_ghijklmnop"
	prefix, ok := credentialKeyPrefix(key)
	if !ok {
		t.Fatal("expected ok")
	}
	if prefix != "nma_my-app_abc_def_ghij" {
		t.Fatalf("prefix=%s", prefix)
	}
}

func TestAppCredentialIssueListRevokeVerify(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	secret := "cred-test-secret"
	owner := "NQ05 TEST OWNER ADDRESS 0001"
	other := "NQ05 TEST OTHER ADDRESS 0002"
	slug := "cred-app-" + time.Now().Format("150405000000")

	if _, err := pool.Exec(ctx, `INSERT INTO users (wallet_address, display_name) VALUES ($1,'Owner'),($2,'Other')`, owner, other); err != nil {
		t.Fatalf("users: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE wallet_address IN ($1,$2)`, owner, other)
	})

	var appID string
	err := pool.QueryRow(ctx, `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline, description, status, declared_scopes)
		VALUES ($1, 'Cred App', 'cred.example.com', 'Utilities', 'cred-dev', 'Cred Dev', 'tag', 'desc', 'submitted', '{achievements:write}')
		RETURNING id::text`, slug).Scan(&appID)
	if err != nil {
		t.Fatalf("insert app: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM apps WHERE id=$1::uuid`, appID)
	})
	if _, err := pool.Exec(ctx, `INSERT INTO app_owners (app_slug, wallet_address) VALUES ($1,$2)`, slug, owner); err != nil {
		t.Fatalf("owner: %v", err)
	}

	s := &server{
		pool:                    pool,
		walletAuthSecret:        secret,
		nimconnectServiceToken:  "registry-token",
		credentialIssueLimiter:  newRateLimiter(10, time.Hour),
		credentialVerifyLimiter: newRateLimiter(120, time.Minute),
	}

	issue := func(wallet string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/apps/"+slug+"/credentials", strings.NewReader(`{"label":"prod"}`))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("slug", slug)
		req.AddCookie(&http.Cookie{Name: walletCookieName, Value: signWalletCookie(secret, wallet, time.Now().Add(time.Hour))})
		rec := httptest.NewRecorder()
		walletAuthMiddleware(secret, s.createAppCredential)(rec, req)
		return rec
	}

	rec := issue(owner)
	if rec.Code != http.StatusConflict {
		t.Fatalf("submitted issue status %d body %s", rec.Code, rec.Body.String())
	}

	if _, err := pool.Exec(ctx, `UPDATE apps SET status='approved' WHERE id=$1::uuid`, appID); err != nil {
		t.Fatalf("approve: %v", err)
	}

	rec = issue(other)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-owner status %d", rec.Code)
	}

	rec = issue(owner)
	if rec.Code != http.StatusCreated {
		t.Fatalf("issue status %d body %s", rec.Code, rec.Body.String())
	}
	var created AppCredentialCreated
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID < 1 || created.Key == "" || created.KeyPrefix == "" || created.Label != "prod" || created.CreatedBy != owner {
		t.Fatalf("created=%+v", created)
	}
	if !strings.HasPrefix(created.Key, "nma_"+slug+"_") {
		t.Fatalf("key=%s", created.Key)
	}

	// List must never return the secret.
	req := httptest.NewRequest(http.MethodGet, "/api/apps/"+slug+"/credentials", nil)
	req.SetPathValue("slug", slug)
	req.AddCookie(&http.Cookie{Name: walletCookieName, Value: signWalletCookie(secret, owner, time.Now().Add(time.Hour))})
	rec = httptest.NewRecorder()
	walletAuthMiddleware(secret, s.listAppCredentials)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status %d body %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), created.Key) {
		t.Fatal("list leaked plaintext key")
	}
	var listed struct {
		Credentials []AppCredentialMeta `json:"credentials"`
	}
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Credentials) != 1 || listed.Credentials[0].ID != created.ID || listed.Credentials[0].KeyPrefix != created.KeyPrefix {
		t.Fatalf("listed=%+v", listed.Credentials)
	}

	verify := func(key string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]string{"key": key})
		req := httptest.NewRequest(http.MethodPost, "/registry/credentials/verify", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer registry-token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		s.registryServiceAuth(s.verifyAppCredential)(rec, req)
		return rec
	}

	rec = verify(created.Key)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify status %d body %s", rec.Code, rec.Body.String())
	}
	var client AppClientRecord
	if err := json.NewDecoder(rec.Body).Decode(&client); err != nil {
		t.Fatalf("decode client: %v", err)
	}
	if client.Slug != slug || client.AppID != appID || len(client.DeclaredScopes) != 1 || client.DeclaredScopes[0] != "achievements:write" {
		t.Fatalf("client=%+v", client)
	}

	rec = verify(created.Key + "x")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad key status %d", rec.Code)
	}

	credID := strconv.FormatInt(created.ID, 10)
	req = httptest.NewRequest(http.MethodDelete, "/api/apps/"+slug+"/credentials/"+credID, nil)
	req.SetPathValue("slug", slug)
	req.SetPathValue("id", credID)
	req.AddCookie(&http.Cookie{Name: walletCookieName, Value: signWalletCookie(secret, owner, time.Now().Add(time.Hour))})
	rec = httptest.NewRecorder()
	walletAuthMiddleware(secret, s.revokeAppCredential)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("revoke status %d body %s", rec.Code, rec.Body.String())
	}

	// Idempotent revoke
	req = httptest.NewRequest(http.MethodDelete, "/api/apps/"+slug+"/credentials/"+credID, nil)
	req.SetPathValue("slug", slug)
	req.SetPathValue("id", credID)
	req.AddCookie(&http.Cookie{Name: walletCookieName, Value: signWalletCookie(secret, owner, time.Now().Add(time.Hour))})
	rec = httptest.NewRecorder()
	walletAuthMiddleware(secret, s.revokeAppCredential)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("revoke again status %d", rec.Code)
	}

	rec = verify(created.Key)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked verify status %d", rec.Code)
	}

	var reason string
	err = pool.QueryRow(ctx, `
		SELECT reason FROM app_client_changes WHERE app_id=$1::uuid AND reason=$2 ORDER BY id DESC LIMIT 1`,
		appID, clientChangeCredentialRevoked).Scan(&reason)
	if err != nil || reason != clientChangeCredentialRevoked {
		t.Fatalf("change feed reason=%s err=%v", reason, err)
	}

	// Unlisted apps stop verifying even with an active credential.
	rec = issue(owner)
	if rec.Code != http.StatusCreated {
		t.Fatalf("reissue status %d body %s", rec.Code, rec.Body.String())
	}
	var created2 AppCredentialCreated
	if err := json.NewDecoder(rec.Body).Decode(&created2); err != nil {
		t.Fatalf("decode created2: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE apps SET status='rejected' WHERE id=$1::uuid`, appID); err != nil {
		t.Fatalf("reject: %v", err)
	}
	rec = verify(created2.Key)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("rejected verify status %d", rec.Code)
	}
	rec = issue(owner)
	if rec.Code != http.StatusConflict {
		t.Fatalf("rejected issue status %d", rec.Code)
	}
}
