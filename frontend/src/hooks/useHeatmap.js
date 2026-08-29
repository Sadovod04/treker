import { useEffect, useState } from 'react';
import { api } from '../services/api.js';

// Loads the gridded heatmap for a player/session.
export function useHeatmap(playerId, sessionId) {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!playerId) return undefined;
    const ac = new AbortController();
    setLoading(true);
    api
      .heatmap(playerId, sessionId, ac.signal)
      .then((h) => {
        setData(h);
        setError(null);
      })
      .catch((e) => {
        if (e.name !== 'AbortError') setError(e.message);
      })
      .finally(() => setLoading(false));
    return () => ac.abort();
  }, [playerId, sessionId]);

  return { data, error, loading };
}
