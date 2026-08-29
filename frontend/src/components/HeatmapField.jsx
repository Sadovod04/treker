import { useEffect, useMemo, useRef, useState } from 'react';

// Broadcast-style heatmap: soft Gaussian blobs per grid cell, colourised through
// a transparent → blue → green → yellow → red ramp, drawn over a striped pitch.

const SS = 2; // supersample factor for crisp output

// 256-entry lookup table for the classic sports heatmap ramp.
const GRADIENT = buildGradient([
  [0.0, [0, 0, 0, 0]],
  [0.18, [40, 60, 220, 90]],
  [0.35, [30, 150, 230, 170]],
  [0.5, [40, 200, 120, 210]],
  [0.65, [120, 215, 60, 235]],
  [0.8, [240, 220, 40, 245]],
  [0.9, [250, 140, 30, 250]],
  [1.0, [235, 30, 20, 255]],
]);

function buildGradient(stops) {
  const lut = new Uint8ClampedArray(256 * 4);
  for (let i = 0; i < 256; i += 1) {
    const t = i / 255;
    let a = stops[0];
    let b = stops[stops.length - 1];
    for (let s = 0; s < stops.length - 1; s += 1) {
      if (t >= stops[s][0] && t <= stops[s + 1][0]) {
        a = stops[s];
        b = stops[s + 1];
        break;
      }
    }
    const span = b[0] - a[0] || 1;
    const k = (t - a[0]) / span;
    for (let c = 0; c < 4; c += 1) {
      lut[i * 4 + c] = Math.round(a[1][c] + k * (b[1][c] - a[1][c]));
    }
  }
  return lut;
}

function drawPitch(ctx, w, h) {
  const line = 'rgba(255,255,255,0.75)';
  ctx.save();
  // turf base + mowing stripes
  ctx.fillStyle = '#1f7a34';
  ctx.fillRect(0, 0, w, h);
  const stripes = 12;
  for (let i = 0; i < stripes; i += 1) {
    ctx.fillStyle = i % 2 ? 'rgba(255,255,255,0.045)' : 'rgba(0,0,0,0.045)';
    ctx.fillRect((i * w) / stripes, 0, w / stripes + 1, h);
  }

  const m = Math.round(h * 0.045); // margin
  const fw = w - m * 2;
  const fh = h - m * 2;
  ctx.strokeStyle = line;
  ctx.lineWidth = Math.max(1.5, h * 0.006);
  ctx.strokeRect(m, m, fw, fh);

  // halfway line + centre circle + spot
  ctx.beginPath();
  ctx.moveTo(w / 2, m);
  ctx.lineTo(w / 2, h - m);
  ctx.stroke();
  ctx.beginPath();
  ctx.arc(w / 2, h / 2, fh * 0.13, 0, Math.PI * 2);
  ctx.stroke();
  ctx.beginPath();
  ctx.arc(w / 2, h / 2, ctx.lineWidth * 1.2, 0, Math.PI * 2);
  ctx.fillStyle = line;
  ctx.fill();

  const boxH = fh * 0.6;
  const boxW = fw * 0.16;
  const sixH = fh * 0.28;
  const sixW = fw * 0.06;
  // both ends: penalty box, goal box, penalty spot, arc
  [0, 1].forEach((side) => {
    const x0 = side === 0 ? m : w - m - boxW;
    ctx.strokeRect(x0, h / 2 - boxH / 2, boxW, boxH);
    const gx0 = side === 0 ? m : w - m - sixW;
    ctx.strokeRect(gx0, h / 2 - sixH / 2, sixW, sixH);
    const spotX = side === 0 ? m + fw * 0.11 : w - m - fw * 0.11;
    ctx.beginPath();
    ctx.arc(spotX, h / 2, ctx.lineWidth * 1.1, 0, Math.PI * 2);
    ctx.fill();
    ctx.beginPath();
    const a0 = side === 0 ? -Math.PI * 0.3 : Math.PI * 0.7;
    const a1 = side === 0 ? Math.PI * 0.3 : Math.PI * 1.3;
    ctx.arc(spotX, h / 2, fh * 0.13, a0, a1);
    ctx.stroke();
  });

  // corner arcs
  const cr = h * 0.025;
  [[m, m, 0], [w - m, m, Math.PI / 2], [w - m, h - m, Math.PI], [m, h - m, -Math.PI / 2]].forEach(
    ([cx, cy, rot]) => {
      ctx.beginPath();
      ctx.arc(cx, cy, cr, rot, rot + Math.PI / 2);
      ctx.stroke();
    },
  );
  ctx.restore();
}

