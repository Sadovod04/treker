import { useEffect, useState } from 'react';
import { openLiveStream } from '../services/websocket.js';

// Subscribes to WS /ws/live for a session and returns the latest message payload
// by type, e.g. live.metrics.
export function useLiveData(sessionId) {
  const [live, setLive] = useState({});

  useEffect(() => {
    if (!sessionId) return undefined;
    const close = openLiveStream(sessionId, (msg) => {
      if (!msg || !msg.type) return;
      setLive((prev) => ({ ...prev, [msg.type]: msg.data, _ts: Date.now() }));
    });
    return close;
  }, [sessionId]);

  return live;
}
