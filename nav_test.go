package gox

import "testing"

func screenA() *Node { return T("A") }
func screenB() *Node { return T("B") }
func screenC() *Node { return T("C") }

func resetNav() {
	nav.stack = nil
	nav.pending = ""
}

func TestSetRootScreen(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)

	if NavStackDepth() != 1 {
		t.Fatalf("expected depth 1, got %d", NavStackDepth())
	}
	node := CurrentScreen()
	if node == nil || node.Text != "A" {
		t.Fatal("expected screen A")
	}
	if action := PendingNavAction(); action != "" {
		t.Fatalf("expected no pending action, got %q", action)
	}
}

func TestNavigatePush(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB)

	if NavStackDepth() != 2 {
		t.Fatalf("expected depth 2, got %d", NavStackDepth())
	}
	node := CurrentScreen()
	if node == nil || node.Text != "B" {
		t.Fatal("expected screen B")
	}
	if action := PendingNavAction(); action != "push" {
		t.Fatalf("expected push, got %q", action)
	}
	// pending should be cleared after read
	if action := PendingNavAction(); action != "" {
		t.Fatalf("expected empty after read, got %q", action)
	}
}

func TestGoBack(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB)
	PendingNavAction() // clear

	GoBack()

	if NavStackDepth() != 1 {
		t.Fatalf("expected depth 1, got %d", NavStackDepth())
	}
	node := CurrentScreen()
	if node == nil || node.Text != "A" {
		t.Fatal("expected screen A")
	}
	if action := PendingNavAction(); action != "pop" {
		t.Fatalf("expected pop, got %q", action)
	}
}

func TestGoBackAtRoot(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)

	GoBack() // should be no-op

	if NavStackDepth() != 1 {
		t.Fatalf("expected depth 1, got %d", NavStackDepth())
	}
	if action := PendingNavAction(); action != "" {
		t.Fatalf("expected no pending action, got %q", action)
	}
}

func TestHandleBack(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB)
	PendingNavAction() // clear

	HandleBack()

	if NavStackDepth() != 1 {
		t.Fatalf("expected depth 1, got %d", NavStackDepth())
	}
	// HandleBack should NOT set pending (bridge already handled native pop)
	if action := PendingNavAction(); action != "" {
		t.Fatalf("expected no pending action, got %q", action)
	}
}

func TestMultipleNavigations(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB)
	Navigate(screenC) // last one wins

	if NavStackDepth() != 3 {
		t.Fatalf("expected depth 3, got %d", NavStackDepth())
	}
	node := CurrentScreen()
	if node == nil || node.Text != "C" {
		t.Fatal("expected screen C")
	}
	// last pending action wins
	if action := PendingNavAction(); action != "push" {
		t.Fatalf("expected push, got %q", action)
	}
}

func TestCurrentScreenEmptyStack(t *testing.T) {
	resetNav()
	if CurrentScreen() != nil {
		t.Fatal("expected nil for empty stack")
	}
}

func TestHandleBackAtRoot(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)

	HandleBack() // should be no-op at root

	if NavStackDepth() != 1 {
		t.Fatalf("expected depth 1, got %d", NavStackDepth())
	}
}
