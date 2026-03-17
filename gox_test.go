package gox

import "testing"

func TestE(t *testing.T) {
	node := E("View", nil,
		E("Text", nil,
			T("Hello World"),
		),
	)

	if node.Type != NodeElement {
		t.Errorf("expected NodeElement, got %d", node.Type)
	}
	if node.Tag != "View" {
		t.Errorf("expected tag View, got %s", node.Tag)
	}
	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(node.Children))
	}

	text := node.Children[0]
	if text.Tag != "Text" {
		t.Errorf("expected tag Text, got %s", text.Tag)
	}
	if len(text.Children) != 1 {
		t.Fatalf("expected 1 child in Text, got %d", len(text.Children))
	}
	if text.Children[0].Text != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", text.Children[0].Text)
	}
}

func TestEWithProps(t *testing.T) {
	node := E("Image", P{"src": "photo.jpg", "style": Style{Width: 100}})

	if node.Props["src"] != "photo.jpg" {
		t.Error("missing src prop")
	}
	s, ok := node.Props["style"].(Style)
	if !ok || s.Width != 100 {
		t.Error("missing or wrong style prop")
	}
}

func TestT(t *testing.T) {
	node := T("hello")
	if node.Type != NodeText {
		t.Error("expected NodeText")
	}
	if node.Text != "hello" {
		t.Errorf("expected 'hello', got %q", node.Text)
	}
}

func TestFragment(t *testing.T) {
	node := Fragment(T("one"), T("two"), T("three"))
	if node.Type != NodeFragment {
		t.Error("expected NodeFragment")
	}
	if len(node.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(node.Children))
	}
}

func TestIf(t *testing.T) {
	// True
	node := If(true, func() *Node { return T("yes") })
	if node == nil || node.Text != "yes" {
		t.Error("If(true) should return the node")
	}

	// False
	node = If(false, func() *Node { return T("no") })
	if node != nil {
		t.Error("If(false) should return nil")
	}
}

func TestForEach(t *testing.T) {
	items := []string{"a", "b", "c"}
	node := ForEach(func() []*Node {
		var nodes []*Node
		for _, item := range items {
			nodes = append(nodes, T(item))
		}
		return nodes
	})

	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.Children))
	}
	if node.Children[0].Text != "a" {
		t.Errorf("expected 'a', got %q", node.Children[0].Text)
	}
}

func TestFilterNil(t *testing.T) {
	node := Fragment(T("a"), nil, T("b"), nil)
	if len(node.Children) != 2 {
		t.Errorf("expected 2 children after nil filter, got %d", len(node.Children))
	}
}

func TestMerge(t *testing.T) {
	base := Style{Padding: 16, BackgroundColor: "#FFF", FontSize: 14}
	override := Style{BackgroundColor: "#000", FontWeight: "bold"}

	merged := Merge(base, override)

	if merged.Padding != 16 {
		t.Error("Padding should be preserved from base")
	}
	if merged.BackgroundColor != "#000" {
		t.Error("BackgroundColor should be overridden")
	}
	if merged.FontSize != 14 {
		t.Error("FontSize should be preserved from base")
	}
	if merged.FontWeight != "bold" {
		t.Error("FontWeight should come from override")
	}
}

func TestWhen(t *testing.T) {
	if When(true, "a", "b") != "a" {
		t.Error("When(true) should return first value")
	}
	if When(false, "a", "b") != "b" {
		t.Error("When(false) should return second value")
	}
	if When(true, 1.0, 0.5) != 1.0 {
		t.Error("When should work with float64")
	}
}
