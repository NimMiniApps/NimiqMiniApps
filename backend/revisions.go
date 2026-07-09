package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type AppRevision struct {
	ID              string       `json:"id"`
	AppSlug         string       `json:"app_slug"`
	Status          string       `json:"status"`
	Name            string       `json:"name"`
	Domain          string       `json:"domain"`
	Category        string       `json:"category"`
	DeveloperSlug   string       `json:"developer_slug"`
	DeveloperName   string       `json:"developer_name"`
	Tagline         string       `json:"tagline"`
	Description     string       `json:"description"`
	LongDescription string       `json:"long_description"`
	Tags            []string     `json:"tags"`
	Assets          []string     `json:"assets"`
	ReleaseStage    string       `json:"release_stage"`
	WebsiteURL      *string      `json:"website_url"`
	GithubURL       *string      `json:"github_url"`
	IconURL         *string      `json:"icon_url"`
	BannerURL       *string      `json:"banner_url"`
	Media           []MediaItem  `json:"media"`
	Socials         []SocialLink `json:"socials"`
	AuthorNote      string       `json:"author_note"`
	CreatedAt       time.Time    `json:"created_at"`
	ReviewedAt      *time.Time   `json:"reviewed_at"`
}

const revisionColumns = `id, app_slug, status, name, domain, category, developer_slug, developer_name,
	tagline, description, long_description, tags, assets, release_stage,
	website_url, github_url, icon_url, banner_url, media, socials, author_note, created_at, reviewed_at`

func scanRevision(row pgx.Row) (AppRevision, error) {
	var rev AppRevision
	var mediaJSON, socialsJSON []byte
	err := row.Scan(&rev.ID, &rev.AppSlug, &rev.Status, &rev.Name, &rev.Domain, &rev.Category,
		&rev.DeveloperSlug, &rev.DeveloperName, &rev.Tagline, &rev.Description, &rev.LongDescription,
		&rev.Tags, &rev.Assets, &rev.ReleaseStage, &rev.WebsiteURL, &rev.GithubURL, &rev.IconURL,
		&rev.BannerURL, &mediaJSON, &socialsJSON, &rev.AuthorNote, &rev.CreatedAt, &rev.ReviewedAt)
	if err != nil {
		return rev, err
	}
	if len(mediaJSON) > 0 {
		if err := json.Unmarshal(mediaJSON, &rev.Media); err != nil {
			return rev, err
		}
	}
	if len(socialsJSON) > 0 {
		if err := json.Unmarshal(socialsJSON, &rev.Socials); err != nil {
			return rev, err
		}
	}
	if rev.Media == nil {
		rev.Media = []MediaItem{}
	}
	if rev.Socials == nil {
		rev.Socials = []SocialLink{}
	}
	return rev, nil
}

func revisionToApp(rev AppRevision, keep App) App {
	a := keep
	a.Name = rev.Name
	a.Domain = rev.Domain
	a.Category = rev.Category
	a.DeveloperSlug = rev.DeveloperSlug
	a.DeveloperName = rev.DeveloperName
	a.Tagline = rev.Tagline
	a.Description = rev.Description
	a.LongDescription = rev.LongDescription
	a.Tags = rev.Tags
	a.Assets = rev.Assets
	a.ReleaseStage = rev.ReleaseStage
	a.WebsiteURL = rev.WebsiteURL
	a.GithubURL = rev.GithubURL
	a.IconURL = rev.IconURL
	a.BannerURL = rev.BannerURL
	a.Media = rev.Media
	a.Socials = rev.Socials
	return a
}

type updateRequestBody struct {
	App
	AuthorNote string `json:"author_note"`
}

