package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	ID            string    `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Domain        string    `json:"domain"`
	Category      string    `json:"category"`
	DeveloperSlug string    `json:"developer_slug"`
	DeveloperName string    `json:"developer_name"`
	Tagline       string    `json:"tagline"`
	Description   string    `json:"description"`
	Tags          []string  `json:"tags"`
	Assets        []string  `json:"assets"`
	Status        string    `json:"status"`
	Featured      bool      `json:"featured"`
	WebsiteURL    *string   `json:"website_url"`
	GithubURL     *string   `json:"github_url"`
	IconURL       *string   `json:"icon_url"`
	BannerURL     *string   `json:"banner_url"`
	Screenshots   []string  `json:"screenshots"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	OpenURL       string    `json:"open_url"`
}

const appColumns = `id, slug, name, domain, category, developer_slug, developer_name, tagline,
	description, tags, assets, status, featured, website_url, github_url, icon_url, banner_url,
	screenshots, created_at, updated_at`

func scanApp(row pgx.Row) (App, error) {
	var a App
	err := row.Scan(&a.ID, &a.Slug, &a.Name, &a.Domain, &a.Category, &a.DeveloperSlug,
		&a.DeveloperName, &a.Tagline, &a.Description, &a.Tags, &a.Assets, &a.Status,
		&a.Featured, &a.WebsiteURL, &a.GithubURL, &a.IconURL, &a.BannerURL,
		&a.Screenshots, &a.CreatedAt, &a.UpdatedAt)
	a.OpenURL = "https://nimpay.app/miniapps/open/" + a.Domain
	return a, err
}

type server struct {
	pool *pgxpool.Pool
}

// visibility filter for public endpoints
const publicStatuses = "status IN ('approved', 'verified', 'experimental')"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database unreachable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) listApps(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	where := []string{}
	args := []any{}
	arg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}
	if v := q.Get("q"); v != "" {
		p := arg("%" + v + "%")
		where = append(where, "(name ILIKE "+p+" OR tagline ILIKE "+p+" OR description ILIKE "+p+")")
	}
	if v := q.Get("category"); v != "" {
		where = append(where, "category = "+arg(v))
	}
	if v := q.Get("status"); v != "" {
		where = append(where, "status = "+arg(v))
	} else {
		where = append(where, publicStatuses) // hide submitted/rejected unless a status is asked for explicitly
	}
	if v := q.Get("featured"); v == "true" {
		where = append(where, "featured = true")
	}
	sql := "SELECT " + appColumns + " FROM apps WHERE " + where[0]
	for _, wc := range where[1:] {
		sql += " AND " + wc
	}
	switch q.Get("sort") {
	case "newest":
		sql += " ORDER BY created_at DESC"
	case "name":
		sql += " ORDER BY name ASC"
	default: // featured
		sql += " ORDER BY featured DESC, created_at DESC"
	}
	rows, err := s.pool.Query(r.Context(), sql, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	apps := []App{}
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		apps = append(apps, a)
	}
	writeJSON(w, http.StatusOK, apps)
}

func (s *server) fetchApp(r *http.Request, slug string) (App, error) {
	return scanApp(s.pool.QueryRow(r.Context(), "SELECT "+appColumns+" FROM apps WHERE slug=$1", slug))
}

func (s *server) getApp(w http.ResponseWriter, r *http.Request) {
	a, err := s.fetchApp(r, r.PathValue("slug"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *server) listCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(),
		`SELECT category, count(*) FROM apps WHERE `+publicStatuses+` GROUP BY category ORDER BY category`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type cat struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	cats := []cat{}
	for rows.Next() {
		var c cat
		if err := rows.Scan(&c.Name, &c.Count); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		cats = append(cats, c)
	}
	writeJSON(w, http.StatusOK, cats)
}

func (s *server) getDeveloper(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	rows, err := s.pool.Query(r.Context(),
		"SELECT "+appColumns+" FROM apps WHERE developer_slug=$1 AND "+publicStatuses+" ORDER BY featured DESC, created_at DESC", slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	apps := []App{}
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		apps = append(apps, a)
	}
	if len(apps) == 0 {
		writeError(w, http.StatusNotFound, "developer not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"slug": slug,
		"name": apps[0].DeveloperName,
		"apps": apps,
	})
}

// adminListApps returns every app regardless of status.
func (s *server) adminListApps(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), "SELECT "+appColumns+" FROM apps ORDER BY name ASC")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	apps := []App{}
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		apps = append(apps, a)
	}
	writeJSON(w, http.StatusOK, apps)
}

