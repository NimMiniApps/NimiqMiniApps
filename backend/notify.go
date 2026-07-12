package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func notifySubmission(app App) {
	notifyWebhook("app.submitted", map[string]any{
		"submitted_at": time.Now().UTC().Format(time.RFC3339),
		"app":          appSummary(app),
	})
}

func notifyRejected(app App, note string) {
	payload := map[string]any{
		"rejected_at": time.Now().UTC().Format(time.RFC3339),
		"app":         appSummary(app),
	}
	if note != "" {
		payload["note"] = note
	}
	notifyWebhook("app.rejected", payload)
}

func notifyUpdateRequest(current App, rev AppRevision) {
	notifyWebhook("app.update_requested", map[string]any{
		"requested_at": time.Now().UTC().Format(time.RFC3339),
		"app_slug":     current.Slug,
		"revision_id":  rev.ID,
		"author_note":  rev.AuthorNote,
		"current":      appSummary(current),
		"proposed": map[string]any{
			"name":    rev.Name,
			"domain":  rev.Domain,
			"tagline": rev.Tagline,
		},
	})
}

func appSummary(app App) map[string]any {
	return map[string]any{
		"slug":              app.Slug,
		"name":              app.Name,
		"domain":            app.Domain,
		"category":          app.Category,
		"developer_name":    app.DeveloperName,
		"tagline":           app.Tagline,
		"submitter_contact": app.SubmitterContact,
	}
}

func notifyWebhook(event string, data map[string]any) {
	url := os.Getenv("SUBMIT_WEBHOOK_URL")
	if url == "" {
		return
	}
	body := map[string]any{"event": event}
	for k, v := range data {
		body[k] = v
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			slog.Warn("submit webhook: build request failed", "error", err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.Warn("submit webhook: post failed", "error", err.Error())
			return
		}
		res.Body.Close()
		if res.StatusCode >= 300 {
			slog.Warn("submit webhook: unexpected status", "status", res.StatusCode)
		}
	}()
}
