package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	// Demo default: accept any origin. Restrict in production.
	CheckOrigin: func(*http.Request) bool { return true },
}

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
)

// handleWSLive streams live metric updates for ?session_id=<id>.
func (s *Server) handleWSLive(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := queryInt64(r, "session_id")
	if !ok {
		writeError(w, http.StatusBadRequest, "session_id is required")
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade already wrote an error response.
	}
	defer conn.Close()

	msgs, unsubscribe := s.hub.Subscribe(sessionID)
	defer unsubscribe()

	// Reader: drain control frames and honour pongs; exit on close.
	go func() {
		conn.SetReadLimit(512)
		_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(wsPongWait))
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				unsubscribe()
				return
			}
		}
	}()

	ping := time.NewTicker(wsPingPeriod)
	defer ping.Stop()

	for {
		select {
		case msg, open := <-msgs:
			if !open {
				return
			}
			_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := conn.WriteJSON(msg); err != nil {
				slog.Debug("ws write", "err", err)
				return
			}
		case <-ping.C:
			_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
