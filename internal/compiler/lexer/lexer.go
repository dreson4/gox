// Package lexer tokenizes .gox source files.
//
// The lexer operates in multiple modes:
//   - Go mode: collects raw Go code, looking for `view {` to switch
//   - View mode: tokenizes JSX-like syntax inside the view block
//   - Tag mode: scanning props/attributes inside an opening tag
//   - Expr mode: collecting a Go expression inside { }
//
// Expressions can contain JSX (e.g. {if cond { <Foo/> }}), so the lexer
// maintains a stack of contexts to return to after nested mode switches.
package lexer

import (
	"gox/internal/compiler/token"
	"strings"
	"unicode"
)

// mode tracks what part of a .gox file we're scanning.
type mode int

const (
	modeGo   mode = iota // outside view block
	modeView             // inside view block, between tags
	modeTag              // inside an opening tag (scanning props)
	modeExpr             // inside {expression}
)

// exprContext saves expression state when we switch to view mode mid-expression.
type exprContext struct {
	braceDepth int
	returnMode mode
}

// Lexer tokenizes .gox source code.
type Lexer struct {
	src        []byte
	file       string
	pos        int
	line       int
	col        int
	mode       mode
	returnMode mode // where to go after expression ends
	braceDepth int  // tracks nested {} in expressions
	exprStack  []exprContext // saved expr contexts when switching to view mid-expr
	tokens     []token.Token
}

// New creates a lexer for the given source.
func New(src []byte, file string) *Lexer {
	return &Lexer{
		src:  src,
		file: file,
		line: 1,
		col:  1,
	}
}

// Tokenize processes the entire source and returns all tokens.
func (l *Lexer) Tokenize() []token.Token {
	for !l.atEnd() {
		switch l.mode {
		case modeGo:
			l.scanGo()
		case modeView:
			l.scanView()
		case modeTag:
			l.scanTagProps()
		case modeExpr:
			l.scanExpr()
		}
	}
	l.emit(token.EOF, "")
	return l.tokens
}

// --- Go mode ---

func (l *Lexer) scanGo() {
	start := l.pos
	startPos := l.position()

	for !l.atEnd() {
		if l.matchViewKeyword() {
			// Emit Go code collected before `view`
			if l.pos-len("view") > start {
				l.tokens = append(l.tokens, token.Token{
					Type:  token.GoCode,
					Value: string(l.src[start : l.pos-len("view")]),
					Pos:   startPos,
				})
			}
			l.emitAt(token.ViewKeyword, "view", l.positionAt(l.pos-len("view")))
			l.skipWhitespace()
			if l.peek() == '{' {
				l.advance()
				l.emit(token.LBrace, "{")
				l.mode = modeView
				return
			}
			continue
		}
		l.advance()
	}

	// Remaining Go code (skip if only whitespace)
	if l.pos > start {
		code := string(l.src[start:l.pos])
		if strings.TrimSpace(code) != "" {
			l.tokens = append(l.tokens, token.Token{
				Type:  token.GoCode,
				Value: code,
				Pos:   startPos,
			})
		}
	}
}

// matchViewKeyword checks if current position starts a `view` keyword
// that begins a view block. Must be at start of a line (only whitespace before it)
// and followed by whitespace or `{`.
func (l *Lexer) matchViewKeyword() bool {
	remaining := string(l.src[l.pos:])
	if !strings.HasPrefix(remaining, "view") {
		return false
	}

	// Must be at line start (only whitespace since last newline)
	if l.pos > 0 {
		i := l.pos - 1
		for i > 0 && (l.src[i] == ' ' || l.src[i] == '\t') {
			i--
		}
		if i > 0 && l.src[i] != '\n' && l.src[i] != '\r' {
			return false
		}
	}

	// Character after "view" must be whitespace or '{'
	afterView := l.pos + len("view")
	if afterView >= len(l.src) {
		return false
	}
	ch := l.src[afterView]
	if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' && ch != '{' {
		return false
	}

	for range len("view") {
		l.advance()
	}
	return true
}

// --- View mode (between tags) ---

