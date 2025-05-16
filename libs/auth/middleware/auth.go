package middleware

import (
	"net/http"

	"github.com/detect-viz/shared-lib/auth"
)

type AuthMiddleware struct {
	auth auth.Authenticator
}

func NewAuthMiddleware(auth auth.Authenticator) *AuthMiddleware {
	return &AuthMiddleware{auth: auth}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		valid, err := m.auth.Verify(r.Context(), token)
		if err != nil || !valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
