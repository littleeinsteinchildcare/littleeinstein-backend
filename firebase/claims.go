package firebase

import (
	"context"
	"log"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
)

// SyncAdminClaims checks Firestore for admin users and updates their custom claims
func SyncAdminClaims(app *firebase.App) error {
	ctx := context.Background()

	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return err
	}
	defer firestoreClient.Close()

	authClient, err := app.Auth(ctx)
	if err != nil {
		return err
	}

	iter := firestoreClient.Collection("invitedUsers").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		data := doc.Data()
		isAdmin, _ := data["admin"].(bool)
		signedUp, _ := data["signedUp"].(bool)
		if !signedUp || !isAdmin {
			continue
		}

		email := doc.Ref.ID // You could also use a "uid" field if preferred

		user, err := authClient.GetUserByEmail(ctx, email)
		if err != nil {
			log.Printf("Could not find user %s: %v", email, err)
			continue
		}

		err = authClient.SetCustomUserClaims(ctx, user.UID, map[string]interface{}{"admin": true})
		if err != nil {
			log.Printf("Error setting admin claim for %s: %v", email, err)
		} else {
			log.Printf("Admin claim set for user: %s", email)
		}
	}

	return nil
}

func SetAdminClaimForEmail(ctx context.Context, app *firebase.App, email string) error {
	// Initialize Firestore
	fsClient, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to init Firestore: %w", err)
	}
	defer fsClient.Close()

	// Initialize Auth
	authClient, err := app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to init Firebase Auth: %w", err)
	}

	// Lookup invitedUsers/{email}
	doc, err := fsClient.Collection("invitedUsers").Doc(email).Get(ctx)
	if err != nil {
		log.Printf("No invitedUsers record found for %s", email)
		return nil
	}

	data := doc.Data()
	isAdmin, _ := data["admin"].(bool)
	signedUp, _ := data["signedUp"].(bool)

	if !isAdmin || !signedUp {
		log.Printf("Skipping claim for %s: admin=%v, signedUp=%v", email, isAdmin, signedUp)
		return nil
	}

	// Get Firebase user
	user, err := authClient.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("firebase user not found: %w", err)
	}

	// Set claim
	err = authClient.SetCustomUserClaims(ctx, user.UID, map[string]interface{}{"admin": true})
	if err != nil {
		return fmt.Errorf("failed to set admin claim: %w", err)
	}

	log.Printf("✅ Admin claim set for: %s", email)
	return nil
}