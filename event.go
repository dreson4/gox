package gox

import "sync"

// Event registry — maps view IDs to callback functions.
// Populated during layout computation, called when native events fire.

var (
	eventMu            sync.Mutex
	eventCallbacks     map[int]func()
	textEventCallbacks map[int]func(string)
	submitCallbacks    map[int]func()
	focusCallbacks     map[int]func()
	blurCallbacks      map[int]func()
	loadCallbacks      map[int]func()
	errorCallbacks     map[int]func()
	rerenderFn         func()
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

// RegisterTextEvent stores a text callback (onChange) for a view ID.
func RegisterTextEvent(id int, callback func(string)) {
	eventMu.Lock()
	defer eventMu.Unlock()
	if textEventCallbacks == nil {
		textEventCallbacks = make(map[int]func(string))
	}
	textEventCallbacks[id] = callback
}

// RegisterSubmitEvent stores a submit callback for a view ID.
func RegisterSubmitEvent(id int, callback func()) {
	eventMu.Lock()
	defer eventMu.Unlock()
	if submitCallbacks == nil {
		submitCallbacks = make(map[int]func())
	}
	submitCallbacks[id] = callback
}

// RegisterFocusEvent stores a focus callback for a view ID.
func RegisterFocusEvent(id int, callback func()) {
	eventMu.Lock()
	defer eventMu.Unlock()
	if focusCallbacks == nil {
		focusCallbacks = make(map[int]func())
	}
	focusCallbacks[id] = callback
}

// RegisterBlurEvent stores a blur callback for a view ID.
func RegisterBlurEvent(id int, callback func()) {
	eventMu.Lock()
	defer eventMu.Unlock()
	if blurCallbacks == nil {
		blurCallbacks = make(map[int]func())
	}
	blurCallbacks[id] = callback
}

// RegisterLoadEvent stores an onLoad callback for a view ID.
func RegisterLoadEvent(id int, callback func()) {
	eventMu.Lock()
	defer eventMu.Unlock()
	if loadCallbacks == nil {
		loadCallbacks = make(map[int]func())
	}
	loadCallbacks[id] = callback
}

// RegisterErrorEvent stores an onError callback for a view ID.
func RegisterErrorEvent(id int, callback func()) {
	eventMu.Lock()
	defer eventMu.Unlock()
	if errorCallbacks == nil {
		errorCallbacks = make(map[int]func())
	}
	errorCallbacks[id] = callback
}

// ClearEvents removes all registered callbacks.
// Called before each render to avoid stale references.
func ClearEvents() {
	eventMu.Lock()
	defer eventMu.Unlock()
	eventCallbacks = make(map[int]func())
	textEventCallbacks = make(map[int]func(string))
	submitCallbacks = make(map[int]func())
	focusCallbacks = make(map[int]func())
	blurCallbacks = make(map[int]func())
	loadCallbacks = make(map[int]func())
	errorCallbacks = make(map[int]func())
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

// HandleTextEvent fires the text callback for a view ID with the new text value.
// Called from the native bridge when a text input's content changes.
func HandleTextEvent(id int, text string) {
	eventMu.Lock()
	cb, ok := textEventCallbacks[id]
	eventMu.Unlock()

	if ok && cb != nil {
		cb(text)
	}
}

// HandleSubmitEvent fires the submit callback for a view ID.
// Called from the native bridge when return key is pressed in a text input.
func HandleSubmitEvent(id int) {
	eventMu.Lock()
	cb, ok := submitCallbacks[id]
	eventMu.Unlock()

	if ok && cb != nil {
		cb()
	}
}

// HandleFocusEvent fires the focus callback for a view ID.
// Called from the native bridge when a text input gains focus.
func HandleFocusEvent(id int) {
	eventMu.Lock()
	cb, ok := focusCallbacks[id]
	eventMu.Unlock()

	if ok && cb != nil {
		cb()
	}
}

// HandleBlurEvent fires the blur callback for a view ID.
// Called from the native bridge when a text input loses focus.
func HandleBlurEvent(id int) {
	eventMu.Lock()
	cb, ok := blurCallbacks[id]
	eventMu.Unlock()

	if ok && cb != nil {
		cb()
	}
}

// HandleLoadEvent fires the onLoad callback for a view ID.
func HandleLoadEvent(id int) {
	eventMu.Lock()
	cb, ok := loadCallbacks[id]
	eventMu.Unlock()

	if ok && cb != nil {
		cb()
	}
}

// HandleErrorEvent fires the onError callback for a view ID.
func HandleErrorEvent(id int) {
	eventMu.Lock()
	cb, ok := errorCallbacks[id]
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
