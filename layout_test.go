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

// --- Deep nesting + row coordinate tests ---
// These verify that absolute coordinates are correct at every depth,
// which is critical for the bridge's absolute→relative conversion.

func TestLayoutNestedRowCoordinates(t *testing.T) {
	// SafeArea → View → Row(header) → children
	// This is the exact pattern that was broken: 3+ levels of nesting with row layout
	screen := ScreenInfo{
		Width:      390,
		Height:     844,
		SafeTop:    47,
		SafeBottom: 34,
	}

	tree := E("SafeArea", nil,
		E("View", P{"style": Style{Flex: 1}},
			E("View", P{"style": Style{
				FlexDirection: "row",
				Height:        60,
				Padding:       10,
				Gap:           8,
			}},
				E("View", P{"style": Style{Width: 40, Height: 40}}),  // "avatar"
				E("View", P{"style": Style{Flex: 1}}),                 // "info"
				E("View", P{"style": Style{Width: 60, Height: 30}}),   // "button"
			),
		),
	)

	frames := ComputeLayout(tree, screen)

	// Find the row container (3rd frame: SafeArea=0, View=1, Row=2)
	row := findFrameByID(frames, 2)
	if row == nil {
		t.Fatal("row frame not found")
	}

	// Row should be at top of View, which is at safe area top
	if !approxF(row.Y, 47) {
		t.Errorf("row Y expected 47 (safe top), got %f", row.Y)
	}

	// Avatar (id=3): should be inside row, offset by padding
	avatar := findFrameByID(frames, 3)
	if avatar == nil {
		t.Fatal("avatar frame not found")
	}
	if !approxF(avatar.X, 10) {
		t.Errorf("avatar X expected 10 (row padding), got %f", avatar.X)
	}
	if !approxF(avatar.Y, 47+10) {
		t.Errorf("avatar Y expected %f (row.Y + padding), got %f", 47.0+10, avatar.Y)
	}

	// Info (id=4): should be right of avatar + gap
	info := findFrameByID(frames, 4)
	if info == nil {
		t.Fatal("info frame not found")
	}
	expectedInfoX := 10.0 + 40 + 8 // padding + avatar width + gap
	if !approxF(info.X, expectedInfoX) {
		t.Errorf("info X expected %f, got %f", expectedInfoX, info.X)
	}

	// Button (id=5): should be rightmost
	button := findFrameByID(frames, 5)
	if button == nil {
		t.Fatal("button frame not found")
	}
	if button.X <= info.X {
		t.Errorf("button should be right of info: button.X=%f info.X=%f", button.X, info.X)
	}

	// Verify parent-child relative coordinates are correct
	// (this is what the bridge computes: child_abs - parent_abs)
	avatarRelX := avatar.X - row.X
	avatarRelY := avatar.Y - row.Y
	if !approxF(avatarRelX, 10) {
		t.Errorf("avatar relative X expected 10, got %f", avatarRelX)
	}
	if !approxF(avatarRelY, 10) {
		t.Errorf("avatar relative Y expected 10, got %f", avatarRelY)
	}
}

