package apix

import (
	"net/http"
	"strings"

	"parent-child-api/ux"
)

type AuthedHandler func(http.ResponseWriter, *http.Request, *ux.AuthTokenClaims)

func WithAuth(jwtSecret string, next AuthedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawToken := strings.TrimSpace(r.Header.Get("Authorization"))
		rawToken = strings.TrimPrefix(rawToken, "Bearer ")
		if rawToken == "" {
			WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		token, err := ux.JwtUtil.ParseAuthToken(rawToken, jwtSecret)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		next(w, r, token)
	}
}
