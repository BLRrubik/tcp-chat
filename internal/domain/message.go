package domain

import (
	"fmt"
	"time"
)

const (
	MsgTypeUser   = "user"
	MsgTypeSystem = "system"
)

type ChatMessage struct {
	Timestamp   time.Time
	ClientID    string
	Content     string
	MessageType string
}

func FormatMessage(msg ChatMessage) string {
	switch msg.MessageType {
	case MsgTypeSystem:
		return fmt.Sprintf(
			"[%s] *** %s",
			msg.Timestamp.Format("15:04:05"),
			msg.Content,
		)
	default:
		return fmt.Sprintf(
			"[%s] <%s>: %s",
			msg.Timestamp.Format("15:04:05"),
			msg.ClientID,
			msg.Content,
		)
	}
}

func ParseIncomingMessage(raw string, senderID string) ChatMessage {
	return ChatMessage{
		Timestamp:   time.Now(),
		ClientID:    senderID,
		Content:     raw,
		MessageType: MsgTypeUser,
	}
}
