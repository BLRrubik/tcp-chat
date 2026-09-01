package server

import (
	"sync"
	"tcp-chat/internal/domain"
	"testing"
)

func TestHub_ConcurrentBroadcastAndReads_NoRace(t *testing.T) {
	h := NewHub()
	go h.Run()

	var wg sync.WaitGroup

	for i := 0; i < 200; i++ {
		wg.Add(3)
		go func(n int) {
			defer wg.Done()
			h.Broadcast(domain.ChatMessage{Content: "hi", MessageType: domain.MsgTypeUser})
		}(i)
		go func() {
			defer wg.Done()
			_ = h.GetMessageHistory()
		}()
		go func() {
			defer wg.Done()
			_ = h.GetActiveClients()
		}()
	}
	wg.Wait()
}