func TestLayoutDeeplyNestedCoordinates(t *testing.T) {
	// 5 levels deep: Root → A → B → C → Leaf
	// Verify absolute coordinates accumulate correctly
	tree := E("View", P{"style": Style{Width: 400, Height: 400, Padding: 20}},
		E("View", P{"style": Style{Flex: 1, Padding: 15}},
			E("View", P{"style": Style{Flex: 1, Padding: 10}},
				E("View", P{"style": Style{Width: 50, Height: 30}}),
			),
		),
	)

	frames := ComputeLayout(tree, testScreen)

	if len(frames) < 4 {
		t.Fatalf("expected at least 4 frames, got %d", len(frames))
	}

	leaf := findFrameByID(frames, 3)
	if leaf == nil {
		t.Fatal("leaf frame not found")
	}

	// Leaf X should be 20 + 15 + 10 = 45 (accumulated padding)
	expectedX := 20.0 + 15 + 10
	if !approxF(leaf.X, expectedX) {
		t.Errorf("deep nested: leaf X expected %f, got %f", expectedX, leaf.X)
	}
	if !approxF(leaf.Y, expectedX) { // same padding on all sides
		t.Errorf("deep nested: leaf Y expected %f, got %f", expectedX, leaf.Y)
	}

	// Verify relative-to-parent is correct at each level
	for i := 1; i < len(frames); i++ {
		f := frames[i]
		parent := findFrameByID(frames, f.ParentID)
		if parent == nil {
			continue
		}
		relX := f.X - parent.X
		relY := f.Y - parent.Y
		if relX < 0 || relY < 0 {
			t.Errorf("frame %d (%s): negative relative position (%.1f, %.1f) vs parent %d (%.1f, %.1f)",
				f.ID, f.Tag, f.X, f.Y, parent.ID, parent.X, parent.Y)
		}
	}
}

func TestLayoutRowWithGap(t *testing.T) {
	// Row with gap — verify horizontal spacing
	tree := E("View", P{"style": Style{
		FlexDirection: "row",
		Width:         300,
		Height:        50,
		Gap:           12,
	}},
		E("View", P{"style": Style{Width: 60, Height: 50}}),
		E("View", P{"style": Style{Width: 60, Height: 50}}),
		E("View", P{"style": Style{Width: 60, Height: 50}}),
	)

	frames := ComputeLayout(tree, testScreen)

	c1 := findFrameByID(frames, 1)
	c2 := findFrameByID(frames, 2)
	c3 := findFrameByID(frames, 3)

	if !approxF(c1.X, 0) {
		t.Errorf("row gap: c1.X expected 0, got %f", c1.X)
	}
	if !approxF(c2.X, 72) { // 60 + 12 gap
		t.Errorf("row gap: c2.X expected 72, got %f", c2.X)
	}
	if !approxF(c3.X, 144) { // 60 + 12 + 60 + 12
		t.Errorf("row gap: c3.X expected 144, got %f", c3.X)
	}
}

func TestLayoutRowAlignCenter(t *testing.T) {
	// Row with items of different heights, centered vertically
	tree := E("View", P{"style": Style{
		FlexDirection: "row",
		AlignItems:    "center",
		Width:         200,
		Height:        80,
	}},
		E("View", P{"style": Style{Width: 40, Height: 20}}),
		E("View", P{"style": Style{Width: 40, Height: 60}}),
	)

	frames := ComputeLayout(tree, testScreen)

	small := findFrameByID(frames, 1)
	tall := findFrameByID(frames, 2)

	// Small item centered: (80 - 20) / 2 = 30
	if !approxF(small.Y, 30) {
		t.Errorf("row align center: small Y expected 30, got %f", small.Y)
	}
	// Tall item centered: (80 - 60) / 2 = 10
	if !approxF(tall.Y, 10) {
		t.Errorf("row align center: tall Y expected 10, got %f", tall.Y)
	}
}

func TestLayoutRowJustifySpaceBetween(t *testing.T) {
	tree := E("View", P{"style": Style{
		FlexDirection:  "row",
		JustifyContent: "between",
		Width:          300,
		Height:         50,
	}},
		E("View", P{"style": Style{Width: 50, Height: 50}}),
		E("View", P{"style": Style{Width: 50, Height: 50}}),
		E("View", P{"style": Style{Width: 50, Height: 50}}),
	)

	frames := ComputeLayout(tree, testScreen)

	c1 := findFrameByID(frames, 1)
	c3 := findFrameByID(frames, 3)

	// First at 0, last at 300-50=250
	if !approxF(c1.X, 0) {
		t.Errorf("space-between: c1.X expected 0, got %f", c1.X)
	}
	if !approxF(c3.X, 250) {
		t.Errorf("space-between: c3.X expected 250, got %f", c3.X)
	}
}