func (l *Lexer) scanView() {
	l.skipWhitespace()
	if l.atEnd() {
		return
	}

	switch ch := l.peek(); {
	case ch == '}':
		// Check if we're inside a nested expression that switched to view mode
		if len(l.exprStack) > 0 {
			// Return to expression mode — this } closes the control flow body
			ctx := l.exprStack[len(l.exprStack)-1]
			l.exprStack = l.exprStack[:len(l.exprStack)-1]
			l.mode = modeExpr
			l.braceDepth = ctx.braceDepth - 1 // the } decrements depth
			l.returnMode = ctx.returnMode
			l.advance() // consume the }
			return
		}
		// End of view block
		l.advance()
		l.emit(token.RBrace, "}")
		l.mode = modeGo

	case ch == '<':
		l.scanTagOpen()

	case ch == '{':
		l.advance()
		l.emit(token.ExprStart, "{")
		l.mode = modeExpr
		l.returnMode = modeView
		l.braceDepth = 1

	default:
		l.scanText()
	}
}

// scanTagOpen handles `<` or `</` and the tag name, then enters tag mode.
func (l *Lexer) scanTagOpen() {
	l.advance() // consume '<'

	if l.peek() == '/' {
		l.advance()
		l.emit(token.LAngleSlash, "</")
	} else {
		l.emit(token.LAngle, "<")
	}

	l.skipWhitespace()
	l.scanTagName()
	l.mode = modeTag
}

// --- Tag mode (scanning props inside an opening/closing tag) ---

func (l *Lexer) scanTagProps() {
	l.skipWhitespace()
	if l.atEnd() {
		return
	}

	switch ch := l.peek(); {
	case ch == '>':
		l.advance()
		l.emit(token.RAngle, ">")
		l.mode = modeView

	case ch == '/' && l.peekAt(1) == '>':
		l.advance()
		l.advance()
		l.emit(token.SlashRAngle, "/>")
		l.mode = modeView

	case ch == '=':
		l.advance()
		l.emit(token.Equals, "=")

	case ch == '{':
		l.advance()
		l.emit(token.ExprStart, "{")
		l.mode = modeExpr
		l.returnMode = modeTag
		l.braceDepth = 1

	case ch == '"':
		l.scanString()

	case isIdentStart(ch):
		l.scanIdent()

	default:
		l.advance()
		l.emit(token.Illegal, string(rune(ch)))
	}
}

func (l *Lexer) scanTagName() {
	if l.atEnd() || !isIdentStart(l.peek()) {
		return
	}
	l.scanIdent()

	// Dotted name: gox.View
	if !l.atEnd() && l.peek() == '.' {
		l.advance()
		l.emit(token.Dot, ".")
		if !l.atEnd() && isIdentStart(l.peek()) {
			l.scanIdent()
		}
	}
}

func (l *Lexer) scanIdent() {
	start := l.pos
	for !l.atEnd() && isIdentPart(l.peek()) {
		l.advance()
	}
	value := string(l.src[start:l.pos])

	switch value {
	case "if":
		l.emit(token.If, value)
	case "for":
		l.emit(token.For, value)
	case "switch":
		l.emit(token.Switch, value)
	default:
		l.emit(token.Ident, value)
	}
}

func (l *Lexer) scanString() {
	l.advance() // opening "
	start := l.pos
	for !l.atEnd() && l.peek() != '"' {
		if l.peek() == '\\' {
			l.advance()
		}
		l.advance()
	}
	value := string(l.src[start:l.pos])
	if !l.atEnd() {
		l.advance() // closing "
	}
	l.emit(token.String, value)
}

func (l *Lexer) scanText() {
	start := l.pos
	for !l.atEnd() {
		ch := l.peek()
		if ch == '<' || ch == '{' || ch == '}' {
			break
		}
		l.advance()
	}
	value := strings.TrimSpace(string(l.src[start:l.pos]))
	if value != "" {
		l.emit(token.Text, value)
	}
}

// --- Expression mode ---

// controlFlowKeywords that get their own token when they appear at the
// start of an expression (right after ExprStart).
var controlFlowKeywords = []string{"if", "for", "switch"}

func (l *Lexer) scanExpr() {
	// At expression start, check for control flow keywords
	l.skipWhitespace()
	if l.braceDepth == 1 { // top-level of expression
		for _, kw := range controlFlowKeywords {
			if l.matchKeyword(kw) {
				var typ token.Type
				switch kw {
				case "if":
					typ = token.If
				case "for":
					typ = token.For
				case "switch":
					typ = token.Switch
				}
				l.emit(typ, kw)
				// Collect the condition/clause up to the body-opening {
				l.scanControlFlowCondition()
				return
			}
		}
	}

	l.scanExprBody()
}

