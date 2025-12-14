package main

// @title Music Server API
// @version 1.0
// @description Local network API for music management
// @host localhost:8080
// @BasePath /api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "bottomley.ian/musicserver/docs"
	_ "modernc.org/sqlite"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"bottomley.ian/musicserver/internal/app"
	"bottomley.ian/musicserver/internal/db"
	"bottomley.ian/musicserver/internal/handlers"
	"bottomley.ian/musicserver/internal/services/fs"
	"bottomley.ian/musicserver/internal/services/scanner"
	"bottomley.ian/musicserver/internal/store"
)

func main() {
	dbPath := getenv("DB_PATH", "./data.sqlite")

	sqlite, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer sqlite.Close()

	// pragmatic defaults for a local server
	if _, err := sqlite.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		log.Fatal(err)
	}
	if _, err := sqlite.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		log.Fatal(err)
	}
	if _, err := sqlite.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		log.Fatal(err)
	}

	if err := store.ApplyMigrations(sqlite); err != nil {
		log.Fatal(err)
	}

	a := &app.App{
		DB:      sqlite,
		Queries: db.New(sqlite),
	}
	s := scanner.New(a.Queries, fs.OSFS{})
	h := handlers.New(a, s)
	r := chi.NewRouter()

	// global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// api
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", h.Health)

		r.Route("/folders", func(r chi.Router) {
			r.Get("/", h.ListFolders)
			r.Get("/{id}", h.GetFolder)
			r.Delete("/{id}", h.DeleteFolder)
			r.Post("/{id}/scan", h.ScanFolder)
			r.Get("/{id}/scan", h.ScanStatus)
			r.Post("/", h.CreateFolder)
		})
		r.Route("/tracks", func(r chi.Router) {
			r.Get("/", h.ListTracks)
			r.Get("/{id}", h.GetTrack)
			r.Put("/{id}", h.UpdateTrack)
			r.Get("/{id}/play", h.StreamTrack)
			r.Get("/{id}/download", h.DownloadTrack)
		})
		r.Route("/artists", func(r chi.Router) {
			r.Get("/", h.ListArtists)
			r.Get("/{id}", h.GetArtist)
			r.Put("/{id}", h.UpdateArtist)
			r.Delete("/{id}", h.DeleteArtist)
		})
		r.Route("/albums", func(r chi.Router) {
			r.Get("/", h.ListAlbums)
			r.Get("/{id}", h.GetAlbum)
			r.Put("/{id}", h.UpdateAlbum)
			r.Delete("/{id}", h.DeleteAlbum)
		})
	})

	addr := ":8080"
	log.Printf("Listening on http://0.0.0.0%s (db=%s)\n", addr, dbPath)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
