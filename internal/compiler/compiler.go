// Package compiler provides the top-level API for compiling .gox files to .go.
package compiler

import (
	"fmt"
	"github.com/dreson4/gox/internal/compiler/codegen"
	"github.com/dreson4/gox/internal/compiler/lexer"
	"github.com/dreson4/gox/internal/compiler/parser"
	"path/filepath"
	"strings"
)

// Result holds the output of a compilation.
type Result struct {
	GoSource string // the generated .go source code
	Errors   []Error
}

// Error represents a compilation error with position info.
type Error struct {
	File    string
	Line    int
	Column  int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
}

// Compile transforms .gox source into .go source.
func Compile(src []byte, filename string) Result {
	// Lex
	l := lexer.New(src, filename)
	tokens := l.Tokenize()

	// Parse
	p := parser.New(tokens)
	file, parseErrors := p.Parse()

	if len(parseErrors) > 0 {
		var errs []Error
		for _, pe := range parseErrors {
			errs = append(errs, Error{
				File:    pe.Pos.File,
				Line:    pe.Pos.Line,
				Column:  pe.Pos.Column,
				Message: pe.Message,
			})
		}
		return Result{Errors: errs}
	}

	// Detect component files (has type Props struct)
	for _, section := range file.GoSections {
		if strings.Contains(section.Code, "type Props struct") {
			file.IsComponent = true
			// Derive component name from filename: "comment.gox" → "Comment"
			base := filepath.Base(filename)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			file.ComponentName = strings.ToUpper(name[:1]) + name[1:]
			break
		}
	}

	// Generate
	g := codegen.New()
	goSource := g.Generate(file)

	return Result{GoSource: goSource}
}
