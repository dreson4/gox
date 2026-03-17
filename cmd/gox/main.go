// Command gox is the GOX compiler and build tool CLI.
//
// Usage:
//
//	gox compile <file.gox|dir>    Compile .gox files to .go
//	gox generate ios              Generate iOS project in ios/
//	gox run ios                   Compile, generate, build, and run on iOS simulator
package main

import (
	"fmt"
	"gox/internal/compiler"
	"gox/internal/generator"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "compile":
		err = cmdCompile()
	case "generate":
		err = cmdGenerate()
	case "run":
		err = cmdRun()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gox <command> [args]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  compile <file.gox|dir>    Compile .gox files to .go")
	fmt.Fprintln(os.Stderr, "  generate ios              Generate iOS native project")
	fmt.Fprintln(os.Stderr, "  run ios                   Build and run on iOS simulator")
}

// --- compile ---

func cmdCompile() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: gox compile <file.gox|dir>")
	}
	return runCompile(os.Args[2])
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

	outPath := outputPath(path)
	if err := os.WriteFile(outPath, []byte(result.GoSource), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Printf("  %s → %s\n", path, outPath)
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

func outputPath(goxPath string) string {
	ext := filepath.Ext(goxPath)
	base := strings.TrimSuffix(goxPath, ext)
	return base + "_gox.go"
}

// --- generate ---

func cmdGenerate() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: gox generate <ios|android>")
	}

	switch os.Args[2] {
	case "ios":
		return generateIOS()
	default:
		return fmt.Errorf("unknown platform: %s (supported: ios)", os.Args[2])
	}
}

func generateIOS() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	appName := filepath.Base(cwd)
	iosDir := filepath.Join(cwd, "ios")

	fmt.Printf("Generating iOS project: %s\n", appName)
	return generator.GenerateIOS(generator.IOSConfig{
		AppName:   appName,
		OutputDir: iosDir,
	})
}

// --- run ---

func cmdRun() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: gox run <ios|android>")
	}

	switch os.Args[2] {
	case "ios":
		return runIOS()
	default:
		return fmt.Errorf("unknown platform: %s (supported: ios)", os.Args[2])
	}
}

func runIOS() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	appName := filepath.Base(cwd)
	iosDir := filepath.Join(cwd, "ios")

	// Step 1: Compile all .gox files
	fmt.Println("[1/5] Compiling .gox files...")
	if err := compileAllGox(cwd); err != nil {
		return fmt.Errorf("compile: %w", err)
	}

	// Step 2: Generate bootstrap (main_gox_bootstrap.go with GoxGetTree export)
	fmt.Println("[2/5] Generating bootstrap...")
	if err := generator.GenerateBootstrap(cwd); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}

	// Step 3: Generate iOS project
	fmt.Println("[3/5] Generating iOS project...")
	if err := generator.GenerateIOS(generator.IOSConfig{
		AppName:   appName,
		OutputDir: iosDir,
	}); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// Step 4: Build Go as C archive for iOS simulator
	fmt.Println("[4/5] Building Go → C archive...")
	if err := buildGoArchive(cwd, iosDir); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	// Step 5: Build with clang and run on simulator
	fmt.Println("[5/5] Building iOS app...")
	if err := buildAndRunIOS(cwd, iosDir, appName); err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func compileAllGox(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip ios/, android/, build/ directories
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "ios" || base == "android" || base == "build" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".gox") {
			return compileFile(path)
		}
		return nil
	})
}

func buildGoArchive(projectDir, iosDir string) error {
	outPath := filepath.Join(iosDir, "libgox.a")

	sdkPath := iosSimulatorSDK()

	cmd := exec.Command("go", "build",
		"-buildmode=c-archive",
		"-o", outPath,
		".",
	)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"GOOS=ios",
		"GOARCH=arm64",
		"CGO_ENABLED=1",
		"CC=clang",
		fmt.Sprintf("CGO_CFLAGS=-isysroot %s -miphonesimulator-version-min=16.0 -arch arm64 -fembed-bitcode", sdkPath),
		fmt.Sprintf("CGO_LDFLAGS=-isysroot %s -miphonesimulator-version-min=16.0 -arch arm64", sdkPath),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func buildAndRunIOS(projectDir, iosDir, appName string) error {
	sdkPath := iosSimulatorSDK()
	bundleID := "com.gox." + strings.ToLower(appName)

	// Build the .app bundle
	appDir := filepath.Join(iosDir, "build", appName+".app")
	os.MkdirAll(appDir, 0755)

	// Copy Info.plist
	plistSrc := filepath.Join(iosDir, appName, "Info.plist")
	plistData, err := os.ReadFile(plistSrc)
	if err != nil {
		return fmt.Errorf("reading Info.plist: %w", err)
	}
	os.WriteFile(filepath.Join(appDir, "Info.plist"), plistData, 0644)

	// Compile bridge.m + link with libgox.a → executable
	bridgeSrc := filepath.Join(iosDir, appName, "bridge.m")
	libPath := filepath.Join(iosDir, "libgox.a")
	headerPath := filepath.Join(iosDir, "libgox.h")
	outputBin := filepath.Join(appDir, appName)

	clangArgs := []string{
		"-isysroot", sdkPath,
		"-miphonesimulator-version-min=16.0",
		"-arch", "arm64",
		"-framework", "UIKit",
		"-framework", "Foundation",
		"-framework", "CoreGraphics",
		"-framework", "Security",
		"-I", filepath.Dir(headerPath),
		bridgeSrc,
		libPath,
		"-lresolv",
		"-o", outputBin,
	}

	cmd := exec.Command("clang", clangArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clang: %w", err)
	}

	fmt.Printf("Built %s\n", appDir)

	// Boot simulator
	fmt.Println("Booting simulator...")
	_ = exec.Command("open", "-a", "Simulator").Run()

	// Find a booted simulator or boot one
	bootCmd := exec.Command("xcrun", "simctl", "boot", "iPhone 16")
	bootCmd.Run() // ignore error if already booted

	// Install
	install := exec.Command("xcrun", "simctl", "install", "booted", appDir)
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	// Launch
	launch := exec.Command("xcrun", "simctl", "launch", "--console-pty", "booted", bundleID)
	launch.Stdout = os.Stdout
	launch.Stderr = os.Stderr
	return launch.Run()
}

func iosSimulatorSDK() string {
	out, err := exec.Command("xcrun", "--sdk", "iphonesimulator", "--show-sdk-path").Output()
	if err != nil {
		return "/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneSimulator.platform/Developer/SDKs/iPhoneSimulator.sdk"
	}
	return strings.TrimSpace(string(out))
}
