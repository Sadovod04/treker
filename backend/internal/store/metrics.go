package store

import (
	"context"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

// UpsertMetrics writes (or replaces) the aggregate metrics for a session.
func (s *Store) UpsertMetrics(ctx context.Context, m domain.Metrics) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO metrics (
			session_id, player_id, speed_max, speed_avg, speed_min,
			distance_total, sprint_count, sprint_avg_length, sprint_max_length,
			acceleration_count, player_load,
			low_intensity_time, med_intensity_time, high_intensity_time,
			jump_height_max, jump_count, hr_max, hr_avg, hr_min, duration_minutes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		ON CONFLICT (session_id) DO UPDATE SET
			speed_max = EXCLUDED.speed_max,
			speed_avg = EXCLUDED.speed_avg,
			speed_min = EXCLUDED.speed_min,
			distance_total = EXCLUDED.distance_total,
			sprint_count = EXCLUDED.sprint_count,
			sprint_avg_length = EXCLUDED.sprint_avg_length,
			sprint_max_length = EXCLUDED.sprint_max_length,
			acceleration_count = EXCLUDED.acceleration_count,
			player_load = EXCLUDED.player_load,
			low_intensity_time = EXCLUDED.low_intensity_time,
			med_intensity_time = EXCLUDED.med_intensity_time,
			high_intensity_time = EXCLUDED.high_intensity_time,
			jump_height_max = EXCLUDED.jump_height_max,
			jump_count = EXCLUDED.jump_count,
			hr_max = EXCLUDED.hr_max,
			hr_avg = EXCLUDED.hr_avg,
			hr_min = EXCLUDED.hr_min,
			duration_minutes = EXCLUDED.duration_minutes,
			created_at = now()`,
		m.SessionID, m.PlayerID, m.SpeedMaxKmh, m.SpeedAvgKmh, m.SpeedMinKmh,
		m.DistanceTotalM, m.SprintCount, m.SprintAvgLength, m.SprintMaxLength,
		m.AccelerationCount, m.PlayerLoad,
		m.Zones.LowSeconds, m.Zones.MediumSeconds, m.Zones.HighSeconds,
		m.JumpHeightMaxCm, m.JumpCount, m.HRMax, m.HRAvg, m.HRMin, m.DurationMinutes,
	)
	return err
}

// GetMetrics reads the stored metrics for a session.
func (s *Store) GetMetrics(ctx context.Context, sessionID int64) (domain.Metrics, error) {
	var m domain.Metrics
	err := s.pool.QueryRow(ctx,
		`SELECT session_id, player_id, speed_max, speed_avg, speed_min,
		        distance_total, sprint_count, sprint_avg_length, sprint_max_length,
		        acceleration_count, player_load,
		        low_intensity_time, med_intensity_time, high_intensity_time,
		        jump_height_max, jump_count, hr_max, hr_avg, hr_min,
		        duration_minutes, created_at
		 FROM metrics WHERE session_id = $1`, sessionID).Scan(
		&m.SessionID, &m.PlayerID, &m.SpeedMaxKmh, &m.SpeedAvgKmh, &m.SpeedMinKmh,
		&m.DistanceTotalM, &m.SprintCount, &m.SprintAvgLength, &m.SprintMaxLength,
		&m.AccelerationCount, &m.PlayerLoad,
		&m.Zones.LowSeconds, &m.Zones.MediumSeconds, &m.Zones.HighSeconds,
		&m.JumpHeightMaxCm, &m.JumpCount, &m.HRMax, &m.HRAvg, &m.HRMin,
		&m.DurationMinutes, &m.ComputedAt,
	)
	if err != nil {
		return domain.Metrics{}, mapErr(err)
	}
	return m, nil
}
