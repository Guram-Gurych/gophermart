package middleware

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

func AuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				http.Error(w, "User authorization failed", http.StatusUnauthorized)
				return
			}

			claims := &domain.TokenClaims{}

			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "User authorization failed", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), domain.UserIDKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
