package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"littleeinsteinchildcare/backend/internal/models"
	"log"
	"net/http"
)

// UserService interface implemented in services package
type UserService interface {
	CreateUser(user models.User) error
	GetUserByID(id string) (models.User, error)
	GetAllUsers() ([]models.User, error)
	DeleteUserByID(id string) error
	UpdateUser(user models.User) (models.User, error)
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
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("UserHandler.GetUser: Failed to find User with ID %s", id), err)
		return
	}

	response := BuildResponse(user)

	// Return JSON response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("UserHandler.GetUser: Failed to retrieve list of users"), err)
	}

	var responses []map[string]interface{}
	// For idx, value
	for _, user := range users {
		// Create response with a list of user data
		resp := BuildResponse(user)
		responses = append(responses, resp)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// UpdateUser handles PUT requests for a specific user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	newData, err := io.ReadAll(r.Body)

	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "UserHandler.UpdateUser: Failed to read request body", err)
		return
	}

	defer r.Body.Close()

	var user models.User
	if err := json.Unmarshal(newData, &user); err != nil {
		WriteJSONError(w, http.StatusBadRequest, "UserHandler.UpdateUser: Attempt to unpack invalid JSON object", err)
		return
	}

	user, err = h.userService.UpdateUser(user)
	if err != nil {
		WriteJSONError(w, http.StatusNotFound, "UserHandler.UpdateUser: User does not exist", err)
		return
	}
	response := BuildResponse(user)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser handles DELETE requests to remove an existing user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.userService.DeleteUserByID(id)
	if err != nil {
		WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("UserHandler.DeleteUser: Failed to delete User with ID %s", id), err)
		return
	}

	// Standard RestAPI response on successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// CreateUser handles POST requests to create a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {

	userData, err := DecodeUserRequest(r)
	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "UserHandler.DecodeUserRequest: Failed to decode JSON request", err)
		return
	}

	user := models.User{
		ID:    userData["id"].(string),
		Name:  userData["username"].(string),
		Email: userData["email"].(string),
		Role:  userData["role"].(string),
	}

	err = h.userService.CreateUser(user)

	if err != nil {
		WriteJSONError(w, http.StatusBadRequest, "UserHandler.CreateUser: Failed to create User", err)
		return
	}

	response := BuildResponse(user)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function - Unpack JSON Object
func DecodeUserRequest(r *http.Request) (map[string]interface{}, error) {
	var userData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		return nil, err
	}
	return userData, nil
}

// Helper - build response object from User data
func BuildResponse(user models.User) map[string]interface{} {
	response := map[string]interface{}{
		"ID":       user.ID,
		"Username": user.Name,
		"Email":    user.Email,
		"Role":     user.Role,
	}
	return response
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
