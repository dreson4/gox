// Package platform defines the interface between the GOX render tree
// and native platform backends (iOS, Android).
//
// The renderer walks a *gox.Node tree and calls Backend methods to
// create and configure native views. Each platform provides its own
// Backend implementation (e.g. iOS uses UIKit via cgo).
package platform

import "gox"

// ViewHandle is an opaque reference to a native view.
// Each platform maps these to concrete types (UIView, android.View, etc.).
type ViewHandle int64

// Backend is the interface that platform-specific code implements.
// It provides primitive operations for creating and configuring native views.
type Backend interface {
	// View creation
	CreateView() ViewHandle
	CreateText(content string) ViewHandle

	// View hierarchy
	AddChild(parent, child ViewHandle)
	SetRootView(handle ViewHandle)

	// Styling
	SetBackgroundColor(handle ViewHandle, color string)
	SetFrame(handle ViewHandle, x, y, width, height float64)
	SetPadding(handle ViewHandle, top, right, bottom, left float64)
	SetFontSize(handle ViewHandle, size float64)
	SetFontWeight(handle ViewHandle, weight string)
	SetTextColor(handle ViewHandle, color string)
	SetTextAlign(handle ViewHandle, align string)
	SetBorderRadius(handle ViewHandle, radius float64)
	SetOpacity(handle ViewHandle, opacity float64)
	SetFlexLayout(handle ViewHandle, direction string, justify string, align string)

	// App lifecycle
	RunApp()
}

// Renderer walks a gox.Node tree and creates native views via a Backend.
type Renderer struct {
	backend Backend
}

// NewRenderer creates a renderer targeting the given backend.
func NewRenderer(b Backend) *Renderer {
	return &Renderer{backend: b}
}

// Render processes a node tree and returns the root native view handle.
func (r *Renderer) Render(node *gox.Node) ViewHandle {
	if node == nil {
		return 0
	}
	return r.renderNode(node)
}

// RenderToScreen renders a node tree and sets it as the root view.
func (r *Renderer) RenderToScreen(node *gox.Node) {
	handle := r.Render(node)
	if handle != 0 {
		r.backend.SetRootView(handle)
	}
}

func (r *Renderer) renderNode(node *gox.Node) ViewHandle {
	switch node.Type {
	case gox.NodeElement:
		return r.renderElement(node)
	case gox.NodeText:
		return r.renderText(node)
	case gox.NodeFragment:
		return r.renderFragment(node)
	default:
		return 0
	}
}

func (r *Renderer) renderElement(node *gox.Node) ViewHandle {
	var handle ViewHandle

	switch node.Tag {
	case "Text":
		// Text elements: create a label, collect text from children
		text := r.collectText(node)
		handle = r.backend.CreateText(text)
	default:
		// All other elements: create a generic view container
		handle = r.backend.CreateView()
	}

	if handle == 0 {
		return 0
	}

	// Apply style from props
	if style, ok := r.getStyleProp(node); ok {
		r.applyStyle(handle, style, node.Tag)
	}

	// For non-Text elements, render and attach children
	if node.Tag != "Text" {
		for _, child := range node.Children {
			childHandle := r.renderNode(child)
			if childHandle != 0 {
				r.backend.AddChild(handle, childHandle)
			}
		}
	}

	return handle
}

func (r *Renderer) renderText(node *gox.Node) ViewHandle {
	return r.backend.CreateText(node.Text)
}

func (r *Renderer) renderFragment(node *gox.Node) ViewHandle {
	// Fragments need a wrapper view to hold children
	if len(node.Children) == 1 {
		return r.renderNode(node.Children[0])
	}

	wrapper := r.backend.CreateView()
	for _, child := range node.Children {
		childHandle := r.renderNode(child)
		if childHandle != 0 {
			r.backend.AddChild(wrapper, childHandle)
		}
	}
	return wrapper
}

// collectText recursively extracts text content from a Text element's children.
func (r *Renderer) collectText(node *gox.Node) string {
	var text string
	for _, child := range node.Children {
		switch child.Type {
		case gox.NodeText:
			text += child.Text
		}
	}
	return text
}

// getStyleProp extracts the Style from a node's props, if present.
func (r *Renderer) getStyleProp(node *gox.Node) (gox.Style, bool) {
	if node.Props == nil {
		return gox.Style{}, false
	}
	s, ok := node.Props["style"].(gox.Style)
	return s, ok
}

// applyStyle sets native view properties from a Style struct.
func (r *Renderer) applyStyle(handle ViewHandle, s gox.Style, tag string) {
	if s.BackgroundColor != "" {
		r.backend.SetBackgroundColor(handle, s.BackgroundColor)
	}
	if s.Opacity != 0 {
		r.backend.SetOpacity(handle, s.Opacity)
	}
	if s.BorderRadius != 0 {
		r.backend.SetBorderRadius(handle, s.BorderRadius)
	}

	// Padding
	top, right, bottom, left := s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft
	if s.Padding != 0 {
		if top == 0 {
			top = s.Padding
		}
		if right == 0 {
			right = s.Padding
		}
		if bottom == 0 {
			bottom = s.Padding
		}
		if left == 0 {
			left = s.Padding
		}
	}
	if top != 0 || right != 0 || bottom != 0 || left != 0 {
		r.backend.SetPadding(handle, top, right, bottom, left)
	}

	// Flex layout
	if s.FlexDirection != "" || s.JustifyContent != "" || s.AlignItems != "" {
		r.backend.SetFlexLayout(handle, s.FlexDirection, s.JustifyContent, s.AlignItems)
	}

	// Text-specific styles (only apply to Text elements)
	if tag == "Text" {
		if s.FontSize != 0 {
			r.backend.SetFontSize(handle, s.FontSize)
		}
		if s.FontWeight != "" {
			r.backend.SetFontWeight(handle, s.FontWeight)
		}
		if s.Color != "" {
			r.backend.SetTextColor(handle, s.Color)
		}
		if s.TextAlign != "" {
			r.backend.SetTextAlign(handle, s.TextAlign)
		}
	}
}
