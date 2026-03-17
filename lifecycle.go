package gox

import "context"

// ScreenCallbacks holds lifecycle functions for a screen.
// All are optional — only set the ones your screen needs.
type ScreenCallbacks struct {
	OnMount     func(ctx context.Context) // once, when screen first renders after push
	OnUnmount   func()                    // once, when screen is popped (ctx already cancelled)
	OnAppear    func(ctx context.Context) // each time screen becomes visible
	OnDisappear func()                    // each time screen is hidden
}

// screenEntry wraps a Screen with its lifecycle state.
type screenEntry struct {
	screen       Screen
	callbacks    ScreenCallbacks
	lifecycleSet bool               // true after UseLifecycle registers callbacks
	pendingMount bool               // true until OnMount fires via FirePendingMount
	ctx          context.Context    // per-screen context, cancelled on unmount
	cancel       context.CancelFunc // cancels ctx
}

// UseLifecycle registers lifecycle callbacks on the current screen entry.
// Idempotent — on re-renders, returns the existing context without re-registering.
func UseLifecycle(cb ScreenCallbacks) context.Context {
	if len(nav.stack) == 0 {
		return context.Background()
	}
	top := &nav.stack[len(nav.stack)-1]
	if top.lifecycleSet {
		return top.ctx // already registered, just return ctx
	}
	top.callbacks = cb
	top.lifecycleSet = true
	top.pendingMount = true
	return top.ctx
}

// FirePendingMount fires OnMount and OnAppear for newly pushed screens.
// Called by bootstrap after layout computation completes.
func FirePendingMount() {
	if len(nav.stack) == 0 {
		return
	}
	top := &nav.stack[len(nav.stack)-1]
	if !top.pendingMount {
		return
	}
	top.pendingMount = false
	if top.callbacks.OnMount != nil {
		top.callbacks.OnMount(top.ctx)
	}
	if top.callbacks.OnAppear != nil {
		top.callbacks.OnAppear(top.ctx)
	}
}

// App-level lifecycle callbacks.
var appLifecycle struct {
	foreground []func()
	background []func()
}

// OnAppForeground registers a callback that fires when the app enters the foreground.
func OnAppForeground(fn func()) {
	appLifecycle.foreground = append(appLifecycle.foreground, fn)
}

// OnAppBackground registers a callback that fires when the app enters the background.
func OnAppBackground(fn func()) {
	appLifecycle.background = append(appLifecycle.background, fn)
}

// AppDidEnterForeground fires all registered foreground callbacks.
// Called by the bridge when the app becomes active.
func AppDidEnterForeground() {
	for _, fn := range appLifecycle.foreground {
		fn()
	}
}

// AppDidEnterBackground fires all registered background callbacks.
// Called by the bridge when the app enters the background.
func AppDidEnterBackground() {
	for _, fn := range appLifecycle.background {
		fn()
	}
}
