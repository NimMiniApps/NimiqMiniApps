package main

import (
	"context"
	"testing"
	"time"
)

func TestAllowSubmit(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	s := &server{pool: pool}
	ip := "test-submit-" + time.Now().Format("150405.000000")
	now := time.Now()

	for i := 0; i < submitLimit; i++ {
		ok, err := s.allowSubmit(ctx, ip, now)
		if err != nil {
			t.Fatalf("submission %d: %v", i+1, err)
		}
		if !ok {
			t.Fatalf("submission %d should be allowed", i+1)
		}
	}
	ok, err := s.allowSubmit(ctx, ip, now)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("submission over the limit should be rejected")
	}

	otherIP := ip + "-other"
	ok, err = s.allowSubmit(ctx, otherIP, now)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("different IP should be unaffected")
	}

	ok, err = s.allowSubmit(ctx, ip, now.Add(submitWindow+time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("limit should reset after the window passes")
	}
}
