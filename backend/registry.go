package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	clientChangeOwnerTransfer   = "owner_transfer"
	clientChangeScopesChanged   = "scopes_changed"
	clientChangeStatusChanged   = "status_changed"
	clientChangeMetadataChanged = "metadata_changed"
)

type AppClientRecord struct {
	AppID          string    `json:"app_id"`
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	IconURL        *string   `json:"icon_url"`
	Verified       bool      `json:"verified"`
	Status         string    `json:"status"`
	DeclaredScopes []string  `json:"declared_scopes"`
	LaunchOrigins  []string  `json:"launch_origins"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AppClientChange struct {
	AppID     string    `json:"app_id"`
	Slug      string    `json:"slug"`
	UpdatedAt time.Time `json:"updated_at"`
	Reasons   []string  `json:"reasons"`
}

func (s *server) registryServiceAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.nimconnectServiceToken == "" {
			writeError(w, http.StatusUnauthorized, "registry service access required")
			return
		}
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(s.nimconnectServiceToken)) != 1 {
			writeError(w, http.StatusUnauthorized, "registry service access required")
			return
		}
		next(w, r)
	}
}

func launchOrigins(domain string, websiteURL *string) []string {
	origins := make([]string, 0, 2)
	seen := map[string]struct{}{}
	add := func(origin string) {
		if origin == "" {
			return
		}
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	domain = strings.TrimSpace(domain)
	if domain != "" {
		add("https://" + domain)
	}
	if websiteURL != nil && strings.TrimSpace(*websiteURL) != "" {
		if u, err := url.Parse(strings.TrimSpace(*websiteURL)); err == nil && u.Scheme != "" && u.Host != "" {
			add(u.Scheme + "://" + u.Host)
		}
	}
	return origins
}

func clientIconURL(a App) *string {
	if a.IconURL != nil && strings.TrimSpace(*a.IconURL) != "" {
		return a.IconURL
	}
	if a.DiscoveredIconURL != nil && strings.TrimSpace(*a.DiscoveredIconURL) != "" {
		return a.DiscoveredIconURL
	}
	return nil
}

func toAppClientRecord(a App) AppClientRecord {
	scopes := a.DeclaredScopes
	if scopes == nil {
		scopes = []string{}
	}
	return AppClientRecord{
		AppID:          a.ID,
		Slug:           a.Slug,
		Name:           a.Name,
		IconURL:        clientIconURL(a),
		Verified:       a.Status == "verified",
		Status:         a.Status,
		DeclaredScopes: scopes,
		LaunchOrigins:  launchOrigins(a.Domain, a.WebsiteURL),
		UpdatedAt:      a.UpdatedAt,
	}
}

func (s *server) getAppClient(w http.ResponseWriter, r *http.Request) {
	a, err := s.fetchApp(r, r.PathValue("slug"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isPublicStatus(a.Status) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	writeJSON(w, http.StatusOK, toAppClientRecord(a))
}

func (s *server) listAppClientChanges(w http.ResponseWriter, r *http.Request) {
	sinceRaw := strings.TrimSpace(r.URL.Query().Get("since"))
	if sinceRaw == "" {
		writeError(w, http.StatusBadRequest, "since is required")
		return
	}
	since, err := time.Parse(time.RFC3339Nano, sinceRaw)
	if err != nil {
		since, err = time.Parse(time.RFC3339, sinceRaw)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "since must be an RFC3339 timestamp")
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT c.app_id::text, a.slug, c.reason, c.occurred_at
		FROM app_client_changes c
		JOIN apps a ON a.id = c.app_id
		WHERE c.occurred_at > $1
		ORDER BY c.occurred_at ASC`, since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	byApp := map[string]*AppClientChange{}
	order := []string{}
	for rows.Next() {
		var appID, slug, reason string
		var occurredAt time.Time
		if err := rows.Scan(&appID, &slug, &reason, &occurredAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item, ok := byApp[appID]
		if !ok {
			item = &AppClientChange{
				AppID:     appID,
				Slug:      slug,
				UpdatedAt: occurredAt,
				Reasons:   []string{},
			}
			byApp[appID] = item
			order = append(order, appID)
		}
		if !slices.Contains(item.Reasons, reason) {
			item.Reasons = append(item.Reasons, reason)
		}
		if occurredAt.After(item.UpdatedAt) {
			item.UpdatedAt = occurredAt
		}
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]AppClientChange, 0, len(order))
	for _, appID := range order {
		out = append(out, *byApp[appID])
	}
	writeJSON(w, http.StatusOK, map[string]any{"changes": out})
}

func recordClientChangeTx(ctx context.Context, tx pgx.Tx, appID, reason string) error {
	_, err := tx.Exec(ctx, `INSERT INTO app_client_changes (app_id, reason) VALUES ($1::uuid, $2)`, appID, reason)
	return err
}

func (s *server) recordClientChange(ctx context.Context, appID, reason string) {
	if appID == "" || reason == "" {
		return
	}
	_, _ = s.pool.Exec(ctx, `INSERT INTO app_client_changes (app_id, reason) VALUES ($1::uuid, $2)`, appID, reason)
}

func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left := append([]string(nil), a...)
	right := append([]string(nil), b...)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}

func (s *server) recordClientChangesFromAppUpdate(ctx context.Context, before, after App) {
	if before.Status != after.Status {
		s.recordClientChange(ctx, after.ID, clientChangeStatusChanged)
	}
	if !sameStringSet(before.DeclaredScopes, after.DeclaredScopes) {
		s.recordClientChange(ctx, after.ID, clientChangeScopesChanged)
	}
	if before.Name != after.Name ||
		!stringPtrEqual(before.IconURL, after.IconURL) ||
		before.Domain != after.Domain ||
		!stringPtrEqual(before.WebsiteURL, after.WebsiteURL) {
		s.recordClientChange(ctx, after.ID, clientChangeMetadataChanged)
	}
}

func (s *server) appIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id::text FROM apps WHERE slug=$1`, slug).Scan(&id)
	return id, err
}
