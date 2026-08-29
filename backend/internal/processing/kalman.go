package processing

import "math"

// LocalPoint is a pitch-local position in metres relative to a session origin,
// together with the sample time in seconds since session start.
type LocalPoint struct {
	TS float64 // seconds since session start
	X  float64 // metres east of origin
	Y  float64 // metres north of origin
	VX float64 // m/s, filter estimate
	VY float64 // m/s
}

// Speed returns the planar speed in m/s.
func (p LocalPoint) Speed() float64 { return math.Hypot(p.VX, p.VY) }

const earthRadiusM = 6_371_000.0

// geoToLocal projects lat/lng to a flat local frame anchored at (lat0, lng0)
// using the equirectangular approximation — accurate to well under a metre
// across a football pitch.
func geoToLocal(lat0, lng0, lat, lng float64) (x, y float64) {
	rad := math.Pi / 180
	x = (lng - lng0) * rad * earthRadiusM * math.Cos(lat0*rad)
	y = (lat - lat0) * rad * earthRadiusM
	return x, y
}

// Kalman2D is a constant-velocity Kalman filter for planar position.
// State: [x, y, vx, vy]. GPS supplies position; the IMU-derived acceleration
// magnitude is folded into the process noise so hard cuts are tracked without
// lag. It is intentionally compact — good enough for pitch analytics at 10 Hz.
type Kalman2D struct {
	x      [4]float64 // state estimate
	p      [4]float64 // diagonal covariance (x, y, vx, vy)
	q      float64    // base process noise
	inited bool
}

// NewKalman2D builds a filter. processNoise defaults sensibly around 0.5.
func NewKalman2D(processNoise float64) *Kalman2D {
	if processNoise <= 0 {
		processNoise = 0.5
	}
	return &Kalman2D{q: processNoise}
}

// Update advances the filter by dt seconds and folds in a position measurement
// (mx, my) with reported standard deviation measStd metres. accelMag is the
// IMU acceleration magnitude (m/s^2) minus gravity, used to inflate process
// noise during accelerations. It returns the fused position and velocity.
func (k *Kalman2D) Update(dt, mx, my, measStd, accelMag float64) (px, py, vx, vy float64) {
	if !k.inited {
		k.x = [4]float64{mx, my, 0, 0}
		k.p = [4]float64{measStd * measStd, measStd * measStd, 100, 100}
		k.inited = true
		return mx, my, 0, 0
	}
	if dt <= 0 {
		dt = 1e-3
	}

	// Predict: constant-velocity model.
	k.x[0] += k.x[2] * dt
	k.x[1] += k.x[3] * dt

	q := k.q * (1 + accelMag) // more manoeuvre => trust the model less
	k.p[0] += k.p[2]*dt*dt + q*dt
	k.p[1] += k.p[3]*dt*dt + q*dt
	k.p[2] += q * dt
	k.p[3] += q * dt

	// Update: position measurement only.
	r := measStd * measStd
	if r <= 0 {
		r = 9 // ~3 m default GPS sigma
	}
	kx := k.p[0] / (k.p[0] + r)
	ky := k.p[1] / (k.p[1] + r)

	resX := mx - k.x[0]
	resY := my - k.x[1]

	k.x[0] += kx * resX
	k.x[1] += ky * resY
	// Feed the position residual back into the velocity estimate.
	k.x[2] += (kx * resX) / dt
	k.x[3] += (ky * resY) / dt

	k.p[0] *= (1 - kx)
	k.p[1] *= (1 - ky)
	k.p[2] *= (1 - kx)
	k.p[3] *= (1 - ky)

	return k.x[0], k.x[1], k.x[2], k.x[3]
}
