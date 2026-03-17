package gox

import "context"

// Screen is a function that returns a rendered node tree.
// Every screen in a GOX app is just a component function.
type Screen func() *Node

// Navigation stack — manages screen history for push/pop navigation.
var nav struct {
	stack   []screenEntry
	pending string // "push", "pop", ""
	title   string // title for the next pushed screen
}

// Navigate pushes a new screen onto the navigation stack.
// An optional title sets the native navigation bar title.
// Fires OnDisappear on the current top screen before pushing.
func Navigate(screen Screen, title ...string) {
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
	if len(title) > 0 {
		nav.title = title[0]
	} else {
		nav.title = ""
	}
}

// PendingNavTitle returns and clears the pending navigation title.
func PendingNavTitle() string {
	t := nav.title
	nav.title = ""
	return t
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
