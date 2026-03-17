// Package ast defines the abstract syntax tree for .gox files.
//
// A GoxFile consists of Go code sections (passed through) and a single
// view block containing a tree of UI nodes.
package ast

import "gox/internal/compiler/token"

// File represents a parsed .gox file.
type File struct {
	Package     string      // package name
	GoSections  []GoSection // raw Go code blocks (imports, types, funcs, vars)
	View        *ViewBlock  // the view { } block, nil if none
	IsComponent   bool   // true if file defines a reusable component (has type Props struct)
	ComponentName string // exported function name, derived from filename (e.g. "Comment" from comment.gox)
}

// GoSection is a chunk of raw Go source code that passes through unchanged.
type GoSection struct {
	Code string
	Pos  token.Position
}

// ViewBlock is the `view { ... }` block containing UI nodes.
type ViewBlock struct {
	Children []Node
	Pos      token.Position
}

// Node is the interface for all AST nodes inside a view block.
type Node interface {
	nodeType() string
	Position() token.Position
}

// Element represents a JSX-like element: <pkg.Name prop={val}>children</pkg.Name>
type Element struct {
	Tag         string          // full tag name, e.g. "gox.View", "gox.Text"
	Props       []Prop          // attributes on the element
	Children    []Node          // child nodes
	SelfClosing bool            // true for <Foo />
	SpreadExpr  string          // spread props expression: {...expr} → "expr"
	Pos         token.Position
}

func (e *Element) nodeType() string       { return "Element" }
func (e *Element) Position() token.Position { return e.Pos }

// Prop is a single attribute on an element.
type Prop struct {
	Name  string         // prop name, e.g. "style", "onPress"
	Value PropValue      // the value
	Pos   token.Position
}

// PropValue is either a string literal or a Go expression.
type PropValue struct {
	StringValue *string // non-nil for quoted strings: "hello"
	ExprValue   *string // non-nil for expressions: {someVar}
}

// IsString returns true if this prop value is a string literal.
func (pv PropValue) IsString() bool { return pv.StringValue != nil }

// TextNode represents raw text content between tags.
type TextNode struct {
	Content string
	Pos     token.Position
}

func (t *TextNode) nodeType() string       { return "Text" }
func (t *TextNode) Position() token.Position { return t.Pos }

// ExprNode represents a Go expression inside {}: {fmt.Sprintf(...)}
type ExprNode struct {
	Expr string
	Pos  token.Position
}

func (e *ExprNode) nodeType() string       { return "Expr" }
func (e *ExprNode) Position() token.Position { return e.Pos }

// IfNode represents {if cond { ... }} inside a view.
type IfNode struct {
	Cond    string   // Go condition expression
	Body    []Node   // nodes when true
	Else    []Node   // nodes when false (optional)
	Pos     token.Position
}

func (i *IfNode) nodeType() string       { return "If" }
func (i *IfNode) Position() token.Position { return i.Pos }

// ForNode represents {for _, item := range items { ... }} inside a view.
type ForNode struct {
	Clause  string   // full for clause: "_, item := range items"
	Body    []Node
	Pos     token.Position
}

func (f *ForNode) nodeType() string       { return "For" }
func (f *ForNode) Position() token.Position { return f.Pos }

// SwitchNode represents {switch expr { case: ... }} inside a view.
type SwitchNode struct {
	Expr  string       // switch expression
	Cases []SwitchCase
	Pos   token.Position
}

func (s *SwitchNode) nodeType() string       { return "Switch" }
func (s *SwitchNode) Position() token.Position { return s.Pos }

// SwitchCase is a single case within a switch.
type SwitchCase struct {
	Expr    string   // case expression, empty for default
	Body    []Node
	Default bool
}
