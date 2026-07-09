package main

import (
	"net/url"
	"strconv"
)

const (
	defaultPageLimit = 50
	maxPageLimit     = 100
)

type paginatedApps struct {
	Items  []App `json:"items"`
	Total  int   `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

func parsePagination(q url.Values) (limit, offset int, paginate bool, errMsg string) {
	if v := q.Get("paginate"); v == "1" || v == "true" {
		paginate = true
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return 0, 0, false, "invalid offset"
		}
		offset = n
		paginate = true
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return 0, 0, false, "invalid limit"
		}
		if n > maxPageLimit {
			n = maxPageLimit
		}
		limit = n
	} else if paginate {
		limit = defaultPageLimit
	}
	return limit, offset, paginate, ""
}
