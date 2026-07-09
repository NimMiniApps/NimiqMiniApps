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
	{ID: "games", Title: "Games", Description: "Play inside Nimiq Pay"},
	{ID: "usdt", Title: "Uses USDT", Description: "Apps that support USDT"},
}

func applyCollection(where *[]string, arg func(any) string, id string) bool {
	switch id {
	case "new-week":
		*where = append(*where, "created_at >= now() - interval '7 days'")
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
