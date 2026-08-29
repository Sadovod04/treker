package httpapi

import (
	"math"
	"net/http"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	p1, ok1 := queryInt64(r, "player1_id")
	p2, ok2 := queryInt64(r, "player2_id")
	if !ok1 || !ok2 {
		writeError(w, http.StatusBadRequest, "player1_id and player2_id are required")
		return
	}

	ctx := r.Context()
	sharedSession, hasShared := queryInt64(r, "session_id")

	load := func(playerID int64) (*domain.Metrics, error) {
		sid := sharedSession
		if !hasShared {
			var err error
			sid, err = s.store.LatestSessionID(ctx, playerID)
			if err != nil {
				return nil, err
			}
		}
		m, err := s.store.GetMetrics(ctx, sid)
		if err != nil {
			return nil, err
		}
		return &m, nil
	}

	m1, err := load(p1)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	m2, err := load(p2)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	resp := domain.Comparison{
		Player1: m1,
		Player2: m2,
		Rows:    compareRows(m1, m2),
	}
	if hasShared {
		resp.SessionID = sharedSession
	}
	writeJSON(w, http.StatusOK, resp)
}

func compareRows(a, b *domain.Metrics) []domain.ComparisonRow {
	type spec struct {
		name, unit string
		get        func(*domain.Metrics) float64
	}
	specs := []spec{
		{"Speed (Max)", "км/ч", func(m *domain.Metrics) float64 { return m.SpeedMaxKmh }},
		{"Speed (Avg)", "км/ч", func(m *domain.Metrics) float64 { return m.SpeedAvgKmh }},
		{"Distance", "км", func(m *domain.Metrics) float64 { return m.DistanceTotalM / 1000 }},
		{"Sprints", "шт", func(m *domain.Metrics) float64 { return float64(m.SprintCount) }},
		{"Accelerations", "шт", func(m *domain.Metrics) float64 { return float64(m.AccelerationCount) }},
		{"Player Load", "ед.", func(m *domain.Metrics) float64 { return m.PlayerLoad }},
		{"Jump Height", "см", func(m *domain.Metrics) float64 { return m.JumpHeightMaxCm }},
		{"Duration", "мин", func(m *domain.Metrics) float64 { return float64(m.DurationMinutes) }},
	}

	rows := make([]domain.ComparisonRow, 0, len(specs))
	for _, sp := range specs {
		v1, v2 := round2(sp.get(a)), round2(sp.get(b))
		diff := round2(v1 - v2)
		pct := 0.0
		if v2 != 0 {
			pct = round2((v1 - v2) / math.Abs(v2) * 100)
		}
		rows = append(rows, domain.ComparisonRow{
			Metric:      sp.name,
			Unit:        sp.unit,
			Player1:     v1,
			Player2:     v2,
			Diff:        diff,
			DiffPercent: pct,
		})
	}
	return rows
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
