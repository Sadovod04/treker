// Thin fetch wrapper around the Go REST API.

const BASE = import.meta.env.VITE_API_URL || '';

function token() {
  try {
    return localStorage.getItem('st.token') || '';
  } catch {
    return '';
  }
}

async function request(path, { method = 'GET', body, signal } = {}) {
  const res = await fetch(BASE + path, {
    method,
    signal,
    headers: {
      'Content-Type': 'application/json',
      ...(token() ? { Authorization: `Bearer ${token()}` } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`${res.status}: ${text}`);
  }
  if (res.status === 204) return null;
  return res.json();
}

export const api = {
  login: (username, password) =>
    request('/api/auth/login', { method: 'POST', body: { username, password } }),

  listPlayers: (signal) => request('/api/players', { signal }),
  createPlayer: (player) => request('/api/players', { method: 'POST', body: player }),
  deletePlayer: (id) => request(`/api/players/${id}`, { method: 'DELETE' }),

  metrics: (playerId, sessionId, signal) =>
    request(`/api/players/${playerId}/metrics${sessionId ? `?session_id=${sessionId}` : ''}`, { signal }),

  heatmap: (playerId, sessionId, signal) =>
    request(`/api/players/${playerId}/heatmap${sessionId ? `?session_id=${sessionId}` : ''}`, { signal }),

  sessions: (playerId, { limit = 20, offset = 0 } = {}, signal) =>
    request(`/api/players/${playerId}/sessions?limit=${limit}&offset=${offset}`, { signal }),

  compare: (p1, p2, sessionId, signal) => {
    const q = new URLSearchParams({ player1_id: p1, player2_id: p2 });
    if (sessionId) q.set('session_id', sessionId);
    return request(`/api/compare?${q.toString()}`, { signal });
  },
};
