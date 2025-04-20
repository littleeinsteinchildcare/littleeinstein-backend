package handlers

import (
	"encoding/json"
	"fmt"
	"littleeinsteinchildcare/backend/internal/models"
	"net/http"
)

// UserService interface implemented in services package
type UserService interface {
	CreateUser(user models.User) error
	GetUserByID(id string) (models.User, error)
}

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userService UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(s UserService) *UserHandler {
	return &UserHandler{
		userService: s,
	}
}

// GetUser handles GET requests for a specific user
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from request
	id := r.PathValue("id")

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		fmt.Printf("Error retrieving user: %v", err)
	}

	response := map[string]interface{}{
		"id":       id,
		"username": user.Name,
		"email":    user.Email,
		"role":     user.Role,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateUser handles POST requests to create a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	success := true
	msg := "User Created Successfully"
	userData, err := DecodeUserRequest(r)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	user := models.User{
		ID:    userData["id"].(string),
		Name:  userData["username"].(string),
		Email: userData["email"].(string),
		Role:  userData["role"].(string),
	}

	err = h.userService.CreateUser(user)
	if err != nil {
		msg = fmt.Sprintf("Error Creating User: %v\n", err)
		success = false
	}

	response := map[string]interface{}{
		"success":  success,
		"message":  msg,
		"userId":   user.ID,
		"username": user.Name,
		"email":    user.Email,
		"role":     user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func DecodeUserRequest(r *http.Request) (map[string]interface{}, error) {
	var userData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		fmt.Printf("Failed to decode json request")
		return userData, err
	}
	return userData, nil
}
