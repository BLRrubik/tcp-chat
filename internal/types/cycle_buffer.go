package types

import "sync"

type CycleBuffer[T any] struct {
	readPtr  int
	writePtr int
	buf      []T

	mu sync.Mutex
}

func NewCycleBuffer[T any](size int) *CycleBuffer[T] {
	return &CycleBuffer[T]{
		readPtr:  -1,
		writePtr: -1,
		buf:      make([]T, size),
	}
}

func (cb *CycleBuffer[T]) Push(val T) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.writePtr == -1 {
		cb.writePtr = 0
	}

	cb.buf[cb.writePtr] = val

	cb.writePtr = (cb.writePtr + 1) % len(cb.buf)
}

func (cb *CycleBuffer[T]) Pop() T {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.isEmpty() {
		var zero T

		return zero
	}

	if cb.readPtr == -1 {
		cb.readPtr = 0
	}

	val := cb.buf[cb.readPtr]

	cb.readPtr = (cb.readPtr + 1) % len(cb.buf)

	if cb.readPtr == cb.writePtr {
		cb.readPtr = -1
		cb.writePtr = -1
	}

	return val
}

func (cb *CycleBuffer[T]) IsEmpty() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return cb.isEmpty()
}

func (cb *CycleBuffer[T]) isEmpty() bool {
	return cb.writePtr == -1
}
