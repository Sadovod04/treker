const statusLabel = {
  online: '🟢 Online',
  offline: '⚪ Offline',
  disconnected: '🔴 Disconnected',
};

export default function PlayerSelector({ players, value, onChange, label = 'Игрок' }) {
  return (
    <label className="player-selector">
      <span>{label}</span>
      <select value={value ?? ''} onChange={(e) => onChange(Number(e.target.value) || null)}>
        <option value="">— выберите —</option>
        {players.map((p) => (
          <option key={p.id} value={p.id}>
            {p.number ? `#${p.number} ` : ''}
            {p.name} · {statusLabel[p.status] || p.status}
          </option>
        ))}
      </select>
    </label>
  );
}
