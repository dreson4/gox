// Package yoga provides Go bindings for the Yoga layout engine (facebook/yoga).
//
// Yoga is a C++ library with a C API that implements CSS Flexbox layout.
// This package wraps the C API via cgo, providing a Go-idiomatic interface.
//
// The Yoga C++ sources are vendored in c/ and pre-compiled to libyoga.a.
package yoga

/*
#cgo CFLAGS: -I${SRCDIR}/lib/include
#cgo LDFLAGS: -L${SRCDIR}/lib -lyoga -lc++
#include <yoga/Yoga.h>
*/
import "C"

// Direction
type Direction int

const (
	DirectionInherit Direction = C.YGDirectionInherit
	DirectionLTR     Direction = C.YGDirectionLTR
	DirectionRTL     Direction = C.YGDirectionRTL
)

// FlexDirection
type FlexDirection int

const (
	FlexDirectionColumn        FlexDirection = C.YGFlexDirectionColumn
	FlexDirectionColumnReverse FlexDirection = C.YGFlexDirectionColumnReverse
	FlexDirectionRow           FlexDirection = C.YGFlexDirectionRow
	FlexDirectionRowReverse    FlexDirection = C.YGFlexDirectionRowReverse
)

// Justify
type Justify int

const (
	JustifyFlexStart    Justify = C.YGJustifyFlexStart
	JustifyCenter       Justify = C.YGJustifyCenter
	JustifyFlexEnd      Justify = C.YGJustifyFlexEnd
	JustifySpaceBetween Justify = C.YGJustifySpaceBetween
	JustifySpaceAround  Justify = C.YGJustifySpaceAround
	JustifySpaceEvenly  Justify = C.YGJustifySpaceEvenly
)

// Align
type Align int

const (
	AlignAuto      Align = C.YGAlignAuto
	AlignFlexStart Align = C.YGAlignFlexStart
	AlignCenter    Align = C.YGAlignCenter
	AlignFlexEnd   Align = C.YGAlignFlexEnd
	AlignStretch   Align = C.YGAlignStretch
	AlignBaseline  Align = C.YGAlignBaseline
)

// Wrap
type Wrap int

const (
	WrapNoWrap  Wrap = C.YGWrapNoWrap
	WrapWrap    Wrap = C.YGWrapWrap
	WrapReverse Wrap = C.YGWrapWrapReverse
)

// Edge
type Edge int

const (
	EdgeLeft   Edge = C.YGEdgeLeft
	EdgeTop    Edge = C.YGEdgeTop
	EdgeRight  Edge = C.YGEdgeRight
	EdgeBottom Edge = C.YGEdgeBottom
	EdgeAll    Edge = C.YGEdgeAll
)

// Gutter
type Gutter int

const (
	GutterColumn Gutter = C.YGGutterColumn
	GutterRow    Gutter = C.YGGutterRow
	GutterAll    Gutter = C.YGGutterAll
)

// PositionType
type PositionType int

const (
	PositionTypeStatic   PositionType = C.YGPositionTypeStatic
	PositionTypeRelative PositionType = C.YGPositionTypeRelative
	PositionTypeAbsolute PositionType = C.YGPositionTypeAbsolute
)

// Overflow
type Overflow int

const (
	OverflowVisible Overflow = C.YGOverflowVisible
	OverflowHidden  Overflow = C.YGOverflowHidden
	OverflowScroll  Overflow = C.YGOverflowScroll
)

// Node wraps a YGNodeRef.
type Node struct {
	ref C.YGNodeRef
}

// NewNode creates a new Yoga node.
func NewNode() *Node {
	return &Node{ref: C.YGNodeNew()}
}

// Free releases the Yoga node.
func (n *Node) Free() {
	if n.ref != nil {
		C.YGNodeFree(n.ref)
		n.ref = nil
	}
}

// FreeRecursive frees this node and all its children.
func (n *Node) FreeRecursive() {
	if n.ref != nil {
		C.YGNodeFreeRecursive(n.ref)
		n.ref = nil
	}
}

// --- Children ---

// InsertChild adds a child at the given index.
func (n *Node) InsertChild(child *Node, index int) {
	C.YGNodeInsertChild(n.ref, child.ref, C.size_t(index))
}

