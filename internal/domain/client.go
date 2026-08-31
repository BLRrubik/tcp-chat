package domain

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

var counter atomic.Uint32

type Client struct {
	ID       string
	Conn     net.Conn
	JoinTime time.Time
}

func GenerateClientID() string {
	counter.Add(1)

	return fmt.Sprintf("User_%d", counter.Load())
}
