package main

import (
	"database/sql"
	"embed"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"

	"github.com/adamararcane/d2optifarm/backend/internal/database"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type apiConfig struct {
	DB *database.Queries
}

//go:embed static/*
var staticFiles embed.FS

var (
	oauth2Config *oauth2.Config
	store        *sessions.CookieStore
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURL := os.Getenv("REDIRECT_URL")
	sessionKey := os.Getenv("SESSION_KEY")

	if clientID == "" || clientSecret == "" || redirectURL == "" || sessionKey == "" {
		log.Fatal("Missing CLIENT_ID, CLIENT_SECRET, REDIRECT_URL, or sessionKey environment variables")
	}

	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		// Scopes:       []string{"ReadBasicUserProfile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.bungie.net/en/OAuth/Authorize",
			TokenURL: "https://www.bungie.net/platform/app/oauth/token/",
		},
	}

	key := []byte(sessionKey)
	store = sessions.NewCookieStore(key)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // Session expires after 7 days
		HttpOnly: true,      // Prevent JavaScript access
		Secure:   false,     // Set to true in production
		SameSite: http.SameSiteLaxMode,
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	apiCfg := apiConfig{}

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

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := staticFiles.Open("static/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	v1Router := chi.NewRouter()

	if apiCfg.DB != nil {
		v1Router.Get("/auth/login", handleLogin)
		v1Router.Get("/auth/callback", apiCfg.handleCallback)
		v1Router.With(AuthMiddleware).Get("/api/inventory", apiCfg.inventoryHandler)
	}

	v1Router.Get("/healthz", handlerReadiness)

	router.Mount("/v1", v1Router)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
