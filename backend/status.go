package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *server) adminStats(w http.ResponseWriter, r *http.Request) {
	var pending, unreachable, pendingUpdates int
	err := s.pool.QueryRow(r.Context(), `
		SELECT
			count(*) FILTER (WHERE status = 'submitted'),
			count(*) FILTER (WHERE domain_reachable = false),
			(SELECT count(*) FROM app_revisions WHERE status = 'pending')
		FROM apps`).Scan(&pending, &unreachable, &pendingUpdates)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{
		"pending":         pending,
		"unreachable":     unreachable,
		"pending_updates": pendingUpdates,
	})
}

func publicStatusLabel(status string) string {
	switch status {
	case "submitted":
		return "pending"
	case "approved", "verified", "experimental":
		return "live"
	case "rejected":
		return "rejected"
	default:
		return status
	}
}

func (s *server) getSubmissionStatus(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var name, status string
	var rejectionNote *string
	var updatedAt time.Time
	err := s.pool.QueryRow(r.Context(),
		`SELECT name, status, rejection_note, updated_at FROM apps WHERE slug=$1`, slug).
		Scan(&name, &status, &rejectionNote, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updatePending, _ := s.hasPendingRevision(r.Context(), slug)
	resp := map[string]any{
		"slug":            slug,
		"name":            name,
		"status":          publicStatusLabel(status),
		"raw_status":      status,
		"public":          isPublicStatus(status),
		"updated_at":      updatedAt,
		"update_pending":  updatePending,
	}
	if rejectionNote != nil && strings.TrimSpace(*rejectionNote) != "" {
		resp["rejection_note"] = strings.TrimSpace(*rejectionNote)
	}
	writeJSON(w, http.StatusOK, resp)
}
