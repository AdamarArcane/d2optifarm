package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"

	"github.com/adamararcane/d2optifarm/backend/internal/database"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type apiConfig struct {
	DB            *database.Queries
	API_KEY       string
	CLIENT_ID     string
	CLIENT_SECRET string
}

var oauth2Config *oauth2.Config

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURL := os.Getenv("REDIRECT_URL")
	sessionKey := os.Getenv("SESSION_KEY")
	apiKey := os.Getenv("API_KEY")

	if clientID == "" || clientSecret == "" || redirectURL == "" || sessionKey == "" || apiKey == "" {
		log.Fatal("Missing required environment variables")
	}

	// Initialize oauth2Config
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		/// Scopes:       []string{"ReadBasicUserProfile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.bungie.net/en/OAuth/Authorize",
			TokenURL: "https://www.bungie.net/platform/app/oauth/token/",
		},
	}

	// Initialize apiConfig
	apiCfg := apiConfig{
		API_KEY:       apiKey,
		CLIENT_ID:     clientID,
		CLIENT_SECRET: clientSecret,
	}

	// Set up database connection if needed
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			log.Fatal(err)
		}
		dbQueries := database.New(db)
		apiCfg.DB = dbQueries
		log.Println("Connected to database!")
	}

	// Set up router
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/", handleMain)

	v1Router := chi.NewRouter()

	v1Router.Get("/auth/login", handleLogin)
	v1Router.Get("/auth/callback", apiCfg.handleCallback)

	router.Mount("/v1", v1Router)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><body><a href="/v1/auth/login">Login with Bungie</a></body></html>`)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate the authorization URL without the state parameter
	url := oauth2Config.AuthCodeURL("", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (api *apiConfig) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code from the URL
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	// Exchange the code for an access token
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create an HTTP client using the access token
	client := oauth2Config.Client(context.Background(), token)

	// Make a request to the Bungie API
	req, err := http.NewRequest("GET", "https://www.bungie.net/Platform/User/GetMembershipsForCurrentUser/", nil)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("X-API-Key", api.API_KEY)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get membership data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Output the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
