import { useEffect, useState } from 'react';
import PlayerSelector from '../components/PlayerSelector.jsx';
import HeatmapField from '../components/HeatmapField.jsx';
import ComparisonTable from '../components/ComparisonTable.jsx';
import IntensityChart from '../components/IntensityChart.jsx';
import { api } from '../services/api.js';

export default function Compare() {
  const [players, setPlayers] = useState([]);
  const [p1, setP1] = useState(null);
  const [p2, setP2] = useState(null);
  const [cmp, setCmp] = useState(null);
  const [hm1, setHm1] = useState(null);
  const [hm2, setHm2] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    const ac = new AbortController();
    api.listPlayers(ac.signal).then(setPlayers).catch(() => {});
    return () => ac.abort();
  }, []);

  useEffect(() => {
    if (!p1 || !p2) return undefined;
    const ac = new AbortController();
    setError(null);
    Promise.all([
      api.compare(p1, p2, null, ac.signal),
      api.heatmap(p1, null, ac.signal),
      api.heatmap(p2, null, ac.signal),
    ])
      .then(([c, h1, h2]) => {
        setCmp(c);
        setHm1(h1);
        setHm2(h2);
      })
      .catch((e) => e.name !== 'AbortError' && setError(e.message));
    return () => ac.abort();
  }, [p1, p2]);

  const name = (id) => players.find((p) => p.id === id)?.name || `Игрок ${id}`;

  return (
    <section className="page">
      <h1>Сравнение игроков</h1>
      <div className="toolbar">
        <PlayerSelector players={players} value={p1} onChange={setP1} label="Игрок 1" />
        <PlayerSelector players={players} value={p2} onChange={setP2} label="Игрок 2" />
      </div>

      {error && <p className="error">{error}</p>}
      {(!p1 || !p2) && <p className="muted">Выберите двух игроков для сравнения.</p>}

      {cmp && (
        <>
          <div className="side-by-side">
            <HeatmapField heatmap={hm1} title={name(p1)} />
            <HeatmapField heatmap={hm2} title={name(p2)} />
          </div>

          <h2>Сравнительная таблица</h2>
          <ComparisonTable rows={cmp.rows} name1={name(p1)} name2={name(p2)} />

          <h2>Интенсивность по зонам</h2>
          <IntensityChart
            series={[
              { name: name(p1), zones: cmp.player1?.zones },
              { name: name(p2), zones: cmp.player2?.zones },
            ]}
          />
        </>
      )}
    </section>
  );
}
