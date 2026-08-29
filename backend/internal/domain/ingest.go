package domain

import (
	"errors"
	"time"
)

// IngestRequest is the body of POST /api/data/ingest sent by an ESP32 tracker.
type IngestRequest struct {
	DeviceID  string         `json:"device_id"`
	Timestamp time.Time      `json:"timestamp"`
	Data      []IngestSample `json:"data"`
}

// IngestSample is one point inside an ingest batch.
type IngestSample struct {
	// T is milliseconds since the batch timestamp (monotonic device clock).
	T           int64      `json:"t"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	Alt         float64    `json:"alt"`
	GPSSpeed    float64    `json:"gps_speed"`
	GPSAccuracy float64    `json:"gps_accuracy"`
	Accel       [3]float64 `json:"accel"`
	Gyro        [3]float64 `json:"gyro"`
	HR          *int       `json:"hr,omitempty"`
}

// Validate performs cheap structural checks before the batch is queued.
func (r IngestRequest) Validate() error {
	if r.DeviceID == "" {
		return errors.New("device_id is required")
	}
	if len(r.Data) == 0 {
		return errors.New("data must contain at least one sample")
	}
	if len(r.Data) > 5000 {
		return errors.New("batch too large (max 5000 samples)")
	}
	for i, s := range r.Data {
		if s.Lat < -90 || s.Lat > 90 || s.Lng < -180 || s.Lng > 180 {
			return &SampleError{Index: i, Msg: "coordinates out of range"}
		}
	}
	return nil
}

// SampleError points at a bad sample inside a batch.
type SampleError struct {
	Index int
	Msg   string
}

func (e *SampleError) Error() string { return e.Msg }
