package middleware

import "net/http"

// corsMiddleware creates a middleware function that handles CORS for incoming requests
// It takes a handler as input and returns a new handler that adds CORS headers
func CorsMiddleware(next http.Handler) http.Handler {
	// Return a new handler by converting our function to http.HandlerFunc type
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set which origins (domains) are allowed to make requests to this server
		// In this case, only http://localhost:3000 can access our API
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

		// Specify which HTTP methods the client is allowed to use
		// These are the standard REST methods plus OPTIONS for preflight requests
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Allow all headers to be sent with the request
		// "*" is a wildcard meaning any header is permitted
		w.Header().Set("Access-Control-Allow-Headers", "*")

		// Check if this is a preflight request (browsers send OPTIONS before actual request)
		if r.Method == "OPTIONS" {
			// Respond with 200 OK to indicate the actual request is allowed
			w.WriteHeader(http.StatusOK)
			// Exit early - don't process this as a normal request
			return
		}

		// If not a preflight request, pass control to the next handler in the chain
		// This executes the actual API endpoint handler
		next.ServeHTTP(w, r)
	})
}
