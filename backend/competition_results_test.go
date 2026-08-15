package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateCompetitionResults(t *testing.T) {
	place := 0
	if err := validateCompetitionResults([]CompetitionResult{{Cycle: 0, Score: 1}}); err == nil {
		t.Fatal("expected error for cycle 0")
	}
	if err := validateCompetitionResults([]CompetitionResult{{Cycle: 1, Score: -1}}); err == nil {
		t.Fatal("expected error for negative score")
	}
	if err := validateCompetitionResults([]CompetitionResult{{Cycle: 1, Score: 0, Place: &place}}); err == nil {
		t.Fatal("expected error for place 0")
	}
	if err := validateCompetitionResults([]CompetitionResult{
		{Cycle: 1, Score: 0},
		{Cycle: 1, Score: 10},
	}); err == nil {
		t.Fatal("expected duplicate cycle error")
	}
	if err := validateCompetitionResults([]CompetitionResult{{Cycle: 1, Score: 0}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDisplayCompetitionResult(t *testing.T) {
	cycle := 2
	a := App{
		CompetitionCycle: &cycle,
		CompetitionResults: []CompetitionResult{
			{Cycle: 1, Score: 10},
			{Cycle: 2, Score: 20},
		},
	}
	got := displayCompetitionResult(a)
	if got == nil || got.Cycle != 2 || got.Score != 20 {
		t.Fatalf("got %#v, want cycle 2 score 20", got)
	}

	a.CompetitionCycle = nil
	got = displayCompetitionResult(a)
	if got == nil || got.Cycle != 2 {
		t.Fatalf("got %#v, want highest cycle 2", got)
	}
}

func TestCompetitionResultsSeedAndAdminUpsert(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	var count int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM app_competition_results r
		JOIN apps a ON a.id = r.app_id
		WHERE a.competition_cycle = 1 AND r.cycle = 1`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("expected seeded cycle 1 results for competition apps")
	}

	for _, tc := range []struct {
		slug  string
		place int
	}{
		{"nimiq-space", 1},
		{"nimjump", 2},
		{"nimquest", 3},
	} {
		var place *int
		var score int
		err := pool.QueryRow(ctx, `
			SELECT r.score, r.place FROM app_competition_results r
			JOIN apps a ON a.id = r.app_id
			WHERE a.slug = $1 AND r.cycle = 1`, tc.slug).Scan(&score, &place)
		if err != nil {
			t.Fatalf("%s: %v", tc.slug, err)
		}
		if score != 0 {
			t.Fatalf("%s score = %d, want 0", tc.slug, score)
		}
		if place == nil || *place != tc.place {
			t.Fatalf("%s place = %v, want %d", tc.slug, place, tc.place)
		}
	}

	s := &server{pool: pool, adminToken: "test-admin"}
	body := map[string]any{
		"competition_results": []map[string]any{
			{"cycle": 1, "score": 42, "place": 2},
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/apps/nimjump", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer test-admin")
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("slug", "nimjump")
	rec := httptest.NewRecorder()
	s.adminAuth(s.updateApp)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update status %d: %s", rec.Code, rec.Body.String())
	}
	var updated App
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if len(updated.CompetitionResults) == 0 || updated.CompetitionResults[0].Score != 42 {
		t.Fatalf("updated results = %#v", updated.CompetitionResults)
	}

	// Restore placeholder score for local DBs reused across tests.
	_, _ = pool.Exec(ctx, `
		UPDATE app_competition_results r
		SET score = 0
		FROM apps a
		WHERE a.id = r.app_id AND a.slug = 'nimjump' AND r.cycle = 1`)
}

func TestGetAppIncludesCompetitionResults(t *testing.T) {
	pool := testPool(t)
	s := &server{pool: pool}
	req := httptest.NewRequest(http.MethodGet, "/api/apps/nimjump", nil)
	req.SetPathValue("slug", "nimjump")
	rec := httptest.NewRecorder()
	s.getApp(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var a App
	if err := json.Unmarshal(rec.Body.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	if a.CompetitionResults == nil {
		t.Fatal("competition_results should be [] not null")
	}
	found := false
	for _, r := range a.CompetitionResults {
		if r.Cycle == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cycle 1 result, got %#v", a.CompetitionResults)
	}
}
