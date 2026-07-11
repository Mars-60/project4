package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/Mars-60/project4/backend/internal/auth"
	"github.com/Mars-60/project4/backend/internal/core"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				WriteError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			claims, err := authService.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				WriteError(w, http.StatusUnauthorized, err.Error())
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsContextKey, claims)))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok || !allowed[claims.Role] {
				WriteError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (auth.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(auth.Claims)
	return claims, ok
}

func UserIDFromContext(ctx context.Context) core.ID {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return ""
	}
	return claims.UserID
}
