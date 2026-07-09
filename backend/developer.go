package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// slugifyDisplayName mirrors the app-slug generation already used on the submit
// form client-side: lowercase, collapse non-alphanumeric runs into single hyphens,
// trim leading/trailing hyphens. May return "" for a name with no ASCII letters/digits.
func slugifyDisplayName(name string) string {
	lower := strings.ToLower(name)
	var b strings.Builder
	prevSep := true // suppress a leading hyphen
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			prevSep = false
		case !prevSep:
			b.WriteByte('-')
			prevSep = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

var errLastOwner = errors.New("can't remove the last owner")

func (s *server) isOwner(ctx context.Context, slug, wallet string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM app_owners WHERE app_slug=$1 AND wallet_address=$2)`,
		slug, wallet).Scan(&ok)
	return ok, err
}

// addOwner links wallet to slug's ownership set. The wallet must already have a
// profile with a display name set (same bar the old single-owner link enforced).
// Adding an already-current owner again is a no-op.
func (s *server) addOwner(ctx context.Context, slug, wallet string) error {
	wallet = strings.TrimSpace(wallet)
	if wallet == "" {
		return errors.New("wallet_address is required")
	}
	var displayName *string
	err := s.pool.QueryRow(ctx,
		`SELECT display_name FROM users WHERE wallet_address=$1`, wallet).
		Scan(&displayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("wallet must have logged in at least once")
	}
	if err != nil {
		return err
	}
	if displayName == nil || strings.TrimSpace(*displayName) == "" {
		return errors.New("wallet must set a display name on their profile first")
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO app_owners (app_slug, wallet_address) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		slug, wallet)
	return err
}

// removeOwner unlinks wallet from slug's ownership set. When allowEmpty is false
// (self-service), it refuses to remove the last remaining owner.
func (s *server) removeOwner(ctx context.Context, slug, wallet string, allowEmpty bool) error {
	if !allowEmpty {
		var count int
		if err := s.pool.QueryRow(ctx,
			`SELECT count(*) FROM app_owners WHERE app_slug=$1`, slug).Scan(&count); err != nil {
			return err
		}
		if count <= 1 {
			return errLastOwner
		}
	}
	_, err := s.pool.Exec(ctx,
		`DELETE FROM app_owners WHERE app_slug=$1 AND wallet_address=$2`, slug, wallet)
	return err
}

// resolveDeveloperSlug returns the developer_slug a wallet should submit under.
// A wallet that already owns an app reuses that app's developer_slug (identity is
// assigned once, at first submission — see the developer portal spec). Otherwise it
// derives one from displayName, appending -2, -3, ... on collision with a different
// wallet's developer_slug.
func (s *server) resolveDeveloperSlug(ctx context.Context, address, displayName string) (string, error) {
	var existing string
	err := s.pool.QueryRow(ctx, `
		SELECT a.developer_slug FROM apps a
		JOIN app_owners o ON o.app_slug = a.slug
		WHERE o.wallet_address = $1 LIMIT 1`, address).
		Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	base := slugifyDisplayName(displayName)
	if base == "" {
		return "", errors.New("display name must contain at least one letter or number")
	}
	slug := base
	for i := 2; ; i++ {
		var count int
		if err := s.pool.QueryRow(ctx, `
			SELECT count(*) FROM apps a
			WHERE a.developer_slug = $1
			  AND NOT EXISTS (SELECT 1 FROM app_owners o WHERE o.app_slug = a.slug AND o.wallet_address = $2)`,
			slug, address).Scan(&count); err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		slug = base + "-" + strconv.Itoa(i)
	}
}

func (s *server) myApps(w http.ResponseWriter, r *http.Request, address string) {
	rows, err := s.pool.Query(r.Context(),
		`SELECT `+appColumns+` FROM apps
		 WHERE EXISTS (SELECT 1 FROM app_owners o WHERE o.app_slug = apps.slug AND o.wallet_address = $1)
		 ORDER BY created_at DESC`, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type myApp struct {
		App
		HasPendingRevision bool `json:"has_pending_revision"`
	}
	items := []myApp{}
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		pending, err := s.hasPendingRevision(r.Context(), a.Slug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, myApp{App: a, HasPendingRevision: pending})
	}
	writeJSON(w, http.StatusOK, items)
}

type ownerRequestBody struct {
	WalletAddress string `json:"wallet_address"`
}

func (s *server) addAppOwnerSelf(w http.ResponseWriter, r *http.Request, address string) {
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
	var body ownerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := s.addOwner(r.Context(), slug, body.WalletAddress); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (s *server) removeAppOwnerSelf(w http.ResponseWriter, r *http.Request, address string) {
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
	target := r.PathValue("wallet")
	if err := s.removeOwner(r.Context(), slug, target, false); err != nil {
		if errors.Is(err, errLastOwner) {
			writeError(w, http.StatusConflict, "can't remove the last owner — ask an admin to unclaim this app instead")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (s *server) adminAddAppOwner(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var body ownerRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := s.addOwner(r.Context(), slug, body.WalletAddress); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "added"})
}

func (s *server) adminRemoveAppOwner(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	target := r.PathValue("wallet")
	if err := s.removeOwner(r.Context(), slug, target, true); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (s *server) adminSearchUsers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, []struct{}{})
		return
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT wallet_address, display_name FROM users
		WHERE display_name ILIKE $1 OR wallet_address ILIKE $1
		ORDER BY display_name ASC NULLS LAST LIMIT 20`, q+"%")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type userResult struct {
		WalletAddress string  `json:"wallet_address"`
		DisplayName   *string `json:"display_name"`
	}
	items := []userResult{}
	for rows.Next() {
		var it userResult
		if err := rows.Scan(&it.WalletAddress, &it.DisplayName); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, items)
}
