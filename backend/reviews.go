package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Review struct {
	ID            string    `json:"id"`
	AppID         string    `json:"app_id"`
	WalletAddress string    `json:"wallet_address"`
	DisplayName   *string   `json:"display_name"`
	Rating        int       `json:"rating"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func validateReviewInput(rating int, body string) string {
	if rating < 1 || rating > 5 {
		return "rating must be between 1 and 5"
	}
	if len(body) > 1000 {
		return "body must be at most 1000 characters"
	}
	return ""
}

// rateLimiter is a per-key fixed-window limiter (in-memory; single backend instance).
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{attempts: map[string][]time.Time{}, limit: limit, window: window}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-rl.window)
	kept := rl.attempts[key][:0]
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.limit {
		rl.attempts[key] = kept
		return false
	}
	rl.attempts[key] = append(kept, now)
	return true
}

func (s *server) appIDForSlug(ctx context.Context, slug string) (string, error) {
	var id string
	err := s.pool.QueryRow(ctx, `SELECT id FROM apps WHERE slug=$1`, slug).Scan(&id)
	return id, err
}

func (s *server) listReviews(w http.ResponseWriter, r *http.Request) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT ar.id, ar.app_id, ar.wallet_address, u.display_name, ar.rating, ar.body, ar.created_at, ar.updated_at
		FROM app_reviews ar
		LEFT JOIN users u ON u.wallet_address = ar.wallet_address
		WHERE ar.app_id=$1 ORDER BY ar.created_at DESC`, appID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	reviews := []Review{}
	for rows.Next() {
		var rv Review
		if err := rows.Scan(&rv.ID, &rv.AppID, &rv.WalletAddress, &rv.DisplayName, &rv.Rating, &rv.Body, &rv.CreatedAt, &rv.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		reviews = append(reviews, rv)
	}
	var average float64
	if len(reviews) > 0 {
		var sum int
		for _, rv := range reviews {
			sum += rv.Rating
		}
		average = float64(sum) / float64(len(reviews))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":   reviews,
		"average": average,
		"count":   len(reviews),
	})
}

func (s *server) upsertReview(w http.ResponseWriter, r *http.Request, address string) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if !s.reviewLimiter.allow(address) {
		writeError(w, http.StatusTooManyRequests, "too many reviews, try again later")
		return
	}
	var req struct {
		Rating int    `json:"rating"`
		Body   string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	body := strings.TrimSpace(req.Body)
	if msg := validateReviewInput(req.Rating, body); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	var rv Review
	err = s.pool.QueryRow(r.Context(), `
		INSERT INTO app_reviews (app_id, wallet_address, rating, body)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (app_id, wallet_address)
		DO UPDATE SET rating=$3, body=$4, updated_at=now()
		RETURNING id, app_id, wallet_address, rating, body, created_at, updated_at`,
		appID, address, req.Rating, body).
		Scan(&rv.ID, &rv.AppID, &rv.WalletAddress, &rv.Rating, &rv.Body, &rv.CreatedAt, &rv.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rv)
}

func (s *server) deleteOwnReview(w http.ResponseWriter, r *http.Request, address string) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	tag, err := s.pool.Exec(r.Context(), `DELETE FROM app_reviews WHERE app_id=$1 AND wallet_address=$2`, appID, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) adminDeleteReview(w http.ResponseWriter, r *http.Request) {
	tag, err := s.pool.Exec(r.Context(), `DELETE FROM app_reviews WHERE id=$1`, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "review not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
