package gox

import (
	"math"
	"testing"
)

var testScreen = ScreenInfo{
	Width:  390,
	Height: 844,
}

func approxF(a, b float64) bool {
	return math.Abs(a-b) < 1.0
}

func findFrame(frames []LayoutFrame, tag string) *LayoutFrame {
	for i := range frames {
		if frames[i].Tag == tag {
			return &frames[i]
		}
	}
	return nil
}

func findFrameByID(frames []LayoutFrame, id int) *LayoutFrame {
	for i := range frames {
		if frames[i].ID == id {
			return &frames[i]
		}
	}
	return nil
}

func TestLayoutBasicColumn(t *testing.T) {
	tree := E("View", P{"style": Style{Flex: 1}},
		E("Text", P{"style": Style{FontSize: 28}}, T("Hello")),
		E("Text", P{"style": Style{FontSize: 16}}, T("World")),
	)

	frames := ComputeLayout(tree, testScreen)

	if len(frames) < 3 {
		t.Fatalf("expected at least 3 frames, got %d", len(frames))
	}

	// Root should fill screen
	root := frames[0]
	if !approxF(root.Width, 390) {
		t.Errorf("root width: expected 390, got %f", root.Width)
	}

	// Children should be stacked vertically
	child1 := frames[1]
	child2 := frames[2]
	if child1.Y >= child2.Y {
		t.Errorf("child1 should be above child2: y1=%f y2=%f", child1.Y, child2.Y)
	}
}

func TestLayoutRow(t *testing.T) {
	tree := E("View", P{"style": Style{
		FlexDirection: "row",
		Width:         200,
		Height:        50,
	}},
		E("View", P{"style": Style{Width: 80, Height: 50}}),
		E("View", P{"style": Style{Width: 60, Height: 50}}),
	)

	frames := ComputeLayout(tree, testScreen)

	if len(frames) < 3 {
		t.Fatalf("expected 3 frames, got %d", len(frames))
	}

	child1 := frames[1]
	child2 := frames[2]

	// In a row, child2.X should be to the right of child1
	if child2.X <= child1.X {
		t.Errorf("row: child2 should be right of child1: x1=%f x2=%f", child1.X, child2.X)
	}
	if !approxF(child2.X-child1.X, 80) {
		t.Errorf("row: child2.X expected 80 from child1, got %f", child2.X-child1.X)
	}
}

func TestLayoutPadding(t *testing.T) {
	tree := E("View", P{"style": Style{
		Width:   200,
		Height:  200,
		Padding: 20,
	}},
		E("View", P{"style": Style{Height: 30}}),
	)

	frames := ComputeLayout(tree, testScreen)

	child := frames[1]
	// Child should be offset by padding (20) from parent origin
	if !approxF(child.X, 20) {
		t.Errorf("padding: child X expected 20, got %f", child.X)
	}
	if !approxF(child.Y, 20) {
		t.Errorf("padding: child Y expected 20, got %f", child.Y)
	}
	// Child width should be 200 - 20*2 = 160
	if !approxF(child.Width, 160) {
		t.Errorf("padding: child width expected 160, got %f", child.Width)
	}
}

func TestLayoutGap(t *testing.T) {
	tree := E("View", P{"style": Style{
		Width:  100,
		Height: 200,
		Gap:    10,
	}},
		E("View", P{"style": Style{Height: 30}}),
		E("View", P{"style": Style{Height: 30}}),
	)

	frames := ComputeLayout(tree, testScreen)

	child1 := frames[1]
	child2 := frames[2]
	// child2 should be at 30 (child1 height) + 10 (gap) = 40
	gap := child2.Y - (child1.Y + child1.Height)
	if !approxF(gap, 10) {
		t.Errorf("gap: expected 10px between children, got %f", gap)
	}
}

func TestLayoutFlexGrow(t *testing.T) {
	tree := E("View", P{"style": Style{
		Width:         300,
		Height:        100,
		FlexDirection: "row",
	}},
		E("View", P{"style": Style{Flex: 1}}),
		E("View", P{"style": Style{Flex: 2}}),
	)

	frames := ComputeLayout(tree, testScreen)

	child1 := frames[1]
	child2 := frames[2]
	// child1 gets 1/3, child2 gets 2/3
	if !approxF(child1.Width, 100) {
		t.Errorf("flex grow: child1 width expected 100, got %f", child1.Width)
	}
	if !approxF(child2.Width, 200) {
		t.Errorf("flex grow: child2 width expected 200, got %f", child2.Width)
	}
}

func TestLayoutAbsolutePosition(t *testing.T) {
	tree := E("View", P{"style": Style{Width: 300, Height: 300}},
		E("View", P{"style": Style{
			Position: "absolute",
			Top:      50,
			Left:     30,
			Width:    80,
			Height:   40,
		}}),
	)

	frames := ComputeLayout(tree, testScreen)

	child := frames[1]
	if !approxF(child.X, 30) {
		t.Errorf("absolute: X expected 30, got %f", child.X)
	}
	if !approxF(child.Y, 50) {
		t.Errorf("absolute: Y expected 50, got %f", child.Y)
	}
}

func TestLayoutSafeArea(t *testing.T) {
	screen := ScreenInfo{
		Width:      390,
		Height:     844,
		SafeTop:    47,
		SafeBottom: 34,
	}

	tree := E("SafeArea", nil,
		E("View", P{"style": Style{Height: 50}}),
	)

	frames := ComputeLayout(tree, screen)

	child := frames[1]
	// Child should be offset by safe area top (47)
	if !approxF(child.Y, 47) {
		t.Errorf("safe area: child Y expected 47, got %f", child.Y)
	}
}

func TestLayoutJustifyCenter(t *testing.T) {
	tree := E("View", P{"style": Style{
		Width:          200,
		Height:         200,
		JustifyContent: "center",
	}},
		E("View", P{"style": Style{Height: 40}}),
	)

	frames := ComputeLayout(tree, testScreen)

	child := frames[1]
	// Centered: (200 - 40) / 2 = 80
	if !approxF(child.Y, 80) {
		t.Errorf("justify center: child Y expected 80, got %f", child.Y)
	}
}

func TestLayoutAlignCenter(t *testing.T) {
	tree := E("View", P{"style": Style{
		Width:      200,
		Height:     200,
		AlignItems: "center",
	}},
		E("View", P{"style": Style{Width: 80, Height: 40}}),
	)

	frames := ComputeLayout(tree, testScreen)

	child := frames[1]
	// Centered on cross axis: (200 - 80) / 2 = 60
	if !approxF(child.X, 60) {
		t.Errorf("align center: child X expected 60, got %f", child.X)
	}
}

func TestLayoutNil(t *testing.T) {
	frames := ComputeLayout(nil, testScreen)
	if frames != nil {
		t.Error("nil root should return nil frames")
	}
}

func TestLayoutParentIDs(t *testing.T) {
	tree := E("View", nil,
		E("View", nil,
			E("Text", nil, T("Nested")),
		),
	)

	frames := ComputeLayout(tree, testScreen)

	if len(frames) < 3 {
		t.Fatalf("expected at least 3 frames, got %d", len(frames))
	}

	// Root has parentID -1
	if frames[0].ParentID != -1 {
		t.Errorf("root parentID expected -1, got %d", frames[0].ParentID)
	}
	// Inner view's parent should be root (0)
	if frames[1].ParentID != 0 {
		t.Errorf("inner view parentID expected 0, got %d", frames[1].ParentID)
	}
}
