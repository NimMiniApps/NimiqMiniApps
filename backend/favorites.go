package main

import "net/http"

func (s *server) addFavorite(w http.ResponseWriter, r *http.Request, address string) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	_, err = s.pool.Exec(r.Context(),
		`INSERT INTO app_favorites (app_id, wallet_address) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		appID, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) removeFavorite(w http.ResponseWriter, r *http.Request, address string) {
	appID, err := s.appIDForSlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	_, err = s.pool.Exec(r.Context(), `DELETE FROM app_favorites WHERE app_id=$1 AND wallet_address=$2`, appID, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) myFavorites(w http.ResponseWriter, r *http.Request, address string) {
	rows, err := s.pool.Query(r.Context(),
		`SELECT `+appColumns+` FROM apps
		 JOIN app_favorites f ON f.app_id = apps.id
		 WHERE f.wallet_address = $1 AND `+publicStatuses+`
		 ORDER BY f.created_at DESC`, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	apps := []App{}
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		stripPrivateAppFields(&a)
		apps = append(apps, a)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, apps)
}
