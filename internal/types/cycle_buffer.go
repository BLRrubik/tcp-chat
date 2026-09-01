package types

type CycleBuffer[T any] struct {
	readPtr  int
	writePtr int
	count    int
	buf      []T
}

func NewCycleBuffer[T any](size int) *CycleBuffer[T] {
	return &CycleBuffer[T]{
		buf: make([]T, size),
	}
}

func (cb *CycleBuffer[T]) Push(val T) {
	cb.buf[cb.writePtr] = val
	cb.writePtr = (cb.writePtr + 1) % len(cb.buf)

	if cb.count == len(cb.buf) {
		cb.readPtr = cb.writePtr
	} else {
		cb.count++
	}
}

func (cb *CycleBuffer[T]) Pop() T {
	if cb.count == 0 {
		var zero T

		return zero
	}

	val := cb.buf[cb.readPtr]
	cb.readPtr = (cb.readPtr + 1) % len(cb.buf)
	cb.count--

	return val
}

func (cb *CycleBuffer[T]) IsEmpty() bool {
	return cb.count == 0
}

func (cb *CycleBuffer[T]) GetValues() []T {
	vals := make([]T, 0, cb.count)
	idx := cb.readPtr

	for range cb.count {
		vals = append(vals, cb.buf[idx])
		idx = (idx + 1) % len(cb.buf)
	}

	return vals
}
