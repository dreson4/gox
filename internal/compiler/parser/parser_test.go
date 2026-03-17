package parser

import (
	"github.com/dreson4/gox/internal/compiler/ast"
	"github.com/dreson4/gox/internal/compiler/lexer"
	"testing"
)

func parse(t *testing.T, src string) *ast.File {
	t.Helper()
	l := lexer.New([]byte(src), "test.gox")
	tokens := l.Tokenize()
	p := New(tokens)
	file, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	return file
}

func TestParseHelloWorld(t *testing.T) {
	file := parse(t, `package app

import "gox"

view {
    <gox.View>
        <gox.Text>Hello World</gox.Text>
    </gox.View>
}
`)

	if file.View == nil {
		t.Fatal("expected view block")
	}
	if len(file.View.Children) != 1 {
		t.Fatalf("expected 1 child in view, got %d", len(file.View.Children))
	}

	// Root element: gox.View
	view, ok := file.View.Children[0].(*ast.Element)
	if !ok {
		t.Fatal("expected Element node")
	}
	if view.Tag != "gox.View" {
		t.Errorf("expected tag gox.View, got %s", view.Tag)
	}
	if len(view.Children) != 1 {
		t.Fatalf("expected 1 child in View, got %d", len(view.Children))
	}

	// Child: gox.Text
	text, ok := view.Children[0].(*ast.Element)
	if !ok {
		t.Fatal("expected Element node for Text")
	}
	if text.Tag != "gox.Text" {
		t.Errorf("expected tag gox.Text, got %s", text.Tag)
	}
	if len(text.Children) != 1 {
		t.Fatalf("expected 1 child in Text, got %d", len(text.Children))
	}

	// Text content
	content, ok := text.Children[0].(*ast.TextNode)
	if !ok {
		t.Fatal("expected TextNode")
	}
	if content.Content != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", content.Content)
	}
}

func TestParseSelfClosing(t *testing.T) {
	file := parse(t, `package app

import "gox"

view {
    <gox.Image src="photo.jpg" />
}
`)

	if file.View == nil {
		t.Fatal("expected view block")
	}

	img, ok := file.View.Children[0].(*ast.Element)
	if !ok {
		t.Fatal("expected Element")
	}
	if img.Tag != "gox.Image" {
		t.Errorf("expected gox.Image, got %s", img.Tag)
	}
	if !img.SelfClosing {
		t.Error("expected self-closing")
	}
	if len(img.Props) != 1 {
		t.Fatalf("expected 1 prop, got %d", len(img.Props))
	}
	if img.Props[0].Name != "src" {
		t.Errorf("expected prop name 'src', got %q", img.Props[0].Name)
	}
	if !img.Props[0].Value.IsString() || *img.Props[0].Value.StringValue != "photo.jpg" {
		t.Errorf("expected string prop 'photo.jpg'")
	}
}

func TestParseExprProp(t *testing.T) {
	file := parse(t, `package app

import "gox"

view {
    <gox.Text style={styles["title"]}>Hello</gox.Text>
}
`)

	text := file.View.Children[0].(*ast.Element)
	if len(text.Props) != 1 {
		t.Fatalf("expected 1 prop, got %d", len(text.Props))
	}
	prop := text.Props[0]
	if prop.Name != "style" {
		t.Errorf("expected prop 'style', got %q", prop.Name)
	}
	if prop.Value.ExprValue == nil || *prop.Value.ExprValue != `styles["title"]` {
		t.Errorf("expected expression prop, got %+v", prop.Value)
	}
}

func TestParseTextExpression(t *testing.T) {
	file := parse(t, `package app

import "gox"

view {
    <gox.Text>{user.Name}</gox.Text>
}
`)

	text := file.View.Children[0].(*ast.Element)
	if len(text.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(text.Children))
	}
	expr, ok := text.Children[0].(*ast.ExprNode)
	if !ok {
		t.Fatalf("expected ExprNode, got %T", text.Children[0])
	}
	if expr.Expr != "user.Name" {
		t.Errorf("expected 'user.Name', got %q", expr.Expr)
	}
}

func TestParseGoSections(t *testing.T) {
	file := parse(t, `package posts

import (
    "gox"
    "fmt"
)

type Props struct {
    UserID string
}

var styles = gox.Styles{
    "title": gox.Style{FontSize: 28},
}

view {
    <gox.Text>Hello</gox.Text>
}
`)

	if len(file.GoSections) == 0 {
		t.Fatal("expected Go sections")
	}
	code := file.GoSections[0].Code
	if !containsStr(code, "package posts") {
		t.Error("Go section should contain package")
	}
	if !containsStr(code, "type Props struct") {
		t.Error("Go section should contain Props")
	}
}

func TestParseNestedElements(t *testing.T) {
	file := parse(t, `package app

import "gox"

view {
    <gox.View>
        <gox.View>
            <gox.Text>Deep</gox.Text>
        </gox.View>
    </gox.View>
}
`)

	outer := file.View.Children[0].(*ast.Element)
	inner := outer.Children[0].(*ast.Element)
	text := inner.Children[0].(*ast.Element)
	content := text.Children[0].(*ast.TextNode)
	if content.Content != "Deep" {
		t.Errorf("expected 'Deep', got %q", content.Content)
	}
}

func TestParseIfControlFlow(t *testing.T) {
	file := parse(t, `package app

import "gox"

view {
    {if loading {
        <gox.Text>Loading...</gox.Text>
    }}
}
`)

	if file.View == nil {
		t.Fatal("expected view block")
	}
	if len(file.View.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(file.View.Children))
	}

	ifNode, ok := file.View.Children[0].(*ast.IfNode)
	if !ok {
		t.Fatalf("expected IfNode, got %T", file.View.Children[0])
	}
	if ifNode.Cond != "loading" {
		t.Errorf("expected condition 'loading', got %q", ifNode.Cond)
	}
	if len(ifNode.Body) != 1 {
		t.Fatalf("expected 1 body node, got %d", len(ifNode.Body))
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && findStr(s, sub)
}

func findStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
