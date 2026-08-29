// Renders the diff table from GET /api/compare.
export default function ComparisonTable({ rows, name1 = 'Игрок 1', name2 = 'Игрок 2' }) {
  return (
    <div className="table-scroll">
      <table className="compare-table">
        <thead>
          <tr>
            <th>Метрика</th>
            <th>{name1}</th>
            <th>{name2}</th>
            <th>Разница</th>
            <th>%</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => {
            const cls = r.diff > 0 ? 'better' : r.diff < 0 ? 'worse' : 'neutral';
            return (
              <tr key={r.metric}>
                <td>{r.metric}</td>
                <td>{r.player1} {r.unit}</td>
                <td>{r.player2} {r.unit}</td>
                <td className={cls}>{r.diff > 0 ? '+' : ''}{r.diff} {r.unit}</td>
                <td className={cls}>{r.diff_percent > 0 ? '+' : ''}{r.diff_percent}%</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
