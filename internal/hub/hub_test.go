package hub

import (
	"net"
	"sync"
	"tcp-chat/internal/domain"
	"tcp-chat/internal/logging"
	"testing"
	"time"
)

func TestHub_ConcurrentBroadcastAndReads_NoRace(t *testing.T) {
	h := NewHub(logging.SetupLogging("info"), 50, 0)
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

type panicOnWriteConn struct {
	net.Conn
}

func (c *panicOnWriteConn) Write(_ []byte) (int, error) {
	panic("boom: simulated client write failure")
}

func TestHandleConnection_PanicRecovered_HubKeepsRunning(t *testing.T) {
	h := NewHub(logging.SetupLogging("info"), 50, 0)
	go h.Run()

	client, srv := net.Pipe()
	defer client.Close()

	done := make(chan struct{})
	go func() {
		h.HandleConnection(&panicOnWriteConn{srv})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("HandleConnection did not return after panic; recover() missing or broken")
	}

	if count := h.GetClientCount(); count != 0 {
		t.Errorf("expected 0 clients after panicking connection, got %d", count)
	}

	if got := h.GetActiveClients(); len(got) != 0 {
		t.Errorf("hub unresponsive or corrupted after panic: %v", got)
	}
}
