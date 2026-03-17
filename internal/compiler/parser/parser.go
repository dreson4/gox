// Package parser converts a token stream from the lexer into an AST.
package parser

import (
	"fmt"
	"gox/internal/compiler/ast"
	"gox/internal/compiler/token"
	"regexp"
	"strings"
)

// Parser builds an AST from a token stream.
type Parser struct {
	tokens []token.Token
	pos    int
	errors []Error
}

// Error represents a parse error with position.
type Error struct {
	Message string
	Pos     token.Position
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Pos, e.Message)
}

// New creates a parser for the given tokens.
func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse processes all tokens and returns the AST.
func (p *Parser) Parse() (*ast.File, []Error) {
	file := &ast.File{}

	for !p.atEnd() {
		tok := p.current()

		switch tok.Type {
		case token.GoCode:
			file.GoSections = append(file.GoSections, ast.GoSection{
				Code: tok.Value,
				Pos:  tok.Pos,
			})
			p.advance()

		case token.ViewKeyword:
			p.advance() // skip `view`
			p.expect(token.LBrace)
			view := p.parseViewBlock()
			file.View = view

		case token.EOF:
			p.advance()

		default:
			p.error("unexpected token: %s %q", tok.Type, tok.Value)
			p.advance()
		}
	}

	// Detect lifecycle functions in Go sections
	file.LifecycleFuncs = detectLifecycleFuncs(file.GoSections)

	return file, p.errors
}

// lifecycleFuncNames are the recognized lifecycle function names.
var lifecycleFuncNames = []string{"onMount", "onUnmount", "onAppear", "onDisappear"}

// lifecycleRe matches "func onMount(", "func onUnmount(", etc.
var lifecycleRe = regexp.MustCompile(`\bfunc\s+(onMount|onUnmount|onAppear|onDisappear)\s*\(`)

// detectLifecycleFuncs scans GoSections for lifecycle function declarations.
func detectLifecycleFuncs(sections []ast.GoSection) []string {
	var found []string
	for _, section := range sections {
		matches := lifecycleRe.FindAllStringSubmatch(section.Code, -1)
		for _, m := range matches {
			found = append(found, m[1])
		}
	}
	return found
}

// parseViewBlock parses the children inside view { ... }.
func (p *Parser) parseViewBlock() *ast.ViewBlock {
	block := &ast.ViewBlock{Pos: p.current().Pos}
	block.Children = p.parseChildren()
	p.expect(token.RBrace)
	return block
}

// parseChildren parses a sequence of nodes until a closing tag or block end.
func (p *Parser) parseChildren() []ast.Node {
	var nodes []ast.Node

	for !p.atEnd() {
		tok := p.current()

		switch tok.Type {
		case token.RBrace, token.ExprEnd, token.EOF:
			return nodes

		case token.LAngleSlash:
			// Closing tag — parent handles this
			return nodes

		case token.LAngle:
			elem := p.parseElement()
			if elem != nil {
				nodes = append(nodes, elem)
			}

		case token.Text:
			nodes = append(nodes, &ast.TextNode{
				Content: tok.Value,
				Pos:     tok.Pos,
			})
			p.advance()

		case token.ExprStart:
			node := p.parseExprOrControlFlow()
			if node != nil {
				nodes = append(nodes, node)
			}

		default:
			p.error("unexpected token in view body: %s %q", tok.Type, tok.Value)
			p.advance()
		}
	}
	return nodes
}

// parseElement parses <pkg.Name props>children</pkg.Name> or <pkg.Name props />.
func (p *Parser) parseElement() *ast.Element {
	p.expect(token.LAngle) // consume <
	pos := p.current().Pos

	tag := p.parseTagName()
	if tag == "" {
		p.error("expected tag name")
		return nil
	}

	elem := &ast.Element{Tag: tag, Pos: pos}
	elem.Props, elem.SpreadExpr = p.parseProps()

	// Self-closing?
	if p.check(token.SlashRAngle) {
		p.advance()
		elem.SelfClosing = true
		return elem
	}

	// Opening tag close
	p.expect(token.RAngle)

	// Children
	elem.Children = p.parseChildren()

	// Closing tag
	p.parseClosingTag(tag)

	return elem
}

// parseTagName reads "Ident" or "Ident.Ident".
func (p *Parser) parseTagName() string {
	if !p.check(token.Ident) {
		return ""
	}
	name := p.current().Value
	p.advance()

	if p.check(token.Dot) {
		p.advance()
		if p.check(token.Ident) {
			name += "." + p.current().Value
			p.advance()
		}
	}
	return name
}

