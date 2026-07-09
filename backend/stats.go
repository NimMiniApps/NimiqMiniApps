package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errUnknownTrackEvent = errors.New("unknown track event")

func trackEvent(ctx context.Context, pool *pgxpool.Pool, appID, event string) error {
	day := time.Now().UTC().Format("2006-01-02")
	switch event {
	case "open":
		_, err := pool.Exec(ctx, `
			INSERT INTO app_stats_daily (app_id, day, opens, views)
			VALUES ($1, $2::date, 1, 0)
			ON CONFLICT (app_id, day) DO UPDATE SET opens = app_stats_daily.opens + 1`,
			appID, day)
		return err
	case "view":
		_, err := pool.Exec(ctx, `
			INSERT INTO app_stats_daily (app_id, day, opens, views)
			VALUES ($1, $2::date, 0, 1)
			ON CONFLICT (app_id, day) DO UPDATE SET views = app_stats_daily.views + 1`,
			appID, day)
		return err
	default:
		return errUnknownTrackEvent
	}
}

type trackRequest struct {
	Event string `json:"event"`
}

func (s *server) trackApp(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !s.statsLimiter.allow(clientIP(r) + ":" + slug) {
		writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	var body trackRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	event := body.Event
	if event != "open" && event != "view" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	appID, err := s.appIDForSlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = trackEvent(r.Context(), s.pool, appID, event)
	w.WriteHeader(http.StatusNoContent)
}

type statsTotals struct {
	Opens int `json:"opens"`
	Views int `json:"views"`
}

type statsDailyRow struct {
	Date  string `json:"date"`
	Opens int    `json:"opens"`
	Views int    `json:"views"`
}

type appStatsResponse struct {
	Totals statsTotals     `json:"totals"`
	Daily  []statsDailyRow `json:"daily"`
}

func (s *server) appStats(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	appID, err := s.appIDForSlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "app not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT day, opens, views
		FROM app_stats_daily
		WHERE app_id = $1 AND day >= (CURRENT_DATE - INTERVAL '29 days')
		ORDER BY day ASC`, appID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	resp := appStatsResponse{Daily: []statsDailyRow{}}
	for rows.Next() {
		var day time.Time
		var row statsDailyRow
		if err := rows.Scan(&day, &row.Opens, &row.Views); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		row.Date = day.Format("2006-01-02")
		resp.Totals.Opens += row.Opens
		resp.Totals.Views += row.Views
		resp.Daily = append(resp.Daily, row)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
