package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAPIEmbeddedJSON(t *testing.T) {
	data, err := openapiFS.ReadFile("openapi.json")
	if err != nil {
		t.Fatalf("read openapi.json: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if doc["openapi"] != "3.1.0" {
		t.Fatalf("expected openapi 3.1.0, got %v", doc["openapi"])
	}
	paths, ok := doc["paths"].(map[string]any)
	if !ok || paths["/api/apps/submit"] == nil {
		t.Fatal("expected /api/apps/submit in paths")
	}
}

func TestOpenAPIHandlers(t *testing.T) {
	s := &server{}
	for _, tc := range []struct {
		name string
		path string
		ct   string
		fn   func(http.ResponseWriter, *http.Request)
	}{
		{"json", "/openapi.json", "application/json", s.openAPIJSON},
		{"yaml", "/openapi.yaml", "text/yaml", s.openAPIYAML},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			tc.fn(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != tc.ct && ct != tc.ct+"; charset=utf-8" {
				t.Fatalf("content-type %q want %q", ct, tc.ct)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("empty body")
			}
		})
	}
}
