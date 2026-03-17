package gox

import "testing"

func TestRegisterAndHandleEvent(t *testing.T) {
	ClearEvents()
	called := false
	RegisterEvent(1, func() { called = true })
	HandleEvent(1)
	if !called {
		t.Error("expected event callback to be called")
	}
}

func TestHandleEventMissingID(t *testing.T) {
	ClearEvents()
	// Should not panic
	HandleEvent(999)
}

func TestRegisterAndHandleTextEvent(t *testing.T) {
	ClearEvents()
	var received string
	RegisterTextEvent(1, func(text string) { received = text })
	HandleTextEvent(1, "hello world")
	if received != "hello world" {
		t.Errorf("expected 'hello world', got %q", received)
	}
}

func TestHandleTextEventMissingID(t *testing.T) {
	ClearEvents()
	// Should not panic
	HandleTextEvent(999, "test")
}

func TestRegisterAndHandleSubmitEvent(t *testing.T) {
	ClearEvents()
	called := false
	RegisterSubmitEvent(1, func() { called = true })
	HandleSubmitEvent(1)
	if !called {
		t.Error("expected submit callback to be called")
	}
}

func TestRegisterAndHandleFocusEvent(t *testing.T) {
	ClearEvents()
	called := false
	RegisterFocusEvent(1, func() { called = true })
	HandleFocusEvent(1)
	if !called {
		t.Error("expected focus callback to be called")
	}
}

func TestRegisterAndHandleBlurEvent(t *testing.T) {
	ClearEvents()
	called := false
	RegisterBlurEvent(1, func() { called = true })
	HandleBlurEvent(1)
	if !called {
		t.Error("expected blur callback to be called")
	}
}

func TestClearEventsAllTypes(t *testing.T) {
	RegisterEvent(1, func() {})
	RegisterTextEvent(2, func(string) {})
	RegisterSubmitEvent(3, func() {})
	RegisterFocusEvent(4, func() {})
	RegisterBlurEvent(5, func() {})
	RegisterLoadEvent(6, func() {})
	RegisterErrorEvent(7, func() {})

	ClearEvents()

	// After clear, no callbacks should fire
	eventMu.Lock()
	if len(eventCallbacks) != 0 {
		t.Error("eventCallbacks should be empty after clear")
	}
	if len(textEventCallbacks) != 0 {
		t.Error("textEventCallbacks should be empty after clear")
	}
	if len(submitCallbacks) != 0 {
		t.Error("submitCallbacks should be empty after clear")
	}
	if len(focusCallbacks) != 0 {
		t.Error("focusCallbacks should be empty after clear")
	}
	if len(blurCallbacks) != 0 {
		t.Error("blurCallbacks should be empty after clear")
	}
	if len(loadCallbacks) != 0 {
		t.Error("loadCallbacks should be empty after clear")
	}
	if len(errorCallbacks) != 0 {
		t.Error("errorCallbacks should be empty after clear")
	}
	eventMu.Unlock()
}

func TestRegisterAndHandleLoadEvent(t *testing.T) {
	ClearEvents()
	called := false
	RegisterLoadEvent(1, func() { called = true })
	HandleLoadEvent(1)
	if !called {
		t.Error("expected load callback to be called")
	}
}

func TestRegisterAndHandleErrorEvent(t *testing.T) {
	ClearEvents()
	called := false
	RegisterErrorEvent(1, func() { called = true })
	HandleErrorEvent(1)
	if !called {
		t.Error("expected error callback to be called")
	}
}

func TestMultipleEventsSameViewID(t *testing.T) {
	ClearEvents()
	var textVal string
	submitCalled := false
	focusCalled := false
	blurCalled := false

	RegisterTextEvent(5, func(text string) { textVal = text })
	RegisterSubmitEvent(5, func() { submitCalled = true })
	RegisterFocusEvent(5, func() { focusCalled = true })
	RegisterBlurEvent(5, func() { blurCalled = true })

	HandleTextEvent(5, "typed")
	HandleSubmitEvent(5)
	HandleFocusEvent(5)
	HandleBlurEvent(5)

	if textVal != "typed" {
		t.Errorf("text event: expected 'typed', got %q", textVal)
	}
	if !submitCalled {
		t.Error("submit event not called")
	}
	if !focusCalled {
		t.Error("focus event not called")
	}
	if !blurCalled {
		t.Error("blur event not called")
	}
}
