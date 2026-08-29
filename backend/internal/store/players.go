package store

import (
	"context"
	"time"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

const playerCols = `id, name, number, position, team_id, device_id, status,
	last_seen_at, created_at, updated_at`

func scanPlayer(row rowScanner) (domain.Player, error) {
	var p domain.Player
	err := row.Scan(&p.ID, &p.Name, &p.Number, &p.Position, &p.TeamID,
		&p.DeviceID, &p.Status, &p.LastSeenAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

// ListPlayers returns every player ordered by shirt number then name.
func (s *Store) ListPlayers(ctx context.Context) ([]domain.Player, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+playerCols+` FROM players ORDER BY number NULLS LAST, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Player
	for rows.Next() {
		p, err := scanPlayer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetPlayer looks up a single player by id.
func (s *Store) GetPlayer(ctx context.Context, id int64) (domain.Player, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT `+playerCols+` FROM players WHERE id = $1`, id)
	p, err := scanPlayer(row)
	if err != nil {
		return domain.Player{}, mapErr(err)
	}
	return p, nil
}

// GetPlayerByDevice looks up a player by tracker device id.
func (s *Store) GetPlayerByDevice(ctx context.Context, deviceID string) (domain.Player, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT `+playerCols+` FROM players WHERE device_id = $1`, deviceID)
	p, err := scanPlayer(row)
	if err != nil {
		return domain.Player{}, mapErr(err)
	}
	return p, nil
}

// CreatePlayer inserts a new player and returns it.
func (s *Store) CreatePlayer(ctx context.Context, p domain.Player) (domain.Player, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO players (name, number, position, team_id, device_id, status)
		 VALUES ($1,$2,$3,$4,$5,COALESCE(NULLIF($6,''),'offline'))
		 RETURNING `+playerCols,
		p.Name, p.Number, p.Position, p.TeamID, p.DeviceID, string(p.Status))
	created, err := scanPlayer(row)
	if err != nil {
		return domain.Player{}, mapErr(err)
	}
	return created, nil
}

// DeletePlayer removes a player and all dependent rows.
func (s *Store) DeletePlayer(ctx context.Context, id int64) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM players WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// TouchDevice marks a device online and stamps last_seen_at.
func (s *Store) TouchDevice(ctx context.Context, deviceID string, at time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE players SET status = 'online', last_seen_at = $2, updated_at = now()
		 WHERE device_id = $1`, deviceID, at)
	return err
}

// MarkStaleOffline flips players with no data for the cutoff to 'offline'.
func (s *Store) MarkStaleOffline(ctx context.Context, olderThan time.Duration) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE players SET status = 'offline', updated_at = now()
		 WHERE status = 'online'
		   AND (last_seen_at IS NULL OR last_seen_at < now() - $1::interval)`,
		olderThan.String())
	return err
}
