package main

import (
	"testing"
	"time"
)

func TestAllowSubmit(t *testing.T) {
	now := time.Now()
	for i := 0; i < submitLimit; i++ {
		if !allowSubmit("1.2.3.4", now) {
			t.Fatalf("submission %d should be allowed", i+1)
		}
	}
	if allowSubmit("1.2.3.4", now) {
		t.Error("submission over the limit should be rejected")
	}
	if !allowSubmit("5.6.7.8", now) {
		t.Error("different IP should be unaffected")
	}
	if !allowSubmit("1.2.3.4", now.Add(submitWindow+time.Minute)) {
		t.Error("limit should reset after the window passes")
	}
}
