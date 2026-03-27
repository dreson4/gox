package clipboard

import "testing"

func TestCopyAndRead(t *testing.T) {
	Reset()

	Copy("hello world")
	got := Read()
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestReadEmpty(t *testing.T) {
	Reset()

	got := Read()
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestCopyOverwrite(t *testing.T) {
	Reset()

	Copy("first")
	Copy("second")
	got := Read()
	if got != "second" {
		t.Errorf("got %q, want %q", got, "second")
	}
}
