package httpapi

import (
	"errors"
	"net/http"

	"github.com/sadovod04/sports-tracker/internal/store"
)

// resolveSession picks the session id from ?session_id= or falls back to the
// player's most recent session.
func (s *Server) resolveSession(r *http.Request, playerID int64) (int64, error) {
	if sid, ok := queryInt64(r, "session_id"); ok {
		return sid, nil
	}
	return s.store.LatestSessionID(r.Context(), playerID)
}

func (s *Server) handlePlayerMetrics(w http.ResponseWriter, r *http.Request) {
	playerID, err := pathInt64(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid player id")
		return
	}
	sessionID, err := s.resolveSession(r, playerID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	m, err := s.store.GetMetrics(r.Context(), sessionID)
	if errors.Is(err, store.ErrNotFound) {
		// Metrics not materialised yet — compute synchronously once.
		s.recomputeSession(r.Context(), sessionID)
		m, err = s.store.GetMetrics(r.Context(), sessionID)
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