export default function HeatmapField({ heatmap, title, width = 420 }) {
  const [zoom, setZoom] = useState(1);
  const [hover, setHover] = useState(null);
  const canvasRef = useRef(null);

  const model = useMemo(() => {
    if (!heatmap) return null;
    const cols = heatmap.grid_cols || 11;
    const rows = heatmap.grid_rows || 7;
    const cells = heatmap.cells || [];
    const maxFromCells = cells.reduce((mx, c) => Math.max(mx, c.time_seconds || 0), 0);
    const max = heatmap.max_seconds || maxFromCells || 1;
    const map = new Map();
    cells.forEach((c) => map.set(`${c.grid_x}:${c.grid_y}`, c));
    return { cols, rows, max, map, cells };
  }, [heatmap]);

  const pitchW = width;
  const pitchH = model ? Math.round((width * model.rows) / model.cols) : Math.round(width * 0.62);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !model) return;
    const w = pitchW * SS;
    const h = pitchH * SS;
    canvas.width = w;
    canvas.height = h;
    const ctx = canvas.getContext('2d');

    drawPitch(ctx, w, h);

    // 1. accumulate greyscale blobs on an offscreen layer
    const heat = document.createElement('canvas');
    heat.width = w;
    heat.height = h;
    const hctx = heat.getContext('2d');
    const cw = w / model.cols;
    const ch = h / model.rows;
    const radius = Math.max(cw, ch) * 1.9;

    model.cells.forEach((c) => {
      const t = Math.pow(Math.min(1, (c.time_seconds || 0) / model.max), 0.65);
      if (t <= 0) return;
      const cx = (c.grid_x + 0.5) * cw;
      const cy = h - (c.grid_y + 0.5) * ch; // grid_y grows upward
      const g = hctx.createRadialGradient(cx, cy, 0, cx, cy, radius);
      g.addColorStop(0, `rgba(0,0,0,${t})`);
      g.addColorStop(1, 'rgba(0,0,0,0)');
      hctx.fillStyle = g;
      hctx.fillRect(cx - radius, cy - radius, radius * 2, radius * 2);
    });

    // 2. colourise by alpha through the gradient LUT
    const img = hctx.getImageData(0, 0, w, h);
    const d = img.data;
    for (let i = 0; i < d.length; i += 4) {
      const a = d[i + 3];
      if (a === 0) continue;
      const o = a * 4;
      d[i] = GRADIENT[o];
      d[i + 1] = GRADIENT[o + 1];
      d[i + 2] = GRADIENT[o + 2];
      d[i + 3] = Math.round((GRADIENT[o + 3] * a) / 255);
    }
    hctx.putImageData(img, 0, 0);

    // 3. composite heat over the pitch
    ctx.drawImage(heat, 0, 0);
  }, [model, pitchW, pitchH]);

  if (!model) return <div className="heatmap-empty">Нет данных тепловой карты</div>;

  const onMove = (e) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const fx = (e.clientX - rect.left) / rect.width;
    const fy = (e.clientY - rect.top) / rect.height;
    const gx = Math.min(model.cols - 1, Math.max(0, Math.floor(fx * model.cols)));
    const gy = Math.min(model.rows - 1, Math.max(0, Math.floor((1 - fy) * model.rows)));
    const cell = model.map.get(`${gx}:${gy}`);
    setHover({ gx, gy, secs: cell?.time_seconds || 0, spd: cell?.avg_speed_kmh });
  };

  return (
    <figure className="heatmap">
      {title && <figcaption>{title}</figcaption>}
      <div className="heatmap__controls">
        <button className="ghost" onClick={() => setZoom((z) => Math.max(1, z - 0.25))}>−</button>
        <span>{Math.round(zoom * 100)}%</span>
        <button className="ghost" onClick={() => setZoom((z) => Math.min(3, z + 0.25))}>+</button>
        <span className="heatmap__legend" aria-hidden="true">
          <i />низкая<i className="mid" />средняя<i className="hot" />высокая
        </span>
      </div>
      <div className="heatmap__viewport" style={{ maxWidth: pitchW }}>
        <canvas
          ref={canvasRef}
          style={{ width: pitchW * zoom, height: pitchH * zoom, display: 'block' }}
          onMouseMove={onMove}
          onMouseLeave={() => setHover(null)}
          role="img"
          aria-label={title ? `Тепловая карта: ${title}` : 'Тепловая карта поля'}
        />
      </div>
      <div className="heatmap__hint">
        {hover && hover.secs
          ? `Зона (${hover.gx}, ${hover.gy}): ${hover.secs} c${hover.spd ? `, ${hover.spd.toFixed(1)} км/ч` : ''}`
          : `Макс. время в зоне: ${model.max} c`}
      </div>
    </figure>
  );
}
