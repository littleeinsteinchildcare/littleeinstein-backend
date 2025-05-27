package middleware

import (
	"context"
	"net/http"
	"strings"

	"littleeinsteinchildcare/backend/firebase"
	"littleeinsteinchildcare/backend/internal/utils"
)

func FirebaseAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondUnauthorized(w, "Missing or invalid Authorization header")
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		app := firebase.Init()
		authClient, err := app.Auth(r.Context())
		if err != nil {
			utils.RespondUnauthorized(w, "Failed to initialize Firebase Auth")
			return
		}

		token, err := authClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			utils.RespondUnauthorized(w, "Your session is invalid or has expired. Please sign in again.")
			return
		}

		ctx := context.WithValue(r.Context(), utils.ContextUID, token.UID)
		log.Printf("DEBUG: Token verified successfully, UID: %s", token.UID)
		if email, ok := token.Claims["email"].(string); ok {
		ctx = context.WithValue(ctx, utils.ContextEmail, email)
		log.Printf("DEBUG: Email from token: %s", email)
		}
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