// matchKeyword checks if the current position starts with the given keyword
// followed by a non-identifier character.
func (l *Lexer) matchKeyword(kw string) bool {
	if l.pos+len(kw) > len(l.src) {
		return false
	}
	if string(l.src[l.pos:l.pos+len(kw)]) != kw {
		return false
	}
	// Must be followed by non-ident char
	after := l.pos + len(kw)
	if after < len(l.src) && isIdentPart(l.src[after]) {
		return false
	}
	// Advance past keyword
	for range len(kw) {
		l.advance()
	}
	return true
}

// scanControlFlowCondition collects the condition/clause of an if/for/switch
// up to the body-opening `{`, emits it as GoCode, then transitions to view
// mode for the body content.
func (l *Lexer) scanControlFlowCondition() {
	l.skipWhitespace()
	start := l.pos
	startPos := l.position()

	for !l.atEnd() {
		ch := l.peek()
		if ch == '{' {
			// Emit condition
			l.emitExprContent(start, startPos)
			// The `{` opens the body — increment depth and switch to view mode
			l.braceDepth++
			l.advance()
			l.exprStack = append(l.exprStack, exprContext{
				braceDepth: l.braceDepth,
				returnMode: l.returnMode,
			})
			l.mode = modeView
			return
		}
		l.advance()
	}

	// If we get here without finding {, emit what we have
	l.emitExprContent(start, startPos)
}

func (l *Lexer) scanExprBody() {
	l.skipWhitespace()
	start := l.pos
	startPos := l.position()

	for !l.atEnd() {
		switch ch := l.peek(); ch {
		case '{':
			l.braceDepth++
			l.advance()
		case '}':
			l.braceDepth--
			if l.braceDepth == 0 {
				l.emitExprContent(start, startPos)
				l.advance()
				l.emit(token.ExprEnd, "}")
				l.mode = l.returnMode
				return
			}
			l.advance()
		case '"':
			l.skipQuoted('"')
		case '\'':
			l.skipQuoted('\'')
		case '`':
			l.skipRawString()
		case '<':
			// JSX element inside expression (e.g. inside {if cond { <Foo /> }})
			// Save current expression context and switch to view mode
			l.emitExprContent(start, startPos)
			l.exprStack = append(l.exprStack, exprContext{
				braceDepth: l.braceDepth,
				returnMode: l.returnMode,
			})
			l.mode = modeView
			return
		default:
			l.advance()
		}
	}
}

func (l *Lexer) emitExprContent(start int, startPos token.Position) {
	expr := strings.TrimSpace(string(l.src[start:l.pos]))
	if expr != "" {
		l.tokens = append(l.tokens, token.Token{
			Type:  token.GoCode,
			Value: expr,
			Pos:   startPos,
		})
	}
}

func (l *Lexer) skipQuoted(quote byte) {
	l.advance()
	for !l.atEnd() && l.peek() != quote {
		if l.peek() == '\\' {
			l.advance()
		}
		l.advance()
	}
	if !l.atEnd() {
		l.advance()
	}
}

func (l *Lexer) skipRawString() {
	l.advance()
	for !l.atEnd() && l.peek() != '`' {
		l.advance()
	}
	if !l.atEnd() {
		l.advance()
	}
}

// --- Character helpers ---

func (l *Lexer) peek() byte {
	if l.atEnd() {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peekAt(offset int) byte {
	p := l.pos + offset
	if p >= len(l.src) {
		return 0
	}
	return l.src[p]
}

func (l *Lexer) advance() byte {
	if l.atEnd() {
		return 0
	}
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.src)
}

func (l *Lexer) skipWhitespace() {
	for !l.atEnd() && isWhitespace(l.peek()) {
		l.advance()
	}
}

func (l *Lexer) position() token.Position {
	return token.Position{File: l.file, Line: l.line, Column: l.col, Offset: l.pos}
}

func (l *Lexer) positionAt(offset int) token.Position {
	line, col := 1, 1
	for i := range offset {
		if i < len(l.src) && l.src[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return token.Position{File: l.file, Line: line, Column: col, Offset: offset}
}

func (l *Lexer) emit(typ token.Type, value string) {
	l.tokens = append(l.tokens, token.Token{
		Type:  typ,
		Value: value,
		Pos:   l.position(),
	})
}

func (l *Lexer) emitAt(typ token.Type, value string, pos token.Position) {
	l.tokens = append(l.tokens, token.Token{
		Type:  typ,
		Value: value,
		Pos:   pos,
	})
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isIdentStart(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isIdentPart(ch byte) bool {
	r := rune(ch)
	return unicode.IsLetter(r) || unicode.IsDigit(r) || ch == '_'
}
