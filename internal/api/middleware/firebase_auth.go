package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"littleeinsteinchildcare/backend/firebase"
	"littleeinsteinchildcare/backend/internal/utils"
)
func FirebaseAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: FirebaseAuthMiddleware called for %s %s", r.Method, r.URL.Path)
		
		authHeader := r.Header.Get("Authorization")
		log.Printf("DEBUG: Auth header: %s", authHeader)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("DEBUG: Missing or invalid Authorization header")
			utils.RespondUnauthorized(w, "Missing or invalid Authorization header")
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		log.Printf("DEBUG: Extracted token: %s...", idToken[:min(len(idToken), 20)])

		app := firebase.Init()
		authClient, err := app.Auth(r.Context())
		if err != nil {
			log.Printf("DEBUG: Failed to initialize Firebase Auth: %v", err)
			utils.RespondUnauthorized(w, "Failed to initialize Firebase Auth")
			return
		}

		token, err := authClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("DEBUG: Failed to verify token: %v", err)
			utils.RespondUnauthorized(w, "Your session is invalid or has expired. Please sign in again.")
			return
		}

		log.Printf("DEBUG: Token verified successfully, UID: %s", token.UID)
		ctx := context.WithValue(r.Context(), utils.ContextUID, token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			ctx = context.WithValue(ctx, utils.ContextEmail, email)
			log.Printf("DEBUG: Email from token: %s", email)
		}
		
		log.Printf("DEBUG: Calling next handler with UID in context")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
