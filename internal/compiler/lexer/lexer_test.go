package lexer

import (
	"gox/internal/compiler/token"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	src := `package app

import "gox"

view {
    <gox.View>
        <gox.Text>Hello World</gox.Text>
    </gox.View>
}
`
	l := New([]byte(src), "test.gox")
	tokens := l.Tokenize()

	// Debug: print all tokens
	for i, tok := range tokens {
		t.Logf("token[%d]: %s %q", i, tok.Type, tok.Value)
	}

	expected := []struct {
		typ token.Type
		val string
	}{
		{token.GoCode, "package app\n\nimport \"gox\"\n\n"},
		{token.ViewKeyword, "view"},
		{token.LBrace, "{"},
		{token.LAngle, "<"},
		{token.Ident, "gox"},
		{token.Dot, "."},
		{token.Ident, "View"},
		{token.RAngle, ">"},
		{token.LAngle, "<"},
		{token.Ident, "gox"},
		{token.Dot, "."},
		{token.Ident, "Text"},
		{token.RAngle, ">"},
		{token.Text, "Hello World"},
		{token.LAngleSlash, "</"},
		{token.Ident, "gox"},
		{token.Dot, "."},
		{token.Ident, "Text"},
		{token.RAngle, ">"},
		{token.LAngleSlash, "</"},
		{token.Ident, "gox"},
		{token.Dot, "."},
		{token.Ident, "View"},
		{token.RAngle, ">"},
		{token.RBrace, "}"},
		{token.EOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		got := tokens[i]
		if got.Type != exp.typ || got.Value != exp.val {
			t.Errorf("token[%d]: expected %s %q, got %s %q",
				i, exp.typ, exp.val, got.Type, got.Value)
		}
	}
}

func TestSelfClosingTag(t *testing.T) {
	src := `package app

import "gox"

view {
    <gox.Image src="photo.jpg" />
}
`
	l := New([]byte(src), "test.gox")
	tokens := l.Tokenize()

	for i, tok := range tokens {
		t.Logf("token[%d]: %s %q", i, tok.Type, tok.Value)
	}

	// Find the self-closing token
	found := false
	for _, tok := range tokens {
		if tok.Type == token.SlashRAngle {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected /> token for self-closing tag")
	}
}

func TestPropExpression(t *testing.T) {
	src := `package app

import "gox"

view {
    <gox.Text style={styles["title"]}>Hello</gox.Text>
}
`
	l := New([]byte(src), "test.gox")
	tokens := l.Tokenize()

	for i, tok := range tokens {
		t.Logf("token[%d]: %s %q", i, tok.Type, tok.Value)
	}

	// Verify we get ExprStart, GoCode (the expression), ExprEnd
	foundExpr := false
	for i, tok := range tokens {
		if tok.Type == token.ExprStart && i+2 < len(tokens) {
			if tokens[i+1].Type == token.GoCode && tokens[i+1].Value == `styles["title"]` {
				if tokens[i+2].Type == token.ExprEnd {
					foundExpr = true
				}
			}
		}
	}
	if !foundExpr {
		t.Error("expected expression {styles[\"title\"]} to be tokenized as ExprStart GoCode ExprEnd")
	}
}

func TestTextExpression(t *testing.T) {
	src := `package app

import "gox"

view {
    <gox.Text>{user.Name}</gox.Text>
}
`
	l := New([]byte(src), "test.gox")
	tokens := l.Tokenize()

	for i, tok := range tokens {
		t.Logf("token[%d]: %s %q", i, tok.Type, tok.Value)
	}

	// Should have ExprStart, GoCode("user.Name"), ExprEnd between > and </
	foundExpr := false
	for i, tok := range tokens {
		if tok.Type == token.GoCode && tok.Value == "user.Name" {
			if i > 0 && tokens[i-1].Type == token.ExprStart {
				foundExpr = true
			}
		}
	}
	if !foundExpr {
		t.Error("expected expression {user.Name} in text content")
	}
}

func TestGoCodeBeforeView(t *testing.T) {
	src := `package posts

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

var styles = gox.Styles{
    "title": gox.Style{FontSize: 28},
}

view {
    <gox.Text>Hello</gox.Text>
}
`
	l := New([]byte(src), "test.gox")
	tokens := l.Tokenize()

	// First token should be GoCode containing everything before view
	if tokens[0].Type != token.GoCode {
		t.Fatalf("expected first token to be GoCode, got %s", tokens[0].Type)
	}

	goCode := tokens[0].Value
	if !contains(goCode, "package posts") {
		t.Error("GoCode should contain package declaration")
	}
	if !contains(goCode, "type Props struct") {
		t.Error("GoCode should contain Props struct")
	}
	if !contains(goCode, "type State struct") {
		t.Error("GoCode should contain State struct")
	}
	if !contains(goCode, "var styles") {
		t.Error("GoCode should contain styles var")
	}
}

func TestIfControlFlow(t *testing.T) {
	src := `package app

import "gox"

view {
    {if loading {
        <gox.Text>Loading...</gox.Text>
    }}
}
`
	l := New([]byte(src), "test.gox")
	tokens := l.Tokenize()
	for i, tok := range tokens {
		t.Logf("token[%d]: %s %q", i, tok.Type, tok.Value)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
