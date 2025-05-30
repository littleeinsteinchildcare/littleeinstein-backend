package handlers

import (
	"encoding/json"
	"net/http"
	"context"
	"littleeinsteinchildcare/backend/internal/models"
	"littleeinsteinchildcare/backend/internal/services"
	"littleeinsteinchildcare/backend/firebase"
	"cloud.google.com/go/firestore"
)

type EmailHandler struct {
	EmailService services.EmailService
	  FirestoreClient *firestore.Client
}

func NewEmailHandler(emailService services.EmailService) *EmailHandler {
	return &EmailHandler{EmailService: emailService}
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

	_, err = fsClient.Collection("invitedUsers").Doc(req.Email).Set(ctx, map[string]interface{}{
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