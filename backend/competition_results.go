package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompetitionResult struct {
	Cycle int  `json:"cycle"`
	Score int  `json:"score"`
	Place *int `json:"place"`
}

func validateCompetitionResults(results []CompetitionResult) error {
	if results == nil {
		return nil
	}
	var problems []string
	seen := map[int]struct{}{}
	for i, r := range results {
		if r.Cycle < 1 {
			problems = append(problems, "competition_results["+strconv.Itoa(i)+"].cycle must be a positive integer")
		}
		if r.Score < 0 {
			problems = append(problems, "competition_results["+strconv.Itoa(i)+"].score must be >= 0")
		}
		if r.Place != nil && *r.Place < 1 {
			problems = append(problems, "competition_results["+strconv.Itoa(i)+"].place must be a positive integer")
		}
		if r.Cycle >= 1 {
			if _, ok := seen[r.Cycle]; ok {
				problems = append(problems, "competition_results duplicate cycle "+strconv.Itoa(r.Cycle))
			}
			seen[r.Cycle] = struct{}{}
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func attachCompetitionResults(ctx context.Context, pool *pgxpool.Pool, apps []App) error {
	if len(apps) == 0 {
		return nil
	}
	ids := make([]string, len(apps))
	index := make(map[string]int, len(apps))
	for i := range apps {
		ids[i] = apps[i].ID
		index[apps[i].ID] = i
		apps[i].CompetitionResults = []CompetitionResult{}
	}
	rows, err := pool.Query(ctx, `
		SELECT app_id::text, cycle, score, place
		FROM app_competition_results
		WHERE app_id = ANY($1::uuid[])
		ORDER BY cycle ASC`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var appID string
		var r CompetitionResult
		if err := rows.Scan(&appID, &r.Cycle, &r.Score, &r.Place); err != nil {
			return err
		}
		i, ok := index[appID]
		if !ok {
			continue
		}
		apps[i].CompetitionResults = append(apps[i].CompetitionResults, r)
	}
	return rows.Err()
}

func attachCompetitionResult(ctx context.Context, pool *pgxpool.Pool, a *App) error {
	apps := []App{*a}
	if err := attachCompetitionResults(ctx, pool, apps); err != nil {
		return err
	}
	*a = apps[0]
	return nil
}

func upsertCompetitionResults(ctx context.Context, pool *pgxpool.Pool, appID string, results []CompetitionResult) error {
	if len(results) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, r := range results {
		batch.Queue(`
			INSERT INTO app_competition_results (app_id, cycle, score, place)
			VALUES ($1::uuid, $2, $3, $4)
			ON CONFLICT (app_id, cycle) DO UPDATE
			SET score = EXCLUDED.score, place = EXCLUDED.place`,
			appID, r.Cycle, r.Score, r.Place)
	}
	br := pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert competition result: %w", err)
		}
	}
	return nil
}

// displayCompetitionResult picks the result to show on cards: matching current
// cycle if present, otherwise the highest cycle.
func displayCompetitionResult(a App) *CompetitionResult {
	if len(a.CompetitionResults) == 0 {
		return nil
	}
	if a.CompetitionCycle != nil {
		for i := range a.CompetitionResults {
			if a.CompetitionResults[i].Cycle == *a.CompetitionCycle {
				return &a.CompetitionResults[i]
			}
		}
	}
	best := &a.CompetitionResults[0]
	for i := 1; i < len(a.CompetitionResults); i++ {
		if a.CompetitionResults[i].Cycle > best.Cycle {
			best = &a.CompetitionResults[i]
		}
	}
	return best
}
