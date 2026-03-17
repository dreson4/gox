package gox

import "testing"

func TestSetStateRunsMutation(t *testing.T) {
	count := 0
	SetState(func() {
		count = 42
	})
	if count != 42 {
		t.Fatalf("expected 42, got %d", count)
	}
}

func TestSetStateTriggersRerender(t *testing.T) {
	rendered := false
	SetRerender(func() {
		rendered = true
	})
	defer SetRerender(nil)

	SetState(func() {})
	if !rendered {
		t.Fatal("SetState should trigger Rerender")
	}
}

func TestSetStateMultipleMutations(t *testing.T) {
	rerenderCount := 0
	SetRerender(func() {
		rerenderCount++
	})
	defer SetRerender(nil)

	a, b := 0, 0
	SetState(func() {
		a = 1
		b = 2
	})
	if a != 1 || b != 2 {
		t.Fatalf("expected a=1 b=2, got a=%d b=%d", a, b)
	}
	if rerenderCount != 1 {
		t.Fatalf("expected 1 rerender, got %d", rerenderCount)
	}
}
