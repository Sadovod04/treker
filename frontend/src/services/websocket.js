// Auto-reconnecting WebSocket client for the live metrics stream.

function wsBase() {
  if (import.meta.env.VITE_WS_URL) return import.meta.env.VITE_WS_URL;
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${proto}//${window.location.host}`;
}

export function openLiveStream(sessionId, onMessage) {
  let socket;
  let closed = false;
  let retry = 1000;

  const connect = () => {
    if (closed) return;
    socket = new WebSocket(`${wsBase()}/ws/live?session_id=${sessionId}`);

    socket.onopen = () => {
      retry = 1000;
    };
    socket.onmessage = (ev) => {
      try {
        onMessage(JSON.parse(ev.data));
      } catch {
        /* ignore malformed frame */
      }
    };
    socket.onclose = () => {
      if (closed) return;
      setTimeout(connect, retry);
      retry = Math.min(retry * 2, 15000);
    };
    socket.onerror = () => socket && socket.close();
  };

  connect();

  return () => {
    closed = true;
    if (socket) socket.close();
  };
}
