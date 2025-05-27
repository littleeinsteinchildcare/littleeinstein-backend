package utils

import (
	"context"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

)

type ContextKey interface{}

type contextKey string

const (
	ContextUID   contextKey = "uid"
	ContextEmail contextKey = "email"
)

type ContextKey interface{}

func DecodeJSONRequest(r *http.Request) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Error helper for additional frontend information
func WriteJSONError(w http.ResponseWriter, status int, msg string, err error) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
		"error":  msg,
	})
	if err != nil {
		log.Printf("HTTP %d - %s\nError: %v\n", status, msg, err)
	}

}

func GetUserIDFromAuth(r *http.Request) (string, error) {
	log.Printf("DEBUG: GetUserIDFromAuth called")
	uid, ok := r.Context().Value(ContextUID).(string)
	log.Printf("DEBUG: Context UID lookup - found: %v, value: '%s'", ok, uid)
	if !ok || uid == "" {
		return "", errors.New("User ID not found in context")
	}
	log.Printf("DEBUG: Returning UID: %s", uid)
	return uid, nil
}

func GetContextString(ctx context.Context, key ContextKey) (string, bool) {
    val, ok := ctx.Value(key).(string)
    return val, ok
}

func RespondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func GetContextString(ctx context.Context, key ContextKey) (string, bool) {
	val, ok := ctx.Value(key).(string)
	return val, ok
}
