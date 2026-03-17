package gox

import "context"

// Screen is a function that returns a rendered node tree.
// Every screen in a GOX app is just a component function.
type Screen func() *Node

// NavigateOptions configures the navigation bar and presentation for a pushed screen.
// All fields are optional — zero value means "use default".
type NavigateOptions struct {
	Title          string       // nav bar title
	LargeTitle     *bool        // nil=inherit, true=always, false=never
	HeaderShown    *bool        // nil/true=show nav bar, false=hide
	BackTitle      string       // override back button text on NEXT pushed screen
	BackVisible    *bool        // nil/true=show back button, false=hide
	GestureEnabled *bool        // nil/true=allow swipe-back, false=disable
	Presentation   string       // "push" (default) or "modal" (future)
	HeaderStyle    *HeaderStyle // nav bar appearance
	RightButtons   []BarButton  // right side bar button items
	LeftButtons    []BarButton  // left side bar button items
}

// HeaderStyle configures the navigation bar's visual appearance.
type HeaderStyle struct {
	BackgroundColor string // hex color for nav bar background
	TintColor       string // color for buttons and back arrow
	TitleColor      string // color for title text
}

// BarButton defines a navigation bar button item.
type BarButton struct {
	Title      string // text label (use Title OR SystemItem, not both)
	SystemItem string // system icon: "done", "cancel", "edit", "add", "close", "search", "compose"
	OnPress    func() // callback when tapped
	eventID    int    // assigned by registerNavEvents, used for JSON serialization
}

// EventID returns the assigned event ID for this bar button.
// Used by the bootstrap for JSON serialization.
func (b BarButton) EventID() int {
	return b.eventID
}

// Nav creates NavigateOptions with just a title — the common case.
func Nav(title string) NavigateOptions {
	return NavigateOptions{Title: title}
}

// Bool returns a pointer to a bool value, for use with optional bool fields.
func Bool(v bool) *bool {
	return &v
}

// Navigation stack — manages screen history for push/pop navigation.
var nav struct {
	stack   []screenEntry
	pending string            // "push", "pop", ""
	options *NavigateOptions  // options for the next pushed screen
}

// Counter for bar button event IDs (negative to avoid collision with layout IDs).
var navEventCounter int

// Navigate pushes a new screen onto the navigation stack.
// Accepts optional NavigateOptions to configure the navigation bar.
// Fires OnDisappear on the current top screen before pushing.
func Navigate(screen Screen, opts ...NavigateOptions) {
	// Fire OnDisappear on current top screen
	if len(nav.stack) > 0 {
		top := &nav.stack[len(nav.stack)-1]
		if top.lifecycleSet && top.callbacks.OnDisappear != nil {
			top.callbacks.OnDisappear()
		}
	}

	// Create new context for the new screen
	ctx, cancel := context.WithCancel(context.Background())
	nav.stack = append(nav.stack, screenEntry{
		screen: screen,
		ctx:    ctx,
		cancel: cancel,
	})
	nav.pending = "push"
	if len(opts) > 0 {
		o := opts[0]
		registerNavEvents(&o)
		nav.options = &o
	} else {
		nav.options = nil
	}
}

// registerNavEvents assigns negative event IDs to bar button callbacks
// and registers them in the event system.
func registerNavEvents(opts *NavigateOptions) {
	if opts == nil {
		return
	}
	for i := range opts.RightButtons {
		if opts.RightButtons[i].OnPress != nil {
			navEventCounter--
			RegisterEvent(navEventCounter, opts.RightButtons[i].OnPress)
			opts.RightButtons[i].eventID = navEventCounter
		}
	}
	for i := range opts.LeftButtons {
		if opts.LeftButtons[i].OnPress != nil {
			navEventCounter--
			RegisterEvent(navEventCounter, opts.LeftButtons[i].OnPress)
			opts.LeftButtons[i].eventID = navEventCounter
		}
	}
}

// PendingNavOptions returns and clears the pending navigation options.
func PendingNavOptions() *NavigateOptions {
	opts := nav.options
	nav.options = nil
	return opts
}

// GoBack pops the current screen and returns to the previous one.
// Does nothing if already at the root screen.
// Fires OnDisappear + OnUnmount on the popped screen, OnAppear on revealed screen.
func GoBack() {
	if len(nav.stack) <= 1 {
		return
	}

	// Pop the top screen
	popped := &nav.stack[len(nav.stack)-1]

	// Cancel context first (goroutines stop)
	if popped.cancel != nil {
		popped.cancel()
	}

	// Fire OnDisappear + OnUnmount
	if popped.lifecycleSet {
		if popped.callbacks.OnDisappear != nil {
			popped.callbacks.OnDisappear()
		}
		if popped.callbacks.OnUnmount != nil {
			popped.callbacks.OnUnmount()
		}
	}

	nav.stack = nav.stack[:len(nav.stack)-1]
	nav.pending = "pop"

	// Fire OnAppear on revealed screen
	if len(nav.stack) > 0 {
		revealed := &nav.stack[len(nav.stack)-1]
		if revealed.lifecycleSet && revealed.callbacks.OnAppear != nil {
			revealed.callbacks.OnAppear(revealed.ctx)
		}
	}
}

// SetRootScreen sets the initial screen of the app.
// Called once during bootstrap initialization.
func SetRootScreen(screen Screen) {
	ctx, cancel := context.WithCancel(context.Background())
	nav.stack = []screenEntry{{
		screen: screen,
		ctx:    ctx,
		cancel: cancel,
	}}
	nav.pending = ""
}

// CurrentScreen renders the top screen on the navigation stack.
// Returns nil if the stack is empty.
func CurrentScreen() *Node {
	if len(nav.stack) == 0 {
		return nil
	}
	return nav.stack[len(nav.stack)-1].screen()
}

// PendingNavAction returns and clears the pending navigation action.
// Returns "push", "pop", or "".
func PendingNavAction() string {
	action := nav.pending
	nav.pending = ""
	return action
}

// HandleBack synchronizes the Go nav stack when the bridge detects
// a native back gesture (iOS swipe-back). Pops without setting pending
// since the bridge already handled the native pop.
// Fires the same lifecycle as GoBack (OnDisappear + OnUnmount on popped, OnAppear on revealed).
func HandleBack() {
	if len(nav.stack) <= 1 {
		return
	}

	// Pop the top screen
	popped := &nav.stack[len(nav.stack)-1]

	// Cancel context
	if popped.cancel != nil {
		popped.cancel()
	}

	// Fire OnDisappear + OnUnmount
	if popped.lifecycleSet {
		if popped.callbacks.OnDisappear != nil {
			popped.callbacks.OnDisappear()
		}
		if popped.callbacks.OnUnmount != nil {
			popped.callbacks.OnUnmount()
		}
	}

	nav.stack = nav.stack[:len(nav.stack)-1]
	// No pending action — bridge already handled native pop

	// Fire OnAppear on revealed screen
	if len(nav.stack) > 0 {
		revealed := &nav.stack[len(nav.stack)-1]
		if revealed.lifecycleSet && revealed.callbacks.OnAppear != nil {
			revealed.callbacks.OnAppear(revealed.ctx)
		}
	}
}

// NavStackDepth returns the current depth of the navigation stack.
func NavStackDepth() int {
	return len(nav.stack)
}
