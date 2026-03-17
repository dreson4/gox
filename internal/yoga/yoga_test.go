package yoga

import (
	"math"
	"testing"
)

func approx(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.5
}

func TestColumnLayout(t *testing.T) {
	root := NewNode()
	defer root.FreeRecursive()

	root.SetWidth(100)
	root.SetHeight(100)
	root.SetFlexDirection(FlexDirectionColumn)

	child1 := NewNode()
	child1.SetHeight(30)
	root.InsertChild(child1, 0)

	child2 := NewNode()
	child2.SetHeight(40)
	root.InsertChild(child2, 1)

	root.CalculateLayout(100, 100, DirectionLTR)

	// Child1 at top
	if !approx(child1.LayoutGetTop(), 0) {
		t.Errorf("child1 top: expected 0, got %f", child1.LayoutGetTop())
	}
	if !approx(child1.LayoutGetHeight(), 30) {
		t.Errorf("child1 height: expected 30, got %f", child1.LayoutGetHeight())
	}

	// Child2 below child1
	if !approx(child2.LayoutGetTop(), 30) {
		t.Errorf("child2 top: expected 30, got %f", child2.LayoutGetTop())
	}
	if !approx(child2.LayoutGetHeight(), 40) {
		t.Errorf("child2 height: expected 40, got %f", child2.LayoutGetHeight())
	}
}

func TestRowLayout(t *testing.T) {
	root := NewNode()
	defer root.FreeRecursive()

	root.SetWidth(200)
	root.SetHeight(100)
	root.SetFlexDirection(FlexDirectionRow)

	child1 := NewNode()
	child1.SetWidth(80)
	root.InsertChild(child1, 0)

	child2 := NewNode()
	child2.SetWidth(60)
	root.InsertChild(child2, 1)

	root.CalculateLayout(200, 100, DirectionLTR)

	if !approx(child1.LayoutGetLeft(), 0) {
		t.Errorf("child1 left: expected 0, got %f", child1.LayoutGetLeft())
	}
	if !approx(child2.LayoutGetLeft(), 80) {
		t.Errorf("child2 left: expected 80, got %f", child2.LayoutGetLeft())
	}
}

func TestPadding(t *testing.T) {
	root := NewNode()
	defer root.FreeRecursive()

	root.SetWidth(100)
	root.SetHeight(100)
	root.SetPadding(EdgeAll, 10)

	child := NewNode()
	child.SetHeight(20)
	root.InsertChild(child, 0)

	root.CalculateLayout(100, 100, DirectionLTR)

	// Child should be offset by padding
	if !approx(child.LayoutGetTop(), 10) {
		t.Errorf("child top: expected 10, got %f", child.LayoutGetTop())
	}
	if !approx(child.LayoutGetLeft(), 10) {
		t.Errorf("child left: expected 10, got %f", child.LayoutGetLeft())
	}
	// Child width should be container - padding*2
	if !approx(child.LayoutGetWidth(), 80) {
		t.Errorf("child width: expected 80, got %f", child.LayoutGetWidth())
	}
}

func TestFlexGrow(t *testing.T) {
	root := NewNode()
	defer root.FreeRecursive()

	root.SetWidth(300)
	root.SetHeight(100)
	root.SetFlexDirection(FlexDirectionRow)

	child1 := NewNode()
	child1.SetFlexGrow(1)
	root.InsertChild(child1, 0)

	child2 := NewNode()
	child2.SetFlexGrow(2)
	root.InsertChild(child2, 1)

	root.CalculateLayout(300, 100, DirectionLTR)

	// child1 gets 1/3, child2 gets 2/3
	if !approx(child1.LayoutGetWidth(), 100) {
		t.Errorf("child1 width: expected 100, got %f", child1.LayoutGetWidth())
	}
	if !approx(child2.LayoutGetWidth(), 200) {
		t.Errorf("child2 width: expected 200, got %f", child2.LayoutGetWidth())
	}
}

func TestGap(t *testing.T) {
	root := NewNode()
	defer root.FreeRecursive()

	root.SetWidth(100)
	root.SetHeight(200)
	root.SetFlexDirection(FlexDirectionColumn)
	root.SetGap(GutterAll, 10)

	child1 := NewNode()
	child1.SetHeight(30)
	root.InsertChild(child1, 0)

	child2 := NewNode()
	child2.SetHeight(30)
	root.InsertChild(child2, 1)

	root.CalculateLayout(100, 200, DirectionLTR)

	// child2 should be at 30 + 10 gap = 40
	if !approx(child2.LayoutGetTop(), 40) {
		t.Errorf("child2 top: expected 40, got %f", child2.LayoutGetTop())
	}
}

func TestJustifyCenter(t *testing.T) {
	root := NewNode()
	defer root.FreeRecursive()

	root.SetWidth(100)
	root.SetHeight(200)
	root.SetFlexDirection(FlexDirectionColumn)
	root.SetJustifyContent(JustifyCenter)

	child := NewNode()
	child.SetHeight(40)
	root.InsertChild(child, 0)

	root.CalculateLayout(100, 200, DirectionLTR)

	// child should be centered: (200-40)/2 = 80
	if !approx(child.LayoutGetTop(), 80) {
		t.Errorf("child top: expected 80, got %f", child.LayoutGetTop())
	}
}

func TestAbsolutePosition(t *testing.T) {
	root := NewNode()
	defer root.FreeRecursive()

	root.SetWidth(200)
	root.SetHeight(200)

	child := NewNode()
	child.SetPositionType(PositionTypeAbsolute)
	child.SetPosition(EdgeTop, 50)
	child.SetPosition(EdgeLeft, 30)
	child.SetWidth(60)
	child.SetHeight(40)
	root.InsertChild(child, 0)

	root.CalculateLayout(200, 200, DirectionLTR)

	if !approx(child.LayoutGetTop(), 50) {
		t.Errorf("child top: expected 50, got %f", child.LayoutGetTop())
	}
	if !approx(child.LayoutGetLeft(), 30) {
		t.Errorf("child left: expected 30, got %f", child.LayoutGetLeft())
	}
}
