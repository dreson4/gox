// Package gox provides the runtime for GOX applications.
//
// Generated .go code from the GOX compiler calls into this package
// to build a render tree of nodes, which the platform bridge then
// maps to native views.
package gox

// NodeType identifies what kind of render tree node this is.
type NodeType int

const (
	NodeElement  NodeType = iota // native view element (View, Text, etc.)
	NodeText                    // raw text content
	NodeFragment                // group of children without a wrapper
)

// Node is a single node in the render tree.
// JSON tags enable serialization for the iOS bridge.
type Node struct {
	Type     NodeType `json:"type"`
	Tag      string   `json:"tag,omitempty"`
	Props    P        `json:"props,omitempty"`
	Text     string   `json:"text,omitempty"`
	Children []*Node  `json:"children,omitempty"`
}

// P is the props map type used by generated code.
type P map[string]any

// E creates an element node with the given tag, props, and children.
// This is the primary function called by generated code.
//
//	gox.E("View", gox.P{"style": s}, child1, child2)
func E(tag string, props P, children ...*Node) *Node {
	return &Node{
		Type:     NodeElement,
		Tag:      tag,
		Props:    props,
		Children: filterNil(children),
	}
}

// T creates a text node.
//
//	gox.T("Hello World")
func T(text string) *Node {
	return &Node{
		Type: NodeText,
		Text: text,
	}
}

// Fragment groups children without adding a wrapper element.
//
//	gox.Fragment(child1, child2, child3)
func Fragment(children ...*Node) *Node {
	return &Node{
		Type:     NodeFragment,
		Children: filterNil(children),
	}
}

// If conditionally renders content. Returns nil if condition is false.
//
//	gox.If(loading, func() *Node { return gox.T("Loading...") })
func If(cond bool, fn func() *Node) *Node {
	if cond {
		return fn()
	}
	return nil
}

// ForEach renders a dynamic list of nodes.
//
//	gox.ForEach(func() []*Node { ... })
func ForEach(fn func() []*Node) *Node {
	nodes := fn()
	return &Node{
		Type:     NodeFragment,
		Children: filterNil(nodes),
	}
}

// filterNil removes nil entries from a node slice.
func filterNil(nodes []*Node) []*Node {
	var out []*Node
	for _, n := range nodes {
		if n != nil {
			out = append(out, n)
		}
	}
	return out
}
