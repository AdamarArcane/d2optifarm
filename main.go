package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"

	_ "github.com/tursodatabase/go-libsql"
)

func main() {
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		log.Printf("Error opening db: %s", err)
		return
	}
	defer db.Close()

	dbQueries := database.New(db)

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	fileServerHandler := http.FileServer(http.Dir("."))
	strippedHandler := http.StripPrefix("/app", fileServerHandler)

	mux.Handle("/app/", cfg.siteHitsMiddleware(strippedHandler))
	mux.Handle("/app/assets/logo.png", strippedHandler)

	server.ListenAndServe()
}

// Data Structures

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}
