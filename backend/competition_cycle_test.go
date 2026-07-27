package main

import "testing"

func TestParseCompetitionCycleFilter(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantCycle   *int
		wantWithout bool
		wantError   bool
	}{
		{name: "omitted"},
		{name: "cycle one", raw: "1", wantCycle: intPointer(1)},
		{name: "without cycle", raw: "none", wantWithout: true},
		{name: "zero", raw: "0", wantError: true},
		{name: "negative", raw: "-1", wantError: true},
		{name: "not a number", raw: "abc", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycle, withoutCycle, err := parseCompetitionCycleFilter(tt.raw)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError %v", err, tt.wantError)
			}
			if withoutCycle != tt.wantWithout {
				t.Errorf("withoutCycle = %v, want %v", withoutCycle, tt.wantWithout)
			}
			if !equalIntPointers(cycle, tt.wantCycle) {
				t.Errorf("cycle = %v, want %v", cycle, tt.wantCycle)
			}
		})
	}
}

func TestRevisionToAppCompetitionCycle(t *testing.T) {
	currentCycle := 1
	revisedCycle := 2

	updated := revisionToApp(
		AppRevision{CompetitionCycle: &revisedCycle},
		App{CompetitionCycle: &currentCycle},
	)

	if updated.CompetitionCycle == nil || *updated.CompetitionCycle != revisedCycle {
		t.Fatalf("competition cycle = %v, want %d", updated.CompetitionCycle, revisedCycle)
	}
}

func intPointer(value int) *int {
	return &value
}

func equalIntPointers(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
