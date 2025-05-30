package common

// contextKey is used for context keys to avoid collisions
type contextKey string

const (
	// ContextUID stores the Firebase user ID in request context
	ContextUID contextKey = "uid"
	// ContextEmail stores the Firebase user email in request context  
	ContextEmail contextKey = "email"
)