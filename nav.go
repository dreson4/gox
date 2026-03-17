package gox

// Screen is a function that returns a rendered node tree.
// Every screen in a GOX app is just a component function.
type Screen func() *Node

// Navigation stack — manages screen history for push/pop navigation.
var nav struct {
	stack   []Screen
	pending string // "push", "pop", ""
	title   string // title for the next pushed screen
}

// Navigate pushes a new screen onto the navigation stack.
// An optional title sets the native navigation bar title.
func Navigate(screen Screen, title ...string) {
	nav.stack = append(nav.stack, screen)
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
func GoBack() {
	if len(nav.stack) <= 1 {
		return
	}
	nav.stack = nav.stack[:len(nav.stack)-1]
	nav.pending = "pop"
}

// SetRootScreen sets the initial screen of the app.
// Called once during bootstrap initialization.
func SetRootScreen(screen Screen) {
	nav.stack = []Screen{screen}
	nav.pending = ""
}

// CurrentScreen renders the top screen on the navigation stack.
// Returns nil if the stack is empty.
func CurrentScreen() *Node {
	if len(nav.stack) == 0 {
		return nil
	}
	return nav.stack[len(nav.stack)-1]()
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
func HandleBack() {
	if len(nav.stack) <= 1 {
		return
	}
	nav.stack = nav.stack[:len(nav.stack)-1]
}

// NavStackDepth returns the current depth of the navigation stack.
func NavStackDepth() int {
	return len(nav.stack)
}
