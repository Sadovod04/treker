package processing

import (
	"math"
	"testing"
	"time"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

// synthTrack walks a player east at a constant ground speed (m/s) for the given
// number of 10 Hz samples, starting from a fixed lat/lng.
func synthTrack(speedMS float64, n int) []domain.RawSample {
	const lat0, lng0 = 55.75, 37.61
	const dt = 0.1
	metrePerDegLat := 111_320.0
	base := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	out := make([]domain.RawSample, n)
	for i := 0; i < n; i++ {
		dist := speedMS * dt * float64(i)
		out[i] = domain.RawSample{
			SessionID:   1,
			PlayerID:    1,
			Timestamp:   base.Add(time.Duration(float64(i)*dt*float64(time.Second))),
			Latitude:    lat0 + dist/metrePerDegLat,
			Longitude:   lng0,
			Altitude:    100,
			GPSAccuracy: 1.0,
			AccelZ:      9.81,
		}
	}
	return out
}

func TestComputeConstantSpeed(t *testing.T) {
	// 6 m/s ≈ 21.6 km/h for 20 s.
	samples := synthTrack(6, 200)
	res := Compute(samples, Params{})

	if got := res.Metrics.DistanceTotalM; math.Abs(got-120) > 20 {
		t.Errorf("distance: got %.1f m, want ~120 m", got)
	}
	if got := res.Metrics.SpeedMaxKmh; got < 15 || got > 30 {
		t.Errorf("speed max: got %.1f km/h, want ~21.6", got)
	}
	if res.Metrics.DurationMinutes != 0 {
		t.Logf("duration rounds to %d min (expected 0 for 20 s)", res.Metrics.DurationMinutes)
	}
	if len(res.Track) != len(samples) {
		t.Errorf("track length: got %d, want %d", len(res.Track), len(samples))
	}
}

func TestComputeDetectsSprint(t *testing.T) {
	// 7 m/s = 25.2 km/h, above the 20 km/h sprint threshold, held for 5 s.
	samples := synthTrack(7, 50)
	res := Compute(samples, Params{SprintSpeedKmh: 20})
	if res.Metrics.SprintCount == 0 {
		t.Fatalf("expected at least one sprint, got 0 (speed max %.1f)", res.Metrics.SpeedMaxKmh)
	}
}

func TestComputeEmpty(t *testing.T) {
	res := Compute(nil, Params{})
	if res.Metrics.DistanceTotalM != 0 || len(res.Track) != 0 {
		t.Fatalf("empty input should yield zero result")
	}
}

func TestBuildHeatmapBucketsTime(t *testing.T) {
	samples := synthTrack(6, 200)
	res := Compute(samples, Params{})
	hm := BuildHeatmap(res.Track, res.Metrics.SpeedMaxKmh/3.6, HeatmapParams{
		FieldLengthM: 105, FieldWidthM: 68, CellSizeM: 10,
	})
	if len(hm.Cells) == 0 {
		t.Fatal("expected non-empty heatmap")
	}
	var total int
	for _, c := range hm.Cells {
		total += c.TimeSeconds
	}
	if total < 10 || total > 30 {
		t.Errorf("summed dwell time %d s, want ~20 s", total)
	}
}
