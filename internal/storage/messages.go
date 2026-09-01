package storage

import (
	"sync"
	"tcp-chat/internal/domain"
	"tcp-chat/internal/types"
)

type MessageHistory struct {
	buf *types.CycleBuffer[domain.ChatMessage]

	mu sync.RWMutex
}

func NewMessageHistory() *MessageHistory {
	return &MessageHistory{
		buf: types.NewCycleBuffer[domain.ChatMessage](50),
	}
}

func (h *MessageHistory) Add(message domain.ChatMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.buf.Push(message)
}

func (h *MessageHistory) GetRecent() []domain.ChatMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.buf.GetValues()
}
