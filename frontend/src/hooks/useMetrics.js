import { useCallback, useEffect, useState } from 'react';
import { api } from '../services/api.js';

// Fetches per-session metrics for a player. Re-runs when player or session change.
export function useMetrics(playerId, sessionId) {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  const reload = useCallback(
    (signal) => {
      if (!playerId) return;
      setLoading(true);
      api
        .metrics(playerId, sessionId, signal)
        .then((m) => {
          setData(m);
          setError(null);
        })
        .catch((e) => {
          if (e.name !== 'AbortError') setError(e.message);
        })
        .finally(() => setLoading(false));
    },
    [playerId, sessionId],
  );

  useEffect(() => {
    const ac = new AbortController();
    reload(ac.signal);
    return () => ac.abort();
  }, [reload]);

  return { data, error, loading, reload, setData };
}
