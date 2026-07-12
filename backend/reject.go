package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *server) rejectApp(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	var body struct {
		Note string `json:"note"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
	}
	note := strings.TrimSpace(body.Note)
	if len(note) > 2000 {
		writeError(w, http.StatusBadRequest, "note must be at most 2000 characters")
		return
	}
	var noteArg any
	if note != "" {
		noteArg = note
	}
	a, err := scanApp(s.pool.QueryRow(r.Context(),
		`UPDATE apps SET status='rejected', rejection_note=$1, updated_at=now() WHERE slug=$2 RETURNING `+appColumns,
		noteArg, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	notifyRejected(a, note)
	writeJSON(w, http.StatusOK, a)
}

func (s *server) setStatus(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearNote := status == "approved" || status == "verified" || status == "experimental"
		var query string
		var args []any
		if clearNote {
			query = `UPDATE apps SET status=$1, rejection_note=NULL, updated_at=now() WHERE slug=$2 RETURNING ` + appColumns
			args = []any{status, r.PathValue("slug")}
		} else {
			query = `UPDATE apps SET status=$1, updated_at=now() WHERE slug=$2 RETURNING ` + appColumns
			args = []any{status, r.PathValue("slug")}
		}
		a, err := scanApp(s.pool.QueryRow(r.Context(), query, args...))
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