func (s *server) requestAppUpdate(w http.ResponseWriter, r *http.Request, address string) {
	slug := r.PathValue("slug")
	if !allowSubmit(clientIP(r), time.Now()) {
		writeError(w, http.StatusTooManyRequests, "too many requests, try again later")
		return
	}

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
	if current.DeveloperWalletAddress == nil || *current.DeveloperWalletAddress != address {
		writeError(w, http.StatusForbidden, "you don't own this app")
		return
	}

	var body updateRequestBody
	body.Tags, body.Assets, body.Media = []string{}, []string{}, []MediaItem{}
	body.Socials = []SocialLink{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Slug != "" && body.Slug != slug {
		writeError(w, http.StatusBadRequest, "slug cannot be changed in an update request")
		return
	}
	body.Slug = slug
	body.Status = current.Status
	body.Featured = current.Featured
	body.FeaturedOrder = current.FeaturedOrder
	body.DeveloperSlug = current.DeveloperSlug
	body.DeveloperName = current.DeveloperName
	if body.Media == nil {
		body.Media = []MediaItem{}
	}
	if body.Socials == nil {
		body.Socials = []SocialLink{}
	}
	if err := validateApp(&body.App); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var pending bool
	if err := s.pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM app_revisions WHERE app_slug=$1 AND status='pending')`, slug).
		Scan(&pending); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if pending {
		writeError(w, http.StatusConflict, "an update request is already pending for this app")
		return
	}

	mediaJSON, err := json.Marshal(body.Media)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	socialsJSON, err := json.Marshal(body.Socials)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rev, err := scanRevision(s.pool.QueryRow(r.Context(), `
		INSERT INTO app_revisions (
			app_slug, name, domain, category, developer_slug, developer_name, tagline,
			description, long_description, tags, assets, release_stage,
			website_url, github_url, icon_url, banner_url, media, socials, author_note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING `+revisionColumns,
		slug, body.Name, body.Domain, body.Category, current.DeveloperSlug, current.DeveloperName,
		body.Tagline, body.Description, body.LongDescription, body.Tags, body.Assets, body.ReleaseStage,
		body.WebsiteURL, body.GithubURL, body.IconURL, body.BannerURL, mediaJSON, socialsJSON, body.AuthorNote))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	notifyUpdateRequest(current, rev)
	writeJSON(w, http.StatusCreated, map[string]any{
		"revision_id": rev.ID,
		"app_slug":    slug,
		"status":      "pending",
	})
}

func (s *server) adminListRevisions(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT `+revisionColumns+` FROM app_revisions WHERE status=$1 ORDER BY created_at ASC`, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type item struct {
		Revision AppRevision `json:"revision"`
		Current  App         `json:"current"`
	}
	items := []item{}
	for rows.Next() {
		rev, err := scanRevision(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		current, err := s.fetchApp(r, rev.AppSlug)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, item{Revision: rev, Current: current})
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *server) approveRevision(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback(ctx)

	rev, err := scanRevision(tx.QueryRow(ctx,
		`SELECT `+revisionColumns+` FROM app_revisions WHERE id=$1 AND status='pending' FOR UPDATE`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "pending revision not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	current, err := scanApp(tx.QueryRow(ctx, "SELECT "+appColumns+" FROM apps WHERE slug=$1 FOR UPDATE", rev.AppSlug))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated := revisionToApp(rev, current)
	if updated.Domain != current.Domain {
		updated.DomainReachable = nil
		updated.DomainCheckedAt = nil
	}
	if err := validateApp(&updated); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	mediaJSON, err := json.Marshal(updated.Media)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	socialsJSON, err := json.Marshal(updated.Socials)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	app, err := scanApp(tx.QueryRow(ctx, `
		UPDATE apps SET name=$1, domain=$2, category=$3, developer_slug=$4, developer_name=$5,
			tagline=$6, description=$7, long_description=$8, tags=$9, assets=$10, release_stage=$11,
			website_url=$12, github_url=$13, icon_url=$14, banner_url=$15, media=$16, socials=$17,
			domain_reachable=$18, domain_checked_at=$19, updated_at=now()
		WHERE id=$20
		RETURNING `+appColumns,
		updated.Name, updated.Domain, updated.Category, updated.DeveloperSlug, updated.DeveloperName,
		updated.Tagline, updated.Description, updated.LongDescription, updated.Tags, updated.Assets,
		updated.ReleaseStage, updated.WebsiteURL, updated.GithubURL, updated.IconURL, updated.BannerURL,
		mediaJSON, socialsJSON, updated.DomainReachable, updated.DomainCheckedAt, updated.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if _, err := tx.Exec(ctx,
		`UPDATE app_revisions SET status='approved', reviewed_at=$1 WHERE id=$2`, now, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *server) rejectRevision(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tag, err := s.pool.Exec(r.Context(),
		`UPDATE app_revisions SET status='rejected', reviewed_at=now() WHERE id=$1 AND status='pending'`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "pending revision not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (s *server) hasPendingRevision(ctx context.Context, slug string) (bool, error) {
	var pending bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM app_revisions WHERE app_slug=$1 AND status='pending')`, slug).
		Scan(&pending)
	return pending, err
}
