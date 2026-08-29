package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/sadovod04/sports-tracker/internal/domain"
)

// InsertRawBatch bulk-inserts fused samples using COPY.
func (s *Store) InsertRawBatch(ctx context.Context, samples []domain.RawSample) (int64, error) {
	if len(samples) == 0 {
		return 0, nil
	}
	rows := make([][]any, len(samples))
	for i, s := range samples {
		var hr any
		if s.HeartRate != nil {
			hr = *s.HeartRate
		}
		rows[i] = []any{
			s.SessionID, s.PlayerID, s.Timestamp,
			s.Latitude, s.Longitude, s.Altitude, s.GPSSpeed, s.GPSAccuracy,
			s.AccelX, s.AccelY, s.AccelZ,
			s.GyroX, s.GyroY, s.GyroZ, hr,
		}
	}
	return s.pool.CopyFrom(ctx,
		pgx.Identifier{"raw_data"},
		[]string{
			"session_id", "player_id", "timestamp",
			"latitude", "longitude", "altitude", "gps_speed", "gps_accuracy",
			"accel_x", "accel_y", "accel_z",
			"gyro_x", "gyro_y", "gyro_z", "heart_rate",
		},
		pgx.CopyFromRows(rows),
	)
}

// LoadSessionSamples returns every raw sample for a session, time-ordered.
func (s *Store) LoadSessionSamples(ctx context.Context, sessionID int64) ([]domain.RawSample, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT session_id, player_id, timestamp, latitude, longitude, altitude,
		        gps_speed, gps_accuracy, accel_x, accel_y, accel_z,
		        gyro_x, gyro_y, gyro_z, heart_rate
		 FROM raw_data WHERE session_id = $1 ORDER BY timestamp`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.RawSample
	for rows.Next() {
		var r domain.RawSample
		if err := rows.Scan(
			&r.SessionID, &r.PlayerID, &r.Timestamp, &r.Latitude, &r.Longitude, &r.Altitude,
			&r.GPSSpeed, &r.GPSAccuracy, &r.AccelX, &r.AccelY, &r.AccelZ,
			&r.GyroX, &r.GyroY, &r.GyroZ, &r.HeartRate,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
