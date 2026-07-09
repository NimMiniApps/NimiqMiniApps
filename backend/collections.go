package main

import (
	"net/http"
)

type collection struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

var catalogCollections = []collection{
	{ID: "new-week", Title: "New this week", Description: "Apps added in the last 7 days"},
	{ID: "popular", Title: "Trending this week", Description: "Most viewed apps in the last 7 days"},
	{ID: "rewards", Title: "Apps with rewards", Description: "Apps that can reward users with crypto assets"},
	{ID: "games", Title: "Games", Description: "Play inside Nimiq Pay"},
	{ID: "usdt", Title: "Uses USDT", Description: "Apps that support USDT"},
}

func applyCollection(where *[]string, arg func(any) string, id string) bool {
	switch id {
	case "new-week":
		*where = append(*where, "created_at >= now() - interval '7 days'")
	case "popular":
		*where = append(*where, `EXISTS (
			SELECT 1 FROM app_stats_daily
			WHERE app_id = apps.id
			AND day >= CURRENT_DATE - INTERVAL '6 days'
			GROUP BY app_id
			HAVING SUM(views) >= 1
		)`)
	case "rewards":
		*where = append(*where, "array_length(reward_assets, 1) > 0")
	case "games":
		*where = append(*where, "category = "+arg("Games"))
	case "usdt":
		*where = append(*where, arg("USDT")+" = ANY(assets)")
	default:
		return false
	}
	return true
}

func (s *server) listCollections(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, catalogCollections)
}
