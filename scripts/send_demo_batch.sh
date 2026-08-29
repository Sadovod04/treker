#!/usr/bin/env bash
# Generate a synthetic GPS/IMU batch and POST it to the ingest endpoint.
# Simulates a player jogging with two sprints across ~60 seconds at 10 Hz.
#
# Usage: ./scripts/send_demo_batch.sh [DEVICE_ID] [API_URL]
set -euo pipefail

DEVICE_ID="${1:-ESP32-TRACKER-001}"
API_URL="${2:-http://localhost:8080}"
API_KEY="${INGEST_API_KEY:-dev-tracker-key}"

PAYLOAD="$(python3 - "$DEVICE_ID" <<'PY'
import json, math, sys, datetime

device_id = sys.argv[1]
lat0, lng0 = 55.751244, 37.618423           # Moscow-ish origin
m_per_deg_lat = 111_320.0
m_per_deg_lng = m_per_deg_lat * math.cos(math.radians(lat0))

hz, secs = 10, 60
samples = []
x = y = 0.0
for i in range(hz * secs):
    t = i / hz
    # base jog 3 m/s; two sprint windows at 7 m/s
    speed = 3.0
    if 12 <= t < 18 or 34 <= t < 41:
        speed = 7.0
    heading = 0.4 * math.sin(t / 6.0)       # gentle curving run
    x += speed * math.cos(heading) / hz
    y += speed * math.sin(heading) / hz
    az = 9.81 + (4.0 if abs((t % 10) - 5) < 0.1 else 0.0)   # periodic "jump"
    samples.append({
        "t": int(t * 1000),
        "lat": lat0 + y / m_per_deg_lat,
        "lng": lng0 + x / m_per_deg_lng,
        "alt": 150.0,
        "gps_speed": round(speed * 3.6, 2),
        "gps_accuracy": 2.5,
        "accel": [round(0.2 * math.cos(t), 3), round(0.2 * math.sin(t), 3), round(az, 3)],
        "gyro": [0.01, -0.01, round(0.05 * math.sin(t), 3)],
        "hr": 120 + (40 if speed > 5 else 0) + (i % 5),
    })

print(json.dumps({
    "device_id": device_id,
    "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    "data": samples,
}))
PY
)"

echo "Posting $(echo "$PAYLOAD" | python3 -c 'import json,sys;print(len(json.load(sys.stdin)["data"]))') samples for $DEVICE_ID ..."
curl -sS -X POST "$API_URL/api/data/ingest" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "$PAYLOAD"
echo
