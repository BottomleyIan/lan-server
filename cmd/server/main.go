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
	"os/exec"

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

	requireFFmpeg()

	a := &app.App{
		DB:      sqlite,
		Queries: db.New(sqlite),
		FS:      fs.OSFS{},
	}
	s := scanner.New(a.Queries, a.FS)
	h := handlers.New(a, s)
	r := chi.NewRouter()

	// global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsAllowAll())
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
			r.Patch("/{id}/rating", h.UpdateTrackRating)
			r.Get("/{id}/play", h.StreamTrack)
			r.Get("/{id}/download", h.DownloadTrack)
			r.Get("/{id}/image", h.GetTrackImage)
			r.Post("/{id}/image", h.UpdateTrackImage)
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
			r.Get("/{id}/tracks", h.ListAlbumTracks)
			r.Put("/{id}", h.UpdateAlbum)
			r.Delete("/{id}", h.DeleteAlbum)
			r.Get("/{id}/image", h.GetAlbumImage)
		})
		r.Route("/calendar", func(r chi.Router) {
			r.Get("/", h.GetCalendarDay)
		})
		r.Route("/journals", func(r chi.Router) {
			r.Get("/", h.ListJournals)
			r.Get("/property-keys", h.ListJournalPropertyKeys)
			r.Get("/property-keys/{key}/values", h.ListJournalPropertyValues)
			r.Route("/entries", func(r chi.Router) {
				r.Get("/", h.ListJournalEntries)
				r.Post("/", h.CreateJournalEntry)
				r.Put("/{year}/{month}/{day}/{position}", h.UpdateJournalEntryByPosition)
				r.Put("/{year}/{month}/{day}/{position}/{status}", h.UpdateJournalEntryStatus)
				r.Delete("/{year}/{month}/{day}/{hash}", h.DeleteJournalEntryByHash)
			})
			r.Get("/assets", h.GetJournalAsset)
			r.Post("/assets", h.UploadJournalAsset)
			r.Get("/tags/graph", h.ListJournalTagGraph)
			r.Get("/tags/graph/{tag}", h.GetJournalTagGraph)
			r.Get("/tags", h.ListJournalTags)
			r.Get("/{year}/{month}", h.ListJournalsByMonth)
			r.Get("/{year}/{month}/{day}", h.GetJournalDay)
		})
		r.Route("/prices", func(r chi.Router) {
			r.Get("/metals", h.GetMetalPrices)
		})
		r.Route("/settings", func(r chi.Router) {
			r.Get("/", h.ListSettings)
			r.Post("/", h.CreateSetting)
			r.Get("/keys", h.ListSettingKeys)
			r.Route("/{key}", func(r chi.Router) {
				r.Get("/", h.GetSetting)
				r.Put("/", h.UpdateSetting)
				r.Delete("/", h.DeleteSetting)
			})
		})
		r.Route("/playlists", func(r chi.Router) {
			r.Get("/", h.ListPlaylists)
			r.Post("/", h.CreatePlaylist)
			r.Get("/{id}", h.GetPlaylist)
			r.Put("/{id}", h.UpdatePlaylist)
			r.Delete("/{id}", h.DeletePlaylist)
			r.Post("/{id}/clear", h.ClearPlaylist)
			r.Post("/{id}/enqueue", h.EnqueuePlaylistTrack)
			r.Route("/{id}/tracks", func(r chi.Router) {
				r.Get("/", h.ListPlaylistTracks)
				r.Post("/", h.AddPlaylistTrack)
				r.Put("/{track_id}", h.UpdatePlaylistTrack)
				r.Delete("/{track_id}", h.DeletePlaylistTrack)
			})
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

func requireFFmpeg() {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Panic("ffmpeg not installed or not on PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		log.Panic("ffprobe not installed or not on PATH")
	}
}