func (s *server) createApp(w http.ResponseWriter, r *http.Request) {
	s.decodeAndInsert(w, r, nil)
}

// decodeAndInsert parses an app from the request, optionally forces fields
// (used by public submissions), validates, and inserts it.
func (s *server) decodeAndInsert(w http.ResponseWriter, r *http.Request, force func(*App)) {
	var a App
	a.Tags, a.Assets, a.Screenshots = []string{}, []string{}, []string{}
	a.Status = "submitted"
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if force != nil {
		force(&a)
	}
	if err := validateApp(&a); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := s.pool.QueryRow(r.Context(), `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline,
			description, tags, assets, status, featured, website_url, github_url, icon_url,
			banner_url, screenshots)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING `+appColumns,
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.Tags, a.Assets, a.Status, a.Featured, a.WebsiteURL, a.GithubURL,
		a.IconURL, a.BannerURL, a.Screenshots).Scan(
		&a.ID, &a.Slug, &a.Name, &a.Domain, &a.Category, &a.DeveloperSlug, &a.DeveloperName,
		&a.Tagline, &a.Description, &a.Tags, &a.Assets, &a.Status, &a.Featured, &a.WebsiteURL,
		&a.GithubURL, &a.IconURL, &a.BannerURL, &a.Screenshots, &a.CreatedAt, &a.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		writeError(w, http.StatusConflict, "slug already exists")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.OpenURL = "https://nimpay.app/miniapps/open/" + a.Domain
	writeJSON(w, http.StatusCreated, a)
}

// updateApp serves both PUT and PATCH: load existing, overlay request JSON, validate, save.
// ponytail: merge semantics for both verbs; strict PUT replacement isn't worth a second code path.
func (s *server) updateApp(w http.ResponseWriter, r *http.Request) {
	a, err := s.fetchApp(r, r.PathValue("slug"))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := validateApp(&a); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = s.pool.QueryRow(r.Context(), `
		UPDATE apps SET slug=$1, name=$2, domain=$3, category=$4, developer_slug=$5,
			developer_name=$6, tagline=$7, description=$8, tags=$9, assets=$10, status=$11,
			featured=$12, website_url=$13, github_url=$14, icon_url=$15, banner_url=$16,
			screenshots=$17, updated_at=now()
		WHERE id=$18
		RETURNING `+appColumns,
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.Tags, a.Assets, a.Status, a.Featured, a.WebsiteURL, a.GithubURL,
		a.IconURL, a.BannerURL, a.Screenshots, a.ID).Scan(
		&a.ID, &a.Slug, &a.Name, &a.Domain, &a.Category, &a.DeveloperSlug, &a.DeveloperName,
		&a.Tagline, &a.Description, &a.Tags, &a.Assets, &a.Status, &a.Featured, &a.WebsiteURL,
		&a.GithubURL, &a.IconURL, &a.BannerURL, &a.Screenshots, &a.CreatedAt, &a.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		writeError(w, http.StatusConflict, "slug already exists")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.OpenURL = "https://nimpay.app/miniapps/open/" + a.Domain
	writeJSON(w, http.StatusOK, a)
}

func (s *server) deleteApp(w http.ResponseWriter, r *http.Request) {
	tag, err := s.pool.Exec(r.Context(), `DELETE FROM apps WHERE slug=$1`, r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) setStatus(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a, err := scanApp(s.pool.QueryRow(r.Context(),
			`UPDATE apps SET status=$1, updated_at=now() WHERE slug=$2 RETURNING `+appColumns,
			status, r.PathValue("slug")))
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "app not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, a)
	}
}
