package gox

// SetState runs a state mutation function and triggers a UI re-render.
// Safe to call from any goroutine — the re-render is dispatched to the
// main thread by the bridge.
//
// Usage:
//
//	go func() {
//	    data := fetchData(ctx)
//	    gox.SetState(func() {
//	        posts = data
//	    })
//	}()
func SetState(fn func()) {
	fn()
	Rerender()
}