func TestLayoutNestedRowInColumn(t *testing.T) {
	// Column with multiple rows — tests that each row's children
	// get correct absolute coordinates
	tree := E("View", P{"style": Style{Width: 300, Height: 300, Padding: 10}},
		E("View", P{"style": Style{FlexDirection: "row", Height: 40, Gap: 5}},
			E("View", P{"style": Style{Width: 50, Height: 40}}),
			E("View", P{"style": Style{Width: 50, Height: 40}}),
		),
		E("View", P{"style": Style{FlexDirection: "row", Height: 40, Gap: 5}},
			E("View", P{"style": Style{Width: 50, Height: 40}}),
			E("View", P{"style": Style{Width: 50, Height: 40}}),
		),
	)

	frames := ComputeLayout(tree, testScreen)

	// Row 1 children (ids 2, 3)
	r1c1 := findFrameByID(frames, 2)
	r1c2 := findFrameByID(frames, 3)
	// Row 2 children (ids 5, 6)
	r2c1 := findFrameByID(frames, 5)
	r2c2 := findFrameByID(frames, 6)

	// Row 1 starts at (10, 10) due to padding
	if !approxF(r1c1.X, 10) {
		t.Errorf("nested rows: r1c1.X expected 10, got %f", r1c1.X)
	}
	if !approxF(r1c1.Y, 10) {
		t.Errorf("nested rows: r1c1.Y expected 10, got %f", r1c1.Y)
	}
	if !approxF(r1c2.X, 65) { // 10 + 50 + 5
		t.Errorf("nested rows: r1c2.X expected 65, got %f", r1c2.X)
	}

	// Row 2 starts at Y=50 (10 padding + 40 row1 height)
	if !approxF(r2c1.X, 10) {
		t.Errorf("nested rows: r2c1.X expected 10, got %f", r2c1.X)
	}
	if !approxF(r2c1.Y, 50) {
		t.Errorf("nested rows: r2c1.Y expected 50, got %f", r2c1.Y)
	}
	if !approxF(r2c2.X, 65) {
		t.Errorf("nested rows: r2c2.X expected 65, got %f", r2c2.X)
	}
}

