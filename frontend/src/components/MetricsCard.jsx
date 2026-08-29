// A single dashboard metric card: title, big value, optional sub-rows.
export default function MetricsCard({ title, value, unit, rows = [], accent }) {
  return (
    <div className="metric-card" style={accent ? { borderTopColor: accent } : undefined}>
      <div className="metric-card__title">{title}</div>
      <div className="metric-card__value">
        {value}
        {unit ? <span className="metric-card__unit"> {unit}</span> : null}
      </div>
      {rows.length > 0 && (
        <dl className="metric-card__rows">
          {rows.map((r) => (
            <div key={r.label}>
              <dt>{r.label}</dt>
              <dd>{r.value}</dd>
            </div>
          ))}
        </dl>
      )}
    </div>
  );
}
