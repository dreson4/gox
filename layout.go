package gox

import (
	"gox/internal/yoga"
)

// LayoutFrame represents a positioned native view with computed coordinates.
// This is the output of ComputeLayout — a flat list that the native bridge
// uses to create and position views without any layout logic.
type LayoutFrame struct {
	ID       int     `json:"id"`
	Tag      string  `json:"tag"`
	Text     string  `json:"text,omitempty"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"w"`
	Height   float64 `json:"h"`
	ParentID int     `json:"pid"`
	Props    P       `json:"props,omitempty"`
}

// ScreenInfo provides screen dimensions and safe area insets.
type ScreenInfo struct {
	Width      float64
	Height     float64
	SafeTop    float64
	SafeRight  float64
	SafeBottom float64
	SafeLeft   float64
}

// ComputeLayout takes a render tree and screen info, runs Yoga layout,
// and returns a flat list of positioned frames for the native bridge.
func ComputeLayout(root *Node, screen ScreenInfo) []LayoutFrame {
	if root == nil {
		return nil
	}

	// Clear previous event callbacks before re-render
	ClearEvents()

	lc := &layoutComputer{
		screen: screen,
		nextID: 0,
	}

	// Build Yoga tree from Node tree
	yogaRoot := lc.buildYogaTree(root, -1)
	if yogaRoot == nil {
		return nil
	}
	defer yogaRoot.FreeRecursive()

	// Set root to screen size only if the root doesn't have explicit dimensions
	rootHasWidth := false
	rootHasHeight := false
	if root.Type == NodeElement {
		if s, ok := lc.getStyle(root); ok {
			rootHasWidth = s.Width > 0
			rootHasHeight = s.Height > 0
		}
	}
	if !rootHasWidth {
		yogaRoot.SetWidth(float32(screen.Width))
	}
	if !rootHasHeight {
		yogaRoot.SetHeight(float32(screen.Height))
	}

	// Compute layout
	yogaRoot.CalculateLayout(float32(screen.Width), float32(screen.Height), yoga.DirectionLTR)

	// Extract computed frames
	lc.extractFrames(yogaRoot, 0, 0)

	return lc.frames
}

// layoutComputer holds state during layout computation.
type layoutComputer struct {
	screen   ScreenInfo
	nextID   int
	frames   []LayoutFrame
	nodeMap  []nodeInfo // parallel to frames, maps ID → original Node
}

type nodeInfo struct {
	node     *Node
	yogaNode *yoga.Node
	parentID int
}

func (lc *layoutComputer) allocID() int {
	id := lc.nextID
	lc.nextID++
	return id
}

// buildYogaTree recursively creates Yoga nodes from the GOX node tree.
func (lc *layoutComputer) buildYogaTree(node *Node, parentID int) *yoga.Node {
	if node == nil {
		return nil
	}

	switch node.Type {
	case NodeElement:
		return lc.buildElement(node, parentID)
	case NodeText:
		return lc.buildTextNode(node, parentID)
	case NodeFragment:
		return lc.buildFragment(node, parentID)
	default:
		return nil
	}
}

func (lc *layoutComputer) buildElement(node *Node, parentID int) *yoga.Node {
	yn := yoga.NewNode()
	id := lc.allocID()

	// Register this node
	lc.nodeMap = append(lc.nodeMap, nodeInfo{
		node:     node,
		yogaNode: yn,
		parentID: parentID,
	})

	// Apply style
	if style, ok := lc.getStyle(node); ok {
		lc.applyStyleToYoga(yn, style)
	}

	// SafeArea: apply screen safe insets as padding
	if node.Tag == "SafeArea" {
		yn.SetPadding(yoga.EdgeTop, float32(lc.screen.SafeTop))
		yn.SetPadding(yoga.EdgeRight, float32(lc.screen.SafeRight))
		yn.SetPadding(yoga.EdgeBottom, float32(lc.screen.SafeBottom))
		yn.SetPadding(yoga.EdgeLeft, float32(lc.screen.SafeLeft))
	}

	// Default intrinsic sizes for native elements
	switch node.Tag {
	case "Switch":
		yn.SetWidth(51)  // iOS UISwitch default size
		yn.SetHeight(31)
	case "TextInput":
		if _, ok := lc.getStyle(node); !ok {
			yn.SetHeight(44) // default iOS text field height
		}
	}

	// For Text elements, estimate size from children's text content
	if node.Tag == "Text" {
		text := collectTextContent(node)
		fontSize := lc.getTextFontSize(node)
		estimateTextSize(yn, text, fontSize)
	}

	// Recursively add children (skip for Text — text content handled above)
	if node.Tag != "Text" {
		childIdx := 0
		for _, child := range node.Children {
			childYoga := lc.buildYogaTree(child, id)
			if childYoga != nil {
				yn.InsertChild(childYoga, childIdx)
				childIdx++
			}
		}
	}

	return yn
}