func TestLayoutTextInputEventProps(t *testing.T) {
	var textVal string
	submitCalled := false

	tree := E("TextInput", P{
		"value":       "hello",
		"placeholder": "Type here",
		"onChange":     func(text string) { textVal = text },
		"onSubmit":    func() { submitCalled = true },
		"onFocus":     func() {},
		"onBlur":      func() {},
		"style":       Style{Height: 44},
	})

	frames := ComputeLayout(tree, testScreen)

	if len(frames) < 1 {
		t.Fatalf("expected at least 1 frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame.Tag != "TextInput" {
		t.Errorf("expected tag TextInput, got %s", frame.Tag)
	}

	// Check event flags are set
	if frame.Props["_hasOnChange"] != true {
		t.Error("expected _hasOnChange=true")
	}
	if frame.Props["_hasOnSubmit"] != true {
		t.Error("expected _hasOnSubmit=true")
	}
	if frame.Props["_hasOnFocus"] != true {
		t.Error("expected _hasOnFocus=true")
	}
	if frame.Props["_hasOnBlur"] != true {
		t.Error("expected _hasOnBlur=true")
	}

	// Check that props from ExtractFrame are passed through
	if frame.Props["value"] != "hello" {
		t.Errorf("expected value=hello, got %v", frame.Props["value"])
	}
	if frame.Props["placeholder"] != "Type here" {
		t.Errorf("expected placeholder='Type here', got %v", frame.Props["placeholder"])
	}

	// Verify the event callbacks were registered
	HandleTextEvent(0, "world")
	if textVal != "world" {
		t.Errorf("expected textVal='world', got %q", textVal)
	}
	HandleSubmitEvent(0)
	if !submitCalled {
		t.Error("expected submit callback to fire")
	}
}

func TestLayoutMargin(t *testing.T) {
	tree := E("View", P{"style": Style{Width: 200, Height: 200}},
		E("View", P{"style": Style{
			Width:     50,
			Height:    50,
			MarginTop: 20,
			MarginLeft: 15,
		}}),
	)

	frames := ComputeLayout(tree, testScreen)

	child := findFrameByID(frames, 1)
	if !approxF(child.X, 15) {
		t.Errorf("margin: X expected 15, got %f", child.X)
	}
	if !approxF(child.Y, 20) {
		t.Errorf("margin: Y expected 20, got %f", child.Y)
	}
}

func TestLayoutFlexRowWithFixedAndFlex(t *testing.T) {
	// Common pattern: fixed-width items with flex spacer
	tree := E("View", P{"style": Style{
		FlexDirection: "row",
		Width:         300,
		Height:        50,
		AlignItems:    "center",
	}},
		E("View", P{"style": Style{Width: 40, Height: 40}}),   // avatar
		E("View", P{"style": Style{Flex: 1}}),                  // spacer
		E("View", P{"style": Style{Width: 80, Height: 30}}),    // button
	)

	frames := ComputeLayout(tree, testScreen)

	avatar := findFrameByID(frames, 1)
	spacer := findFrameByID(frames, 2)
	button := findFrameByID(frames, 3)

	if !approxF(avatar.X, 0) {
		t.Errorf("flex row: avatar.X expected 0, got %f", avatar.X)
	}
	if !approxF(spacer.X, 40) {
		t.Errorf("flex row: spacer.X expected 40, got %f", spacer.X)
	}
	if !approxF(spacer.Width, 180) { // 300 - 40 - 80
		t.Errorf("flex row: spacer.Width expected 180, got %f", spacer.Width)
	}
	if !approxF(button.X, 220) { // 40 + 180
		t.Errorf("flex row: button.X expected 220, got %f", button.X)
	}
	// Button centered vertically: (50 - 30) / 2 = 10
	if !approxF(button.Y, 10) {
		t.Errorf("flex row: button.Y expected 10, got %f", button.Y)
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

func TestLayoutFrameHashing(t *testing.T) {
	tree := E("View", P{"style": Style{Width: 200, Height: 200, BackgroundColor: "#FF0000"}},
		E("Text", P{"style": Style{FontSize: 16}}, T("Hello")),
	)

	// Compute layout twice with same tree — hashes should match
	frames1 := ComputeLayout(tree, testScreen)
	frames2 := ComputeLayout(tree, testScreen)

	if len(frames1) != len(frames2) {
		t.Fatalf("frame count mismatch: %d vs %d", len(frames1), len(frames2))
	}

	for i := range frames1 {
		if frames1[i].Hash == "" {
			t.Errorf("frame %d has empty hash", i)
		}
		if frames1[i].Hash != frames2[i].Hash {
			t.Errorf("frame %d hash mismatch across identical renders: %s vs %s",
				i, frames1[i].Hash, frames2[i].Hash)
		}
	}

	// Change the text — hash should differ for that frame
	tree2 := E("View", P{"style": Style{Width: 200, Height: 200, BackgroundColor: "#FF0000"}},
		E("Text", P{"style": Style{FontSize: 16}}, T("World")),
	)

	frames3 := ComputeLayout(tree2, testScreen)

	// Root frame unchanged (same style, same position)
	if frames1[0].Hash != frames3[0].Hash {
		t.Errorf("root hash should match: %s vs %s", frames1[0].Hash, frames3[0].Hash)
	}

	// Text frame should differ
	text1 := findFrame(frames1, "Text")
	text3 := findFrame(frames3, "Text")
	if text1.Hash == text3.Hash {
		t.Error("text frame hash should differ after content change")
	}
}
