import { useEffect, useState } from 'react';
import { api } from '../services/api.js';

const EMPTY = { name: '', number: '', position: '', device_id: '' };

export default function Admin() {
  const [players, setPlayers] = useState([]);
  const [form, setForm] = useState(EMPTY);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);

  const reload = () => api.listPlayers().then(setPlayers).catch((e) => setError(e.message));

  useEffect(() => {
    reload();
  }, []);

  const submit = async (e) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api.createPlayer({
        name: form.name.trim(),
        device_id: form.device_id.trim(),
        number: form.number ? Number(form.number) : null,
        position: form.position || null,
      });
      setForm(EMPTY);
      reload();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const remove = async (id) => {
    if (!window.confirm('Удалить игрока и все его данные?')) return;
    try {
      await api.deletePlayer(id);
      reload();
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <section className="page">
      <h1>Панель администратора</h1>
      {error && <p className="error">{error}</p>}

      <h2>Добавить игрока</h2>
      <form className="admin-form" onSubmit={submit}>
        <input required placeholder="Имя"
          value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
        <input placeholder="№" type="number" min="1" style={{ width: 80 }}
          value={form.number} onChange={(e) => setForm({ ...form, number: e.target.value })} />
        <input placeholder="Позиция"
          value={form.position} onChange={(e) => setForm({ ...form, position: e.target.value })} />
        <input required placeholder="Device ID (ESP32-...)"
          value={form.device_id} onChange={(e) => setForm({ ...form, device_id: e.target.value })} />
        <button type="submit" disabled={busy}>Добавить</button>
      </form>

      <h2>Игроки</h2>
      <div className="table-scroll">
        <table>
          <thead>
            <tr><th>№</th><th>Имя</th><th>Позиция</th><th>Device</th><th>Статус</th><th /></tr>
          </thead>
          <tbody>
            {players.map((p) => (
              <tr key={p.id}>
                <td>{p.number ?? '—'}</td>
                <td>{p.name}</td>
                <td>{p.position ?? '—'}</td>
                <td><code>{p.device_id}</code></td>
                <td>{p.status}</td>
                <td><button className="ghost" onClick={() => remove(p.id)}>Удалить</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
