// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime settings for the backend service.
type Config struct {
	APIPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	RedisURL string

	JWTSecret string
	// IngestAPIKey authorises ESP32 trackers on POST /api/data/ingest.
	IngestAPIKey string

	// Domain parameters for the metric processor.
	SprintSpeedKmh  float64 // speed threshold that counts as a sprint
	FieldLengthM    float64 // pitch length (X axis)
	FieldWidthM     float64 // pitch width (Y axis)
	HeatmapCellM    float64 // heatmap grid cell size, metres
	SampleRateHz    int     // expected tracker sample rate
	ShutdownTimeout time.Duration
}

// DatabaseURL renders a libpq/pgx connection string.
func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// Load reads configuration, applying defaults. A local .env file is loaded if present.
func Load() Config {
	_ = godotenv.Load()

	return Config{
		APIPort:         env("API_PORT", "8080"),
		DBHost:          env("DB_HOST", "localhost"),
		DBPort:          env("DB_PORT", "5432"),
		DBName:          env("DB_NAME", "sports_tracker"),
		DBUser:          env("DB_USER", "tracker_user"),
		DBPassword:      env("DB_PASSWORD", "tracker_pass"),
		RedisURL:        env("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:       env("JWT_SECRET", "dev-secret-change-me"),
		IngestAPIKey:    env("INGEST_API_KEY", "dev-tracker-key"),
		SprintSpeedKmh:  envFloat("SPRINT_SPEED_KMH", 20.0),
		FieldLengthM:    envFloat("FIELD_LENGTH_M", 105.0),
		FieldWidthM:     envFloat("FIELD_WIDTH_M", 68.0),
		HeatmapCellM:    envFloat("HEATMAP_CELL_M", 10.0),
		SampleRateHz:    envInt("SAMPLE_RATE_HZ", 10),
		ShutdownTimeout: 10 * time.Second,
	}
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
