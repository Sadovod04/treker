package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/sadovod04/sports-tracker/internal/domain"
	"github.com/sadovod04/sports-tracker/internal/live"
	"github.com/sadovod04/sports-tracker/internal/processing"
)

// handleIngest accepts a batch of fused samples from an ESP32 tracker,
// persists them, and refreshes the session's derived metrics + heatmap.
func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	var req domain.IngestRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx := r.Context()
	player, err := s.store.GetPlayerByDevice(ctx, req.DeviceID)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}
	session, err := s.store.GetOrCreateActiveSession(ctx, player.ID, req.Timestamp)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	samples := make([]domain.RawSample, 0, len(req.Data))
	for _, d := range req.Data {
		samples = append(samples, domain.RawSample{
			SessionID:   session.ID,
			PlayerID:    player.ID,
			Timestamp:   req.Timestamp.Add(time.Duration(d.T) * time.Millisecond),
			Latitude:    d.Lat,
			Longitude:   d.Lng,
			Altitude:    d.Alt,
			GPSSpeed:    d.GPSSpeed,
			GPSAccuracy: d.GPSAccuracy,
			AccelX:      d.Accel[0], AccelY: d.Accel[1], AccelZ: d.Accel[2],
			GyroX: d.Gyro[0], GyroY: d.Gyro[1], GyroZ: d.Gyro[2],
			HeartRate: d.HR,
		})
	}

	if _, err := s.store.InsertRawBatch(ctx, samples); err != nil {
		writeStoreError(w, err)
		return
	}
	if err := s.store.TouchDevice(ctx, req.DeviceID, req.Timestamp); err != nil {
		slog.Warn("touch device", "err", err)
	}

	// Recompute in the background so ingest stays fast (<500ms budget).
	go s.recomputeSession(context.WithoutCancel(ctx), session.ID)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":     "ok",
		"session_id": session.ID,
		"accepted":   len(samples),
	})
}

// recomputeSession reloads a session's raw samples, recomputes every metric and
// the heatmap, persists them, and pushes a live update to WebSocket clients.
func (s *Server) recomputeSession(ctx context.Context, sessionID int64) {
	samples, err := s.store.LoadSessionSamples(ctx, sessionID)
	if err != nil || len(samples) == 0 {
		if err != nil {
			slog.Error("recompute: load samples", "session", sessionID, "err", err)
		}
		return
	}

	result := processing.Compute(samples, s.metricParams)
	if err := s.store.UpsertMetrics(ctx, result.Metrics); err != nil {
		slog.Error("recompute: upsert metrics", "session", sessionID, "err", err)
		return
	}

	maxSpeedMS := result.Metrics.SpeedMaxKmh / 3.6
	hm := processing.BuildHeatmap(result.Track, maxSpeedMS, s.heatmapParams)
	hm.SessionID = sessionID
	hm.PlayerID = result.Metrics.PlayerID
	if err := s.store.ReplaceHeatmap(ctx, hm); err != nil {
		slog.Error("recompute: replace heatmap", "session", sessionID, "err", err)
	}

	s.hub.Publish(sessionID, live.Message{Type: "metrics", Data: result.Metrics})
}
