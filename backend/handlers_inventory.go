package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/adamararcane/d2optifarm/backend/internal/auth"
	"github.com/adamararcane/d2optifarm/backend/internal/database"
	"golang.org/x/oauth2"
)

func (api *apiConfig) inventoryHandler(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated HTTP client from the session
	client, err := api.getClientFromSession(r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get the player's membership data
	membershipID, membershipType, err := GetMembershipData(client)
	if err != nil {
		http.Error(w, "Failed to get membership data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the player's profile data, including inventory
	profileData, err := GetProfileData(client, membershipID, membershipType)
	if err != nil {
		http.Error(w, "Failed to get profile data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract the inventory items from the profile data
	inventoryItems := profileData.Response.ProfileInventory.Data.Items

	// Return the inventory items as a JSON response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(inventoryItems)
	if err != nil {
		http.Error(w, "Failed to encode inventory data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (api *apiConfig) getClientFromSession(r *http.Request) (*http.Client, error) {
	// Get the session
	session, err := store.Get(r, "session-name")
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}

	// Retrieve the user ID from the session
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Fetch encrypted tokens and expiry from the database
	user, err := api.DB.GetUser(context.Background(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from DB: %v", err)
	}

	// Decrypt tokens
	accessToken, err := auth.Decrypt(user.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %v", err)
	}
	refreshToken, err := auth.Decrypt(user.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %v", err)
	}

	// Create the token object with expiry
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       user.TokenExpiry, // Assuming user.TokenExpiry is of type time.Time
	}

	// Create a token source
	tokenSource := oauth2Config.TokenSource(context.Background(), token)

	// Obtain a valid token (refresh if necessary)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update the database if the token was refreshed
	if newToken.AccessToken != token.AccessToken {
		// Encrypt the new tokens
		encryptedAccessToken, err := auth.Encrypt(newToken.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt new access token: %v", err)
		}
		encryptedRefreshToken, err := auth.Encrypt(newToken.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt new refresh token: %v", err)
		}

		err = api.DB.UpdateToken(context.Background(), database.UpdateTokenParams{
			UserID:       user.UserID,
			AccessToken:  encryptedAccessToken,
			RefreshToken: encryptedRefreshToken,
			TokenExpiry:  newToken.Expiry,
			UpdatedAt:    time.Now(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update tokens in DB: %v", err)
		}
	}

	// Create an authenticated HTTP client
	client := oauth2Config.Client(context.Background(), newToken)
	return client, nil
}
