package httpapi

import (
	"net/http"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

func (s *Server) handlePlayerHeatmap(w http.ResponseWriter, r *http.Request) {
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

	hm, err := s.store.GetHeatmap(r.Context(), sessionID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if len(hm.Cells) == 0 {
		// Build it on demand if raw data exists.
		s.recomputeSession(r.Context(), sessionID)
		hm, err = s.store.GetHeatmap(r.Context(), sessionID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
	}

	hm.CellSizeM = s.cfg.HeatmapCellM
	if hm.Cells == nil {
		hm.Cells = []domain.HeatmapCell{}
	}
	writeJSON(w, http.StatusOK, hm)
}
