// Command gox is the GOX compiler CLI.
//
// Usage:
//
//	gox compile <file.gox>       Compile a .gox file to .go
//	gox compile <dir>            Compile all .gox files in a directory
package main

import (
	"fmt"
	"gox/internal/compiler"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compile":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: gox compile <file.gox|dir>")
			os.Exit(1)
		}
		if err := runCompile(os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gox <command> [args]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  compile <file.gox|dir>    Compile .gox files to .go")
}

func runCompile(target string) error {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", target, err)
	}

	if info.IsDir() {
		return compileDir(target)
	}
	return compileFile(target)
}

func compileFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	result := compiler.Compile(src, path)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintln(os.Stderr, e)
		}
		return fmt.Errorf("compilation failed with %d error(s)", len(result.Errors))
	}

	// Write output: foo.gox → foo_gox.go
	outPath := outputPath(path)
	if err := os.WriteFile(outPath, []byte(result.GoSource), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Printf("%s → %s\n", path, outPath)
	return nil
}

func compileDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	compiled := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gox") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := compileFile(path); err != nil {
			return err
		}
		compiled++
	}

	if compiled == 0 {
		fmt.Println("no .gox files found")
	}
	return nil
}

// outputPath converts a .gox path to a .go output path.
// app.gox → app_gox.go
func outputPath(goxPath string) string {
	ext := filepath.Ext(goxPath)
	base := strings.TrimSuffix(goxPath, ext)
	return base + "_gox.go"
}
