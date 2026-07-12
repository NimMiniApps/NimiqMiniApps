package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaItem struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type App struct {
	ID                   string       `json:"id"`
	Slug                 string       `json:"slug"`
	Name                 string       `json:"name"`
	Domain               string       `json:"domain"`
	Category             string       `json:"category"`
	DeveloperSlug        string       `json:"developer_slug"`
	DeveloperName        string       `json:"developer_name"`
	OwnerWalletAddresses []string     `json:"owner_wallet_addresses"`
	Tagline              string       `json:"tagline"`
	Description          string       `json:"description"`
	LongDescription      string       `json:"long_description"`
	Tags                 []string     `json:"tags"`
	Assets               []string     `json:"assets"`
	RewardAssets         []string     `json:"reward_assets"`
	Status               string       `json:"status"`
	ReleaseStage         string       `json:"release_stage"`
	Featured             bool         `json:"featured"`
	FeaturedOrder        int          `json:"featured_order"`
	WebsiteURL           *string      `json:"website_url"`
	GithubURL            *string      `json:"github_url"`
	IconURL              *string      `json:"icon_url"`
	DiscoveredIconURL    *string      `json:"discovered_icon_url"`
	BannerURL            *string      `json:"banner_url"`
	Media                []MediaItem  `json:"media"`
	Socials              []SocialLink `json:"socials"`
	DomainReachable      *bool        `json:"domain_reachable"`
	DomainCheckedAt      *time.Time   `json:"domain_checked_at"`
	AvgRating            float64      `json:"avg_rating"`
	ReviewCount          int          `json:"review_count"`
	SubmitterContact     string       `json:"submitter_contact,omitempty"`
	TotalOpens           int          `json:"total_opens,omitempty"`
	TotalViews           int          `json:"total_views,omitempty"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
	OpenURL              string       `json:"open_url"`
}

const appColumns = `id, slug, name, domain, category, developer_slug, developer_name, tagline,
	description, long_description, tags, assets, reward_assets, status, release_stage, featured, featured_order,
	website_url, github_url, icon_url, discovered_icon_url, banner_url, media, socials, domain_reachable, domain_checked_at,
	submitter_contact, created_at, updated_at,
	(ARRAY(SELECT wallet_address FROM app_owners WHERE app_owners.app_slug = apps.slug ORDER BY added_at)) AS owner_wallet_addresses,
	(SELECT COALESCE(AVG(rating), 0) FROM app_reviews WHERE app_reviews.app_id = apps.id) AS avg_rating,
	(SELECT COUNT(*) FROM app_reviews WHERE app_reviews.app_id = apps.id) AS review_count`

func stripPrivateAppFields(a *App) {
	a.SubmitterContact = ""
	a.TotalOpens, a.TotalViews = 0, 0
}

func scanApp(row pgx.Row) (App, error) {
	var a App
	var mediaJSON, socialsJSON []byte
	err := row.Scan(&a.ID, &a.Slug, &a.Name, &a.Domain, &a.Category, &a.DeveloperSlug,
		&a.DeveloperName, &a.Tagline, &a.Description, &a.LongDescription, &a.Tags, &a.Assets, &a.RewardAssets, &a.Status,
		&a.ReleaseStage, &a.Featured, &a.FeaturedOrder, &a.WebsiteURL, &a.GithubURL, &a.IconURL, &a.DiscoveredIconURL, &a.BannerURL,
		&mediaJSON, &socialsJSON, &a.DomainReachable, &a.DomainCheckedAt, &a.SubmitterContact, &a.CreatedAt, &a.UpdatedAt,
		&a.OwnerWalletAddresses, &a.AvgRating, &a.ReviewCount)
	if err != nil {
		return a, err
	}
	if len(mediaJSON) > 0 {
		if err := json.Unmarshal(mediaJSON, &a.Media); err != nil {
			return a, err
		}
	}
	if len(socialsJSON) > 0 {
		if err := json.Unmarshal(socialsJSON, &a.Socials); err != nil {
			return a, err
		}
	}
	if a.Media == nil {
		a.Media = []MediaItem{}
	}
	if a.Socials == nil {
		a.Socials = []SocialLink{}
	}
	if a.RewardAssets == nil {
		a.RewardAssets = []string{}
	}
	if a.OwnerWalletAddresses == nil {
		a.OwnerWalletAddresses = []string{}
	}
	a.OpenURL = "https://nimpay.app/miniapps/open/" + a.Domain
	return a, err
}

type server struct {
	pool             *pgxpool.Pool
	nonces           *nonceStore
	walletAuthSecret string
	adminToken       string
	adminWallets     map[string]struct{}
	reviewLimiter    *rateLimiter
	statsLimiter     *rateLimiter
}

// visibility filter for public endpoints
const publicStatuses = "status IN ('approved', 'verified', 'experimental')"

func isPublicStatus(status string) bool {
	return status == "approved" || status == "verified" || status == "experimental"
}

const featuredOrderSQL = "NULLIF(featured_order, 0) ASC NULLS LAST, created_at DESC"

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
		where = append(where, "(name ILIKE "+p+" OR tagline ILIKE "+p+" OR description ILIKE "+p+
			" OR long_description ILIKE "+p+" OR developer_name ILIKE "+p+
			" OR EXISTS (SELECT 1 FROM unnest(tags) t WHERE t ILIKE "+p+")"+
			" OR EXISTS (SELECT 1 FROM unnest(assets) a WHERE a ILIKE "+p+")"+
			" OR EXISTS (SELECT 1 FROM unnest(reward_assets) ra WHERE ra ILIKE "+p+"))")
	}
	if v := q.Get("tag"); v != "" {
		where = append(where, arg(v)+" = ANY(tags)")
	}
	if v := q.Get("asset"); v != "" {
		where = append(where, arg(v)+" = ANY(assets)")
	}
	if v := q.Get("rewards"); v == "true" {
		where = append(where, "array_length(reward_assets, 1) > 0")
	}
	if v := q.Get("collection"); v != "" {
		if !applyCollection(&where, arg, v) {
			writeError(w, http.StatusBadRequest, "unknown collection")
			return
		}
	}
	if v := q.Get("category"); v != "" {
		where = append(where, "category = "+arg(v))
	}
	if v := q.Get("developer"); v != "" {
		where = append(where, "developer_slug = "+arg(v))
	}
	if v := q.Get("status"); v != "" {
		where = append(where, "status = "+arg(v))
	} else {
		where = append(where, publicStatuses) // hide submitted/rejected unless a status is asked for explicitly
	}
	if v := q.Get("featured"); v == "true" {
		where = append(where, "featured = true")
	}
	whereSQL := where[0]
	for _, wc := range where[1:] {
		whereSQL += " AND " + wc
	}
	orderSQL := ""
	if q.Get("featured") == "true" {
		orderSQL = " ORDER BY " + featuredOrderSQL
	} else if q.Get("collection") == "new-week" {
		orderSQL = " ORDER BY created_at DESC"
	} else if q.Get("collection") == "popular" {
		orderSQL = ` ORDER BY (
			SELECT COALESCE(SUM(views), 0)
			FROM app_stats_daily
			WHERE app_id = apps.id AND day >= CURRENT_DATE - INTERVAL '6 days'
		) DESC, name ASC`
	} else {
		switch q.Get("sort") {
		case "newest":
			orderSQL = " ORDER BY created_at DESC"
		case "name":
			orderSQL = " ORDER BY name ASC"
		case "trending":
			orderSQL = ` ORDER BY (
				SELECT COALESCE(SUM(views), 0)
				FROM app_stats_daily
				WHERE app_id = apps.id AND day >= CURRENT_DATE - INTERVAL '6 days'
			) DESC, name ASC`
		default: // featured
			orderSQL = " ORDER BY featured DESC, " + featuredOrderSQL
		}
	}

	limit, offset, paginate, errMsg := parsePagination(q)
	if errMsg != "" {
		writeError(w, http.StatusBadRequest, errMsg)
		return
	}

	var total int
	if paginate {
		countSQL := "SELECT count(*) FROM apps WHERE " + whereSQL
		if err := s.pool.QueryRow(r.Context(), countSQL, args...).Scan(&total); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	sql := "SELECT " + appColumns + " FROM apps WHERE " + whereSQL + orderSQL
	if paginate && limit > 0 {
		sql += " LIMIT " + strconv.Itoa(limit)
		if offset > 0 {
			sql += " OFFSET " + strconv.Itoa(offset)
		}
	} else if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > maxPageLimit {
				n = maxPageLimit
			}
			sql += " LIMIT " + strconv.Itoa(n)
		}
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
		stripPrivateAppFields(&a)
		apps = append(apps, a)
	}
	if paginate {
		writeJSON(w, http.StatusOK, paginatedApps{
			Items:  apps,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		})
		return
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
	if !isPublicStatus(a.Status) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	stripPrivateAppFields(&a)
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
	counts := map[string]int{}
	for rows.Next() {
		var c cat
		if err := rows.Scan(&c.Name, &c.Count); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		counts[c.Name] = c.Count
	}
	cats := make([]cat, 0, len(validCategories))
	for _, name := range validCategories {
		cats = append(cats, cat{Name: name, Count: counts[name]})
	}
	writeJSON(w, http.StatusOK, cats)
}

func (s *server) getDeveloper(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	rows, err := s.pool.Query(r.Context(),
		"SELECT "+appColumns+" FROM apps WHERE developer_slug=$1 AND "+publicStatuses+" ORDER BY featured DESC, "+featuredOrderSQL, slug)
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
		stripPrivateAppFields(&a)
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

func (s *server) listDevelopers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT developer_slug, MAX(developer_name) AS developer_name, COUNT(*) AS app_count
		FROM apps WHERE `+publicStatuses+`
		GROUP BY developer_slug
		ORDER BY developer_name ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	type dev struct {
		Slug     string `json:"slug"`
		Name     string `json:"name"`
		AppCount int    `json:"app_count"`
	}
	developers := []dev{}
	for rows.Next() {
		var d dev
		if err := rows.Scan(&d.Slug, &d.Name, &d.AppCount); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		developers = append(developers, d)
	}
	writeJSON(w, http.StatusOK, developers)
}

func (s *server) getRelatedApps(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	current, err := s.fetchApp(r, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !isPublicStatus(current.Status) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT `+appColumns+` FROM apps
		WHERE slug != $1 AND `+publicStatuses+`
		  AND (category = $2 OR developer_slug = $3)
		ORDER BY (developer_slug = $3) DESC, featured DESC, `+featuredOrderSQL+`
		LIMIT 4`,
		slug, current.Category, current.DeveloperSlug)
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
		stripPrivateAppFields(&a)
		apps = append(apps, a)
	}
	writeJSON(w, http.StatusOK, apps)
}

