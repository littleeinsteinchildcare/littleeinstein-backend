package middleware

import (
	"context"
	"net/http"
	"strings"

	"littleeinsteinchildcare/backend/firebase"
	"littleeinsteinchildcare/backend/internal/utils"
)

type contextKey string

const (
	ContextUID   contextKey = "uid"
	ContextEmail contextKey = "email"
)

func FirebaseAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondUnauthorized(w, "Missing or invalid Authorization header")
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		authClient, err := firebase.Auth(r.Context())
		if err != nil {
			utils.RespondUnauthorized(w, "Failed to initialize Firebase Auth")
			return
		}

		token, err := authClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			utils.RespondUnauthorized(w, "Your session is invalid or has expired. Please sign in again.")
			return
		}

		ctx := context.WithValue(r.Context(), ContextUID, token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			ctx = context.WithValue(ctx, ContextEmail, email)
		}
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
