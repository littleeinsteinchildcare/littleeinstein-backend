package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"littleeinsteinchildcare/backend/firebase"
	"littleeinsteinchildcare/backend/internal/common"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/utils"
	"net/http"

	"cloud.google.com/go/firestore"
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
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("UserHandler.GetUser: Failed to find User with ID %s", id), err)
		return
	}

	response := buildUserResponse(user)

	// Return JSON response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("UserHandler.GetAllUser: Failed to retrieve list of users"), err)
	}

	var responses []map[string]interface{}
	// For idx, value
	for _, user := range users {
		// Create response with a list of user data
		resp := buildUserResponse(user)
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
		utils.WriteJSONError(w, http.StatusBadRequest, "UserHandler.UpdateUser: Failed to read request body", err)
		return
	}

	defer r.Body.Close()

	var user models.User
	if err := json.Unmarshal(newData, &user); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "UserHandler.UpdateUser: Attempt to unpack invalid JSON object", err)
		return
	}

	user, err = h.userService.UpdateUser(user)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, "UserHandler.UpdateUser: User does not exist", err)
		return
	}
	response := buildUserResponse(user)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser handles DELETE requests to remove an existing user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.userService.DeleteUserByID(id)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, fmt.Sprintf("UserHandler.DeleteUser: Failed to delete User with ID %s", id), err)
		return
	}

	// Standard RestAPI response on successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// CreateUser handles POST requests to create a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user info from middleware
	uid, ok := utils.GetContextString(ctx, common.ContextUID)
	if !ok {
		http.Error(w, "Unauthorized: missing UID in context", http.StatusUnauthorized)
		return
	}

	email, ok := utils.GetContextString(ctx, common.ContextEmail)
	if !ok {
		http.Error(w, "Unauthorized: missing Email in context", http.StatusUnauthorized)
		return
	}

	authClient, err := firebase.Auth(ctx)
	if err != nil {
		http.Error(w, "Failed to initialize Firebase Auth client", http.StatusInternalServerError)
		return
	}

	userRecord, err := authClient.GetUser(ctx, uid)
	if err != nil {
		http.Error(w, "Failed to fetch user info from Firebase", http.StatusInternalServerError)
		return
	}
	fsClient, err := firebase.Firestore(ctx)
	if err != nil {
		http.Error(w, "Failed to initialize Firestore client", http.StatusInternalServerError)
		return
	}
	defer fsClient.Close()

	role := "user" // default
	isAdmin := false

	docSnap, err := fsClient.Collection("invitedUsers").Doc(email).Get(ctx)
	if err == nil {
		if val, ok := docSnap.Data()["admin"].(bool); ok && val {
			role = "admin"
			isAdmin = true
		}
	}

	if isAdmin {
		go func(email string) {
			if err := firebase.SetAdminClaimForEmail(context.Background(), firebase.Init(), email); err != nil {
				fmt.Printf("Failed to set admin claim for %s: %v\n", email, err)
			}
		}(email)

	}
	// Construct user model
	user := models.User{
		ID:    uid,
		Name:  userRecord.DisplayName,
		Email: email,
		Role:  role,
	}

	// Store user in DB
	if err := h.userService.CreateUser(user); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "UserHandler.CreateUser: Failed to create user",
			"details": err.Error(),
		})
		return
	}
	_, err = fsClient.Collection("invitedUsers").Doc(email).Update(ctx, []firestore.Update{
		{Path: "signedUp", Value: true},
	})
	if err != nil {
		fmt.Printf("Warning: failed to update signedUp flag in Firestore: %v\n", err)
	}

	// On success
	response := buildUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// SyncFirebaseUser handles POST requests to sync Firebase users with backend database
func (h *UserHandler) SyncFirebaseUser(w http.ResponseWriter, r *http.Request) {
	userData, err := utils.DecodeJSONRequest(r)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "UserHandler.SyncFirebaseUser: Failed to decode JSON request", err)
		return
	}

	// Extract Firebase user data
	uid, ok := userData["uid"].(string)
	if !ok {
		utils.WriteJSONError(w, http.StatusBadRequest, "UserHandler.SyncFirebaseUser: Missing or invalid uid", nil)
		return
	}

	email, ok := userData["email"].(string)
	if !ok {
		utils.WriteJSONError(w, http.StatusBadRequest, "UserHandler.SyncFirebaseUser: Missing or invalid email", nil)
		return
	}

	name, ok := userData["name"].(string)
	if !ok {
		name = userData["displayName"].(string) // Fallback to displayName
		if name == "" {
			name = "User" // Default name
		}
	}

	// Check if user already exists
	existingUser, err := h.userService.GetUserByID(uid)
	if err == nil {
		// User exists, return existing user
		response := buildUserResponse(existingUser)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// User doesn't exist, create new user
	user := models.User{
		ID:    uid,
		Name:  name,
		Email: email,
		Role:  "parent", // Default role for Firebase users
	}

	err = h.userService.CreateUser(user)
	if err != nil {
		utils.WriteJSONError(w, http.StatusConflict, "UserHandler.SyncFirebaseUser: Failed to create user", err)
		return
	}

	response := buildUserResponse(user)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper - build response object from User data
func buildUserResponse(user models.User) map[string]interface{} {
	response := map[string]interface{}{
		"ID":       user.ID,
		"Username": user.Name,
		"Email":    user.Email,
		"Role":     user.Role,
		"Images":   user.Images,
	}
	return response
}
