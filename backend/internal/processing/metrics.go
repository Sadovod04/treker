package processing

import (
	"math"
	"sort"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

// Params tune the metric processor. Zero values fall back to sane defaults.
type Params struct {
	SprintSpeedKmh   float64 // speed that counts as a sprint (default 20)
	SprintMinSeconds float64 // minimum duration above threshold (default 1)
	AccelThreshMS2   float64 // forward accel that counts as an "acceleration" (default 3)
	GravityMS2       float64 // local gravity for jump/accel de-biasing (default 9.81)
	ProcessNoise     float64 // Kalman process noise (default 0.5)
}

func (p Params) withDefaults() Params {
	if p.SprintSpeedKmh == 0 {
		p.SprintSpeedKmh = 20
	}
	if p.SprintMinSeconds == 0 {
		p.SprintMinSeconds = 1
	}
	if p.AccelThreshMS2 == 0 {
		p.AccelThreshMS2 = 3
	}
	if p.GravityMS2 == 0 {
		p.GravityMS2 = 9.81
	}
	if p.ProcessNoise == 0 {
		p.ProcessNoise = 0.5
	}
	return p
}

// Result bundles the aggregate metrics with the filtered track, which the
// heatmap builder consumes without re-running the filter.
type Result struct {
	Metrics domain.Metrics
	Track   []LocalPoint
}

// Compute runs the Kalman filter over the session's raw samples and derives
// every per-session metric. samples must be sorted by timestamp ascending.
func Compute(samples []domain.RawSample, p Params) Result {
	p = p.withDefaults()
	res := Result{}
	if len(samples) == 0 {
		return res
	}

	lat0, lng0 := samples[0].Latitude, samples[0].Longitude
	t0 := samples[0].Timestamp
	kf := NewKalman2D(p.ProcessNoise)

	track := make([]LocalPoint, 0, len(samples))
	var (
		prevTS   float64
		distM    float64
		speedSum float64
		speedN   int
		speedMax float64
		hrSum    int
		hrN      int
		hrMax    = math.MinInt32
		hrMin    = math.MaxInt32
		accelCnt int
		jumpCnt  int
		jumpMax  float64 // metres
	)

	sprintThreshMS := p.SprintSpeedKmh / 3.6
	var sprintActive bool
	var sprintStartTS, sprintStartDist float64
	var sprintCount int
	var sprintLengths []float64

	for i, s := range samples {
		ts := s.Timestamp.Sub(t0).Seconds()
		dt := ts - prevTS
		mx, my := geoToLocal(lat0, lng0, s.Latitude, s.Longitude)

		accelMag := math.Abs(math.Sqrt(s.AccelX*s.AccelX+s.AccelY*s.AccelY+s.AccelZ*s.AccelZ) - p.GravityMS2)
		px, py, vx, vy := kf.Update(dt, mx, my, clampAccuracy(s.GPSAccuracy), accelMag)

		pt := LocalPoint{TS: ts, X: px, Y: py, VX: vx, VY: vy}
		track = append(track, pt)

		spd := pt.Speed() // m/s
		if i > 0 {
			step := math.Hypot(px-track[i-1].X, py-track[i-1].Y)
			// Reject teleports from GPS glitches.
			if step < 12 {
				distM += step
			}
		}
		speedSum += spd
		speedN++
		if spd > speedMax {
			speedMax = spd
		}

		// Accelerations: sharp positive change in speed.
		if i > 0 && dt > 0 {
			dvdt := (spd - track[i-1].Speed()) / dt
			if dvdt >= p.AccelThreshMS2 {
				accelCnt++
			}
		}

		// Sprints: sustained time above the speed threshold.
		switch {
		case spd >= sprintThreshMS && !sprintActive:
			sprintActive = true
			sprintStartTS = ts
			sprintStartDist = distM
		case spd < sprintThreshMS && sprintActive:
			sprintActive = false
			if ts-sprintStartTS >= p.SprintMinSeconds {
				sprintCount++
				sprintLengths = append(sprintLengths, distM-sprintStartDist)
			}
		}

		// Jump height from vertical accel spike: h = v^2 / 2g, with takeoff
		// velocity estimated from the peak upward acceleration over dt.
		if dt > 0 {
			up := s.AccelZ - p.GravityMS2
			if up > 6 { // strong upward impulse
				v := up * dt
				h := (v * v) / (2 * p.GravityMS2)
				if h > 0.03 {
					jumpCnt++
					if h > jumpMax {
						jumpMax = h
					}
				}
			}
		}

		if s.HeartRate != nil {
			hr := *s.HeartRate
			hrSum += hr
			hrN++
			if hr > hrMax {
				hrMax = hr
			}
			if hr < hrMin {
				hrMin = hr
			}
		}

		prevTS = ts
	}
	if sprintActive && prevTS-sprintStartTS >= p.SprintMinSeconds {
		sprintCount++
		sprintLengths = append(sprintLengths, distM-sprintStartDist)
	}

	durationS := samples[len(samples)-1].Timestamp.Sub(t0).Seconds()
	zones := intensityZones(track, speedMax)

	m := domain.Metrics{
		SessionID:         samples[0].SessionID,
		PlayerID:          samples[0].PlayerID,
		SpeedMaxKmh:       speedMax * 3.6,
		SpeedAvgKmh:       avg(speedSum, speedN) * 3.6,
		SpeedMinKmh:       0,
		DistanceTotalM:    distM,
		SprintCount:       sprintCount,
		SprintAvgLength:   mean(sprintLengths),
		SprintMaxLength:   maxOf(sprintLengths),
		AccelerationCount: accelCnt,
		PlayerLoad:        playerLoad(track, durationS),
		Zones:             zones,
		JumpHeightMaxCm:   jumpMax * 100,
		JumpCount:         jumpCnt,
		DurationMinutes:   int(math.Round(durationS / 60)),
	}
	if hrN > 0 {
		avgHR := int(math.Round(float64(hrSum) / float64(hrN)))
		m.HRAvg = &avgHR
		m.HRMax = &hrMax
		m.HRMin = &hrMin
	}

	res.Metrics = m
	res.Track = track
	return res
}

// intensityZones sums the seconds spent in each speed band relative to the
// session max speed: low 0-70%, medium 70-85%, high 85-100%.
func intensityZones(track []LocalPoint, maxSpeedMS float64) domain.IntensityZones {
	var z domain.IntensityZones
	if len(track) < 2 || maxSpeedMS <= 0 {
		return z
	}
	// Accumulate as float — sample gaps are ~0.1s and would each round to 0.
	var low, medium, high float64
	for i := 1; i < len(track); i++ {
		dt := track[i].TS - track[i-1].TS
		if dt <= 0 || dt > 5 {
			continue
		}
		frac := track[i].Speed() / maxSpeedMS
		switch {
		case frac >= 0.85:
			high += dt
		case frac >= 0.70:
			medium += dt
		default:
			low += dt
		}
	}
	z.LowSeconds = int(math.Round(low))
	z.MediumSeconds = int(math.Round(medium))
	z.HighSeconds = int(math.Round(high))
	return z
}

// playerLoad approximates Catapult-style PlayerLoad: the time-integrated
// magnitude of the acceleration vector derived from the filtered velocity.
func playerLoad(track []LocalPoint, durationS float64) float64 {
	if len(track) < 3 {
		return 0
	}
	var load float64
	for i := 2; i < len(track); i++ {
		dt := track[i].TS - track[i-1].TS
		if dt <= 0 || dt > 5 {
			continue
		}
		ax := (track[i].VX - track[i-1].VX) / dt
		ay := (track[i].VY - track[i-1].VY) / dt
		load += math.Hypot(ax, ay) * dt
	}
	// Scale to a friendly 0-1000-ish range (÷10), matching the TЗ examples.
	return load / 10
}

func clampAccuracy(acc float64) float64 {
	if acc <= 0 {
		return 3.5
	}
	if acc > 30 {
		return 30
	}
	return acc
}

func avg(sum float64, n int) float64 {
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func maxOf(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	return sorted[len(sorted)-1]
}
