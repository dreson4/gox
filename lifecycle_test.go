package gox

import (
	"context"
	"testing"
)

func resetLifecycle() {
	resetNav()
	appLifecycle.foreground = nil
	appLifecycle.background = nil
}

func TestUseLifecycleIdempotent(t *testing.T) {
	resetLifecycle()
	SetRootScreen(screenA)

	// First call registers callbacks
	ctx1 := UseLifecycle(ScreenCallbacks{
		OnMount: func(ctx context.Context) {},
	})

	// Second call returns same context without re-registering
	ctx2 := UseLifecycle(ScreenCallbacks{
		OnMount: func(ctx context.Context) { t.Fatal("should not replace") },
	})

	if ctx1 != ctx2 {
		t.Fatal("UseLifecycle should return same context on re-renders")
	}
}

func TestUseLifecycleEmptyStack(t *testing.T) {
	resetLifecycle()
	ctx := UseLifecycle(ScreenCallbacks{})
	if ctx == nil {
		t.Fatal("should return non-nil context even with empty stack")
	}
}

func TestFirePendingMountOnPush(t *testing.T) {
	resetLifecycle()
	SetRootScreen(screenA)

	mounted := false
	appeared := false

	Navigate(func() *Node {
		UseLifecycle(ScreenCallbacks{
			OnMount:  func(ctx context.Context) { mounted = true },
			OnAppear: func(ctx context.Context) { appeared = true },
		})
		return T("B")
	})

	// Simulate what bootstrap does: render then fire
	CurrentScreen()
	FirePendingMount()

	if !mounted {
		t.Fatal("OnMount should have fired")
	}
	if !appeared {
		t.Fatal("OnAppear should have fired after mount")
	}
}

func TestOnUnmountFiresOnGoBack(t *testing.T) {
	resetLifecycle()
	SetRootScreen(screenA)

	unmounted := false
	disappeared := false

	Navigate(func() *Node {
		UseLifecycle(ScreenCallbacks{
			OnUnmount:   func() { unmounted = true },
			OnDisappear: func() { disappeared = true },
		})
		return T("B")
	})
	CurrentScreen() // trigger UseLifecycle registration
	FirePendingMount()

	GoBack()

	if !disappeared {
		t.Fatal("OnDisappear should fire on GoBack")
	}
	if !unmounted {
		t.Fatal("OnUnmount should fire on GoBack")
	}
}

func TestOnUnmountFiresOnHandleBack(t *testing.T) {
	resetLifecycle()
	SetRootScreen(screenA)

	unmounted := false

	Navigate(func() *Node {
		UseLifecycle(ScreenCallbacks{
			OnUnmount: func() { unmounted = true },
		})
		return T("B")
	})
	CurrentScreen()
	FirePendingMount()

	HandleBack()

	if !unmounted {
		t.Fatal("OnUnmount should fire on HandleBack")
	}
}

func TestOnAppearFiresOnReturnFromPop(t *testing.T) {
	resetLifecycle()

	appearCount := 0
	SetRootScreen(func() *Node {
		UseLifecycle(ScreenCallbacks{
			OnAppear: func(ctx context.Context) { appearCount++ },
		})
		return T("A")
	})
	CurrentScreen()
	FirePendingMount()

	initialAppear := appearCount

	Navigate(screenB)
	GoBack()

	if appearCount != initialAppear+1 {
		t.Fatalf("OnAppear should fire on return from pop, got %d appearances after initial %d", appearCount, initialAppear)
	}
}

func TestOnDisappearFiresOnPush(t *testing.T) {
	resetLifecycle()

	disappeared := false
	SetRootScreen(func() *Node {
		UseLifecycle(ScreenCallbacks{
			OnDisappear: func() { disappeared = true },
		})
		return T("A")
	})
	CurrentScreen()
	FirePendingMount()

	Navigate(screenB)

	if !disappeared {
		t.Fatal("OnDisappear should fire on current screen when new screen is pushed")
	}
}

func TestContextCancelledOnUnmount(t *testing.T) {
	resetLifecycle()
	SetRootScreen(screenA)

	var savedCtx context.Context
	Navigate(func() *Node {
		savedCtx = UseLifecycle(ScreenCallbacks{})
		return T("B")
	})
	CurrentScreen()
	FirePendingMount()

	// Context should not be cancelled yet
	select {
	case <-savedCtx.Done():
		t.Fatal("context should not be cancelled before unmount")
	default:
	}

	GoBack()

	// Context should now be cancelled
	select {
	case <-savedCtx.Done():
		// good
	default:
		t.Fatal("context should be cancelled after unmount")
	}
}

func TestAppForegroundBackground(t *testing.T) {
	resetLifecycle()

	fgCount := 0
	bgCount := 0

	OnAppForeground(func() { fgCount++ })
	OnAppBackground(func() { bgCount++ })

	AppDidEnterBackground()
	if bgCount != 1 {
		t.Fatalf("expected 1 background callback, got %d", bgCount)
	}

	AppDidEnterForeground()
	if fgCount != 1 {
		t.Fatalf("expected 1 foreground callback, got %d", fgCount)
	}
}

func TestMultipleAppCallbacks(t *testing.T) {
	resetLifecycle()

	count := 0
	OnAppForeground(func() { count++ })
	OnAppForeground(func() { count += 10 })

	AppDidEnterForeground()
	if count != 11 {
		t.Fatalf("expected 11, got %d", count)
	}
}

func TestFirePendingMountNoopWhenNoPending(t *testing.T) {
	resetLifecycle()
	SetRootScreen(screenA)

	// No lifecycle registered, should not panic
	FirePendingMount()

	// Register and fire
	mounted := 0
	nav.stack[0].callbacks.OnMount = func(ctx context.Context) { mounted++ }
	nav.stack[0].lifecycleSet = true
	nav.stack[0].pendingMount = true
	FirePendingMount()

	if mounted != 1 {
		t.Fatalf("expected 1, got %d", mounted)
	}

	// Second call should be noop (pendingMount cleared)
	FirePendingMount()
	if mounted != 1 {
		t.Fatalf("expected still 1, got %d", mounted)
	}
}
