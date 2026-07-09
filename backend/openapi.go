package main

import (
	"embed"
	"net/http"
)

//go:embed openapi.yaml openapi.json
var openapiFS embed.FS

func serveOpenAPI(w http.ResponseWriter, r *http.Request, filename, contentType string) {
	data, err := openapiFS.ReadFile(filename)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "openapi spec unavailable")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *server) openAPIJSON(w http.ResponseWriter, r *http.Request) {
	serveOpenAPI(w, r, "openapi.json", "application/json")
}

func (s *server) openAPIYAML(w http.ResponseWriter, r *http.Request) {
	serveOpenAPI(w, r, "openapi.yaml", "text/yaml; charset=utf-8")
}
