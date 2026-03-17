package gox

import "sync"

// Event registry — maps view IDs to callback functions.
// Populated during layout computation, called when native events fire.

var (
	eventMu        sync.Mutex
	eventCallbacks map[int]func()
	rerenderFn     func()
)

// RegisterEvent stores a callback for a view ID.
// Called by the layout computer when it encounters onPress props.
func RegisterEvent(id int, callback func()) {
	eventMu.Lock()
	defer eventMu.Unlock()
	if eventCallbacks == nil {
		eventCallbacks = make(map[int]func())
	}
	eventCallbacks[id] = callback
}

// ClearEvents removes all registered callbacks.
// Called before each render to avoid stale references.
func ClearEvents() {
	eventMu.Lock()
	defer eventMu.Unlock()
	eventCallbacks = make(map[int]func())
}

// HandleEvent fires the callback for a view ID.
// Called from the native bridge when a user interacts with a view.
func HandleEvent(id int) {
	eventMu.Lock()
	cb, ok := eventCallbacks[id]
	eventMu.Unlock()

	if ok && cb != nil {
		cb()
	}
}

// SetRerender registers the function that triggers a full re-render.
// Called once during bootstrap initialization.
func SetRerender(fn func()) {
	rerenderFn = fn
}

// Rerender triggers a full UI re-render.
// Called after state changes to update the screen.
func Rerender() {
	if rerenderFn != nil {
		rerenderFn()
	}
}
