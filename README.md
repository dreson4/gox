# GOX

A native mobile UI framework for Go. Write Go + JSX-like views, compile to native iOS and Android apps. No runtime, no VM, no bridge.

```gox
package app

import "gox"

view {
    <gox.View>
        <gox.Text>Hello World</gox.Text>
    </gox.View>
}
```

## How It Works

GOX adds one concept to Go: the `view` block — JSX-like syntax for declaring UI. Everything else is standard Go.

```
.gox file → GOX Compiler → .go file → go build → native binary
```

- `.gox` files are Go with a `view { }` block containing JSX-like elements
- The compiler transpiles `.gox` → pure `.go` that calls the GOX runtime
- `go build` compiles to a native binary — no interpreter, no virtual DOM

## Status

**Early development.** The compiler pipeline (lexer → parser → codegen) is working. Runtime and platform bridge are next.

## Usage

```bash
# Compile a .gox file to .go
gox compile app.gox

# Compile all .gox files in a directory
gox compile ./screens/home/
```

## Project Structure

```
gox/
├── cmd/gox/                  # CLI tool
├── internal/compiler/
│   ├── token/                # Token types
│   ├── lexer/                # Multi-mode tokenizer
│   ├── ast/                  # Abstract syntax tree
│   ├── parser/               # Recursive descent parser
│   └── codegen/              # Go code generator
├── docs/
│   ├── gox-complete-spec.md  # Language specification
│   └── PLAN.md               # Implementation roadmap
└── testdata/                 # Sample .gox files
```

## Design Principles

- **One new thing** — `view` block is the only non-Go syntax
- **Explicit imports** — `gox.View`, `gox.Text`, no magic globals
- **One way to do everything** — styles are props, events are funcs, state is a struct
- **Go ecosystem works** — `go test`, `go vet`, profiling, all of it
- **Never touch native folders** — iOS/Android projects are auto-generated

See [docs/gox-complete-spec.md](docs/gox-complete-spec.md) for the full language specification.
