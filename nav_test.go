package gox

import "testing"

func screenA() *Node { return T("A") }
func screenB() *Node { return T("B") }
func screenC() *Node { return T("C") }

func resetNav() {
	nav.stack = nil
	nav.pending = ""
	nav.options = nil
	navEventCounter = 0
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

func TestNavigateWithOptions(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB, NavigateOptions{
		Title:      "Details",
		LargeTitle: Bool(true),
	})

	opts := PendingNavOptions()
	if opts == nil {
		t.Fatal("expected non-nil options")
	}
	if opts.Title != "Details" {
		t.Fatalf("expected title 'Details', got %q", opts.Title)
	}
	if opts.LargeTitle == nil || *opts.LargeTitle != true {
		t.Fatal("expected LargeTitle=true")
	}
}

func TestNavigateNoOptions(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB)

	opts := PendingNavOptions()
	if opts != nil {
		t.Fatal("expected nil options when none provided")
	}
}

func TestNavHelper(t *testing.T) {
	opts := Nav("Thread")
	if opts.Title != "Thread" {
		t.Fatalf("expected 'Thread', got %q", opts.Title)
	}
}

func TestNavigateWithNav(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB, Nav("Thread"))

	opts := PendingNavOptions()
	if opts == nil || opts.Title != "Thread" {
		t.Fatal("expected title 'Thread'")
	}
}

func TestBarButtonEventRegistration(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)

	pressed := false
	Navigate(screenB, NavigateOptions{
		Title: "Test",
		RightButtons: []BarButton{
			{Title: "Edit", OnPress: func() { pressed = true }},
		},
	})

	opts := PendingNavOptions()
	if opts == nil {
		t.Fatal("expected options")
	}
	if len(opts.RightButtons) != 1 {
		t.Fatalf("expected 1 right button, got %d", len(opts.RightButtons))
	}
	if opts.RightButtons[0].eventID >= 0 {
		t.Fatalf("expected negative event ID, got %d", opts.RightButtons[0].eventID)
	}

	// Fire the registered event
	HandleEvent(opts.RightButtons[0].eventID)
	if !pressed {
		t.Fatal("bar button OnPress should have fired")
	}
}

func TestPendingNavOptionsClearsAfterRead(t *testing.T) {
	resetNav()
	SetRootScreen(screenA)
	Navigate(screenB, Nav("Test"))

	opts := PendingNavOptions()
	if opts == nil {
		t.Fatal("expected options on first read")
	}

	opts2 := PendingNavOptions()
	if opts2 != nil {
		t.Fatal("expected nil on second read")
	}
}