func (lc *layoutComputer) buildTextNode(node *Node, parentID int) *yoga.Node {
	yn := yoga.NewNode()
	_ = lc.allocID()

	lc.nodeMap = append(lc.nodeMap, nodeInfo{
		node:     node,
		yogaNode: yn,
		parentID: parentID,
	})

	// Estimate text size
	estimateTextSize(yn, node.Text, 17) // default font size

	return yn
}

func (lc *layoutComputer) buildFragment(node *Node, parentID int) *yoga.Node {
	// Fragment with single child: skip the wrapper
	if len(node.Children) == 1 {
		return lc.buildYogaTree(node.Children[0], parentID)
	}

	// Multi-child fragment: create a container
	yn := yoga.NewNode()
	id := lc.allocID()

	lc.nodeMap = append(lc.nodeMap, nodeInfo{
		node:     node,
		yogaNode: yn,
		parentID: parentID,
	})

	childIdx := 0
	for _, child := range node.Children {
		childYoga := lc.buildYogaTree(child, id)
		if childYoga != nil {
			yn.InsertChild(childYoga, childIdx)
			childIdx++
		}
	}

	return yn
}

// extractFrames walks the Yoga tree after layout and collects positioned frames.
func (lc *layoutComputer) extractFrames(yn *yoga.Node, offsetX, offsetY float64) {
	// Find the nodeInfo for this Yoga node
	for i, info := range lc.nodeMap {
		if info.yogaNode != yn {
			continue
		}

		x := float64(yn.LayoutGetLeft()) + offsetX
		y := float64(yn.LayoutGetTop()) + offsetY
		w := float64(yn.LayoutGetWidth())
		h := float64(yn.LayoutGetHeight())

		frame := LayoutFrame{
			ID:       i,
			ParentID: info.parentID,
			X:        x,
			Y:        y,
			Width:    w,
			Height:   h,
		}

		node := info.node
		if node != nil {
			frame.Tag = node.Tag
			if node.Type == NodeText {
				frame.Text = node.Text
				frame.Tag = "_text"
			} else if node.Type == NodeElement && node.Tag == "Text" {
				// Collect text content from children for Text elements
				frame.Text = collectTextContent(node)
			} else if node.Type == NodeElement && node.Tag == "Button" {
				// Collect button label and text style from child Text elements
				for _, child := range node.Children {
					if child.Type == NodeElement && child.Tag == "Text" {
						frame.Text = collectTextContent(child)
					}
				}
			}
			if node.Type == NodeFragment {
				frame.Tag = "_fragment"
			}
			frame.Props = lc.collectVisualProps(node)

			// For buttons, copy child Text styling so bridge can apply to UIButton title
			if node.Type == NodeElement && node.Tag == "Button" {
				for _, child := range node.Children {
					if child.Type == NodeElement && child.Tag == "Text" {
						if childStyle, ok := lc.getStyle(child); ok {
							if frame.Props == nil {
								frame.Props = P{}
							}
							frame.Props["_btnTextColor"] = childStyle.Color
							frame.Props["_btnTextSize"] = childStyle.FontSize
							frame.Props["_btnTextWeight"] = childStyle.FontWeight
						}
						break
					}
				}
			}

			// Register event callbacks (onPress, onChange, etc.)
			if node.Props != nil {
				if onPress, ok := node.Props["onPress"].(func()); ok {
					RegisterEvent(i, onPress)
					// Mark this frame as having an event so bridge can wire it
					if frame.Props == nil {
						frame.Props = P{}
					}
					frame.Props["_hasOnPress"] = true
				}
			}
		}

		lc.frames = append(lc.frames, frame)

		// Recurse into children with accumulated offset
		for ci := 0; ci < yn.ChildCount(); ci++ {
			// Get child yoga node by walking our nodeMap
			childYN := lc.findChildYogaNode(yn, ci)
			if childYN != nil {
				lc.extractFrames(childYN, x, y)
			}
		}

		return
	}
}

