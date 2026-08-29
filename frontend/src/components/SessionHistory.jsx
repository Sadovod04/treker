import { useEffect, useState } from 'react';
import { api } from '../services/api.js';

function fmtDate(iso) {
  try {
    return new Date(iso).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return iso;
  }
}

export default function SessionHistory({ playerId, onSelect }) {
  const [rows, setRows] = useState([]);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!playerId) return undefined;
    const ac = new AbortController();
    api
      .sessions(playerId, { limit: 20 }, ac.signal)
      .then(setRows)
      .catch((e) => e.name !== 'AbortError' && setError(e.message));
    return () => ac.abort();
  }, [playerId]);

  if (error) return <p className="error">История недоступна: {error}</p>;
  if (!rows.length) return <p className="muted">Сессий пока нет.</p>;

  return (
    <div className="table-scroll">
      <table>
        <thead>
          <tr>
            <th>Начало</th>
            <th>Длит., мин</th>
            <th>Дистанция, км</th>
            <th>Player Load</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {rows.map((s) => (
            <tr key={s.id}>
              <td>{fmtDate(s.start_time)}</td>
              <td>{s.duration_minutes}</td>
              <td>{s.distance_km.toFixed(2)}</td>
              <td>{s.player_load.toFixed(0)}</td>
              <td>
                <button className="ghost" onClick={() => onSelect && onSelect(s.id)}>
                  Открыть
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
