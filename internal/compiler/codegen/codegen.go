// Package codegen transforms a GOX AST into valid Go source code.
//
// The generated code calls into the gox runtime to build a render tree.
// For now, it targets the minimal Hello World case: View + Text elements.
package codegen

import (
	"fmt"
	"gox/internal/compiler/ast"
	"strings"
)

// Generator emits Go source from a GOX AST.
type Generator struct {
	buf    strings.Builder
	indent int
}

// New creates a code generator.
func New() *Generator {
	return &Generator{}
}

// Generate produces Go source code from the AST.
func (g *Generator) Generate(file *ast.File) string {
	for _, section := range file.GoSections {
		g.raw(section.Code)
	}

	if file.View != nil {
		g.emitViewFunc(file.View)
	}

	return g.buf.String()
}

func (g *Generator) emitViewFunc(view *ast.ViewBlock) {
	g.line("")
	g.line("func render() *gox.Node {")
	g.indent++

	if len(view.Children) == 1 {
		g.iwrite("return ")
		g.emitNode(view.Children[0])
		g.nl()
	} else if len(view.Children) > 1 {
		g.line("return gox.Fragment(")
		g.indent++
		for _, child := range view.Children {
			g.iwrite("")
			g.emitNode(child)
			g.raw(",\n")
		}
		g.indent--
		g.line(")")
	} else {
		g.line("return nil")
	}

	g.indent--
	g.line("}")
}

func (g *Generator) emitNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.Element:
		g.emitElement(n)
	case *ast.TextNode:
		g.emitText(n)
	case *ast.ExprNode:
		g.emitExpr(n)
	case *ast.IfNode:
		g.emitIf(n)
	case *ast.ForNode:
		g.emitFor(n)
	}
}

func (g *Generator) emitElement(elem *ast.Element) {
	viewName := elemName(elem.Tag)
	hasProps := len(elem.Props) > 0
	hasChildren := len(elem.Children) > 0

	if !hasProps && !hasChildren {
		g.rawf("gox.E(\"%s\", nil)", viewName)
		return
	}

	if !hasChildren {
		g.rawf("gox.E(\"%s\", ", viewName)
		g.emitPropsMap(elem.Props)
		g.raw(")")
		return
	}

	g.rawf("gox.E(\"%s\", ", viewName)
	if hasProps {
		g.emitPropsMap(elem.Props)
	} else {
		g.raw("nil")
	}
	g.raw(",\n")
	g.indent++
	for _, child := range elem.Children {
		g.iwrite("")
		g.emitNode(child)
		g.raw(",\n")
	}
	g.indent--
	g.iwrite(")")
}

func (g *Generator) emitPropsMap(props []ast.Prop) {
	g.raw("gox.P{")
	for i, prop := range props {
		if i > 0 {
			g.raw(", ")
		}
		g.rawf("%q: ", prop.Name)
		if prop.Value.IsString() {
			g.rawf("%q", *prop.Value.StringValue)
		} else if prop.Value.ExprValue != nil {
			g.raw(*prop.Value.ExprValue)
		} else {
			g.raw("true")
		}
	}
	g.raw("}")
}

func (g *Generator) emitText(text *ast.TextNode) {
	g.rawf("gox.T(%q)", text.Content)
}

func (g *Generator) emitExpr(expr *ast.ExprNode) {
	g.rawf("gox.T(fmt.Sprint(%s))", expr.Expr)
}

func (g *Generator) emitIf(node *ast.IfNode) {
	g.rawf("gox.If(%s, func() *gox.Node {\n", node.Cond)
	g.indent++
	if len(node.Body) == 1 {
		g.iwrite("return ")
		g.emitNode(node.Body[0])
		g.nl()
	} else {
		g.line("return gox.Fragment(")
		g.indent++
		for _, child := range node.Body {
			g.iwrite("")
			g.emitNode(child)
			g.raw(",\n")
		}
		g.indent--
		g.line(")")
	}
	g.indent--
	g.iwrite("})")
}

func (g *Generator) emitFor(node *ast.ForNode) {
	g.raw("gox.ForEach(func() []*gox.Node {\n")
	g.indent++
	g.line("var nodes []*gox.Node")
	g.linef("for %s {", node.Clause)
	g.indent++
	for _, child := range node.Body {
		g.iwrite("nodes = append(nodes, ")
		g.emitNode(child)
		g.raw(")\n")
	}
	g.indent--
	g.line("}")
	g.line("return nodes")
	g.indent--
	g.iwrite("})")
}

// elemName extracts the element name from a tag.
// "gox.View" → "View", "posts.PostCard" → "posts.PostCard"
func elemName(tag string) string {
	if strings.HasPrefix(tag, "gox.") {
		return tag[4:]
	}
	return tag
}

// --- Writing primitives ---

// raw writes content inline without indentation.
func (g *Generator) raw(s string) {
	g.buf.WriteString(s)
}

// rawf writes formatted content inline.
func (g *Generator) rawf(format string, args ...any) {
	fmt.Fprintf(&g.buf, format, args...)
}

// nl writes a newline.
func (g *Generator) nl() {
	g.buf.WriteByte('\n')
}

// iwrite writes indentation followed by content.
func (g *Generator) iwrite(s string) {
	for range g.indent {
		g.buf.WriteByte('\t')
	}
	g.buf.WriteString(s)
}

// line writes an indented line with newline.
func (g *Generator) line(s string) {
	g.iwrite(s)
	g.nl()
}

// linef writes a formatted indented line with newline.
func (g *Generator) linef(format string, args ...any) {
	for range g.indent {
		g.buf.WriteByte('\t')
	}
	fmt.Fprintf(&g.buf, format, args...)
	g.nl()
}
