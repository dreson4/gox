package codegen

import (
	"gox/internal/compiler/lexer"
	"gox/internal/compiler/parser"
	"strings"
	"testing"
)

func generate(t *testing.T, src string) string {
	t.Helper()
	l := lexer.New([]byte(src), "test.gox")
	tokens := l.Tokenize()
	p := parser.New(tokens)
	file, errs := p.Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("parse error: %s", e)
		}
		t.FailNow()
	}
	g := New()
	return g.Generate(file)
}

func TestGenerateHelloWorld(t *testing.T) {
	out := generate(t, `package app

import "gox"

view {
    <gox.View>
        <gox.Text>Hello World</gox.Text>
    </gox.View>
}
`)
	t.Log("Generated:\n" + out)

	// Should contain package declaration
	if !strings.Contains(out, "package app") {
		t.Error("missing package declaration")
	}

	// Should contain render function
	if !strings.Contains(out, "func render()") {
		t.Error("missing render function")
	}

	// Should use gox.E for elements
	if !strings.Contains(out, `gox.E("View"`) {
		t.Error("missing gox.E(\"View\")")
	}
	if !strings.Contains(out, `gox.E("Text"`) {
		t.Error("missing gox.E(\"Text\")")
	}

	// Should contain text
	if !strings.Contains(out, `gox.T("Hello World")`) {
		t.Error("missing text node")
	}
}

func TestGenerateSelfClosing(t *testing.T) {
	out := generate(t, `package app

import "gox"

view {
    <gox.Image src="photo.jpg" />
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, `gox.E("Image"`) {
		t.Error("missing Image element")
	}
	if !strings.Contains(out, `"src"`) {
		t.Error("missing src prop")
	}
}

func TestGenerateExprProp(t *testing.T) {
	out := generate(t, `package app

import "gox"

var styles = gox.Styles{
    "title": gox.Style{FontSize: 28},
}

view {
    <gox.Text style={styles["title"]}>Hello</gox.Text>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, `styles["title"]`) {
		t.Error("missing expression prop")
	}
}

func TestGeneratePreservesGoCode(t *testing.T) {
	out := generate(t, `package posts

import (
    "gox"
    "fmt"
)

type Props struct {
    UserID string
}

type State struct {
    count int
}

view {
    <gox.Text>Hello</gox.Text>
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "type Props struct") {
		t.Error("Props struct should be preserved")
	}
	if !strings.Contains(out, "type State struct") {
		t.Error("State struct should be preserved")
	}
}

func TestGenerateIfControlFlow(t *testing.T) {
	out := generate(t, `package app

import "gox"

view {
    {if loading {
        <gox.Text>Loading...</gox.Text>
    }}
}
`)
	t.Log("Generated:\n" + out)

	if !strings.Contains(out, "gox.If(loading") {
		t.Error("missing gox.If call")
	}
	if !strings.Contains(out, `gox.T("Loading...")`) {
		t.Error("missing loading text")
	}
}
