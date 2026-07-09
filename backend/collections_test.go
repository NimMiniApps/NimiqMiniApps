package main

import (
	"strings"
	"testing"
)

func TestApplyPopularCollection(t *testing.T) {
	var where []string
	args := []any{}
	arg := func(v any) string {
		args = append(args, v)
		return "$1"
	}

	if !applyCollection(&where, arg, "popular") {
		t.Fatal("popular collection should be accepted")
	}
	if len(where) != 1 || !strings.Contains(where[0], "app_stats_daily") || !strings.Contains(where[0], "SUM(views)") {
		t.Fatalf("unexpected popular collection clause: %#v", where)
	}
}

func TestApplyRewardsCollection(t *testing.T) {
	var where []string
	args := []any{}
	arg := func(v any) string {
		args = append(args, v)
		return "$1"
	}

	if !applyCollection(&where, arg, "rewards") {
		t.Fatal("rewards collection should be accepted")
	}
	if len(where) != 1 || !strings.Contains(where[0], "array_length(reward_assets, 1) > 0") {
		t.Fatalf("unexpected rewards collection clause: %#v", where)
	}
}
