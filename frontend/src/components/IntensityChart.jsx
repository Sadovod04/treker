import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

const ZONE_COLORS = { low: '#3b82f6', medium: '#f59e0b', high: '#ef4444' };

// Renders the seconds-in-zone breakdown. `series` is an array of
// { name, zones: {low_seconds, medium_seconds, high_seconds} } — one entry for
// the profile view, two for side-by-side comparison.
export default function IntensityChart({ series }) {
  const data = [
    { zone: 'Низкая', key: 'low' },
    { zone: 'Средняя', key: 'medium' },
    { zone: 'Высокая', key: 'high' },
  ].map((row) => {
    const out = { zone: row.zone, _key: row.key };
    series.forEach((s) => {
      out[s.name] = Math.round((s.zones?.[`${row.key}_seconds`] || 0) / 1);
    });
    return out;
  });

  return (
    <div className="chart-box">
      <ResponsiveContainer width="100%" height={260}>
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
          <XAxis dataKey="zone" />
          <YAxis label={{ value: 'сек', angle: -90, position: 'insideLeft' }} />
          <Tooltip formatter={(v) => `${v} с`} />
          <Legend />
          {series.map((s, i) => (
            <Bar key={s.name} dataKey={s.name} fill={i === 0 ? '#2563eb' : '#f97316'}>
              {series.length === 1 &&
                data.map((d) => <Cell key={d._key} fill={ZONE_COLORS[d._key]} />)}
            </Bar>
          ))}
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
