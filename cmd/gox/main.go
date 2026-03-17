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

	// Step 1: Compile all .gox files
	fmt.Println("Compiling .gox files...")
	if err := compileAllGox(cwd); err != nil {
		return fmt.Errorf("compile: %w", err)
	}

	// Step 2: Generate iOS project
	fmt.Println("Generating iOS project...")
	if err := generateIOS(); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// Step 3: Build Go as C archive for iOS simulator
	fmt.Println("Building Go library...")
	iosDir := filepath.Join(cwd, "ios")
	if err := buildGoArchive(cwd, iosDir); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	// Step 4: Build with xcodebuild
	fmt.Println("Building iOS app...")
	appName := filepath.Base(cwd)
	if err := buildXcode(iosDir, appName); err != nil {
		return fmt.Errorf("xcodebuild: %w", err)
	}

	// Step 5: Launch on simulator
	fmt.Println("Launching on simulator...")
	return launchSimulator(iosDir, appName)
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
	appName := filepath.Base(projectDir)
	outPath := filepath.Join(iosDir, "libgox.a")

	// Build for iOS simulator (arm64 on Apple Silicon)
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
		fmt.Sprintf("CGO_CFLAGS=-isysroot %s -mios-simulator-version-min=16.0 -arch arm64", iosSimulatorSDK()),
		fmt.Sprintf("CGO_LDFLAGS=-isysroot %s -mios-simulator-version-min=16.0 -arch arm64", iosSimulatorSDK()),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = appName
	return cmd.Run()
}

func buildXcode(iosDir, appName string) error {
	projPath := filepath.Join(iosDir, appName+".xcodeproj")
	cmd := exec.Command("xcodebuild",
		"-project", projPath,
		"-scheme", appName,
		"-sdk", "iphonesimulator",
		"-configuration", "Debug",
		"-destination", "platform=iOS Simulator,name=iPhone 16",
		"build",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func launchSimulator(iosDir, appName string) error {
	// Boot simulator if needed
	_ = exec.Command("xcrun", "simctl", "boot", "iPhone 16").Run()

	// Find the built .app
	buildDir := filepath.Join(iosDir, "build", "Debug-iphonesimulator")
	appPath := filepath.Join(buildDir, appName+".app")

	// Install and launch
	install := exec.Command("xcrun", "simctl", "install", "booted", appPath)
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr
	if err := install.Run(); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	// TODO: get bundle ID from config
	bundleID := "com.gox." + strings.ToLower(appName)
	launch := exec.Command("xcrun", "simctl", "launch", "booted", bundleID)
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
