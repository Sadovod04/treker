import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import PlayerSelector from '../components/PlayerSelector.jsx';
import MetricsCard from '../components/MetricsCard.jsx';
import IntensityChart from '../components/IntensityChart.jsx';
import HeatmapField from '../components/HeatmapField.jsx';
import { api } from '../services/api.js';
import { storage } from '../services/storage.js';
import { useMetrics } from '../hooks/useMetrics.js';
import { useLiveData } from '../hooks/useLiveData.js';
import { useHeatmap } from '../hooks/useHeatmap.js';

export default function Dashboard() {
  const [players, setPlayers] = useState([]);
  const [playerId, setPlayerId] = useState(storage.getLastPlayer());
  const [loadError, setLoadError] = useState(null);

  useEffect(() => {
    const ac = new AbortController();
    api.listPlayers(ac.signal).then(setPlayers).catch((e) => {
      if (e.name !== 'AbortError') setLoadError(e.message);
    });
    return () => ac.abort();
  }, []);

  useEffect(() => {
    if (playerId) storage.setLastPlayer(playerId);
  }, [playerId]);

  const { data: rest, error, loading, setData } = useMetrics(playerId, null);
  const live = useLiveData(rest?.session_id);
  const { data: heatmap } = useHeatmap(playerId, null);

  // Prefer the freshest metrics pushed over WebSocket.
  useEffect(() => {
    if (live.metrics) setData(live.metrics);
  }, [live.metrics, setData]);

  const m = rest;
  const selected = useMemo(
    () => players.find((p) => p.id === playerId),
    [players, playerId],
  );

  return (
    <section className="page">
      <h1>Dashboard</h1>
      {loadError && <p className="error">Не удалось загрузить игроков: {loadError}</p>}

      <div className="toolbar">
        <PlayerSelector players={players} value={playerId} onChange={setPlayerId} />
        {selected && (
          <div className="quick-stats">
            <span>Статус: {selected.status}</span>
            {selected.last_seen_at && (
              <span>Обновлено: {new Date(selected.last_seen_at).toLocaleTimeString('ru-RU')}</span>
            )}
            <Link to={`/players/${selected.id}`}>Профиль →</Link>
          </div>
        )}
      </div>

      {!playerId && <p className="muted">Выберите игрока, чтобы увидеть метрики.</p>}
      {playerId && loading && !m && <p className="muted">Загрузка метрик…</p>}
      {playerId && error && <p className="error">{error}</p>}

      {m && (
        <>
          <div className="cards-grid">
            <MetricsCard
              title="Скорость"
              value={m.speed_max_kmh?.toFixed(1) ?? '—'}
              unit="км/ч макс"
              accent="#ef4444"
              rows={[
                { label: 'Средняя', value: `${m.speed_avg_kmh?.toFixed(1) ?? '—'} км/ч` },
              ]}
            />
            <MetricsCard
              title="Дистанция"
              value={((m.distance_total_m ?? 0) / 1000).toFixed(2)}
              unit="км"
              accent="#3b82f6"
              rows={[{ label: 'Всего, м', value: (m.distance_total_m ?? 0).toFixed(0) }]}
            />
            <MetricsCard
              title="Спринты"
              value={m.sprint_count ?? 0}
              unit="шт"
              accent="#f59e0b"
              rows={[
                { label: 'Ср. длина', value: `${(m.sprint_avg_length_m ?? 0).toFixed(1)} м` },
                { label: 'Макс. длина', value: `${(m.sprint_max_length_m ?? 0).toFixed(1)} м` },
              ]}
            />
            <MetricsCard
              title="Player Load"
              value={(m.player_load ?? 0).toFixed(0)}
              unit="ед."
              accent="#8b5cf6"
              rows={[{ label: 'Ускорения', value: m.acceleration_count ?? 0 }]}
            />
            <MetricsCard
              title="Прыжки"
              value={(m.jump_height_max_cm ?? 0).toFixed(0)}
              unit="см макс"
              accent="#10b981"
              rows={[{ label: 'Кол-во', value: m.jump_count ?? 0 }]}
            />
            <MetricsCard
              title="Пульс"
              value={m.hr_avg ?? '—'}
              unit="bpm ср."
              accent="#ec4899"
              rows={[{ label: 'Макс', value: m.hr_max ?? '—' }]}
            />
          </div>

          <h2>Интенсивность по зонам</h2>
          <IntensityChart series={[{ name: selected?.name || 'Игрок', zones: m.zones }]} />

          <h2>Тепловая карта поля</h2>
          <HeatmapField heatmap={heatmap} width={640} title={selected?.name} />
        </>
      )}
    </section>
  );
}
