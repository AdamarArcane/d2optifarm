package main

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, "Error getting session", http.StatusInternalServerError)
			return
		}

		userID, ok := session.Values["user_id"].(string)
		if !ok || userID == "" {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		// Proceed to the next handler
		next.ServeHTTP(w, r)
	})
}