// adminListApps returns every app regardless of status.
func (s *server) adminListApps(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT `+appColumns+`, COALESCE(stats.total_opens, 0), COALESCE(stats.total_views, 0)
		FROM apps
		LEFT JOIN LATERAL (
			SELECT SUM(opens) AS total_opens, SUM(views) AS total_views
			FROM app_stats_daily WHERE app_id = apps.id
		) stats ON true
		ORDER BY featured DESC, `+featuredOrderSQL+`, name ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	apps := []App{}
	for rows.Next() {
		a, err := scanAdminApp(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		apps = append(apps, a)
	}
	writeJSON(w, http.StatusOK, apps)
}

func scanAdminApp(row pgx.Row) (App, error) {
	var a App
	var mediaJSON, socialsJSON []byte
	err := row.Scan(&a.ID, &a.Slug, &a.Name, &a.Domain, &a.Category, &a.DeveloperSlug,
		&a.DeveloperName, &a.Tagline, &a.Description, &a.LongDescription, &a.Tags, &a.Assets, &a.RewardAssets, &a.Status,
		&a.ReleaseStage, &a.Featured, &a.FeaturedOrder, &a.WebsiteURL, &a.GithubURL, &a.IconURL, &a.DiscoveredIconURL, &a.BannerURL,
		&mediaJSON, &socialsJSON, &a.DomainReachable, &a.DomainCheckedAt, &a.SubmitterContact, &a.CreatedAt, &a.UpdatedAt,
		&a.OwnerWalletAddresses, &a.AvgRating, &a.ReviewCount, &a.TotalOpens, &a.TotalViews)
	if err != nil {
		return a, err
	}
	if len(mediaJSON) > 0 {
		if err := json.Unmarshal(mediaJSON, &a.Media); err != nil {
			return a, err
		}
	}
	if len(socialsJSON) > 0 {
		if err := json.Unmarshal(socialsJSON, &a.Socials); err != nil {
			return a, err
		}
	}
	if a.Media == nil {
		a.Media = []MediaItem{}
	}
	if a.Socials == nil {
		a.Socials = []SocialLink{}
	}
	if a.RewardAssets == nil {
		a.RewardAssets = []string{}
	}
	if a.OwnerWalletAddresses == nil {
		a.OwnerWalletAddresses = []string{}
	}
	a.OpenURL = "https://nimpay.app/miniapps/open/" + a.Domain
	return a, err
}

func (s *server) createApp(w http.ResponseWriter, r *http.Request) {
	s.decodeAndInsert(w, r, nil, false, nil)
}

// decodeAndInsert parses an app from the request, optionally forces fields
// (used by public submissions), validates, and inserts it.
func (s *server) decodeAndInsert(w http.ResponseWriter, r *http.Request, force func(*App), requireContact bool, ownerAddress *string) {
	var a App
	a.Tags, a.Assets, a.RewardAssets, a.Media = []string{}, []string{}, []string{}, []MediaItem{}
	a.Status = "submitted"
	a.ReleaseStage = "released"
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	a.SubmitterContact = strings.TrimSpace(a.SubmitterContact)
	if force != nil {
		force(&a)
	}
	if requireContact {
		if msg := validateSubmitterContact(a.SubmitterContact); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
	}
	if a.Media == nil {
		a.Media = []MediaItem{}
	}
	if a.Socials == nil {
		a.Socials = []SocialLink{}
	}
	if err := validateApp(&a); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	mediaJSON, err := json.Marshal(a.Media)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	socialsJSON, err := json.Marshal(a.Socials)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	insertSQL := `
		INSERT INTO apps (slug, name, domain, category, developer_slug, developer_name, tagline,
			description, long_description, tags, assets, reward_assets, status, release_stage, featured, featured_order,
			website_url, github_url, icon_url, banner_url, media, socials, submitter_contact)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)
		RETURNING ` + appColumns
	insertArgs := []any{
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.LongDescription, a.Tags, a.Assets, a.RewardAssets, a.Status, a.ReleaseStage, a.Featured, a.FeaturedOrder,
		a.WebsiteURL, a.GithubURL, a.IconURL, a.BannerURL, mediaJSON, socialsJSON, a.SubmitterContact,
	}

	var pgErr *pgconn.PgError
	if ownerAddress != nil {
		tx, err2 := s.pool.Begin(r.Context())
		if err2 != nil {
			writeError(w, http.StatusInternalServerError, err2.Error())
			return
		}
		defer tx.Rollback(r.Context())
		a, err = scanApp(tx.QueryRow(r.Context(), insertSQL, insertArgs...))
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if _, err = tx.Exec(r.Context(),
			`INSERT INTO app_owners (app_slug, wallet_address) VALUES ($1,$2)`, a.Slug, *ownerAddress); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err = tx.Commit(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.OwnerWalletAddresses = []string{*ownerAddress}
	} else {
		a, err = scanApp(s.pool.QueryRow(r.Context(), insertSQL, insertArgs...))
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if a.Status == "submitted" {
		notifySubmission(a)
	}
	if needsIconDiscovery(a.IconURL, a.DiscoveredIconURL) {
		go s.tryDiscoverAppIcon(context.Background(), a.Slug, a.Domain)
	}
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
	domainReachable := a.DomainReachable
	domainCheckedAt := a.DomainCheckedAt
	originalDomain := a.Domain
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if a.Domain != originalDomain {
		a.DomainReachable = nil
		a.DomainCheckedAt = nil
	} else {
		a.DomainReachable = domainReachable
		a.DomainCheckedAt = domainCheckedAt
	}
	if a.Media == nil {
		a.Media = []MediaItem{}
	}
	if a.Socials == nil {
		a.Socials = []SocialLink{}
	}
	if err := validateApp(&a); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	mediaJSON, err := json.Marshal(a.Media)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	socialsJSON, err := json.Marshal(a.Socials)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a, err = scanApp(s.pool.QueryRow(r.Context(), `
		UPDATE apps SET slug=$1, name=$2, domain=$3, category=$4, developer_slug=$5,
			developer_name=$6, tagline=$7, description=$8, long_description=$9, tags=$10, assets=$11, reward_assets=$12,
			status=$13, release_stage=$14, featured=$15, featured_order=$16, website_url=$17, github_url=$18,
			icon_url=$19, banner_url=$20, media=$21, socials=$22, submitter_contact=$23, updated_at=now()
		WHERE id=$24
		RETURNING `+appColumns,
		a.Slug, a.Name, a.Domain, a.Category, a.DeveloperSlug, a.DeveloperName, a.Tagline,
		a.Description, a.LongDescription, a.Tags, a.Assets, a.RewardAssets, a.Status, a.ReleaseStage, a.Featured, a.FeaturedOrder,
		a.WebsiteURL, a.GithubURL, a.IconURL, a.BannerURL, mediaJSON, socialsJSON, a.SubmitterContact, a.ID))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		writeError(w, http.StatusConflict, "slug already exists")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
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

