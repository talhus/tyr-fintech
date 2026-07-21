package notifications

import "sync"

type Hub struct {
	clients map[string]chan string
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]chan string),
	}
}

// REGISTER
func (h *Hub) Register(userID string) chan string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, exists := h.clients[userID]; exists {
		close(ch)
	}

	//buffered new ch
	messageChan := make(chan string, 10)
	h.clients[userID] = messageChan
	return messageChan
}

// UNREGISTER
func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if ch, exists := h.clients[userID]; exists {
		close(ch)
		delete(h.clients, userID)
	}
}

// SEND MSG
func (h *Hub) SendToUser(userID, msg string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ch, exists := h.clients[userID]
	if !exists {
		//user offline
		return false
	}

	select {
	case ch <- msg:
		return true
	default:
		return false
	}

}
