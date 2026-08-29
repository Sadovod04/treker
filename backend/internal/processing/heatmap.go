package processing

import (
	"math"

	"github.com/sadovod04/sports-tracker/internal/domain"
)

// HeatmapParams describe the pitch geometry for gridding.
type HeatmapParams struct {
	FieldLengthM float64 // X extent
	FieldWidthM  float64 // Y extent
	CellSizeM    float64 // square cell edge
}

// BuildHeatmap buckets a filtered track into a grid, accumulating dwell time
// and average speed per cell. The track's local frame is assumed to be roughly
// axis-aligned with the pitch and centred; points are shifted so the pitch
// spans [0,length] x [0,width].
func BuildHeatmap(track []LocalPoint, maxSpeedMS float64, p HeatmapParams) domain.Heatmap {
	if p.CellSizeM <= 0 {
		p.CellSizeM = 10
	}
	if p.FieldLengthM <= 0 {
		p.FieldLengthM = 105
	}
	if p.FieldWidthM <= 0 {
		p.FieldWidthM = 68
	}

	cols := int(math.Ceil(p.FieldLengthM / p.CellSizeM))
	rows := int(math.Ceil(p.FieldWidthM / p.CellSizeM))
	hm := domain.Heatmap{
		GridCols:  cols,
		GridRows:  rows,
		CellSizeM: p.CellSizeM,
	}
	if len(track) < 2 {
		return hm
	}

	// Centre the track over the pitch.
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, pt := range track {
		minX = math.Min(minX, pt.X)
		minY = math.Min(minY, pt.Y)
		maxX = math.Max(maxX, pt.X)
		maxY = math.Max(maxY, pt.Y)
	}
	offX := (p.FieldLengthM-(maxX-minX))/2 - minX
	offY := (p.FieldWidthM-(maxY-minY))/2 - minY

	type acc struct {
		seconds  float64
		speedSum float64
		n        int
	}
	grid := make(map[[2]int]*acc)

	for i := 1; i < len(track); i++ {
		dt := track[i].TS - track[i-1].TS
		if dt <= 0 || dt > 5 {
			continue
		}
		x := track[i].X + offX
		y := track[i].Y + offY
		gx := clampInt(int(x/p.CellSizeM), 0, cols-1)
		gy := clampInt(int(y/p.CellSizeM), 0, rows-1)
		key := [2]int{gx, gy}
		a := grid[key]
		if a == nil {
			a = &acc{}
			grid[key] = a
		}
		a.seconds += dt
		a.speedSum += track[i].Speed()
		a.n++
	}

	for key, a := range grid {
		secs := int(math.Round(a.seconds))
		avgSpeed := 0.0
		if a.n > 0 {
			avgSpeed = a.speedSum / float64(a.n)
		}
		hm.Cells = append(hm.Cells, domain.HeatmapCell{
			GridX:          key[0],
			GridY:          key[1],
			TimeSeconds:    secs,
			AvgSpeedKmh:    avgSpeed * 3.6,
			IntensityLevel: bucket(avgSpeed, maxSpeedMS),
		})
		if secs > hm.MaxSeconds {
			hm.MaxSeconds = secs
		}
	}
	return hm
}

func bucket(speedMS, maxMS float64) domain.IntensityLevel {
	if maxMS <= 0 {
		return domain.IntensityLow
	}
	switch frac := speedMS / maxMS; {
	case frac >= 0.85:
		return domain.IntensityHigh
	case frac >= 0.70:
		return domain.IntensityMedium
	default:
		return domain.IntensityLow
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
