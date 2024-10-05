package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/adamararcane/d2optifarm/backend/internal/auth"
	"github.com/adamararcane/d2optifarm/backend/internal/database"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := GenerateStateToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Falied to generate state token.")
		return
	}

	session, err := store.Get(r, "session-name")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get session")
		return
	}

	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save session")
		return
	}

	url := oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (cfg *apiConfig) handleCallback(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Error getting session", http.StatusInternalServerError)
		return
	}

	savedState, ok := session.Values["state"].(string)
	if !ok || savedState == "" {
		http.Error(w, "Session state missing", http.StatusBadRequest)
		return
	}

	receivedState := r.FormValue("state")
	if receivedState != savedState {
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	delete(session.Values, "state")
	session.Save(r, w)

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauth2Config.Client(context.Background(), token)

	membershipID, membershipType, err := GetMembershipData(client)
	if err != nil {
		http.Error(w, "Failed to get membership ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	encryptedAccessToken, err := auth.Encrypt(token.AccessToken)
	if err != nil {
		fmt.Println("Error encypting access token")
		return
	}
	encryptedRefreshToken, err := auth.Encrypt(token.RefreshToken)
	if err != nil {
		fmt.Println("Error encypting refresh token")
		return
	}

	createUserParams := database.CreateUserParams{
		UserID:         membershipID,
		MembershipType: int64(membershipType),
		AccessToken:    encryptedAccessToken,
		RefreshToken:   encryptedRefreshToken,
		TokenExpiry:    token.Expiry,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = cfg.DB.CreateUser(context.Background(), createUserParams)
	if err != nil {
		fmt.Printf("Error creating user in DB: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create user (user already has account)")
		return
	}

	session.Values["user_id"] = membershipID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
