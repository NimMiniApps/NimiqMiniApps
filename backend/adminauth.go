package main

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func parseAdminWallets(raw string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, part := range strings.Split(raw, ",") {
		addr := normalizeUserFriendlyAddress(part)
		if addr != "" {
			out[addr] = struct{}{}
		}
	}
	return out
}

func isAdminWallet(allowlist map[string]struct{}, address string) bool {
	if len(allowlist) == 0 {
		return false
	}
	_, ok := allowlist[normalizeUserFriendlyAddress(address)]
	return ok
}

// adminAuthMiddleware accepts an allowlisted wallet session cookie or ADMIN_TOKEN bearer.
func adminAuthMiddleware(adminToken string, adminWallets map[string]struct{}, walletSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if walletSecret != "" && len(adminWallets) > 0 {
			if cookie, err := r.Cookie(walletCookieName); err == nil {
				if address, err := verifyWalletCookie(walletSecret, cookie.Value); err == nil {
					if isAdminWallet(adminWallets, address) {
						setWalletCookie(w, r, walletSecret, address)
						next(w, r)
						return
					}
				}
			}
		}
		got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if adminToken != "" && subtle.ConstantTimeCompare([]byte(got), []byte(adminToken)) == 1 {
			next(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, "admin access required")
	}
}

func (s *server) adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return adminAuthMiddleware(s.adminToken, s.adminWallets, s.walletAuthSecret, next)
}
