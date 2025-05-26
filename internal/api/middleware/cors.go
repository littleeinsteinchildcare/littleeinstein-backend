package middleware

import (
	"net/http"
	"os"
)

// CorsMiddleware creates a middleware function that handles CORS for incoming requests
// It takes a handler as input and returns a new handler that adds CORS headers
func CorsMiddleware(next http.Handler) http.Handler {
	// Return a new handler by converting our function to http.HandlerFunc type
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set which origins (domains) are allowed to make requests to this server
		switch environment := os.Getenv("APP_ENV"); environment {
		case "production":
			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CORS_ALLOW_ORIGIN"))
		case "development":
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		case "legacy":
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		}

		// Specify which HTTP methods the client is allowed to use
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		// Add missing CORS headers
		w.Header().Set("Access-Control-Max-Age", "86400")          // Cache preflight response for 24 hours
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow credentials if needed

		// Check if this is a preflight request (browsers send OPTIONS before actual request)
		if r.Method == "OPTIONS" {
			// Respond with 200 OK to indicate the actual request is allowed
			w.WriteHeader(http.StatusOK)
			return
		}

		// If not a preflight request, pass control to the next handler in the chain
		// This executes the actual API endpoint handler
		next.ServeHTTP(w, r)
	})
}