// findChildYogaNode finds the yoga node that is the nth child of parent.
func (lc *layoutComputer) findChildYogaNode(parent *yoga.Node, index int) *yoga.Node {
	count := 0
	for _, info := range lc.nodeMap {
		// Find nodes whose parent's yoga node matches
		if info.parentID >= 0 && info.parentID < len(lc.nodeMap) {
			parentInfo := lc.nodeMap[info.parentID]
			if parentInfo.yogaNode == parent {
				if count == index {
					return info.yogaNode
				}
				count++
			}
		}
	}
	return nil
}

// collectVisualProps extracts non-layout props for the native bridge.
func (lc *layoutComputer) collectVisualProps(node *Node) P {
	if node == nil || node.Props == nil {
		return nil
	}

	// Filter out non-serializable values (functions) —
	// json.Marshal fails on func() types.
	filtered := P{}
	for k, v := range node.Props {
		switch v.(type) {
		case func(), func(string), func(bool), func(float64):
			// Skip function callbacks — they're stored in eventCallbacks
			continue
		default:
			filtered[k] = v
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

// --- Style → Yoga mapping ---

func (lc *layoutComputer) applyStyleToYoga(yn *yoga.Node, s Style) {
	// Dimensions
	if s.Width > 0 {
		yn.SetWidth(float32(s.Width))
	}
	if s.Height > 0 {
		yn.SetHeight(float32(s.Height))
	}
	if s.MinWidth > 0 {
		yn.SetMinWidth(float32(s.MinWidth))
	}
	if s.MinHeight > 0 {
		yn.SetMinHeight(float32(s.MinHeight))
	}
	if s.MaxWidth > 0 {
		yn.SetMaxWidth(float32(s.MaxWidth))
	}
	if s.MaxHeight > 0 {
		yn.SetMaxHeight(float32(s.MaxHeight))
	}

	// Flex
	if s.Flex > 0 {
		yn.SetFlex(float32(s.Flex))
	}
	if s.FlexDirection != "" {
		yn.SetFlexDirection(mapFlexDirection(s.FlexDirection))
	}
	if s.JustifyContent != "" {
		yn.SetJustifyContent(mapJustify(s.JustifyContent))
	}
	if s.AlignItems != "" {
		yn.SetAlignItems(mapAlign(s.AlignItems))
	}
	if s.AlignSelf != "" {
		yn.SetAlignSelf(mapAlign(s.AlignSelf))
	}
	if s.FlexWrap != "" {
		yn.SetFlexWrap(mapWrap(s.FlexWrap))
	}

	// Gap
	if s.Gap > 0 {
		yn.SetGap(yoga.GutterAll, float32(s.Gap))
	}
	if s.RowGap > 0 {
		yn.SetGap(yoga.GutterRow, float32(s.RowGap))
	}
	if s.ColumnGap > 0 {
		yn.SetGap(yoga.GutterColumn, float32(s.ColumnGap))
	}

	// Padding
	if s.Padding > 0 {
		yn.SetPadding(yoga.EdgeAll, float32(s.Padding))
	}
	if s.PaddingTop > 0 {
		yn.SetPadding(yoga.EdgeTop, float32(s.PaddingTop))
	}
	if s.PaddingRight > 0 {
		yn.SetPadding(yoga.EdgeRight, float32(s.PaddingRight))
	}
	if s.PaddingBottom > 0 {
		yn.SetPadding(yoga.EdgeBottom, float32(s.PaddingBottom))
	}
	if s.PaddingLeft > 0 {
		yn.SetPadding(yoga.EdgeLeft, float32(s.PaddingLeft))
	}

	// Margin
	if s.Margin > 0 {
		yn.SetMargin(yoga.EdgeAll, float32(s.Margin))
	}
	if s.MarginTop > 0 {
		yn.SetMargin(yoga.EdgeTop, float32(s.MarginTop))
	}
	if s.MarginRight > 0 {
		yn.SetMargin(yoga.EdgeRight, float32(s.MarginRight))
	}
	if s.MarginBottom > 0 {
		yn.SetMargin(yoga.EdgeBottom, float32(s.MarginBottom))
	}
	if s.MarginLeft > 0 {
		yn.SetMargin(yoga.EdgeLeft, float32(s.MarginLeft))
	}

	// Position
	if s.Position == "absolute" {
		yn.SetPositionType(yoga.PositionTypeAbsolute)
	}
	if s.Top != 0 {
		yn.SetPosition(yoga.EdgeTop, float32(s.Top))
	}
	if s.Right != 0 {
		yn.SetPosition(yoga.EdgeRight, float32(s.Right))
	}
	if s.Bottom != 0 {
		yn.SetPosition(yoga.EdgeBottom, float32(s.Bottom))
	}
	if s.Left != 0 {
		yn.SetPosition(yoga.EdgeLeft, float32(s.Left))
	}

	// Overflow
	if s.Overflow == "hidden" {
		yn.SetOverflow(yoga.OverflowHidden)
	} else if s.Overflow == "scroll" {
		yn.SetOverflow(yoga.OverflowScroll)
	}
}

// --- Text size estimation ---

func estimateTextSize(yn *yoga.Node, text string, fontSize float64) {
	if text == "" {
		return
	}
	// Rough estimation: each character is ~0.5x fontSize wide
	// Height is ~1.4x fontSize per line
	charWidth := fontSize * 0.55
	lineHeight := fontSize * 1.4
	width := float64(len(text)) * charWidth
	yn.SetHeight(float32(lineHeight))
	// Don't set width — let it be determined by parent (stretch)
	_ = width
}

func collectTextContent(node *Node) string {
	var text string
	for _, child := range node.Children {
		if child.Type == NodeText {
			text += child.Text
		}
	}
	return text
}

func (lc *layoutComputer) getTextFontSize(node *Node) float64 {
	if s, ok := lc.getStyle(node); ok && s.FontSize > 0 {
		return s.FontSize
	}
	return 17 // iOS default
}

func (lc *layoutComputer) getStyle(node *Node) (Style, bool) {
	if node.Props == nil {
		return Style{}, false
	}
	s, ok := node.Props["style"].(Style)
	return s, ok
}

// --- Enum mappers ---

func mapFlexDirection(s string) yoga.FlexDirection {
	switch s {
	case "row":
		return yoga.FlexDirectionRow
	case "row-reverse":
		return yoga.FlexDirectionRowReverse
	case "column-reverse":
		return yoga.FlexDirectionColumnReverse
	default:
		return yoga.FlexDirectionColumn
	}
}

func mapJustify(s string) yoga.Justify {
	switch s {
	case "center":
		return yoga.JustifyCenter
	case "end":
		return yoga.JustifyFlexEnd
	case "between":
		return yoga.JustifySpaceBetween
	case "around":
		return yoga.JustifySpaceAround
	case "evenly":
		return yoga.JustifySpaceEvenly
	default:
		return yoga.JustifyFlexStart
	}
}

func mapAlign(s string) yoga.Align {
	switch s {
	case "center":
		return yoga.AlignCenter
	case "end":
		return yoga.AlignFlexEnd
	case "stretch":
		return yoga.AlignStretch
	case "baseline":
		return yoga.AlignBaseline
	default:
		return yoga.AlignFlexStart
	}
}

func mapWrap(s string) yoga.Wrap {
	switch s {
	case "wrap":
		return yoga.WrapWrap
	case "wrap-reverse":
		return yoga.WrapReverse
	default:
		return yoga.WrapNoWrap
	}
}
