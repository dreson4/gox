package platform

import (
	"github.com/dreson4/gox"
	"testing"
)

// mockBackend records all calls for testing.
type mockBackend struct {
	calls    []call
	nextID   ViewHandle
	rootView ViewHandle
}

type call struct {
	method string
	args   []any
}

func newMock() *mockBackend {
	return &mockBackend{nextID: 1}
}

func (m *mockBackend) record(method string, args ...any) {
	m.calls = append(m.calls, call{method: method, args: args})
}

func (m *mockBackend) CreateView() ViewHandle {
	id := m.nextID
	m.nextID++
	m.record("CreateView", id)
	return id
}

func (m *mockBackend) CreateText(content string) ViewHandle {
	id := m.nextID
	m.nextID++
	m.record("CreateText", id, content)
	return id
}

func (m *mockBackend) AddChild(parent, child ViewHandle) {
	m.record("AddChild", parent, child)
}

func (m *mockBackend) SetRootView(handle ViewHandle) {
	m.rootView = handle
	m.record("SetRootView", handle)
}

func (m *mockBackend) SetBackgroundColor(h ViewHandle, c string) {
	m.record("SetBackgroundColor", h, c)
}
func (m *mockBackend) SetFrame(h ViewHandle, x, y, w, ht float64) {
	m.record("SetFrame", h, x, y, w, ht)
}
func (m *mockBackend) SetPadding(h ViewHandle, t, r, b, l float64) {
	m.record("SetPadding", h, t, r, b, l)
}
func (m *mockBackend) SetFontSize(h ViewHandle, s float64) {
	m.record("SetFontSize", h, s)
}
func (m *mockBackend) SetFontWeight(h ViewHandle, w string) {
	m.record("SetFontWeight", h, w)
}
func (m *mockBackend) SetTextColor(h ViewHandle, c string) {
	m.record("SetTextColor", h, c)
}
func (m *mockBackend) SetTextAlign(h ViewHandle, a string) {
	m.record("SetTextAlign", h, a)
}
func (m *mockBackend) SetBorderRadius(h ViewHandle, r float64) {
	m.record("SetBorderRadius", h, r)
}
func (m *mockBackend) SetOpacity(h ViewHandle, o float64) {
	m.record("SetOpacity", h, o)
}
func (m *mockBackend) SetFlexLayout(h ViewHandle, dir, justify, align string) {
	m.record("SetFlexLayout", h, dir, justify, align)
}
func (m *mockBackend) RunApp() {
	m.record("RunApp")
}

func (m *mockBackend) findCall(method string) *call {
	for i := range m.calls {
		if m.calls[i].method == method {
			return &m.calls[i]
		}
	}
	return nil
}

func (m *mockBackend) countCalls(method string) int {
	n := 0
	for _, c := range m.calls {
		if c.method == method {
			n++
		}
	}
	return n
}

func TestRenderHelloWorld(t *testing.T) {
	mock := newMock()
	r := NewRenderer(mock)

	// Simulates generated code: gox.E("View", nil, gox.E("Text", nil, gox.T("Hello World")))
	tree := gox.E("View", nil,
		gox.E("Text", nil,
			gox.T("Hello World"),
		),
	)

	r.RenderToScreen(tree)

	// Should create a View and a Text(label)
	if mock.countCalls("CreateView") != 1 {
		t.Errorf("expected 1 CreateView, got %d", mock.countCalls("CreateView"))
	}
	if mock.countCalls("CreateText") != 1 {
		t.Errorf("expected 1 CreateText, got %d", mock.countCalls("CreateText"))
	}

	// Text should have "Hello World"
	textCall := mock.findCall("CreateText")
	if textCall == nil {
		t.Fatal("missing CreateText call")
	}
	if textCall.args[1] != "Hello World" {
		t.Errorf("expected text 'Hello World', got %v", textCall.args[1])
	}

	// Text should be added as child of View
	if mock.countCalls("AddChild") != 1 {
		t.Errorf("expected 1 AddChild, got %d", mock.countCalls("AddChild"))
	}

	// Root should be set
	if mock.rootView == 0 {
		t.Error("expected root view to be set")
	}
}

func TestRenderWithStyle(t *testing.T) {
	mock := newMock()
	r := NewRenderer(mock)

	tree := gox.E("View", gox.P{
		"style": gox.Style{BackgroundColor: "#F5F5F5", Padding: 16},
	},
		gox.E("Text", gox.P{
			"style": gox.Style{FontSize: 28, FontWeight: "bold", Color: "#111"},
		},
			gox.T("Hello"),
		),
	)

	r.Render(tree)

	// View should have background color set
	bgCall := mock.findCall("SetBackgroundColor")
	if bgCall == nil {
		t.Fatal("missing SetBackgroundColor")
	}
	if bgCall.args[1] != "#F5F5F5" {
		t.Errorf("expected #F5F5F5, got %v", bgCall.args[1])
	}

	// View should have padding
	if mock.findCall("SetPadding") == nil {
		t.Error("missing SetPadding")
	}

	// Text should have font size
	if mock.findCall("SetFontSize") == nil {
		t.Error("missing SetFontSize")
	}

	// Text should have color
	colorCall := mock.findCall("SetTextColor")
	if colorCall == nil {
		t.Fatal("missing SetTextColor")
	}
}

func TestRenderNested(t *testing.T) {
	mock := newMock()
	r := NewRenderer(mock)

	tree := gox.E("View", nil,
		gox.E("View", nil,
			gox.E("Text", nil, gox.T("Deep")),
		),
	)

	r.Render(tree)

	// 2 Views + 1 Text
	if mock.countCalls("CreateView") != 2 {
		t.Errorf("expected 2 CreateView, got %d", mock.countCalls("CreateView"))
	}
	if mock.countCalls("CreateText") != 1 {
		t.Errorf("expected 1 CreateText, got %d", mock.countCalls("CreateText"))
	}
	// 2 AddChild calls (inner view → outer view, text → inner view)
	if mock.countCalls("AddChild") != 2 {
		t.Errorf("expected 2 AddChild, got %d", mock.countCalls("AddChild"))
	}
}

func TestRenderConditional(t *testing.T) {
	mock := newMock()
	r := NewRenderer(mock)

	// If true → should render
	tree := gox.If(true, func() *gox.Node {
		return gox.E("Text", nil, gox.T("Visible"))
	})
	r.Render(tree)
	if mock.countCalls("CreateText") != 1 {
		t.Error("If(true) should create the text")
	}

	// If false → nil, should not crash
	mock2 := newMock()
	r2 := NewRenderer(mock2)
	tree2 := gox.If(false, func() *gox.Node {
		return gox.E("Text", nil, gox.T("Hidden"))
	})
	r2.Render(tree2)
	if mock2.countCalls("CreateText") != 0 {
		t.Error("If(false) should not create any views")
	}
}

func TestRenderNil(t *testing.T) {
	mock := newMock()
	r := NewRenderer(mock)
	handle := r.Render(nil)
	if handle != 0 {
		t.Error("nil node should return 0 handle")
	}
}
