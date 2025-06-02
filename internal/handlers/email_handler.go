package handlers

import (
	"strings"
	"encoding/json"
	"net/http"
	"context"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"
	"littleeinsteinchildcare/backend/firebase"
	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EmailHandler struct {
	EmailService services.EmailService 
	FirestoreClient *firestore.Client
}

func NewEmailHandler(emailService services.EmailService, fsClient *firestore.Client) *EmailHandler {
	return &EmailHandler{EmailService: emailService,
		FirestoreClient: fsClient,
	}
}

func (h *EmailHandler) SendInvite(w http.ResponseWriter, r *http.Request) {
	var req models.InviteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	if err := h.EmailService.SendInviteEmail(req.Email); err != nil {
		http.Error(w, "Failed to send invite", http.StatusInternalServerError)
		return
	}
	
	ctx := context.Background()
	fsClient, err := firebase.Firestore(ctx)
	if err != nil {
		http.Error(w, "Failed to connect to Firestore", http.StatusInternalServerError)
		return
	}
	defer fsClient.Close()

	_, err = h.FirestoreClient.Collection("invitedUsers").Doc(req.Email).Set(ctx, map[string]interface{}{
		"invited": true,
		"signedUp": false,
		"role": "Parent",
	})

	if err != nil {
		http.Error(w, "Failed to record invite", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Invite sent!"))
}

func (h *EmailHandler) CheckIfInvited(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Missing email parameter", http.StatusBadRequest)
		return
	}

	invited, err := h.IsEmailInvited(ctx, email)
	if err != nil {
		http.Error(w, "Failed to check invitation", http.StatusInternalServerError)
		return
	}

	// Send result back to client
	resp := map[string]bool{"invited": invited}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *EmailHandler) IsEmailInvited(ctx context.Context, email string) (bool, error) {
	docRef := h.FirestoreClient.Collection("invitedUsers").Doc(strings.ToLower(email))
	doc, err := docRef.Get(ctx)

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil 
		}
		return false, err 
	}

	return doc.Exists(), nil
}