package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/sadovod04/sports-tracker/internal/domain"
)

// ReplaceHeatmap rewrites all heatmap cells for a session in one transaction.
func (s *Store) ReplaceHeatmap(ctx context.Context, hm domain.Heatmap) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	if _, err := tx.Exec(ctx,
		`DELETE FROM heatmap_data WHERE session_id = $1`, hm.SessionID); err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, c := range hm.Cells {
		batch.Queue(
			`INSERT INTO heatmap_data
			 (session_id, player_id, grid_x, grid_y, time_seconds, avg_speed, intensity_level)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			hm.SessionID, hm.PlayerID, c.GridX, c.GridY,
			c.TimeSeconds, c.AvgSpeedKmh, string(c.IntensityLevel))
	}
	br := tx.SendBatch(ctx, batch)
	for range hm.Cells {
		if _, err := br.Exec(); err != nil {
			_ = br.Close()
			return err
		}
	}
	if err := br.Close(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// GetHeatmap reads a session's stored heatmap cells and geometry.
func (s *Store) GetHeatmap(ctx context.Context, sessionID int64) (domain.Heatmap, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT player_id, grid_x, grid_y, time_seconds, avg_speed, intensity_level
		 FROM heatmap_data WHERE session_id = $1 ORDER BY grid_y, grid_x`, sessionID)
	if err != nil {
		return domain.Heatmap{}, err
	}
	defer rows.Close()

	hm := domain.Heatmap{SessionID: sessionID}
	for rows.Next() {
		var c domain.HeatmapCell
		var lvl string
		if err := rows.Scan(&hm.PlayerID, &c.GridX, &c.GridY, &c.TimeSeconds, &c.AvgSpeedKmh, &lvl); err != nil {
			return domain.Heatmap{}, err
		}
		c.IntensityLevel = domain.IntensityLevel(lvl)
		hm.Cells = append(hm.Cells, c)
		if c.TimeSeconds > hm.MaxSeconds {
			hm.MaxSeconds = c.TimeSeconds
		}
		if c.GridX+1 > hm.GridCols {
			hm.GridCols = c.GridX + 1
		}
		if c.GridY+1 > hm.GridRows {
			hm.GridRows = c.GridY + 1
		}
	}
	return hm, rows.Err()
}