// ChildCount returns the number of children.
func (n *Node) ChildCount() int {
	return int(C.YGNodeGetChildCount(n.ref))
}

// --- Layout computation ---

// CalculateLayout computes the layout for this node and its subtree.
func (n *Node) CalculateLayout(width, height float32, dir Direction) {
	C.YGNodeCalculateLayout(n.ref, C.float(width), C.float(height), C.YGDirection(dir))
}

// --- Layout results ---

func (n *Node) LayoutGetLeft() float32   { return float32(C.YGNodeLayoutGetLeft(n.ref)) }
func (n *Node) LayoutGetTop() float32    { return float32(C.YGNodeLayoutGetTop(n.ref)) }
func (n *Node) LayoutGetWidth() float32  { return float32(C.YGNodeLayoutGetWidth(n.ref)) }
func (n *Node) LayoutGetHeight() float32 { return float32(C.YGNodeLayoutGetHeight(n.ref)) }

// --- Style setters ---

func (n *Node) SetFlexDirection(dir FlexDirection) {
	C.YGNodeStyleSetFlexDirection(n.ref, C.YGFlexDirection(dir))
}

func (n *Node) SetJustifyContent(j Justify) {
	C.YGNodeStyleSetJustifyContent(n.ref, C.YGJustify(j))
}

func (n *Node) SetAlignItems(a Align) {
	C.YGNodeStyleSetAlignItems(n.ref, C.YGAlign(a))
}

func (n *Node) SetAlignSelf(a Align) {
	C.YGNodeStyleSetAlignSelf(n.ref, C.YGAlign(a))
}

func (n *Node) SetFlexWrap(w Wrap) {
	C.YGNodeStyleSetFlexWrap(n.ref, C.YGWrap(w))
}

func (n *Node) SetFlex(flex float32) {
	C.YGNodeStyleSetFlex(n.ref, C.float(flex))
}

func (n *Node) SetFlexGrow(grow float32) {
	C.YGNodeStyleSetFlexGrow(n.ref, C.float(grow))
}

func (n *Node) SetFlexShrink(shrink float32) {
	C.YGNodeStyleSetFlexShrink(n.ref, C.float(shrink))
}

func (n *Node) SetWidth(w float32)  { C.YGNodeStyleSetWidth(n.ref, C.float(w)) }
func (n *Node) SetHeight(h float32) { C.YGNodeStyleSetHeight(n.ref, C.float(h)) }

func (n *Node) SetWidthPercent(w float32)  { C.YGNodeStyleSetWidthPercent(n.ref, C.float(w)) }
func (n *Node) SetHeightPercent(h float32) { C.YGNodeStyleSetHeightPercent(n.ref, C.float(h)) }

func (n *Node) SetMinWidth(w float32)  { C.YGNodeStyleSetMinWidth(n.ref, C.float(w)) }
func (n *Node) SetMinHeight(h float32) { C.YGNodeStyleSetMinHeight(n.ref, C.float(h)) }
func (n *Node) SetMaxWidth(w float32)  { C.YGNodeStyleSetMaxWidth(n.ref, C.float(w)) }
func (n *Node) SetMaxHeight(h float32) { C.YGNodeStyleSetMaxHeight(n.ref, C.float(h)) }

func (n *Node) SetPadding(edge Edge, value float32) {
	C.YGNodeStyleSetPadding(n.ref, C.YGEdge(edge), C.float(value))
}

func (n *Node) SetMargin(edge Edge, value float32) {
	C.YGNodeStyleSetMargin(n.ref, C.YGEdge(edge), C.float(value))
}

func (n *Node) SetGap(gutter Gutter, value float32) {
	C.YGNodeStyleSetGap(n.ref, C.YGGutter(gutter), C.float(value))
}

func (n *Node) SetPositionType(pt PositionType) {
	C.YGNodeStyleSetPositionType(n.ref, C.YGPositionType(pt))
}

func (n *Node) SetPosition(edge Edge, value float32) {
	C.YGNodeStyleSetPosition(n.ref, C.YGEdge(edge), C.float(value))
}

func (n *Node) SetOverflow(o Overflow) {
	C.YGNodeStyleSetOverflow(n.ref, C.YGOverflow(o))
}
