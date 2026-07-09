package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Profile struct {
	WalletAddress string  `json:"wallet_address"`
	DisplayName   *string `json:"display_name"`
}

func validateDisplayName(name string) string {
	if len(name) > 50 {
		return "display_name must be at most 50 characters"
	}
	return ""
}

func (s *server) getProfile(w http.ResponseWriter, r *http.Request, address string) {
	var displayName *string
	err := s.pool.QueryRow(r.Context(),
		`SELECT display_name FROM users WHERE wallet_address=$1`, address,
	).Scan(&displayName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, Profile{WalletAddress: address, DisplayName: displayName})
}

func (s *server) updateProfile(w http.ResponseWriter, r *http.Request, address string) {
	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	name := strings.TrimSpace(req.DisplayName)
	if msg := validateDisplayName(name); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	var namePtr *string
	if name != "" {
		namePtr = &name
	}
	_, err := s.pool.Exec(r.Context(), `
		INSERT INTO users (wallet_address, display_name, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (wallet_address) DO UPDATE SET display_name=$2, updated_at=now()`,
		address, namePtr)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		writeError(w, http.StatusConflict, "display name already taken")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, Profile{WalletAddress: address, DisplayName: namePtr})
}
