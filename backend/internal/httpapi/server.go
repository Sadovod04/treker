// Package httpapi wires the REST + WebSocket surface described in the TЗ.
package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sadovod04/sports-tracker/internal/config"
	"github.com/sadovod04/sports-tracker/internal/live"
	"github.com/sadovod04/sports-tracker/internal/processing"
	"github.com/sadovod04/sports-tracker/internal/store"
)

// Server holds the shared dependencies for every handler.
type Server struct {
	cfg   config.Config
	store *store.Store
	hub   *live.Hub

	metricParams  processing.Params
	heatmapParams processing.HeatmapParams
}

// NewServer builds a Server.
func NewServer(cfg config.Config, st *store.Store, hub *live.Hub) *Server {
	return &Server{
		cfg:   cfg,
		store: st,
		hub:   hub,
		metricParams: processing.Params{
			SprintSpeedKmh: cfg.SprintSpeedKmh,
		},
		heatmapParams: processing.HeatmapParams{
			FieldLengthM: cfg.FieldLengthM,
			FieldWidthM:  cfg.FieldWidthM,
			CellSizeM:    cfg.HeatmapCellM,
		},
	}
}

// Routes returns the fully-wired HTTP handler.
func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", s.handleLogin)

		// Tracker ingest — authenticated with an API key.
		r.With(s.requireAPIKey).Post("/data/ingest", s.handleIngest)

		// Read APIs — JWT-protected (dev mode allows anonymous, see auth.go).
		r.Group(func(r chi.Router) {
			r.Use(s.optionalAuth)
			r.Get("/players", s.handleListPlayers)
			r.Post("/players", s.handleCreatePlayer)
			r.Delete("/players/{id}", s.handleDeletePlayer)
			r.Get("/players/{id}/metrics", s.handlePlayerMetrics)
			r.Get("/players/{id}/heatmap", s.handlePlayerHeatmap)
			r.Get("/players/{id}/sessions", s.handlePlayerSessions)
			r.Get("/compare", s.handleCompare)
		})
	})

	r.Get("/ws/live", s.handleWSLive)
	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
