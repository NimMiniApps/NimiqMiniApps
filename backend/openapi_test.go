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

	components := doc["components"].(map[string]any)
	schemas := components["schemas"].(map[string]any)
	for _, schemaName := range []string{"AppInput", "AppSubmitRequest", "AppPublic", "AppRevision"} {
		schema := schemas[schemaName].(map[string]any)
		properties := schema["properties"].(map[string]any)
		if properties["competition_cycle"] == nil {
			t.Errorf("expected competition_cycle in %s", schemaName)
		}
		if properties["declared_scopes"] == nil {
			t.Errorf("expected declared_scopes in %s", schemaName)
		}
	}
	for _, schemaName := range []string{"AppInput", "AppPublic"} {
		schema := schemas[schemaName].(map[string]any)
		properties := schema["properties"].(map[string]any)
		if properties["competition_results"] == nil {
			t.Errorf("expected competition_results in %s", schemaName)
		}
	}
	if schemas["CompetitionResult"] == nil {
		t.Fatal("expected CompetitionResult schema")
	}
	if paths["/api/apps/{slug}/client"] == nil {
		t.Fatal("expected /api/apps/{slug}/client in paths")
	}
	if paths["/api/apps/changed"] == nil {
		t.Fatal("expected /api/apps/changed in paths")
	}
	if paths["/registry/credentials/verify"] == nil {
		t.Fatal("expected /registry/credentials/verify in paths")
	}
	if paths["/api/apps/{slug}/credentials"] == nil {
		t.Fatal("expected /api/apps/{slug}/credentials in paths")
	}
	if schemas["AppCredentialCreated"] == nil || schemas["VerifyCredentialRequest"] == nil {
		t.Fatal("expected AppCredentialCreated and VerifyCredentialRequest schemas")
	}
	securitySchemes := components["securitySchemes"].(map[string]any)
	if securitySchemes["registryBearer"] == nil {
		t.Fatal("expected registryBearer security scheme")
	}

	parameters := components["parameters"].(map[string]any)
	if parameters["competitionCycle"] == nil {
		t.Error("expected competitionCycle query parameter")
	}
	listApps := paths["/api/apps"].(map[string]any)["get"].(map[string]any)
	listParameters := listApps["parameters"].([]any)
	foundCycleFilter := false
	for _, item := range listParameters {
		parameter := item.(map[string]any)
		if parameter["$ref"] == "#/components/parameters/competitionCycle" {
			foundCycleFilter = true
			break
		}
	}
	if !foundCycleFilter {
		t.Error("expected listApps to reference competitionCycle")
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
