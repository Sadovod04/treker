package store

import (
	"context"
	"time"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

// GetOrCreateActiveSession returns the player's currently active session,
// creating one starting at `start` if none exists.
func (s *Store) GetOrCreateActiveSession(ctx context.Context, playerID int64, start time.Time) (domain.Session, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO sessions (player_id, start_time, session_type, status)
		 VALUES ($1, $2, 'training', 'active')
		 ON CONFLICT (player_id) WHERE status = 'active'
		 DO UPDATE SET player_id = EXCLUDED.player_id
		 RETURNING id, player_id, start_time, end_time, duration_minutes,
		           session_type, status, created_at`,
		playerID, start)
	return scanSession(row)
}

// GetSession fetches one session by id.
func (s *Store) GetSession(ctx context.Context, id int64) (domain.Session, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, player_id, start_time, end_time, duration_minutes,
		        session_type, status, created_at
		 FROM sessions WHERE id = $1`, id)
	sess, err := scanSession(row)
	if err != nil {
		return domain.Session{}, mapErr(err)
	}
	return sess, nil
}

// LatestSessionID returns the most recent session id for a player.
func (s *Store) LatestSessionID(ctx context.Context, playerID int64) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM sessions WHERE player_id = $1
		 ORDER BY start_time DESC LIMIT 1`, playerID).Scan(&id)
	if err != nil {
		return 0, mapErr(err)
	}
	return id, nil
}

// FinishSession closes a session and records its duration.
func (s *Store) FinishSession(ctx context.Context, id int64, end time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions
		 SET status = 'finished', end_time = $2,
		     duration_minutes = GREATEST(0, EXTRACT(EPOCH FROM ($2 - start_time))::int / 60)
		 WHERE id = $1`, id, end)
	return err
}

// ListSessions returns compact history rows joined with their metrics.
func (s *Store) ListSessions(ctx context.Context, playerID int64, limit, offset int) ([]domain.SessionSummary, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx,
		`SELECT s.id, s.start_time,
		        COALESCE(s.duration_minutes, m.duration_minutes, 0),
		        COALESCE(m.distance_total, 0) / 1000.0,
		        COALESCE(m.player_load, 0)
		 FROM sessions s
		 LEFT JOIN metrics m ON m.session_id = s.id
		 WHERE s.player_id = $1
		 ORDER BY s.start_time DESC
		 LIMIT $2 OFFSET $3`, playerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SessionSummary
	for rows.Next() {
		var r domain.SessionSummary
		if err := rows.Scan(&r.ID, &r.StartTime, &r.DurationMinutes, &r.DistanceKm, &r.PlayerLoad); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanSession(row rowScanner) (domain.Session, error) {
	var s domain.Session
	err := row.Scan(&s.ID, &s.PlayerID, &s.StartTime, &s.EndTime,
		&s.DurationMinutes, &s.SessionType, &s.Status, &s.CreatedAt)
	return s, err
}
