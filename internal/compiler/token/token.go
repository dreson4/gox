// Package token defines the token types produced by the GOX lexer.
package token

// Type represents a token classification.
type Type int

const (
	// Special
	EOF     Type = iota
	Illegal      // unrecognized character

	// Go pass-through
	GoCode // raw Go source code (everything outside view block)

	// Structural
	ViewKeyword // `view`
	LBrace      // `{`
	RBrace      // `}`

	// JSX tokens (inside view block)
	LAngle    // `<`
	RAngle    // `>`
	LAngleSlash // `</`
	SlashRAngle // `/>`
	Equals      // `=`
	Dot         // `.`

	// Identifiers and literals
	Ident      // tag name, prop name, package name
	String     // "quoted string"
	Text       // raw text content between tags

	// Expressions
	ExprStart // `{` inside JSX (opens Go expression)
	ExprEnd   // `}` matching ExprStart

	// Control flow keywords (inside view expressions)
	If     // `if`
	For    // `for`
	Switch // `switch`
)

var typeNames = [...]string{
	EOF:         "EOF",
	Illegal:     "Illegal",
	GoCode:      "GoCode",
	ViewKeyword: "view",
	LBrace:      "{",
	RBrace:      "}",
	LAngle:      "<",
	RAngle:      ">",
	LAngleSlash: "</",
	SlashRAngle: "/>",
	Equals:      "=",
	Dot:         ".",
	Ident:       "Ident",
	String:      "String",
	Text:        "Text",
	ExprStart:   "ExprStart",
	ExprEnd:     "ExprEnd",
	If:          "if",
	For:         "for",
	Switch:      "switch",
}

func (t Type) String() string {
	if int(t) < len(typeNames) {
		return typeNames[t]
	}
	return "Unknown"
}

// Token represents a single lexical unit with its position in source.
type Token struct {
	Type    Type
	Value   string
	Pos     Position
}

// Position tracks where a token appears in source.
type Position struct {
	File   string
	Line   int
	Column int
	Offset int
}

func (p Position) String() string {
	if p.File != "" {
		return p.File + ":" + itoa(p.Line) + ":" + itoa(p.Column)
	}
	return itoa(p.Line) + ":" + itoa(p.Column)
}

// itoa avoids importing strconv for a trivial conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
