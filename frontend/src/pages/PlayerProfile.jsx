import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import IntensityChart from '../components/IntensityChart.jsx';
import HeatmapField from '../components/HeatmapField.jsx';
import SessionHistory from '../components/SessionHistory.jsx';
import { api } from '../services/api.js';
import { useMetrics } from '../hooks/useMetrics.js';
import { useHeatmap } from '../hooks/useHeatmap.js';

const METRIC_ROWS = [
  ['speed_max_kmh', 'Speed (Max)', 'км/ч', (v) => v.toFixed(1)],
  ['speed_avg_kmh', 'Speed (Avg)', 'км/ч', (v) => v.toFixed(1)],
  ['distance_total_m', 'Distance', 'км', (v) => (v / 1000).toFixed(2)],
  ['sprint_count', 'Sprints', 'шт', (v) => v],
  ['acceleration_count', 'Accelerations', 'шт', (v) => v],
  ['player_load', 'Player Load', 'ед.', (v) => v.toFixed(0)],
  ['jump_height_max_cm', 'Jump Height', 'см', (v) => v.toFixed(0)],
  ['hr_avg', 'Heart Rate (Avg)', 'bpm', (v) => v ?? '—'],
  ['hr_max', 'Heart Rate (Max)', 'bpm', (v) => v ?? '—'],
  ['duration_minutes', 'Duration', 'мин', (v) => v],
];

export default function PlayerProfile() {
  const { id } = useParams();
  const playerId = Number(id);
  const [player, setPlayer] = useState(null);
  const [sessionId, setSessionId] = useState(null);

  useEffect(() => {
    const ac = new AbortController();
    api.listPlayers(ac.signal)
      .then((ps) => setPlayer(ps.find((p) => p.id === playerId) || null))
      .catch(() => {});
    return () => ac.abort();
  }, [playerId]);

  const { data: m, error, loading } = useMetrics(playerId, sessionId);
  const { data: heatmap } = useHeatmap(playerId, sessionId);

  return (
    <section className="page">
      <h1>{player ? `#${player.number ?? '—'} ${player.name}` : `Игрок ${playerId}`}</h1>
      {player?.position && <p className="muted">Позиция: {player.position}</p>}

      {loading && !m && <p className="muted">Загрузка…</p>}
      {error && <p className="error">{error}</p>}

      {m && (
        <>
          <h2>Метрики за сессию #{m.session_id}</h2>
          <div className="table-scroll">
            <table>
              <thead>
                <tr><th>Метрика</th><th>Значение</th><th>Ед.</th></tr>
              </thead>
              <tbody>
                {METRIC_ROWS.map(([key, label, unit, fmt]) => (
                  <tr key={key}>
                    <td>{label}</td>
                    <td>{m[key] === null || m[key] === undefined ? '—' : fmt(m[key])}</td>
                    <td>{unit}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <h2>Интенсивность по зонам</h2>
          <IntensityChart series={[{ name: player?.name || 'Игрок', zones: m.zones }]} />

          <h2>Тепловая карта поля</h2>
          <HeatmapField heatmap={heatmap} width={640} />
        </>
      )}

      <h2>История сессий</h2>
      <SessionHistory playerId={playerId} onSelect={setSessionId} />
    </section>
  );
}
