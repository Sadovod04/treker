// Package domain contains the core data types shared across the backend.
package domain

import "time"

// ConnectionStatus reflects the last known state of a tracker.
type ConnectionStatus string

const (
	StatusOnline       ConnectionStatus = "online"
	StatusOffline      ConnectionStatus = "offline"
	StatusDisconnected ConnectionStatus = "disconnected"
)

// IntensityLevel buckets a sample by share of the player's max speed.
type IntensityLevel string

const (
	IntensityLow    IntensityLevel = "low"    // 0-70% of max speed
	IntensityMedium IntensityLevel = "medium" // 70-85%
	IntensityHigh   IntensityLevel = "high"   // 85-100%
)

// Player is a tracked athlete.
type Player struct {
	ID         int64            `json:"id"`
	Name       string           `json:"name"`
	Number     *int             `json:"number,omitempty"`
	Position   *string          `json:"position,omitempty"`
	TeamID     *int64           `json:"team_id,omitempty"`
	DeviceID   string           `json:"device_id"`
	Status     ConnectionStatus `json:"status"`
	LastSeenAt *time.Time       `json:"last_seen_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// Session is a single training or match window for one player.
type Session struct {
	ID              int64      `json:"id"`
	PlayerID        int64      `json:"player_id"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	SessionType     string     `json:"session_type"` // training | match | recovery
	Status          string     `json:"status"`       // active | finished
	CreatedAt       time.Time  `json:"created_at"`
}

// SessionSummary is a compact row for the session-history list.
type SessionSummary struct {
	ID              int64     `json:"id"`
	StartTime       time.Time `json:"start_time"`
	DurationMinutes int       `json:"duration_minutes"`
	DistanceKm      float64   `json:"distance_km"`
	PlayerLoad      float64   `json:"player_load"`
}

// RawSample is one fused GPS + IMU reading persisted from a tracker payload.
type RawSample struct {
	SessionID   int64
	PlayerID    int64
	Timestamp   time.Time
	Latitude    float64
	Longitude   float64
	Altitude    float64
	GPSSpeed    float64 // km/h reported by the GPS module
	GPSAccuracy float64 // metres
	AccelX      float64
	AccelY      float64
	AccelZ      float64
	GyroX       float64
	GyroY       float64
	GyroZ       float64
	HeartRate   *int
}

// IntensityZones holds the seconds spent in each speed band.
type IntensityZones struct {
	LowSeconds    int `json:"low_seconds"`
	MediumSeconds int `json:"medium_seconds"`
	HighSeconds   int `json:"high_seconds"`
}

// Metrics is the per-session aggregate shown on the dashboard and profile.
type Metrics struct {
	SessionID int64 `json:"session_id"`
	PlayerID  int64 `json:"player_id"`

	SpeedMaxKmh float64 `json:"speed_max_kmh"`
	SpeedAvgKmh float64 `json:"speed_avg_kmh"`
	SpeedMinKmh float64 `json:"speed_min_kmh"`

	DistanceTotalM float64 `json:"distance_total_m"`

	SprintCount     int     `json:"sprint_count"`
	SprintAvgLength float64 `json:"sprint_avg_length_m"`
	SprintMaxLength float64 `json:"sprint_max_length_m"`

	AccelerationCount int `json:"acceleration_count"`

	PlayerLoad float64 `json:"player_load"`

	Zones IntensityZones `json:"zones"`

	JumpHeightMaxCm float64 `json:"jump_height_max_cm"`
	JumpCount       int     `json:"jump_count"`

	HRMax *int `json:"hr_max,omitempty"`
	HRAvg *int `json:"hr_avg,omitempty"`
	HRMin *int `json:"hr_min,omitempty"`

	DurationMinutes int       `json:"duration_minutes"`
	ComputedAt      time.Time `json:"computed_at"`
}

// HeatmapCell is one grid bucket of the pitch.
type HeatmapCell struct {
	GridX          int            `json:"grid_x"`
	GridY          int            `json:"grid_y"`
	TimeSeconds    int            `json:"time_seconds"`
	AvgSpeedKmh    float64        `json:"avg_speed_kmh"`
	IntensityLevel IntensityLevel `json:"intensity_level"`
}

// Heatmap is the full grid plus its geometry, ready for the frontend.
type Heatmap struct {
	SessionID  int64         `json:"session_id"`
	PlayerID   int64         `json:"player_id"`
	GridCols   int           `json:"grid_cols"`
	GridRows   int           `json:"grid_rows"`
	CellSizeM  float64       `json:"cell_size_m"`
	Cells      []HeatmapCell `json:"cells"`
	MaxSeconds int           `json:"max_seconds"`
}

// ComparisonRow is one metric compared between two players.
type ComparisonRow struct {
	Metric      string  `json:"metric"`
	Unit        string  `json:"unit"`
	Player1     float64 `json:"player1"`
	Player2     float64 `json:"player2"`
	Diff        float64 `json:"diff"`
	DiffPercent float64 `json:"diff_percent"`
}

// Comparison is the payload for GET /api/compare.
type Comparison struct {
	SessionID int64           `json:"session_id"`
	Player1   *Metrics        `json:"player1"`
	Player2   *Metrics        `json:"player2"`
	Rows      []ComparisonRow `json:"rows"`
}
