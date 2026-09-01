package types

import (
	"reflect"
	"testing"
)

func TestGetValues_TwoPushesNeverPopped(t *testing.T) {
	cb := NewCycleBuffer[int](5)
	cb.Push(10)
	cb.Push(20)

	got := cb.GetValues()
	want := []int{10, 20}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGetValues_FullNoPops(t *testing.T) {
	cb := NewCycleBuffer[int](3)
	cb.Push(1)
	cb.Push(2)
	cb.Push(3)

	got := cb.GetValues()
	want := []int{1, 2, 3}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGetValues_PopThenPushWrap(t *testing.T) {
	cb := NewCycleBuffer[int](3)
	cb.Push(1)
	cb.Push(2)
	cb.Push(3)
	cb.Pop()
	cb.Push(4)

	got := cb.GetValues()
	want := []int{2, 3, 4}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGetValues_EmptyBuffer(t *testing.T) {
	cb := NewCycleBuffer[int](3)

	got := cb.GetValues()

	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestGetValues_PushPopBackToEmpty(t *testing.T) {
	cb := NewCycleBuffer[int](3)
	cb.Push(1)
	cb.Pop()

	got := cb.GetValues()

	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestGetValues_SinglePush(t *testing.T) {
	cb := NewCycleBuffer[int](5)
	cb.Push(10)

	got := cb.GetValues()
	want := []int{10}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGetValues_OverflowOverwritesOldest(t *testing.T) {
	cb := NewCycleBuffer[int](3)
	cb.Push(1)
	cb.Push(2)
	cb.Push(3)
	cb.Push(4) // буфер полон, перезаписывает 1

	got := cb.GetValues()
	want := []int{2, 3, 4}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPushPop_Basic(t *testing.T) {
	cb := NewCycleBuffer[int](3)
	cb.Push(1)
	cb.Push(2)

	if v := cb.Pop(); v != 1 {
		t.Errorf("got %d, want 1", v)
	}
	if v := cb.Pop(); v != 2 {
		t.Errorf("got %d, want 2", v)
	}
	if !cb.IsEmpty() {
		t.Errorf("expected empty after draining")
	}
}

func TestPop_OnEmptyReturnsZeroValue(t *testing.T) {
	cb := NewCycleBuffer[int](3)

	if v := cb.Pop(); v != 0 {
		t.Errorf("got %d, want 0 (zero value)", v)
	}
}
