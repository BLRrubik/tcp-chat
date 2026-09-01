package storage

import (
	"tcp-chat/internal/domain"
	"tcp-chat/internal/types"
)

type MessageHistory struct {
	buf *types.CycleBuffer[domain.ChatMessage]
}

func NewMessageHistory(size int) *MessageHistory {
	return &MessageHistory{
		buf: types.NewCycleBuffer[domain.ChatMessage](size),
	}
}

func (h *MessageHistory) Add(message domain.ChatMessage) {
	h.buf.Push(message)
}

func (h *MessageHistory) GetRecent() []domain.ChatMessage {
	return h.buf.GetValues()
}
