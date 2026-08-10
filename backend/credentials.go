package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	clientChangeCredentialRevoked = "credential_revoked"
	appCredentialKeyPrefixLen     = 12
	appCredentialSecretBytes      = 32
	maxCredentialLabelLen         = 80
)

type AppCredentialMeta struct {
	ID         int64      `json:"id"`
	KeyPrefix  string     `json:"key_prefix"`
	Label      string     `json:"label"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  string     `json:"created_by"`
	LastUsedAt *time.Time `json:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

type AppCredentialCreated struct {
	AppCredentialMeta
	Key string `json:"key"`
}

func hashAppCredential(plaintext string) []byte {
	sum := sha256.Sum256([]byte(plaintext))
	return sum[:]
}

func generateAppCredential(slug string) (plaintext, keyPrefix string, hash []byte, err error) {
	slug = strings.TrimSpace(slug)
	if slug == "" || !slugRe.MatchString(slug) {
		return "", "", nil, errors.New("invalid slug")
	}
	buf := make([]byte, appCredentialSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", nil, err
	}
	secret := base64.RawURLEncoding.EncodeToString(buf)
	plaintext = "nma_" + slug + "_" + secret
	keyPrefix, ok := credentialKeyPrefix(plaintext)
	if !ok {
		return "", "", nil, errors.New("failed to derive key prefix")
	}
	return plaintext, keyPrefix, hashAppCredential(plaintext), nil
}

// credentialKeyPrefix returns nma_<slug>_<first 12 of secret>.
// Slugs never contain '_', so the first '_' after "nma_" separates slug from secret
// even though the secret (base64url) may contain '_'.
func credentialKeyPrefix(plaintext string) (string, bool) {
	if !strings.HasPrefix(plaintext, "nma_") {
		return "", false
	}
	rest := plaintext[len("nma_"):]
	idx := strings.IndexByte(rest, '_')
	if idx <= 0 || idx >= len(rest)-1 {
		return "", false
	}
	secret := rest[idx+1:]
	if len(secret) < appCredentialKeyPrefixLen {
		return "", false
	}
	return "nma_" + rest[:idx] + "_" + secret[:appCredentialKeyPrefixLen], true
}

func (s *server) createAppCredential(w http.ResponseWriter, r *http.Request, address string) {
	slug := r.PathValue("slug")
	if s.credentialIssueLimiter != nil && !s.credentialIssueLimiter.allow(address) {
		writeError(w, http.StatusTooManyRequests, "too many credential issues; try again later")
		return
	}
	owner, err := s.isOwner(r.Context(), slug, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !owner {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}

	app, err := s.fetchApp(r, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isPublicStatus(app.Status) {
		writeError(w, http.StatusConflict, "credentials can only be issued for listed apps")
		return
	}

	var body struct {
		Label string `json:"label"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
	}
	label := strings.TrimSpace(body.Label)
	if len(label) > maxCredentialLabelLen {
		writeError(w, http.StatusBadRequest, "label must be at most "+strconv.Itoa(maxCredentialLabelLen)+" characters")
		return
	}

	var (
		plaintext string
		keyPrefix string
		hash      []byte
		id        int64
		createdAt time.Time
	)
	for attempt := 0; attempt < 5; attempt++ {
		plaintext, keyPrefix, hash, err = generateAppCredential(slug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		err = s.pool.QueryRow(r.Context(), `
			INSERT INTO app_credentials (app_id, key_prefix, secret_hash, label, created_by)
			VALUES ($1::uuid, $2, $3, $4, $5)
			RETURNING id, created_at`,
			app.ID, keyPrefix, hash, label, address,
		).Scan(&id, &createdAt)
		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "app_credentials_key_prefix_key") &&
			!strings.Contains(err.Error(), "duplicate key") {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not allocate unique key prefix")
		return
	}

	writeJSON(w, http.StatusCreated, AppCredentialCreated{
		AppCredentialMeta: AppCredentialMeta{
			ID:        id,
			KeyPrefix: keyPrefix,
			Label:     label,
			CreatedAt: createdAt,
			CreatedBy: address,
		},
		Key: plaintext,
	})
}

func (s *server) listAppCredentials(w http.ResponseWriter, r *http.Request, address string) {
	slug := r.PathValue("slug")
	owner, err := s.isOwner(r.Context(), slug, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !owner {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}

	appID, err := s.appIDBySlug(r.Context(), slug)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT id, key_prefix, label, created_at, created_by, last_used_at, revoked_at
		FROM app_credentials
		WHERE app_id=$1::uuid
		ORDER BY created_at DESC`, appID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := []AppCredentialMeta{}
	for rows.Next() {
		var item AppCredentialMeta
		if err := rows.Scan(&item.ID, &item.KeyPrefix, &item.Label, &item.CreatedAt, &item.CreatedBy, &item.LastUsedAt, &item.RevokedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"credentials": items})
}

func (s *server) revokeAppCredential(w http.ResponseWriter, r *http.Request, address string) {
	slug := r.PathValue("slug")
	idRaw := r.PathValue("id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid credential id")
		return
	}

	owner, err := s.isOwner(r.Context(), slug, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !owner {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}

	appID, err := s.appIDBySlug(r.Context(), slug)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx, err := s.pool.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(r.Context())

	tag, err := tx.Exec(r.Context(), `
		UPDATE app_credentials
		SET revoked_at = now()
		WHERE id=$1 AND app_id=$2::uuid AND revoked_at IS NULL`, id, appID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := tx.QueryRow(r.Context(), `
			SELECT EXISTS(SELECT 1 FROM app_credentials WHERE id=$1 AND app_id=$2::uuid)`, id, appID).Scan(&exists); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !exists {
			writeError(w, http.StatusNotFound, "credential not found")
			return
		}
		if err := tx.Commit(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
		return
	}

	if err := recordClientChangeTx(r.Context(), tx, appID, clientChangeCredentialRevoked); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func (s *server) verifyAppCredential(w http.ResponseWriter, r *http.Request) {
	if s.credentialVerifyLimiter != nil {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			token = "anonymous"
		}
		if !s.credentialVerifyLimiter.allow(token) {
			writeError(w, http.StatusTooManyRequests, "too many verify requests; try again later")
			return
		}
	}

	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	key := strings.TrimSpace(body.Key)
	keyPrefix, ok := credentialKeyPrefix(key)
	if !ok {
		writeError(w, http.StatusUnauthorized, "invalid credential")
		return
	}

	var (
		id         int64
		secretHash []byte
		revokedAt  *time.Time
		lastUsedAt *time.Time
		slug       string
		status     string
	)
	err := s.pool.QueryRow(r.Context(), `
		SELECT c.id, c.secret_hash, c.revoked_at, c.last_used_at, a.slug, a.status
		FROM app_credentials c
		JOIN apps a ON a.id = c.app_id
		WHERE c.key_prefix=$1`, keyPrefix).Scan(&id, &secretHash, &revokedAt, &lastUsedAt, &slug, &status)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credential")
		return
	}
	want := hashAppCredential(key)
	if subtle.ConstantTimeCompare(secretHash, want) != 1 || revokedAt != nil || !isPublicStatus(status) {
		writeError(w, http.StatusUnauthorized, "invalid credential")
		return
	}

	s.touchCredentialLastUsed(r.Context(), id, lastUsedAt)

	app, err := s.fetchApp(r, slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toAppClientRecord(app))
}

func (s *server) touchCredentialLastUsed(ctx context.Context, id int64, lastUsedAt *time.Time) {
	if lastUsedAt != nil && time.Since(*lastUsedAt) < time.Minute {
		return
	}
	_, _ = s.pool.Exec(ctx, `
		UPDATE app_credentials
		SET last_used_at = now()
		WHERE id=$1 AND (last_used_at IS NULL OR last_used_at < now() - interval '1 minute')`, id)
}
