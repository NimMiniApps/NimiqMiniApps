package main

import (
	"context"
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

// validateDeveloperWallet checks that a caller-supplied developer_wallet_address (if
// any) points at a wallet that has logged in and set a display name. It does not
// overwrite developer_name/developer_slug — those stay admin-editable even for
// wallet-linked apps (rebrands, anonymous/legacy apps); only the wallet-driven
// submit flow (submitApp) derives them, and it does so itself before calling in here.
func (s *server) validateDeveloperWallet(ctx context.Context, a *App) error {
	if a.DeveloperWalletAddress == nil || strings.TrimSpace(*a.DeveloperWalletAddress) == "" {
		a.DeveloperWalletAddress = nil
		return nil
	}
	addr := strings.TrimSpace(*a.DeveloperWalletAddress)
	a.DeveloperWalletAddress = &addr

	var displayName *string
	err := s.pool.QueryRow(ctx,
		`SELECT display_name FROM users WHERE wallet_address=$1`, addr).
		Scan(&displayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("wallet owner must have logged in at least once")
	}
	if err != nil {
		return err
	}
	if displayName == nil || strings.TrimSpace(*displayName) == "" {
		return errors.New("wallet owner must set a display name on their profile first")
	}
	return nil
}

// resolveDeveloperSlug returns the developer_slug a wallet should submit under.
// A wallet that already owns an app reuses that app's developer_slug (identity is
// assigned once, at first submission — see the developer portal spec). Otherwise it
// derives one from displayName, appending -2, -3, ... on collision with a different
// wallet's developer_slug.
func (s *server) resolveDeveloperSlug(ctx context.Context, address, displayName string) (string, error) {
	var existing string
	err := s.pool.QueryRow(ctx,
		`SELECT developer_slug FROM apps WHERE developer_wallet_address=$1 LIMIT 1`, address).
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
		if err := s.pool.QueryRow(ctx,
			`SELECT count(*) FROM apps WHERE developer_slug=$1 AND developer_wallet_address IS DISTINCT FROM $2`,
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
		"SELECT "+appColumns+" FROM apps WHERE developer_wallet_address=$1 ORDER BY created_at DESC", address)
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
