package storage

import (
	"tcp-chat/internal/domain"
	"tcp-chat/internal/types"
)

type MessageHistory struct {
	buf *types.CycleBuffer[domain.ChatMessage]
}

func NewMessageHistory() *MessageHistory {
	return &MessageHistory{
		buf: types.NewCycleBuffer[domain.ChatMessage](50),
	}
}

func (h *MessageHistory) Add(message domain.ChatMessage) {
	h.buf.Push(message)
}

func (h *MessageHistory) GetRecent() []domain.ChatMessage {
	return h.buf.GetValues()
}
