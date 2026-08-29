package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

func (s *Server) handleListPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := s.store.ListPlayers(r.Context())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if players == nil {
		players = []domain.Player{}
	}
	writeJSON(w, http.StatusOK, players)
}

type createPlayerRequest struct {
	Name     string  `json:"name"`
	Number   *int    `json:"number"`
	Position *string `json:"position"`
	TeamID   *int64  `json:"team_id"`
	DeviceID string  `json:"device_id"`
}

func (s *Server) handleCreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req createPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" || req.DeviceID == "" {
		writeError(w, http.StatusUnprocessableEntity, "name and device_id are required")
		return
	}
	created, err := s.store.CreatePlayer(r.Context(), domain.Player{
		Name:     req.Name,
		Number:   req.Number,
		Position: req.Position,
		TeamID:   req.TeamID,
		DeviceID: req.DeviceID,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleDeletePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid player id")
		return
	}
	if err := s.store.DeletePlayer(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePlayerSessions(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid player id")
		return
	}
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	sessions, err := s.store.ListSessions(r.Context(), id, limit, offset)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if sessions == nil {
		sessions = []domain.SessionSummary{}
	}
	writeJSON(w, http.StatusOK, sessions)
}