// parseProps reads prop="val", prop={expr}, or {...spread} pairs until > or />.
// Returns props list and optional spread expression.
func (p *Parser) parseProps() ([]ast.Prop, string) {
	var props []ast.Prop
	var spread string

	for !p.atEnd() && !p.check(token.RAngle) && !p.check(token.SlashRAngle) {
		// Check for spread: {...expr}
		if p.check(token.ExprStart) {
			p.advance()
			expr := p.collectExprContent()
			p.expect(token.ExprEnd)
			if strings.HasPrefix(expr, "...") {
				spread = strings.TrimPrefix(expr, "...")
			}
			continue
		}

		if !p.check(token.Ident) {
			break
		}

		prop := ast.Prop{
			Name: p.current().Value,
			Pos:  p.current().Pos,
		}
		p.advance()

		if p.check(token.Equals) {
			p.advance()
			prop.Value = p.parsePropValue()
		}
		props = append(props, prop)
	}
	return props, spread
}

// parsePropValue reads either "string" or {expression}.
func (p *Parser) parsePropValue() ast.PropValue {
	if p.check(token.String) {
		s := p.current().Value
		p.advance()
		return ast.PropValue{StringValue: &s}
	}

	if p.check(token.ExprStart) {
		p.advance()
		expr := p.collectExprContent()
		p.expect(token.ExprEnd)
		return ast.PropValue{ExprValue: &expr}
	}

	p.error("expected string or expression for prop value")
	return ast.PropValue{}
}

// collectExprContent collects GoCode tokens as a single expression string.
func (p *Parser) collectExprContent() string {
	if p.check(token.GoCode) {
		val := p.current().Value
		p.advance()
		return val
	}
	return ""
}

// parseExprOrControlFlow handles {expr}, {if ...}, {for ...}, {switch ...}.
func (p *Parser) parseExprOrControlFlow() ast.Node {
	p.expect(token.ExprStart)

	tok := p.current()
	switch tok.Type {
	case token.If:
		return p.parseIf()
	case token.For:
		return p.parseFor()
	case token.Switch:
		return p.parseSwitch()
	case token.GoCode:
		node := &ast.ExprNode{Expr: tok.Value, Pos: tok.Pos}
		p.advance()
		p.expect(token.ExprEnd)
		return node
	default:
		p.error("expected expression or control flow keyword after {")
		p.expect(token.ExprEnd)
		return nil
	}
}

// parseIf parses {if cond { <children> }}.
func (p *Parser) parseIf() *ast.IfNode {
	pos := p.current().Pos
	p.advance() // skip `if`

	// The condition is in a GoCode token
	cond := ""
	if p.check(token.GoCode) {
		cond = p.current().Value
		p.advance()
	}

	// Body: parse children until ExprEnd
	body := p.parseChildren()

	// The closing }} — inner } ends the if body, outer } ends the expression
	p.expect(token.ExprEnd)

	return &ast.IfNode{
		Cond: cond,
		Body: body,
		Pos:  pos,
	}
}

// parseFor parses {for clause { <children> }}.
func (p *Parser) parseFor() *ast.ForNode {
	pos := p.current().Pos
	p.advance() // skip `for`

	clause := ""
	if p.check(token.GoCode) {
		clause = p.current().Value
		p.advance()
	}

	body := p.parseChildren()
	p.expect(token.ExprEnd)

	return &ast.ForNode{
		Clause: clause,
		Body:   body,
		Pos:    pos,
	}
}

// parseSwitch parses {switch expr { case: ... }}.
func (p *Parser) parseSwitch() *ast.SwitchNode {
	pos := p.current().Pos
	p.advance() // skip `switch`

	expr := ""
	if p.check(token.GoCode) {
		expr = p.current().Value
		p.advance()
	}

	// TODO: parse cases properly when needed
	_ = p.parseChildren()
	p.expect(token.ExprEnd)

	return &ast.SwitchNode{
		Expr: expr,
		Pos:  pos,
	}
}

// parseClosingTag expects </tag.Name>.
func (p *Parser) parseClosingTag(expected string) {
	if !p.check(token.LAngleSlash) {
		p.error("expected closing tag </%s>", expected)
		return
	}
	p.advance()

	got := p.parseTagName()
	if got != expected {
		p.error("mismatched closing tag: expected </%s>, got </%s>", expected, got)
	}

	p.expect(token.RAngle)
}

// --- Token navigation ---

func (p *Parser) current() token.Token {
	if p.pos >= len(p.tokens) {
		return token.Token{Type: token.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) check(typ token.Type) bool {
	return p.current().Type == typ
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) expect(typ token.Type) {
	if !p.check(typ) {
		p.error("expected %s, got %s %q", typ, p.current().Type, p.current().Value)
		return
	}
	p.advance()
}

func (p *Parser) atEnd() bool {
	return p.pos >= len(p.tokens) || p.current().Type == token.EOF
}

func (p *Parser) error(format string, args ...any) {
	p.errors = append(p.errors, Error{
		Message: fmt.Sprintf(format, args...),
		Pos:     p.current().Pos,
	})
}
