// Package live is an in-process pub/sub hub that fans real-time metric updates
// out to WebSocket subscribers, keyed by session id.
package live

import "sync"

// Message is what the hub broadcasts to clients.
type Message struct {
	Type string `json:"type"` // "metrics" | "sample" | "status"
	Data any    `json:"data"`
}

// Hub tracks subscribers per session id.
type Hub struct {
	mu   sync.RWMutex
	subs map[int64]map[chan Message]struct{}
}

// New builds an empty hub.
func New() *Hub {
	return &Hub{subs: make(map[int64]map[chan Message]struct{})}
}

// Subscribe registers a buffered channel for a session and returns an
// unsubscribe function.
func (h *Hub) Subscribe(sessionID int64) (<-chan Message, func()) {
	ch := make(chan Message, 32)
	h.mu.Lock()
	if h.subs[sessionID] == nil {
		h.subs[sessionID] = make(map[chan Message]struct{})
	}
	h.subs[sessionID][ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			h.mu.Lock()
			if set := h.subs[sessionID]; set != nil {
				delete(set, ch)
				if len(set) == 0 {
					delete(h.subs, sessionID)
				}
			}
			h.mu.Unlock()
			close(ch)
		})
	}
	return ch, unsub
}

// Publish delivers a message to every subscriber of a session. Slow consumers
// are skipped rather than blocking the publisher.
func (h *Hub) Publish(sessionID int64, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[sessionID] {
		select {
		case ch <- msg:
		default:
		}
	}
}
